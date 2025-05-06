/*
 * MIT License
 *
 * Copyright (c) 2024 VTB-LINK and runstp.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS," WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE, AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS
 * OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES, OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT, OR OTHERWISE, ARISING FROM, OUT OF,
 * OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package basic

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	ierrors "github.com/vtb-link/bianka/errors"
	"github.com/vtb-link/bianka/proto"
	"golang.org/x/exp/slog"
)

const (
	// CloseAuthFailed 鉴权失败
	CloseAuthFailed = 1
	// CloseActively 调用者主动关闭
	CloseActively = 2
	// CloseReadingConnError 读取链接错误
	CloseReadingConnError = 3
	// CloseReceivedShutdownMessage 收到关闭消息
	CloseReceivedShutdownMessage = 4
	// CloseTypeUnknown 未知原因
	CloseTypeUnknown = 5
)

type StartResp interface {
	GetAuthBody() []byte
	GetLinks() []string
}

type WsClientCloseCallback func(wsClient *WsClient, startResp StartResp, closeType int)

// DefaultLoggerGenerator 默认日志生成器
// 如果不设置，会使用 slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
var DefaultLoggerGenerator = func() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

type DispatcherHandle func(wsClient *WsClient, msg *proto.Message) error

type DispatcherHandleMap map[uint32]DispatcherHandle

func (dhm DispatcherHandleMap) Set(op uint32, handle DispatcherHandle) DispatcherHandleMap {
	dhm[op] = handle
	return dhm
}

type WsClient struct {
	logger *slog.Logger
	conn   *websocket.Conn // 实际的链接

	msgChan    chan *proto.Message // 消息队列
	dispatcher DispatcherHandleMap // 调度器

	startResp StartResp // 启动app的返回信息
	authed    bool      // 是否已经鉴权

	onClose WsClientCloseCallback // 关闭回调

	closeWait sync.WaitGroup
	once      *sync.Once
	cancel    context.CancelFunc
}

func NewWsClient(startResp StartResp, dispatcherHandleMap DispatcherHandleMap, logger *slog.Logger) *WsClient {
	return (&WsClient{
		logger:  logger,
		msgChan: make(chan *proto.Message, 1024),

		startResp: startResp,

		closeWait: sync.WaitGroup{},
		once:      &sync.Once{},
	}).initDispatcherHandleMap(dispatcherHandleMap)
}

func (wsClient *WsClient) Logger() *slog.Logger {
	return wsClient.logger
}

func (wsClient *WsClient) AuthSuccess() {
	wsClient.authed = true
}

func (wsClient *WsClient) IsAuthed() bool {
	return wsClient.authed
}

func (wsClient *WsClient) WithOnClose(onClose WsClientCloseCallback) *WsClient {
	wsClient.onClose = onClose
	return wsClient
}

func (wsClient *WsClient) initDispatcherHandleMap(dispatcherHandleMap DispatcherHandleMap) *WsClient {
	wsClient.dispatcher = DispatcherHandleMap{
		proto.OperationUserAuthenticationReply: authResp,
		proto.OperationHeartbeatReply:          heartBeatResp,
	}

	for op, handle := range dispatcherHandleMap {
		wsClient.dispatcher.Set(op, handle)
	}

	return wsClient
}

func (wsClient *WsClient) Close() error {
	return wsClient.CloseWithType(CloseActively)
}

func (wsClient *WsClient) CloseWithType(t int) (err error) {
	wsClient.logger.Info("ws client close", slog.Int("close_type", t))

	wsClient.once.Do(func() {
		_ = wsClient.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		wsClient.cancel()

		// 等待事件处理完毕
		wsClient.closeWait.Wait()
		err = wsClient.conn.Close()

		// 关闭回调
		if wsClient.onClose != nil {
			wsClient.onClose(wsClient, wsClient.startResp, t)
		}
	})

	if err != nil {
		wsClient.logger.Error("close fail", slog.String("err", err.Error()))
	}
	return err
}

func (wsClient *WsClient) Reconnection(startResp StartResp) error {
	wsClient.startResp = startResp
	wsClient.Reset()

	if err := wsClient.Dial(startResp.GetLinks()...); err != nil {
		return err
	}

	if err := wsClient.SendAuth(); err != nil {
		return err
	}

	wsClient.Run()
	return nil
}

func (wsClient *WsClient) Reset() {
	wsClient.closeWait = sync.WaitGroup{}
	wsClient.once = &sync.Once{}
	wsClient.authed = false
	wsClient.cancel = nil
}

// Dial 链接
func (wsClient *WsClient) Dial(links ...string) error {
	var err error
	for _, link := range links {
		wsClient.conn, _, err = websocket.DefaultDialer.Dial(link, nil)
		if err != nil {
			wsClient.logger.Error("websocket dial fail", slog.String("link", link), slog.String("err", err.Error()))
			continue
		}
		break
	}

	if err != nil {
		return errors.Wrapf(err, "websocket dial fail. links:%v", links)
	}

	wsClient.logger.Info("dial success")
	return nil
}

// eventLoop 处理事件
func (wsClient *WsClient) eventLoop(ctx context.Context) {
	wsClient.logger.Info("ws event loop start")
	wsClient.closeWait.Add(1)

	defer func() {
		wsClient.logger.Info("ws event loop stop")
		wsClient.closeWait.Done()
	}()

	heartbeatTicker := time.NewTicker(time.Second * 15)
	authTimer := time.NewTimer(time.Second * 10)
	for {
		select {
		case <-ctx.Done():
			return
		case <-authTimer.C:
			if !wsClient.authed {
				wsClient.logger.Error("auth timeout")
				go wsClient.CloseWithType(CloseAuthFailed)
				return
			}
		case <-heartbeatTicker.C:
			wsClient.logger.Debug("ws send heartbeat")
			if err := wsClient.SendHeartbeat(); err != nil {
				wsClient.logger.Error("send heartbeat fail", slog.String("err", err.Error()))
			}
		case msg := <-wsClient.msgChan:
			if msg == nil {
				continue
			}

			if handle, ok := wsClient.dispatcher[msg.Operation()]; ok && handle != nil {
				if err := handle(wsClient, msg); err != nil {
					wsClient.logger.Error("handle msg fail", slog.String("err", err.Error()))
				}
			}
		}
	}
}

func (wsClient *WsClient) readMessage(ctx context.Context) {
	wsClient.logger.Info("ws read message start")
	wsClient.closeWait.Add(1)

	defer func() {
		wsClient.logger.Info("ws read message stop")
		wsClient.closeWait.Done()
	}()

	// 如果发生读取错误, 先跳出循环, 尝试select ctx.Done() 如果ctx.Done()触发, 则说明的正常关闭
	// 否则, 说明是读取错误, 需要关闭链接
	var isReadingErr error
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if isReadingErr != nil {
				wsClient.logger.Error("read message fail", slog.String("err", errors.Wrap(isReadingErr, "read message fail").Error()))
				go wsClient.CloseWithType(CloseReadingConnError)
				return
			}

			// 读取err or read close message 会导致关闭链接
			msgType, buf, err := wsClient.conn.ReadMessage()
			switch {
			case err != nil:
				isReadingErr = err
				continue
			case msgType == websocket.PongMessage || msgType == websocket.PingMessage:
				wsClient.logger.Debug("read message", slog.String("msg_type", "ping/pong"))
				continue
			case msgType == websocket.CloseMessage:
				wsClient.logger.Info("received shutdown message", slog.Int("msg_type", msgType))
				go wsClient.CloseWithType(CloseReceivedShutdownMessage)
				return
			default:
				msgList, err := proto.UnpackMessage(buf)
				if err != nil {
					wsClient.logger.Error("unpack message fail", slog.String("err", err.Error()))
					continue
				}

				for i := 0; i < len(msgList); i++ {
					wsClient.msgChan <- &msgList[i]
				}
			}
		}
	}
}

func (wsClient *WsClient) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	wsClient.cancel = cancel

	// 读取信息
	go wsClient.readMessage(ctx)
	// 处理事件
	go wsClient.eventLoop(ctx)
}

// SendMessage 发送消息
func (wsClient *WsClient) SendMessage(msg proto.Message) error {
	err := wsClient.conn.WriteMessage(websocket.BinaryMessage, msg.ToBytes())
	if err != nil {
		return errors.Wrapf(err, "send message fail. payload:%s", msg.Payload())
	}

	return nil
}

// SendAuth 发送鉴权信息
func (wsClient *WsClient) SendAuth() error {
	return wsClient.SendMessage(proto.PackMessage(
		proto.HeaderDefaultSequence,
		proto.OperationUserAuthentication,
		wsClient.startResp.GetAuthBody()),
	)
}

// SendHeartbeat 发送心跳
func (wsClient *WsClient) SendHeartbeat() error {
	return wsClient.SendMessage(proto.PackMessage(
		proto.HeaderDefaultSequence,
		proto.OperationHeartbeat,
		nil),
	)
}

// StartWebsocket 启动websocket
// 此方法会一键完成鉴权，心跳，消息分发注册
func StartWebsocket(
	startResp StartResp,
	dispatcherHandleMap DispatcherHandleMap,
	onCloseFunc WsClientCloseCallback,
	logger *slog.Logger,
) (*WsClient, error) {
	wsClient := NewWsClient(
		startResp,
		dispatcherHandleMap,
		logger).
		WithOnClose(onCloseFunc)

	if err := wsClient.Dial(startResp.GetLinks()...); err != nil {
		return nil, err
	}

	if err := wsClient.SendAuth(); err != nil {
		return nil, err
	}

	wsClient.Run()
	return wsClient, nil
}

// authResp  认证结果
func authResp(wsClient *WsClient, msg *proto.Message) error {
	defer func() {
		// 鉴权失败，关闭链接
		if !wsClient.IsAuthed() {
			go wsClient.CloseWithType(CloseAuthFailed)
		}
	}()

	if string(msg.Payload()) != `{"code":0}` {
		return errors.Wrapf(ierrors.BilibiliWebsocketAuthFailed, "auth fail. payload: %s", msg.Payload())
	}

	wsClient.Logger().Info("auth success")
	wsClient.AuthSuccess()
	return nil
}

// heartBeatResp  心跳结果
func heartBeatResp(wsClient *WsClient, _ *proto.Message) (err error) {
	wsClient.Logger().Debug("heartbeat success")
	return
}

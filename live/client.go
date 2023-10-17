/*
 * MIT License
 *
 * Copyright (c) 2023 VTB-LINK and runstp.
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

package live

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"golang.org/x/exp/slog"

	"github.com/vtb-link/bianka/proto"
)

// DefaultLoggerGenerator 默认日志生成器
// 如果不设置，会使用 slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
var DefaultLoggerGenerator = func() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

type DispatcherHandle func(msg *proto.Message) error

type WsClient struct {
	logger *slog.Logger
	conn   *websocket.Conn // 实际的链接

	msgChan    chan *proto.Message         // 消息队列
	dispatcher map[uint32]DispatcherHandle // 调度器

	startResp *AppStartResponse // 启动app的返回信息
	authed    bool              // 是否已经鉴权

	onClose   func(startResp *AppStartResponse) // 关闭回调
	once      sync.Once
	closeChan chan struct{}
	isClose   bool
}

func (wc *WsClient) WithOnClose(onClose func(startResp *AppStartResponse)) *WsClient {
	wc.onClose = onClose
	return wc
}

func NewWsClient(startResp *AppStartResponse, dispatcherHandleMap map[uint32]DispatcherHandle, logger *slog.Logger) *WsClient {
	if logger == nil {
		logger = DefaultLoggerGenerator()
	}

	logger.With(
		slog.Int("uid", startResp.AnchorInfo.Uid),
		slog.Int("room_id", startResp.AnchorInfo.RoomID),
	)

	return (&WsClient{
		logger: logger,

		startResp: startResp,

		msgChan:   make(chan *proto.Message, 1024),
		closeChan: make(chan struct{}),
	}).initDispatcherHandleMap(dispatcherHandleMap)
}

func (wc *WsClient) initDispatcherHandleMap(dispatcherHandleMap map[uint32]DispatcherHandle) *WsClient {
	if dispatcherHandleMap == nil {
		dispatcherHandleMap = make(map[uint32]DispatcherHandle)
	}

	// 注册分发处理函数
	dispatcherHandleMap[proto.OperationUserAuthenticationReply] = wc.authResp
	dispatcherHandleMap[proto.OperationHeartbeatReply] = wc.heartBeatResp

	wc.dispatcher = dispatcherHandleMap
	return wc
}

func (wc *WsClient) Close() error {
	var err error
	wc.once.Do(func() {
		close(wc.closeChan)
		wc.isClose = true

		err = wc.conn.Close()
		if wc.onClose != nil {
			wc.onClose(wc.startResp)
		}
	})

	return err
}

// Dial 链接
func (wc *WsClient) Dial(links []string) error {
	var err error
	for _, link := range links {
		wc.conn, _, err = websocket.DefaultDialer.Dial(link, nil)
		if err != nil {
			wc.logger.Error("websocket dial fail", slog.String("link", link), slog.String("err", err.Error()))
			continue
		}
		break
	}

	if err != nil {
		return errors.Wrapf(err, "websocket dial fail. links:%v", links)
	}

	wc.logger.Info("dial success")
	return nil
}

// eventLoop 处理事件
func (wc *WsClient) eventLoop() {
	defer wc.logger.Info("ws event loop stop")

	ticker := time.NewTicker(time.Second * 15)
	tr := time.NewTimer(time.Second * 10)
	for {
		select {
		case <-wc.closeChan:
			return
		case <-tr.C:
			if !wc.authed {
				wc.logger.Error("auth timeout")
				return
			}
		case <-ticker.C:
			wc.logger.Debug("ws send heartbeat")
			if err := wc.SendHeartbeat(); err != nil {
				wc.logger.Error("send heartbeat fail", slog.String("err", err.Error()))
			}
		case msg := <-wc.msgChan:
			if msg == nil {
				continue
			}

			handle, ok := wc.dispatcher[msg.Operation()]
			if ok {
				if err := handle(msg); err != nil {
					wc.logger.Error("handle msg fail", slog.String("err", err.Error()))
				}
			}
		}
	}
}

func (wc *WsClient) readMessage() {
	defer wc.logger.Info("ws read stop")
	defer wc.Close()

	for {
		// 读取err or read close message 会导致关闭链接
		msgType, buf, err := wc.conn.ReadMessage()
		if err != nil {
			if !wc.isClose {
				wc.logger.Error("read message fail", slog.String("err", errors.Wrapf(err, "msg_type:%d", msgType).Error()))
			}
			return
		} else if msgType == websocket.CloseMessage {
			wc.logger.Info("read message close", slog.Int("msg_type", msgType))
			return
		} else if msgType == websocket.PongMessage || msgType == websocket.PingMessage {
			wc.logger.Debug("read message", slog.String("msg_type", "ping/pong"))
			continue
		}

		msgList, err := proto.UnpackMessage(buf)
		if err != nil {
			wc.logger.Error("unpack message fail", slog.String("err", err.Error()))
			continue
		}

		for _, msg := range msgList {
			wc.msgChan <- &msg
		}
	}
}

func (wc *WsClient) Run() {
	// 读取信息
	go wc.readMessage()
	// 处理事件
	go wc.eventLoop()
}

// SendAuth 发送鉴权信息
func (wc *WsClient) SendAuth() error {
	return wc.SendMessage(proto.PackMessage(proto.HeaderDefaultSequence, proto.OperationUserAuthentication, []byte(wc.startResp.WebsocketInfo.AuthBody)))
}

// SendMessage 发送消息
func (wc *WsClient) SendMessage(msg proto.Message) error {
	err := wc.conn.WriteMessage(websocket.BinaryMessage, msg.ToBytes())
	if err != nil {
		return errors.Wrapf(err, "send message fail. payload:%s", msg.Payload())
	}

	return nil
}

// SendHeartbeat 发送心跳
func (wc *WsClient) SendHeartbeat() error {
	return wc.SendMessage(proto.PackMessage(proto.HeaderDefaultSequence, proto.OperationHeartbeat, nil))
}

// authResp  认证结果
func (wc *WsClient) authResp(msg *proto.Message) error {
	defer func() {
		// 鉴权失败，关闭链接
		if !wc.authed {
			_ = wc.Close()
		}
	}()

	resp := &CmdLiveOpenPlatformAuthData{}
	if err := json.Unmarshal(msg.Payload(), resp); err != nil {
		return errors.Wrapf(err, "json unmarshal fail. payload:%s", msg.Payload())
	}

	if !resp.Success() {
		return errors.Wrapf(BilibiliWebsocketAuthFailed, "auth fail. code:%d", resp.Code)
	}

	wc.logger.Info("auth success")
	wc.authed = true
	return nil
}

// heartBeatResp  心跳结果
func (wc *WsClient) heartBeatResp(msg *proto.Message) (err error) {
	wc.logger.Debug("heartbeat success")
	return
}

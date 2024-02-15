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

	"github.com/pkg/errors"
	"github.com/vtb-link/bianka/basic"
	"github.com/vtb-link/bianka/proto"
	"golang.org/x/exp/slog"
)

const (
	// CloseAuthFailed 鉴权失败
	CloseAuthFailed = basic.CloseAuthFailed
	// CloseActively 调用者主动关闭
	CloseActively = basic.CloseActively
	// CloseReadingConnError 读取链接错误
	CloseReadingConnError = basic.CloseReadingConnError
	// CloseReceivedShutdownMessage 收到关闭消息
	CloseReceivedShutdownMessage = basic.CloseReceivedShutdownMessage
	// CloseTypeUnknown 未知原因
	CloseTypeUnknown = basic.CloseTypeUnknown
)

type WsClientCloseCallback func(wsClient *WsClient, startResp *AppStartResponse, closeType int)

type DispatcherHandle basic.DispatcherHandle

type WsClient struct {
	logger      *slog.Logger
	basicClient *basic.WsClient
	startResp   *AppStartResponse // 启动app的返回信息
}

func (wsClient *WsClient) WithOnClose(onClose WsClientCloseCallback) *WsClient {
	wsClient.basicClient.WithOnClose(func(basicWs *basic.WsClient, closeType int) {
		onClose(wsClient, wsClient.startResp, closeType)
	})
	return wsClient
}

func NewWsClient(startResp *AppStartResponse, dispatcherHandleMap map[uint32]DispatcherHandle, logger *slog.Logger) *WsClient {
	if logger == nil {
		logger = basic.DefaultLoggerGenerator()
	}

	logger.With(
		slog.Int("uid", startResp.AnchorInfo.Uid),
		slog.Int("room_id", startResp.AnchorInfo.RoomID),
	)

	ws := &WsClient{
		logger:    logger,
		startResp: startResp,
	}

	// 注册分发处理函数
	_dispatcherHandleMap := map[uint32]basic.DispatcherHandle{
		proto.OperationUserAuthenticationReply: ws.authResp,
		proto.OperationHeartbeatReply:          ws.heartBeatResp,
	}

	if dispatcherHandleMap != nil {
		for op, handle := range dispatcherHandleMap {
			_dispatcherHandleMap[op] = basic.DispatcherHandle(handle)
		}
	}

	ws.basicClient = basic.NewWsClient(_dispatcherHandleMap, logger)
	return ws
}

func (wsClient *WsClient) Close() error {
	return wsClient.basicClient.Close(basic.CloseActively)
}

func (wsClient *WsClient) Reconnection(startResp *AppStartResponse) error {
	wsClient.startResp = startResp
	wsClient.basicClient.Reset()

	if err := wsClient.Dial(startResp.WebsocketInfo.WssLink); err != nil {
		return err
	}

	if err := wsClient.SendAuth(); err != nil {
		return err
	}

	wsClient.Run()
	return nil
}

// Dial 链接
func (wsClient *WsClient) Dial(links []string) error {
	return wsClient.basicClient.Dial(links...)
}

func (wsClient *WsClient) Run() {
	wsClient.basicClient.Run()
}

// SendAuth 发送鉴权信息
func (wsClient *WsClient) SendAuth() error {
	return wsClient.SendMessage(proto.PackMessage(proto.HeaderDefaultSequence, proto.OperationUserAuthentication, []byte(wsClient.startResp.WebsocketInfo.AuthBody)))
}

// SendMessage 发送消息
func (wsClient *WsClient) SendMessage(msg proto.Message) error {
	return wsClient.basicClient.SendMessage(msg)
}

// SendHeartbeat 发送心跳
func (wsClient *WsClient) SendHeartbeat() error {
	return wsClient.SendMessage(proto.PackMessage(proto.HeaderDefaultSequence, proto.OperationHeartbeat, nil))
}

// authResp  认证结果
func (wsClient *WsClient) authResp(msg *proto.Message) error {
	defer func() {
		// 鉴权失败，关闭链接
		if !wsClient.basicClient.IsAuthed() {
			go wsClient.basicClient.Close(CloseAuthFailed)
		}
	}()

	resp := &CmdLiveOpenPlatformAuthData{}
	if err := json.Unmarshal(msg.Payload(), resp); err != nil {
		return errors.Wrapf(err, "json unmarshal fail. payload:%s", msg.Payload())
	}

	if !resp.Success() {
		return errors.Wrapf(BilibiliWebsocketAuthFailed, "auth fail. code:%d", resp.Code)
	}

	wsClient.logger.Info("auth success")
	wsClient.basicClient.AuthSuccess()
	return nil
}

// heartBeatResp  心跳结果
func (wsClient *WsClient) heartBeatResp(msg *proto.Message) (err error) {
	wsClient.logger.Debug("heartbeat success")
	return
}

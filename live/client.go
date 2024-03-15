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

package live

import (
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

type DispatcherHandle func(msg *proto.Message) error

type WsClient struct {
	basic.WsClient
}

// NewWsClient 创建一个新的WsClient
// Deprecated: use basic.NewWsClient instead
// 由于2024年B站决定在开发者平台启用直播间长链功能,所以重新设计了WsClient,并且将其移动到basic包中
// 请使用 basic.NewWsClient 替代
// 这里仅作为兼容性处理，后续版本会废弃
func NewWsClient(startResp *AppStartResponse, dispatcherHandleMap map[uint32]DispatcherHandle, logger *slog.Logger) *WsClient {
	if logger == nil {
		logger = basic.DefaultLoggerGenerator()
	}

	logger.With(
		slog.Int("uid", startResp.AnchorInfo.Uid),
		slog.Int("room_id", startResp.AnchorInfo.RoomID),
	)

	// 注册分发处理函数
	_dispatcherHandleMap := basic.DispatcherHandleMap{}

	if dispatcherHandleMap != nil {
		for op, handle := range dispatcherHandleMap {
			_dispatcherHandleMap[op] = func(wsClient *basic.WsClient, msg *proto.Message) error {
				return handle(msg)
			}
		}
	}

	return &WsClient{
		WsClient: *basic.NewWsClient(startResp, _dispatcherHandleMap, logger),
	}
}

func (wsClient *WsClient) WithOnClose(onClose WsClientCloseCallback) *WsClient {
	wsClient.WsClient.WithOnClose(func(_ *basic.WsClient, startResp basic.StartResp, closeType int) {
		onClose(wsClient, startResp.(*AppStartResponse), closeType)
	})
	return wsClient
}

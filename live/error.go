package live

import "github.com/vtb-link/bianka/errors"

var (
	// BilibiliRequestFailed 发生在http请求失败
	BilibiliRequestFailed = errors.BilibiliRequestFailed

	// BilibiliResponseNotSuccess 发生在http请求成功，但响应不是success
	BilibiliResponseNotSuccess = errors.BilibiliResponseNotSuccess

	// BilibiliWebsocketAuthFailed 发生在websocket连接建立后，发送auth请求后，收到的响应不是success
	BilibiliWebsocketAuthFailed = errors.BilibiliWebsocketAuthFailed
)

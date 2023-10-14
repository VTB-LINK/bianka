package live

import "errors"

var (
	// BilibiliRequestFailed 发生在http请求失败
	BilibiliRequestFailed = errors.New("bilibili request failed")

	// BilibiliResponseNotSuccess 发生在http请求成功，但响应不是success
	BilibiliResponseNotSuccess = errors.New("bilibili response not success")

	// BilibiliWebsocketAuthFailed 发生在websocket连接建立后，发送auth请求后，收到的响应不是success
	BilibiliWebsocketAuthFailed = errors.New("bilibili websocket auth failed")
)

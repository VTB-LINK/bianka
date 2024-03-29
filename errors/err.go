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

package errors

import "errors"

var (
	// BilibiliRequestFailed 发生在http请求失败
	BilibiliRequestFailed = errors.New("bilibili request failed")

	// BilibiliResponseNotSuccess 发生在http请求成功，但响应不是success
	BilibiliResponseNotSuccess = errors.New("bilibili response not success")

	// BilibiliWebsocketAuthFailed 发生在websocket连接建立后，发送auth请求后，收到的响应不是success
	BilibiliWebsocketAuthFailed = errors.New("bilibili websocket auth failed")
)

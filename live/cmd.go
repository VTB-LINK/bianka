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

// Package live
//
// bilibili 开放平台 cmd 协议
// see https://open-live.bilibili.com/document/f9ce25be-312e-1f4a-85fd-fef21f1637f8
package live

import (
	"encoding/json"

	"github.com/vtb-link/bianka/proto"

	"github.com/pkg/errors"
)

const (
	CmdLiveOpenPlatformDanmu        = proto.CmdLiveOpenPlatformDanmu        // 弹幕
	CmdLiveOpenPlatformSendGift     = proto.CmdLiveOpenPlatformSendGift     // 礼物
	CmdLiveOpenPlatformSuperChat    = proto.CmdLiveOpenPlatformSuperChat    // SC
	CmdLiveOpenPlatformSuperChatDel = proto.CmdLiveOpenPlatformSuperChatDel // SC删除
	CmdLiveOpenPlatformGuard        = proto.CmdLiveOpenPlatformGuard        // 付费大航海
	CmdLiveOpenPlatformLike         = proto.CmdLiveOpenPlatformLike         // 点赞
)

type CmdLiveOpenPlatformDanmuData proto.CmdDanmuData
type CmdLiveOpenPlatformSendGiftData proto.CmdSendGiftData
type CmdLiveOpenPlatformSuperChatData proto.CmdSuperChatData
type CmdLiveOpenPlatformSuperChatDelData proto.CmdSuperChatDelData
type CmdLiveOpenPlatformGuardData proto.CmdGuardData
type CmdLiveOpenPlatformLikeData proto.CmdLikeData

// AutomaticParsingMessageCommand 自动解析消息命令
// 如果是已知的命令，data 会被解析成对应的结构体，否则 data 会被解析成 map[string]interface{}
// Deprecated: use proto.AutomaticParsingMessageCommand instead
func AutomaticParsingMessageCommand(payload []byte) (string, interface{}, error) {
	var _cmd struct {
		Cmd  string          `json:"cmd"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(payload, &_cmd); err != nil {
		return "", nil, errors.Wrapf(err, "json unmarshal fail, payload:%s", payload)
	}

	var err error
	var data interface{}
	switch _cmd.Cmd {
	case CmdLiveOpenPlatformDanmu:
		data = &CmdLiveOpenPlatformDanmuData{}
	case CmdLiveOpenPlatformSendGift:
		data = &CmdLiveOpenPlatformSendGiftData{}
	case CmdLiveOpenPlatformSuperChat:
		data = &CmdLiveOpenPlatformSuperChatData{}
	case CmdLiveOpenPlatformSuperChatDel:
		data = &CmdLiveOpenPlatformSuperChatDelData{}
	case CmdLiveOpenPlatformGuard:
		data = &CmdLiveOpenPlatformGuardData{}
	case CmdLiveOpenPlatformLike:
		data = &CmdLiveOpenPlatformLikeData{}
	default:
		data = map[string]interface{}{}
	}

	if err = json.Unmarshal(_cmd.Data, &data); err != nil {
		return "", nil, errors.Wrapf(err, "json unmarshal fail, payload:%s", payload)
	}

	return _cmd.Cmd, data, nil
}

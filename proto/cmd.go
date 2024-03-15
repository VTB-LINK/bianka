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

package proto

import (
	"encoding/json"

	"github.com/pkg/errors"
)

const (
	CmdLiveOpenPlatformDanmu        = "LIVE_OPEN_PLATFORM_DM"             // 弹幕
	CmdLiveOpenPlatformSendGift     = "LIVE_OPEN_PLATFORM_SEND_GIFT"      // 礼物
	CmdLiveOpenPlatformSuperChat    = "LIVE_OPEN_PLATFORM_SUPER_CHAT"     // SC
	CmdLiveOpenPlatformSuperChatDel = "LIVE_OPEN_PLATFORM_SUPER_CHAT_DEL" // SC删除
	CmdLiveOpenPlatformGuard        = "LIVE_OPEN_PLATFORM_GUARD"          // 付费大航海
	CmdLiveOpenPlatformLike         = "LIVE_OPEN_PLATFORM_LIKE"           // 点赞

	CmdLiveRoomDanmu        = "OPEN_LIVEROOM_DM"             // 弹幕
	CmdLiveRoomSendGift     = "OPEN_LIVEROOM_SEND_GIFT"      // 礼物
	CmdLiveRoomSuperChat    = "OPEN_LIVEROOM_SUPER_CHAT"     // SC
	CmdLiveRoomSuperChatDel = "OPEN_LIVEROOM_SUPER_CHAT_DEL" // SC删除
	CmdLiveRoomSuperGuard   = "OPEN_LIVEROOM_GUARD"          // 付费大航海
	CmdLiveRoomLike         = "OPEN_LIVEROOM_LIKE"           // 点赞
)

// AutomaticParsingMessageCommand 自动解析消息命令
// 如果是已知的命令，data 会被解析成对应的结构体，否则 data 会被解析成 map[string]interface{}
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
	case CmdLiveOpenPlatformDanmu, CmdLiveRoomDanmu:
		data = &CmdDanmuData{}
	case CmdLiveOpenPlatformSendGift, CmdLiveRoomSendGift:
		data = &CmdSendGiftData{}
	case CmdLiveOpenPlatformSuperChat, CmdLiveRoomSuperChat:
		data = &CmdSuperChatData{}
	case CmdLiveOpenPlatformSuperChatDel, CmdLiveRoomSuperChatDel:
		data = &CmdSuperChatDelData{}
	case CmdLiveOpenPlatformGuard, CmdLiveRoomSuperGuard:
		data = &CmdGuardData{}
	case CmdLiveOpenPlatformLike, CmdLiveRoomLike:
		data = &CmdLikeData{}
	default:
		data = map[string]interface{}{}
	}

	if err = json.Unmarshal(_cmd.Data, &data); err != nil {
		return "", nil, errors.Wrapf(err, "json unmarshal fail, payload:%s", payload)
	}

	return _cmd.Cmd, data, nil
}

type Cmd struct {
	Cmd  string      `json:"cmd"`
	Data interface{} `json:"data"`
}

// CmdDanmuData 弹幕数据
type CmdDanmuData struct {
	RoomID                 int    `json:"room_id"`
	OpenID                 string `json:"open_id"`
	Uid                    int    `json:"uid"`
	Uname                  string `json:"uname"`
	Msg                    string `json:"msg"`
	MsgID                  string `json:"msg_id"`
	FansMedalLevel         int    `json:"fans_medal_level"`
	FansMedalName          string `json:"fans_medal_name"`
	FansMedalWearingStatus bool   `json:"fans_medal_wearing_status"`
	GuardLevel             int    `json:"guard_level"`
	Timestamp              int    `json:"timestamp"`
	UFace                  string `json:"uface"`
	EmojiImgUrl            string `json:"emoji_img_url"`
	DmType                 int    `json:"dm_type"`
}

// CmdSendGiftData 礼物数据
type CmdSendGiftData struct {
	RoomID                 int    `json:"room_id"`
	OpenID                 string `json:"open_id"`
	Uid                    int    `json:"uid"`
	Uname                  string `json:"uname"`
	Uface                  string `json:"uface"`
	GiftID                 int    `json:"gift_id"`
	GiftName               string `json:"gift_name"`
	GiftNum                int    `json:"gift_num"`
	Price                  int    `json:"price"`
	Paid                   bool   `json:"paid"`
	FansMedalLevel         int    `json:"fans_medal_level"`
	FansMedalName          string `json:"fans_medal_name"`
	FansMedalWearingStatus bool   `json:"fans_medal_wearing_status"`
	GuardLevel             int    `json:"guard_level"`
	Timestamp              int    `json:"timestamp"`
	MsgID                  string `json:"msg_id"`
	AnchorInfo             struct {
		OpenID string `json:"open_id"`
		Uid    int    `json:"uid"`
		Uname  string `json:"uname"`
		Uface  string `json:"uface"`
	} `json:"anchor_info"`
	GiftIcon  string `json:"gift_icon"`
	ComboGift bool   `json:"combo_gift"`
	ComboInfo struct {
		ComboBaseNum int    `json:"combo_base_num"`
		ComboCount   int    `json:"combo_count"`
		ComboID      string `json:"combo_id"`
		ComboTimeout int    `json:"combo_timeout"`
	} `json:"combo_info"`
}

// CmdSuperChatData SC数据
type CmdSuperChatData struct {
	RoomID                 int    `json:"room_id"`
	OpenID                 string `json:"open_id"`
	Uid                    int    `json:"uid"`
	Uname                  string `json:"uname"`
	Uface                  string `json:"uface"`
	MessageID              int    `json:"message_id"`
	Message                string `json:"message"`
	MsgID                  string `json:"msg_id"`
	Rmb                    int    `json:"rmb"`
	Timestamp              int    `json:"timestamp"`
	StartTime              int    `json:"start_time"`
	EndTime                int    `json:"end_time"`
	GuardLevel             int    `json:"guard_level"`
	FansMedalLevel         int    `json:"fans_medal_level"`
	FansMedalName          string `json:"fans_medal_name"`
	FansMedalWearingStatus bool   `json:"fans_medal_wearing_status"`
}

// CmdSuperChatDelData SC删除数据
type CmdSuperChatDelData struct {
	RoomID     int    `json:"room_id"`
	MessageIds []int  `json:"message_ids"`
	MsgID      string `json:"msg_id"`
}

// CmdGuardData 付费大航海数据
type CmdGuardData struct {
	UserInfo struct {
		OpenID string `json:"open_id"`
		Uid    int    `json:"uid"`
		Uname  string `json:"uname"`
		Uface  string `json:"uface"`
	} `json:"user_info"`
	GuardLevel             int    `json:"guard_level"`
	GuardNum               int    `json:"guard_num"`
	GuardUnit              string `json:"guard_unit"`
	FansMedalLevel         int    `json:"fans_medal_level"`
	FansMedalName          string `json:"fans_medal_name"`
	FansMedalWearingStatus bool   `json:"fans_medal_wearing_status"`
	Timestamp              int    `json:"timestamp"`
	RoomID                 int    `json:"room_id"`
	MsgID                  string `json:"msg_id"`
}

// CmdLikeData 点赞数据
type CmdLikeData struct {
	Uname                  string `json:"uname"`
	OpenID                 string `json:"open_id"`
	Uid                    int    `json:"uid"`
	Uface                  string `json:"uface"`
	Timestamp              int    `json:"timestamp"`
	LikeText               string `json:"like_text"`
	FansMedalWearingStatus bool   `json:"fans_medal_wearing_status"`
	FansMedalName          string `json:"fans_medal_name"`
	FansMedalLevel         int    `json:"fans_medal_level"`
	MsgID                  string `json:"msg_id"`
	RoomID                 int    `json:"room_id"`
}

// CmdAuthData 鉴权数据
type CmdAuthData struct {
	Code int64 `json:"code,omitempty"`
}

func (arp CmdAuthData) Success() bool {
	return arp.Code == 0
}

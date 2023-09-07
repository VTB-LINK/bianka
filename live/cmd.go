package live

/*
 * bilibili 开放平台 cmd 协议
 * see https://open-live.bilibili.com/document/f9ce25be-312e-1f4a-85fd-fef21f1637f8
 */

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
	case CmdLiveOpenPlatformDanmu:
		data = &CmdLiveOpenPlatformDanmuData{}
		err = json.Unmarshal(_cmd.Data, &data)
	default:
		data = map[string]interface{}{}
		err = json.Unmarshal(_cmd.Data, &data)
	}

	if err != nil {
		return "", nil, errors.Wrapf(err, "json unmarshal fail, payload:%s", payload)
	}

	return _cmd.Cmd, data, nil
}

type CmdLiveOpenPlatform struct {
	Cmd  string      `json:"cmd"`
	Data interface{} `json:"data"`
}

// CmdLiveOpenPlatformDanmuData 弹幕数据
type CmdLiveOpenPlatformDanmuData struct {
	RoomID                 int    `json:"room_id"`
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

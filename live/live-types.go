package live

import (
	"encoding/json"
)

const (
	AcceptHeader              = "Accept"
	ContentTypeHeader         = "Content-Type"
	AuthorizationHeader       = "Authorization"
	JsonType                  = "application/json"
	BiliVersion               = "1.0"
	HmacSha256                = "HMAC-SHA256"
	BiliTimestampHeader       = "x-bili-timestamp"
	BiliSignatureMethodHeader = "x-bili-signature-method"
	BiliSignatureNonceHeader  = "x-bili-signature-nonce"
	BiliAccessKeyIdHeader     = "x-bili-accesskeyid"
	BiliSignVersionHeader     = "x-bili-signature-version"
	BiliContentMD5Header      = "x-bili-content-md5"
)

type CommonHeader struct {
	ContentType       string
	ContentAcceptType string
	Timestamp         string
	SignatureMethod   string
	SignatureVersion  string
	Authorization     string
	Nonce             string
	AccessKeyID       string
	ContentMD5        string
}

// ToMap 所有字段转map<string, string>
func (h *CommonHeader) ToMap() map[string]string {
	return map[string]string{
		BiliTimestampHeader:       h.Timestamp,
		BiliSignatureMethodHeader: h.SignatureMethod,
		BiliSignatureNonceHeader:  h.Nonce,
		BiliAccessKeyIdHeader:     h.AccessKeyID,
		BiliSignVersionHeader:     h.SignatureVersion,
		BiliContentMD5Header:      h.ContentMD5,
		AuthorizationHeader:       h.Authorization,
		ContentTypeHeader:         h.ContentType,
		AcceptHeader:              h.ContentAcceptType,
	}
}

// ToSortMap 参与签名的字段转map<string, string>
func (h *CommonHeader) ToSortMap() map[string]string {
	return map[string]string{
		BiliTimestampHeader:       h.Timestamp,
		BiliSignatureMethodHeader: h.SignatureMethod,
		BiliSignatureNonceHeader:  h.Nonce,
		BiliAccessKeyIdHeader:     h.AccessKeyID,
		BiliSignVersionHeader:     h.SignatureVersion,
		BiliContentMD5Header:      h.ContentMD5,
	}
}

// CreateSignature 生成签名
func (h *CommonHeader) CreateSignature(accessKeySecret string) string {
	signStr := ToSortedString(h.ToSortMap())
	return HmacSHA256(accessKeySecret, signStr)
}

type BaseResp struct {
	Code      int64           `json:"code"`
	Message   string          `json:"message"`
	RequestID string          `json:"request_id"`
	Data      json.RawMessage `json:"data"`
}

func (resp BaseResp) Success() bool {
	return resp.Code == 0
}

type WebSocketInfo struct {
	//  长连使用的请求json体 第三方无需关注内容,建立长连时使用即可
	AuthBody string `json:"auth_body"`
	//  wss 长连地址
	WssLink []string `json:"wss_link"`
}

type AnchorInfo struct {
	RoomID int    `json:"room_id"`
	Uname  string `json:"uname"`
	UFace  string `json:"uface"`
	Uid    int    `json:"uid"`
}

type GameInfo struct {
	GameID string `json:"game_id"`
}

type AppStartRequest struct {
	// 主播身份码
	Code string `json:"code"`
	// 项目id
	AppID int64 `json:"app_id"`
}

type AppStartResponse struct {
	// 主播信息
	AnchorInfo AnchorInfo `json:"anchor_info"`
	// 场次信息
	GameInfo GameInfo `json:"game_info"`
	// 长连信息
	WebsocketInfo WebSocketInfo `json:"websocket_info"`
}

type AppEndRequest struct {
	// 场次id
	GameID string `json:"game_id"`
	// 项目id
	AppID int64 `json:"app_id"`
}

type AppHeartbeatRequest struct {
	GameID string `json:"game_id"`
}

type AppBatchHeartbeatRequest struct {
	GameIDs []string `json:"game_ids"`
}

type AppBatchHeartbeatResponse struct {
	FailedGameIds []string `json:"failed_game_ids"` // 失败的场次id
}

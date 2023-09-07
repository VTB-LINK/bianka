package live

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
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

type BaseResp struct {
	Code      int64           `json:"code"`
	Message   string          `json:"message"`
	RequestId string          `json:"request_id"`
	Data      json.RawMessage `json:"data"`
}

func (resp BaseResp) Success() bool {
	return resp.Code == 0
}

type Client struct {
	rCfg *RoomConfig
}

func NewClient(rCfg *RoomConfig) *Client {
	return &Client{
		rCfg: rCfg,
	}
}

type WebSocketInfo struct {
	//  长连使用的请求json体 第三方无需关注内容,建立长连时使用即可
	AuthBody string `json:"auth_body"`
	//  wss 长连地址
	WssLink []string `json:"wss_link"`
	// 主播信息
	AnchorInfo AnchorInfo `json:"anchor_info"`
}

type AnchorInfo struct {
	RoomId int    `json:"room_id"`
	Uname  string `json:"uname"`
	UFace  string `json:"uface"`
	Uid    int    `json:"uid"`
}

type GameInfo struct {
	GameId string `json:"game_id"`
}

type StartAppRequest struct {
	// 主播身份码
	Code string `json:"code"`
	// 项目id
	AppID int64 `json:"app_id"`
}

type StartAppRespData struct {
	// 场次信息
	GameInfo GameInfo `json:"game_info"`
	// 长连信息
	WebsocketInfo WebSocketInfo `json:"websocket_info"`
}

// StartApp 启动app
func (c *Client) StartApp() (*StartAppRespData, error) {
	startAppReq := StartAppRequest{
		Code:  c.rCfg.IdCode,
		AppID: c.rCfg.AppID,
	}

	reqJson, err := json.Marshal(startAppReq)
	if err != nil {
		err = errors.Wrap(err, "json marshal fail")
	}

	resp, err := c.ApiRequest(string(reqJson), "/v2/app/start")
	if err != nil {
		return nil, err
	}

	startAppRespData := &StartAppRespData{}
	if err = json.Unmarshal(resp.Data, &startAppRespData); err != nil {
		return nil, errors.Wrapf(err, "json unmarshal fail, data:%s", resp.Data)
	}

	return startAppRespData, nil
}

type EndAppRequest struct {
	// 场次id
	GameId string `json:"game_id"`
	// 项目id
	AppId int64 `json:"app_id"`
}

// EndApp 关闭app
func (c *Client) EndApp(startResp *StartAppRespData) error {
	endAppReq := EndAppRequest{
		GameId: startResp.GameInfo.GameId,
		AppId:  c.rCfg.AppID,
	}

	reqJson, err := json.Marshal(endAppReq)
	if err != nil {
		err = errors.Wrap(err, "json marshal fail")
	}

	_, err = c.ApiRequest(string(reqJson), "/v2/app/end")
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) StartWebsocket(startResp *StartAppRespData, dispatcherHandleMap map[uint32]DispatcherHandle) (*WsClient, error) {
	wc := NewWsClient(c.rCfg, dispatcherHandleMap)

	if err := wc.Dial(startResp.WebsocketInfo.WssLink); err != nil {
		return nil, err
	}

	if err := wc.SendAuth(startResp.WebsocketInfo.AuthBody); err != nil {
		return nil, err
	}

	// 读取信息
	go wc.ReadMsg()
	// 处理事件
	go wc.EventLoop()

	return wc, nil
}

// ApiRequest http request demo方法
func (c *Client) ApiRequest(reqJson, requestUrl string) (*BaseResp, error) {
	header := &CommonHeader{
		ContentType:       JsonType,
		ContentAcceptType: JsonType,
		Timestamp:         strconv.FormatInt(time.Now().Unix(), 10),
		SignatureMethod:   HmacSha256,
		SignatureVersion:  BiliVersion,
		Authorization:     "",
		Nonce:             strconv.FormatInt(time.Now().UnixNano(), 10), //用于幂等,记得替换
		AccessKeyID:       c.rCfg.AccessKey,
		ContentMD5:        Md5(reqJson),
	}
	header.Authorization = CreateSignature(header, c.rCfg.AccessKeySecret)

	result := BaseResp{}
	resp, err := resty.New().R().
		SetHeaders(header.ToMap()).
		SetBody(reqJson).
		SetResult(&result).
		Post(fmt.Sprintf("%s%s", c.rCfg.OpenPlatformHttpHost, requestUrl))

	if err != nil {
		return nil, errors.Wrapf(err, "request fail, url:%s req: %v", requestUrl, reqJson)
	}

	if resp.StatusCode() >= http.StatusAccepted {
		return nil, errors.Wrapf(err, "request response not ok, url:%s req: %v code:%d", requestUrl, reqJson, resp.StatusCode())
	}

	if !result.Success() {
		return nil, fmt.Errorf("bilbil response code not ok, url:%s req: %v resp:%v", requestUrl, reqJson, result)
	}

	return &result, nil
}

// CreateSignature 生成Authorization加密串
func CreateSignature(header *CommonHeader, accessKeySecret string) string {
	sStr := header.ToSortedString()
	return HmacSHA256(accessKeySecret, sStr)
}

// Md5 md5加密
func Md5(str string) (md5str string) {
	data := []byte(str)
	has := md5.Sum(data)
	md5str = fmt.Sprintf("%x", has)
	return md5str
}

// HmacSHA256 HMAC-SHA256算法
func HmacSHA256(key string, data string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
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

// ToSortMap 参与加密的字段转map<string, string>
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

// ToSortedString 生成需要加密的文本
func (h *CommonHeader) ToSortedString() (sign string) {
	hMap := h.ToSortMap()
	var hSil []string
	for k := range hMap {
		hSil = append(hSil, k)
	}
	sort.Strings(hSil)
	for _, v := range hSil {
		sign += v + ":" + hMap[v] + "\n"
	}
	sign = strings.TrimRight(sign, "\n")
	return
}

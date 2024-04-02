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
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/vtb-link/bianka/basic"
)

const (
	HostProdLiveOpen = "https://live-open.biliapi.com" // 开放平台 (线上环境)
)

type Config struct {
	AccessKey            string // access_key
	AccessKeySecret      string // access_key_secret
	OpenPlatformHttpHost string // 开放平台 (线上环境)
	AppID                int64  // 应用id
}

func NewConfig(accessKey, accessKeySecret string, appID int64) *Config {
	return &Config{
		AccessKey:            accessKey,
		AccessKeySecret:      accessKeySecret,
		OpenPlatformHttpHost: HostProdLiveOpen,
		AppID:                appID,
	}
}

type Client struct {
	rCfg *Config
}

func NewClient(rCfg *Config) *Client {
	return &Client{
		rCfg: rCfg,
	}
}

// AppStart 启动app
func (c *Client) AppStart(code string) (*AppStartResponse, error) {
	startAppReq := AppStartRequest{
		Code:  code,
		AppID: c.rCfg.AppID,
	}

	reqJSON, err := json.Marshal(startAppReq)
	if err != nil {
		return nil, errors.Wrap(err, "json marshal fail")
	}

	resp, err := c.doRequest(string(reqJSON), "/v2/app/start")
	if err != nil {
		return nil, errors.WithMessage(err, "start app fail")
	}

	startAppRespData := &AppStartResponse{}
	if err = json.Unmarshal(resp.Data, &startAppRespData); err != nil {
		return nil, errors.Wrapf(err, "json unmarshal fail, data:%s", resp.Data)
	}

	return startAppRespData, nil
}

// AppEnd 关闭app
func (c *Client) AppEnd(gameID string) error {
	endAppReq := AppEndRequest{
		GameID: gameID,
		AppID:  c.rCfg.AppID,
	}

	reqJSON, err := json.Marshal(endAppReq)
	if err != nil {
		return errors.Wrap(err, "json marshal fail")
	}

	_, err = c.doRequest(string(reqJSON), "/v2/app/end")
	if err != nil {
		return errors.WithMessage(err, "end app fail")
	}

	return nil
}

// AppHeartbeat 心跳
func (c *Client) AppHeartbeat(gameID string) error {
	heartbeatReq := AppHeartbeatRequest{
		GameID: gameID,
	}

	reqJSON, err := json.Marshal(heartbeatReq)
	if err != nil {
		return errors.Wrap(err, "json marshal fail")
	}

	respBody, err := c.doRequest(string(reqJSON), "/v2/app/heartbeat")
	if err != nil {
		return errors.Wrap(err, "heartbeat fail")
	}

	// 直接检查respBody的Code属性
	if respBody.Code != 0 {
		// 当Code不为0时返回错误
		return errors.Errorf("heartbeat fail with code %d and message %s", respBody.Code, respBody.Message)
	}

	return nil
}

// AppBatchHeartbeat 批量心跳
func (c *Client) AppBatchHeartbeat(gameIDs []string) (*AppBatchHeartbeatResponse, error) {
	heartbeatReq := AppBatchHeartbeatRequest{
		GameIDs: gameIDs,
	}

	reqJSON, err := json.Marshal(heartbeatReq)
	if err != nil {
		return nil, errors.Wrap(err, "json marshal fail")
	}

	resp, err := c.doRequest(string(reqJSON), "/v2/app/batchHeartbeat")
	if err != nil {
		return nil, errors.WithMessage(err, "heartbeat fail")
	}

	heartbeatResp := &AppBatchHeartbeatResponse{}
	if err = json.Unmarshal(resp.Data, &heartbeatResp); err != nil {
		return nil, errors.Wrapf(err, "json unmarshal fail, data:%s", resp.Data)
	}

	return heartbeatResp, nil
}

// StartWebsocket 启动websocket
// Deprecated: use basic.StartWebsocket instead
// 此方法会一键完成鉴权，心跳，消息分发
// 由于2024年B站决定在开发者平台启用直播间长链功能,所以重新设计了WsClient,并且将其移动到basic包中
// 这里仅作为兼容性处理，后续版本会废弃
func (c *Client) StartWebsocket(startResp *AppStartResponse, dispatcherHandleMap map[uint32]DispatcherHandle, onCloseFunc WsClientCloseCallback) (*WsClient, error) {
	wsClient := NewWsClient(
		startResp,
		dispatcherHandleMap,
		nil).
		WithOnClose(onCloseFunc)

	if err := wsClient.Dial(startResp.GetLinks()...); err != nil {
		return nil, err
	}

	if err := wsClient.SendAuth(); err != nil {
		return nil, err
	}

	wsClient.Run()
	return wsClient, nil
}

func (c *Client) doRequest(reqJSON, reqPath string) (*BaseResp, error) {
	return c.DoRequest(reqJSON, reqPath, basic.RandStringBytes(32))
}

// DoRequest 发起请求
// 用于用户自定义请求
func (c *Client) DoRequest(reqJSON, reqPath, nonce string) (*BaseResp, error) {
	header := &CommonHeader{
		ContentType:       JsonType,
		ContentAcceptType: JsonType,
		Timestamp:         strconv.FormatInt(time.Now().Unix(), 10),
		SignatureMethod:   HmacSha256,
		SignatureVersion:  BiliVersion,
		Nonce:             nonce, // 用于幂等
		AccessKeyID:       c.rCfg.AccessKey,
		ContentMD5:        Md5(reqJSON),
	}
	header.Authorization = header.CreateSignature(c.rCfg.AccessKeySecret)

	result := BaseResp{}
	resp, err := resty.New().R().
		SetHeaders(header.ToMap()).
		SetBody(reqJSON).
		SetResult(&result).
		Post(c.rCfg.OpenPlatformHttpHost + reqPath)

	if err != nil {
		return nil, errors.Wrapf(err, "request fail, url:%s body: %s", reqPath, reqJSON)
	}

	if resp.StatusCode() >= http.StatusBadRequest {
		return nil, errors.Wrapf(BilibiliRequestFailed, "request response not ok, url:%s req: %v code:%d", reqPath, reqJSON, resp.StatusCode())
	}

	if !result.Success() {
		return &result, errors.Wrapf(BilibiliResponseNotSuccess, "bilbil response code not ok, url:%s  body: %s result: %v", reqPath, reqJSON, result)
	}

	return &result, nil
}

// VerifyH5RequestSignature 验证h5请求签名
func (c *Client) VerifyH5RequestSignature(req *http.Request) bool {
	h5sp := ParseH5SignatureParamsWithRequest(req)

	return c.VerifyH5RequestSignatureWithParams(h5sp)
}

// VerifyH5RequestSignatureWithParams 验证h5请求签名
func (c *Client) VerifyH5RequestSignatureWithParams(h5sp *H5SignatureParams) bool {
	return h5sp.ValidateSignature(c.rCfg.AccessKeySecret)
}

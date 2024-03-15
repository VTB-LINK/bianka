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

package openhome

import (
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
)

type Live basicService

type RoomInfoResp struct {
	Openid      string `json:"open_id"`
	RoomID      int    `json:"room_id"`
	Title       string `json:"title"`        // 直播标题
	IsStreaming bool   `json:"is_streaming"` // 当前是否开播
	IsBanned    bool   `json:"is_banned"`    // 房间是否被封禁
}

func (ri RoomInfoResp) IsNotEnable() bool {
	return ri.Openid == "" || ri.RoomID == 0
}

func (l *Live) GetRoomInfo(accessToken string) (*RoomInfoResp, error) {
	result := NewBaseResp(&RoomInfoResp{})

	resp, err := resty.New().R().
		SetResult(result).
		SetQueryParams(map[string]string{
			"client_id":    l.app.appCfg.ClientID,
			"access_token": accessToken,
		}).
		Get(`https://member.bilibili.com/arcopen/fn/live/room/info`)

	if err != nil {
		return nil, errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return nil, err
	}

	return result.Data.(*RoomInfoResp), nil
}

type WsStartResp struct {
	ConnID        string `json:"conn_id"`
	WebsocketInfo struct {
		AuthBody string   `json:"auth_body"`
		WssLink  []string `json:"wss_link"`
	} `json:"websocket_info"`
}

func (wr *WsStartResp) GetAuthBody() []byte {
	return []byte(wr.WebsocketInfo.AuthBody)
}

func (wr *WsStartResp) GetLinks() []string {
	return wr.WebsocketInfo.WssLink
}

func (l *Live) WsStart(accessToken string) (*WsStartResp, error) {
	result := NewBaseResp(&WsStartResp{})

	resp, err := resty.New().R().
		SetResult(result).
		SetQueryParams(map[string]string{
			"client_id":    l.app.appCfg.ClientID,
			"access_token": accessToken,
		}).
		Post(`https://member.bilibili.com/arcopen/fn/live/room/ws-start`)

	if err != nil {
		return nil, errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return nil, err
	}

	return result.Data.(*WsStartResp), nil
}

func (l *Live) WsHeartbeat(accessToken, connID string) error {
	result := NewBaseResp(&WsStartResp{})

	resp, err := resty.New().R().
		SetResult(result).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"client_id":    l.app.appCfg.ClientID,
			"access_token": accessToken,
		}).
		SetBody(map[string]string{
			"conn_id": connID,
		}).
		Post(`https://member.bilibili.com/arcopen/fn/live/room/ws-heartbeat`)

	if err != nil {
		return errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return err
	}

	return nil
}

type WsBatchHeartbeatResp struct {
	FailedConnIds []string `json:"failed_conn_ids"`
}

func (l *Live) WsBatchHeartbeat(accessToken string, connIDs ...string) (*WsBatchHeartbeatResp, error) {
	result := NewBaseResp(&WsBatchHeartbeatResp{})

	resp, err := resty.New().R().
		SetResult(result).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"client_id":    l.app.appCfg.ClientID,
			"access_token": accessToken,
		}).
		SetBody(map[string]interface{}{
			"conn_ids": connIDs,
		}).
		Post(`https://member.bilibili.com/arcopen/fn/live/room/ws-batch-heartbeat`)

	if err != nil {
		return nil, errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return nil, err
	}

	return result.Data.(*WsBatchHeartbeatResp), nil
}

/*
 * MIT License
 *
 * Copyright (c) 2023 VTB-LINK and runstp.
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
	"math/rand"
	"time"
)

var _random = rand.New(rand.NewSource(time.Now().UnixNano()))

type PageResp struct {
	PageNumber int `json:"pn"`
	PageSize   int `json:"ps"`
	Total      int `json:"total"`
}

type BaseResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    interface{}
}

func NewBaseResp(data interface{}) *BaseResp {
	return &BaseResp{
		Data: data,
	}
}

func (b *BaseResp) IsSuccess() bool {
	return b.Code == 0
}

type basicService struct {
	app *AppClient
}

type AppConfig struct {
	ClientID     string `json:"client_id"`     // 应用id
	ClientSecret string `json:"client_secret"` // 应用密钥
}

type AppClient struct {
	appCfg *AppConfig

	OAuth   *OAuth
	User    *User
	Live    *Live
	Archive *Archive
}

func NewAppClient(cfg *AppConfig) *AppClient {
	app := &AppClient{
		appCfg: cfg,
	}

	bs := &basicService{
		app: app,
	}

	app.OAuth = (*OAuth)(bs)
	app.User = (*User)(bs)
	app.Live = (*Live)(bs)
	app.Archive = (*Archive)(bs)

	return app
}

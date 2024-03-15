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

type User basicService

type AccountScopesResp struct {
	Scopes []string `json:"scopes"`
	Openid string   `json:"openid"`
}

func (u *User) GetAccountScopes(accessToken string) (*AccountScopesResp, error) {
	result := NewBaseResp(&AccountScopesResp{})

	resp, err := resty.New().R().
		SetResult(result).
		SetQueryParams(map[string]string{
			"client_id":    u.app.appCfg.ClientID,
			"access_token": accessToken,
		}).
		Get(`https://member.bilibili.com/arcopen/fn/user/account/scopes`)

	if err != nil {
		return nil, errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return nil, err
	}

	return result.Data.(*AccountScopesResp), nil
}

type AccountInfoResp struct {
	Name   string `json:"name"`
	Face   string `json:"face"`
	Openid string `json:"openid"`
}

func (u *User) GetAccountInfo(accessToken string) (*AccountInfoResp, error) {
	result := NewBaseResp(&AccountInfoResp{})

	resp, err := resty.New().R().
		SetResult(result).
		SetQueryParams(map[string]string{
			"client_id":    u.app.appCfg.ClientID,
			"access_token": accessToken,
		}).
		Get(`https://member.bilibili.com/arcopen/fn/user/account/info`)

	if err != nil {
		return nil, errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return nil, err
	}

	return result.Data.(*AccountInfoResp), nil
}

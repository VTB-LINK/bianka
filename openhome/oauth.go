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
	"net/url"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
)

const (
	ScopesUserinfo          = "USER_INFO"           // 用户信息能力 可获取用户公开信息, 默认开启
	ScopesUserdata          = "USER_DATA"           // 用户数据权限 可用于查询用户的关注数、粉丝数、投稿数
	ScopesArcBase           = "ARC_BASE"            // 视频稿件基础能力 支持稿件的发布、编辑、查询和删除功能
	ScopesArcData           = "ARC_DATA"            // 视频稿件数据权限 可用于查询视频稿件的播放数、点赞数、评论数、弹幕数等相关数据
	ScopesAtcBase           = "ATC_BASE"            // 专栏稿件管理能力 支持专栏稿件（文章和文集）的发布、编辑、查询和删除功能
	ScopesAtcData           = "ATC_DATA"            // 专栏稿件数据权限 可用于查询专栏稿件的阅读数、点赞数、评论数、收藏数等相关数据
	ScopesUserActivity      = "USER_ACTIVITY"       // 活动数据权限 授权并使用您的活动数据 (包括任务和奖励)，用于第三方进行奖励发放
	ScopesShopStoreInfo     = "SHOP_STORE_INFO"     // 服务市场店铺信息 授权第三方获得您的店铺基本信息（店铺名称、店铺头像、shop_id）
	ScopesShopOrderInfo     = "SHOP_ORDER_INFO"     // 服务市场订单信息 授权第三方获得您服务市场店铺的订单数据
	ScopesShopCommodityInfo = "SHOP_COMMODITY_INFO" // 服务市场商品信息 授权第三方获取您服务市场店铺的商品数据
	ScopesLiveRoomData      = "LIVE_ROOM_DATA"      // 直播内容数据权限 用于授权第三方获得直播间基本信息(基本信息、开播状态)、直播间弹幕消息相关数据
)

type OAuthGenerator struct {
	clientID    string // 应用id
	mobileUI    bool   // 移动端UI 默认pc
	callbackURL string // 回调地址
	state       string // 防CSRF攻击
}

func (og *OAuthGenerator) GetState() string {
	return og.state
}

func (og *OAuthGenerator) WithEnableMobileUI() *OAuthGenerator {
	og.mobileUI = true
	return og
}

func (og *OAuthGenerator) WithCallbackURL(callbackURL string) *OAuthGenerator {
	og.callbackURL = callbackURL
	return og
}

func (og *OAuthGenerator) WithState(state string) *OAuthGenerator {
	og.state = state
	return og
}

func (og *OAuthGenerator) GenerateAuthorizationURL() string {
	if og.state == "" {
		og.state = strconv.Itoa(_random.Intn(1000000) + 1000000)
	}

	basicURL := "https://passport.bilibili.com/register/pc_oauth2.html#/?"
	if og.mobileUI {
		basicURL = "https://passport.bilibili.com/register/oauth2.html#/?"
	}

	params := url.Values{}
	params.Add("client_id", og.clientID)
	params.Add("return_url", og.callbackURL)
	params.Add("response_type", "code")
	params.Add("state", og.state)

	return basicURL + params.Encode()
}

type OAuth basicService

func (o *OAuth) GetOAuthGenerator() *OAuthGenerator {
	return &OAuthGenerator{
		clientID: o.app.appCfg.ClientID,
	}
}

type AccessToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"` // 过期时间
	RefreshToken string `json:"refresh_token"`
}

func (at AccessToken) IsExpired() bool {
	return int64(at.ExpiresIn) <= time.Now().Unix()
}

type Code2AccessTokenResp struct {
	AccessToken
	Scopes []string `json:"scopes"`
}

func (o *OAuth) Code2AccessToken(code string) (*Code2AccessTokenResp, error) {
	result := NewBaseResp(&Code2AccessTokenResp{})

	resp, err := resty.New().R().
		SetResult(result).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"client_id":     o.app.appCfg.ClientID,
			"client_secret": o.app.appCfg.ClientSecret,
			"grant_type":    "authorization_code",
			"code":          code,
		}).
		Post(`https://api.bilibili.com/x/account-oauth2/v1/token`)

	if err != nil {
		return nil, errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return nil, err
	}

	return result.Data.(*Code2AccessTokenResp), nil
}

type RefreshTokenResp struct {
	AccessToken
}

func (o *OAuth) RefreshToken(refreshToken string) (*RefreshTokenResp, error) {
	result := NewBaseResp(&Code2AccessTokenResp{})

	resp, err := resty.New().R().
		SetResult(result).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"client_id":     o.app.appCfg.ClientID,
			"client_secret": o.app.appCfg.ClientSecret,
			"grant_type":    "refresh_token",
			"refresh_token": refreshToken,
		}).
		Post(`https://api.bilibili.com/x/account-oauth2/v1/refresh_token`)

	if err != nil {
		return nil, errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return nil, err
	}

	return result.Data.(*RefreshTokenResp), nil
}

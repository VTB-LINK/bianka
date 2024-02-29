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
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
)

type Data basicService

type UserStatResp struct {
	Following      int `json:"following"`
	Follower       int `json:"follower"`
	ArcPassedTotal int `json:"arc_passed_total"`
}

func (d *Data) UserStat(accessToken string) (*UserStatResp, error) {
	result := NewBaseResp(&UserStatResp{})

	resp, err := resty.New().R().
		SetResult(result).
		SetQueryParams(map[string]string{
			"client_id":    d.app.appCfg.ClientID,
			"access_token": accessToken,
		}).
		Get(`https://member.bilibili.com/arcopen/fn/data/user/stat`)

	if err != nil {
		return nil, errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return nil, err
	}

	return result.Data.(*UserStatResp), nil
}

type ArcStatResp struct {
	View     int `json:"view"`
	Danmaku  int `json:"danmaku"`
	Reply    int `json:"reply"`
	Favorite int `json:"favorite"`
	Coin     int `json:"coin"`
	Share    int `json:"share"`
	Like     int `json:"like"`
}

func (d *Data) ArcStat(accessToken, resourceID string) (*ArcStatResp, error) {
	result := NewBaseResp(&ArcStatResp{})

	resp, err := resty.New().R().
		SetResult(result).
		SetQueryParams(map[string]string{
			"client_id":    d.app.appCfg.ClientID,
			"access_token": accessToken,
			"resource_id":  resourceID,
		}).
		Get(`https://member.bilibili.com/arcopen/fn/data/arc/stat`)

	if err != nil {
		return nil, errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return nil, err
	}

	return result.Data.(*ArcStatResp), nil
}

type ArcIncStatsResp struct {
	IncClick int `json:"inc_click"`
	IncDm    int `json:"inc_dm"`
	IcnReply int `json:"icn_reply"`
	IncFav   int `json:"inc_fav"`
	IncCoin  int `json:"inc_coin"`
	IncShare int `json:"inc_share"`
	IncLike  int `json:"inc_like"`
	IncElec  int `json:"inc_elec"`
}

func (d *Data) ArcIncStats(accessToken string) (*ArcIncStatsResp, error) {
	result := NewBaseResp(&ArcIncStatsResp{})

	resp, err := resty.New().R().
		SetResult(result).
		SetQueryParams(map[string]string{
			"client_id":    d.app.appCfg.ClientID,
			"access_token": accessToken,
		}).
		Get(`https://member.bilibili.com/arcopen/fn/data/arc/inc-stats`)

	if err != nil {
		return nil, errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return nil, err
	}

	return result.Data.(*ArcIncStatsResp), nil
}

type ArtStartData struct {
	ID       int `json:"id"`
	Category struct {
		ID       int    `json:"id"`
		ParentID int    `json:"parent_id"`
		Name     string `json:"name"`
	} `json:"category"`
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	BannerUrl   string   `json:"banner_url"`
	TemplateID  int      `json:"template_id"`
	State       int      `json:"state"`
	ImageUrls   []string `json:"image_urls"`
	PublishTime int      `json:"publish_time"`
	Ctime       int      `json:"ctime"`
	Stats       struct {
		View     int `json:"view"`
		Favorite int `json:"favorite"`
		Like     int `json:"like"`
		Dislike  int `json:"dislike"`
		Reply    int `json:"reply"`
		Share    int `json:"share"`
		Coin     int `json:"coin"`
	} `json:"stats"`
	Reason string `json:"reason"`
	Words  int    `json:"words"`
	List   struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		ImageUrl    string `json:"image_url"`
		UpdateTime  int    `json:"update_time"`
		Ctime       int    `json:"ctime"`
		PublishTime int    `json:"publish_time"`
		Summary     string `json:"summary"`
		Words       int    `json:"words"`
	} `json:"list"`
	TopVideoBvid string `json:"top_video_bvid"`
}

type ArtStatResp map[string]ArtStartData

func (d *Data) ArtStat(accessToken string, ids []string) (*ArtStatResp, error) {
	result := NewBaseResp(&ArtStatResp{})

	resp, err := resty.New().R().
		SetResult(result).
		SetHeader("x1-bilispy-color", "article-open").
		SetQueryParams(map[string]string{
			"client_id":    d.app.appCfg.ClientID,
			"access_token": accessToken,
		}).
		SetFormData(map[string]string{
			"ids": strings.Join(ids, ","),
		}).
		Get(`https://member.bilibili.com/arcopen/fn/data/art/stat`)

	if err != nil {
		return nil, errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return nil, err
	}

	return result.Data.(*ArtStatResp), nil
}

type ArtIncStatsResp struct {
	IncReply int `json:"inc_reply"`
	IncRead  int `json:"inc_read"`
	IncFav   int `json:"inc_fav"`
	IncLikes int `json:"inc_likes"`
	IncShare int `json:"inc_share"`
	IncCoin  int `json:"inc_coin"`
}

func (d *Data) ArtIncStats(accessToken string) (*ArtIncStatsResp, error) {
	result := NewBaseResp(&ArtIncStatsResp{})

	resp, err := resty.New().R().
		SetResult(result).
		SetHeader("x1-bilispy-color", "article-open").
		SetQueryParams(map[string]string{
			"client_id":    d.app.appCfg.ClientID,
			"access_token": accessToken,
		}).
		Get(`https://member.bilibili.com/arcopen/fn/data/art/inc-stats`)

	if err != nil {
		return nil, errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return nil, err
	}

	return result.Data.(*ArtIncStatsResp), nil
}

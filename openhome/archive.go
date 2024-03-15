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
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
)

type Archive basicService

type ArchiveEditReq struct {
	ResourceID string `json:"resource_id"`
	Title      string `json:"title"`
	Cover      string `json:"cover"`
	Tid        int    `json:"tid"`
	Desc       string `json:"desc"`
	NoReprint  int    `json:"no_reprint"`
}

type ArchiveEditResp struct {
	ResourceID string `json:"resource_id"`
}

func (a *Archive) Edit(accessToken string, req ArchiveEditReq) (*ArchiveEditResp, error) {
	result := NewBaseResp(&ArchiveEditResp{})

	resp, err := resty.New().R().
		SetResult(result).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"client_id":    a.app.appCfg.ClientID,
			"access_token": accessToken,
		}).
		SetBody(req).
		Post(`https://member.bilibili.com/arcopen/fn/archive/edit`)

	if err != nil {
		return nil, errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return nil, err
	}

	return result.Data.(*ArchiveEditResp), nil
}

func (a *Archive) Delete(accessToken, resourceID string) error {
	result := NewBaseResp(nil)

	resp, err := resty.New().R().
		SetResult(result).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"client_id":    a.app.appCfg.ClientID,
			"access_token": accessToken,
		}).
		SetBody(map[string]string{
			"resource_id": resourceID,
		}).
		Post(`https://member.bilibili.com/arcopen/fn/archive/delete`)

	if err != nil {
		return errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return err
	}

	return nil
}

type ArchiveViewResp struct {
	ResourceID string `json:"resource_id"`
	Title      string `json:"title"`
	Cover      string `json:"cover"`
	Tid        int    `json:"tid"`
	NoReprint  int    `json:"no_reprint"`
	Desc       string `json:"desc"`
	Tag        string `json:"tag"`
	Copyright  int    `json:"copyright"`
	Ctime      int    `json:"ctime"`
	Ptime      int    `json:"ptime"`
	VideoInfo  struct {
		Cid       interface{} `json:"cid"`
		Filename  string      `json:"filename"`
		Duration  int         `json:"duration"`
		ShareUrl  string      `json:"share_url"`
		IframeUrl string      `json:"iframe_url"`
	} `json:"video_info"`
	AdditInfo struct {
		State        int    `json:"state"`
		StateDesc    string `json:"state_desc"`
		RejectReason string `json:"reject_reason"`
	} `json:"addit_info"`
}

func (a *Archive) View(accessToken, resourceID string) (*ArchiveViewResp, error) {
	result := NewBaseResp(&ArchiveViewResp{})

	resp, err := resty.New().R().
		SetResult(result).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"client_id":    a.app.appCfg.ClientID,
			"access_token": accessToken,
			"resource_id":  resourceID,
		}).
		Get(`https://member.bilibili.com/arcopen/fn/archive/view`)

	if err != nil {
		return nil, errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return nil, err
	}

	return result.Data.(*ArchiveViewResp), nil
}

type ArchiveViewListResp struct {
	Page PageResp           `json:"page"`
	List []*ArchiveViewResp `json:"list"`
}

func (l ArchiveViewListResp) IsEmpty() bool {
	return l.List == nil || len(l.List) == 0
}

const (
	ArchiveStatusAll      = "all"
	ArchiveStatusIsPubing = "is_pubing"
	ArchiveStatusPubed    = "pubed"
	ArchiveStatusNotPubed = "not_pubed"
)

type ArchiveViewListReq struct {
	PageNumber int    `json:"pn"`
	PageSize   int    `json:"ps"`
	Status     string `json:"status"`
}

func (a *Archive) ViewList(accessToken string, req ArchiveViewListReq) (*ArchiveViewListResp, error) {
	result := NewBaseResp(&ArchiveViewListResp{})

	resp, err := resty.New().R().
		SetResult(result).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"client_id":    a.app.appCfg.ClientID,
			"access_token": accessToken,
			"pn":           strconv.Itoa(req.PageNumber),
			"ps":           strconv.Itoa(req.PageSize),
			"status":       req.Status,
		}).
		Get(`https://member.bilibili.com/arcopen/fn/archive/view/list`)

	if err != nil {
		return nil, errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return nil, err
	}

	return result.Data.(*ArchiveViewListResp), nil
}

type ArchiveType struct {
	ID          int    `json:"id"`
	Parent      int    `json:"parent"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ArchiveTypeListResp struct {
	ArchiveType
	Children []*ArchiveType `json:"children"`
}

func (a *Archive) TypeList(accessToken string) (*ArchiveTypeListResp, error) {
	result := NewBaseResp(&ArchiveTypeListResp{})

	resp, err := resty.New().R().
		SetResult(result).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"client_id":    a.app.appCfg.ClientID,
			"access_token": accessToken,
		}).
		Get(`https://member.bilibili.com/arcopen/fn/archive/type/list`)

	if err != nil {
		return nil, errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return nil, err
	}

	return result.Data.(*ArchiveTypeListResp), nil
}

type UploadInitResp struct {
	UploadToken string `json:"upload_token"`
}

const (
	UploadUTypeSmallFile = 0
	UploadUTypePart      = 1
)

type UploadInitReq struct {
	Name  string `json:"name"`
	UType int    `json:"utype"`
}

func NewUploadInitReq(name string) UploadInitReq {
	return UploadInitReq{
		Name:  name,
		UType: UploadUTypeSmallFile,
	}
}

func (ui UploadInitReq) WithUType(uType int) UploadInitReq {
	ui.UType = uType
	return ui
}

func (a *Archive) UploadInit(accessToken string, req UploadInitReq) (*UploadInitResp, error) {
	result := NewBaseResp(&UploadInitResp{})

	resp, err := resty.New().R().
		SetResult(result).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"client_id":    a.app.appCfg.ClientID,
			"access_token": accessToken,
		}).
		SetBody(req).
		Post(`https://member.bilibili.com/arcopen/fn/archive/video/init`)

	if err != nil {
		return nil, errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return nil, err
	}

	return result.Data.(*UploadInitResp), nil
}

func (a *Archive) UploadPart(uploadToken string, partNumber int, fileReader io.Reader) error {
	result := NewBaseResp(nil)

	resp, err := resty.New().R().
		SetResult(result).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"upload_token": uploadToken,
			"part_number":  strconv.Itoa(partNumber),
		}).
		SetFileReader("file", fmt.Sprintf("%s-%d-%d", uploadToken, partNumber, time.Now().Unix()), fileReader).
		Post(`https://openupos.bilivideo.com/video/v2/part/upload`)

	if err != nil {
		return errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return err
	}

	return nil
}

func (a *Archive) UploadComplete(uploadToken string) error {
	result := NewBaseResp(nil)

	resp, err := resty.New().R().
		SetResult(result).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"upload_token": uploadToken,
		}).
		Post(`https://member.bilibili.com/arcopen/fn/archive/video/complete`)

	if err != nil {
		return errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return err
	}
	return nil
}

type ArchiveSubmitResp struct {
	ResourceID string `json:"resource_id"`
}

type ArchiveSubmitReq struct {
	//Mid       int    `json:"mid"`
	Title     string `json:"title"`      // 稿件标题，长度小于80，短时间内标题不能相同
	Cover     string `json:"cover"`      // 封面地址，必须由上传封面接口得到
	Tid       int    `json:"tid"`        // 分区id，由获取分区信息接口得到
	Tag       string `json:"tag"`        // 视频标签，多个标签用英文逗号分隔，总长度小于200
	Desc      string `json:"desc"`       // 视频描述，长度小于250
	Copyright int    `json:"copyright"`  // 1-原创，2-转载(转载时source必填)
	Source    string `json:"source"`     // 如果copyright为转载，则此字段表示转载来源
	NoReprint int    `json:"no_reprint"` // 是否允许转载 0-允许，1-不允许。默认0
	TopicID   int    `json:"topic_id"`   // 参加的话题ID，默认情况下不填写，需要填写和运营联系
}

// Submit 提交稿件
// 调用该接口完成稿件投递
// 调用此接口前需要保证视频/封面上传完成
// 稿件提交之后会存在审核过程，期间不对外开放
// 非正式会员单日最多投递5个稿件，可在主站通过答题转为正式会员解除限制
func (a *Archive) Submit(accessToken, uploadToken string, req ArchiveSubmitReq) (*ArchiveSubmitResp, error) {
	result := NewBaseResp(&ArchiveSubmitResp{})

	resp, err := resty.New().R().
		SetResult(result).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"client_id":    a.app.appCfg.ClientID,
			"access_token": accessToken,
			"upload_token": uploadToken,
		}).
		SetBody(req).
		Post(`https://member.bilibili.com/arcopen/fn/archive/add-by-utoken`)

	if err != nil {
		return nil, errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return nil, err
	}

	return result.Data.(*ArchiveSubmitResp), nil
}

type ArchiveUploadCoverResp struct {
	Url string `json:"url"`
}

func (a *Archive) UploadCover(accessToken string, fileReader io.Reader) (*ArchiveUploadCoverResp, error) {
	result := NewBaseResp(&ArchiveUploadCoverResp{})

	resp, err := resty.New().R().
		SetResult(result).
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{
			"client_id":    a.app.appCfg.ClientID,
			"access_token": accessToken,
		}).
		SetFileReader("file", fmt.Sprintf("%s-%d", accessToken, time.Now().Unix()), fileReader).
		Post(`https://uat-member.bilibili.com/arcopen/fn/archive/cover/upload`)

	if err != nil {
		return nil, errors.Wrapf(err, "do request fail")
	}

	if err = checkResp(resp, result); err != nil {
		return nil, err
	}

	return result.Data.(*ArchiveUploadCoverResp), nil
}

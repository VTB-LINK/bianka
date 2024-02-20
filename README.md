# Bianka (bilibili开放平台go-sdk)

*For Bianka·Ataegina*

## 介绍

**Bianka** 是一个基于go语言的bilibili开放平台sdk，目前支持以下功能：

- [live（互动开放平台）](https://open-live.bilibili.com/)
    - [x] 项目开启
    - [x] 项目关闭
    - [x] 项目心跳
    - [x] 项目批量心跳
    - [x] 直播间长连接(全事件支持)
    - H5-API
        - [x] 请求签名解析
        - [x] 请求签名验证
- [openhome(开放平台)](https://openhome.bilibili.com/)
  - 直播能力
    - [x] 获取直播间基础信息
    - [x] 直播间长连接及心跳ID
    - [x] 直播间心跳
    - [x] 直播间批量心跳
    - [x] 直播间长连接(全事件支持)
  - 账号管理
    - [x] 账号授权
    - [x] 令牌刷新
  - 用户管理
    - [x] 查询用户已授权权限列表
    - [x] 获取用户公开信息
  - 视频稿件管理
    - [x] 视频稿件投递
    - [x] 视频稿件查询
    - [x] 视频稿件编辑
    - [x] 视频稿件删除
  - 专栏稿件管理
    - [ ] 文章管理
    - [ ] 文集管理
    - [ ] 图片上传
  - 数据开放服务
    - [x] 用户数据
    - [x] 视频数据
    - [x] 专栏数据 
  - [ ] 活动接入
  - [ ] 服务市场
  - [ ] 数据资源接入
  - [ ] WebHook

## 先决条件

- go 1.20+
- bilibili 互动开放平台
    - 申请方式：[成为开发者并获取开发密钥](https://open-live.bilibili.com/document/849b924b-b421-8586-3e5e-765a72ec3840)
- bilibili 开放平台
  - 申请方式：[成为开发者并获取开发密钥](https://openhome.bilibili.com/company/add)

## 安装

```shell
  go get github.com/vtb-link/bianka
```

## 快速开始

具体使用方法可以参考[`example`](https://github.com/VTB-LINK/bianka/tree/main/example)目录下的例子

### 直播间长连接

```go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/vtb-link/bianka/basic"
    "github.com/vtb-link/bianka/live"
    "github.com/vtb-link/bianka/openhome"
    "github.com/vtb-link/bianka/proto"
)

func messageHandle(wsClient *basic.WsClient, msg *proto.Message) error {
    // 单条消息raw 如果需要自己解析可以使用
    log.Println(string(msg.Payload()))

    // sdk提供了自动解析消息的方法，可以快速解析为对应的cmd和data
    // 具体的cmd 可以参考 proto/cmd.go
    cmd, data, err := proto.AutomaticParsingMessageCommand(msg.Payload())
    if err != nil {
        return err
    }

    // 你可以使用cmd进行switch
    switch cmd {
    case proto.CmdLiveOpenPlatformDanmu:
        log.Println(cmd, data.(*proto.CmdDanmuData))
    }

    // 也可以使用data进行switch
    switch v := data.(type) {
    case *proto.CmdGuardData:
        log.Println(cmd, v)
    }

    return nil
}

func main() {
    appID := 123456 // 申请的appID
    sdkConfig := live.NewConfig("申请的key", "申请的secret", int64(appID))

    // 创建sdk实例
    sdk := live.NewClient(sdkConfig)

    code := "获取的主播身份码" // 申请的主播身份码

    // app start 
    startResp, err := sdk.AppStart(code)
    if err != nil {
        panic(err)
    }

    // 启用项目心跳 20s一次
    // see https://open-live.bilibili.com/document/eba8e2e1-847d-e908-2e5c-7a1ec7d9266f
    tk := time.NewTicker(time.Second * 20)
    go func() {
        for {
            select {
            case <-tk.C:
                // 心跳
                if err := sdk.AppHeartbeat(startResp.GameInfo.GameID); err != nil {
                    log.Println("Heartbeat fail", err)
                }

                // 如果需要批量心跳，可以使用以下方法
                // gameIDs := []string{}
                // if err := sdk.AppBatchHeartbeat(gameIDs); err != nil {
            }
        }
    }()

    // app end
    defer func() {
        tk.Stop()
        sdk.AppEnd(startResp.GameInfo.GameID)
    }()
    
    //{
    //    // 如果想使用openhome开放平台
    //    appClient := openhome.NewAppClient(&openhome.AppConfig{
    //        ClientID:     "申请的clientID",
    //        ClientSecret: "申请的clientSecret",
    //    })
    //
    //    startResp, err := appClient.Live.WsStart("用户accessToken")
    //    if err != nil {
    //        panic(err)
    //    }
    //
    //    // 注意开放平台必须保持心跳
    //    tk := time.NewTicker(time.Second * 20)
    //    go func() {
    //        for {
    //            select {
    //            case <-ctx.Done():
    //                return
    //            case <-tk.C:
    //                // 心跳
    //                log.Println("WsHeartbeat")
    //                if err := appClient.Live.WsHeartbeat(accessToken.AccessToken, startResp.ConnID); err != nil {
    //                    log.Println("Heartbeat fail", err)
    //                }
    //            }
    //        }
    //    }()
    //}

    // 注册事件处理器
    // 如果需要注册其他事件，可以参考 proto/op.go
    // SDK 已经默认处理了心跳和鉴权事件，所以目前为止只需要注册 proto.OperationMessage
    // 注意：注册的事件处理器内不要做耗时操作，如果需要做耗时操作，请创建新的goroutine
    dispatcherHandleMap := basic.DispatcherHandleMap{
        proto.OperationMessage: messageHandle,
    }

    // 关闭回调事件
    // 此事件会在websocket连接关闭后触发
    // 时序如下：
    // 0. send close message // 主动发送关闭消息
    // 1. close eventLoop // 不再处理任何消息
    // 2. close websocket // 关闭websocket连接
    // 3. onCloseCallback // 触发关闭回调事件
    // 增加了closeType 参数, 用于区分关闭类型
    onCloseCallback := func(wcs *basic.WsClient, startResp basic.StartResp, closeType int) {
        // 注册关闭回调
        log.Println("WebsocketClient onClose", startResp)

        // 注意检查关闭类型, 避免无限重连
        if closeType == live.CloseActively || closeType == live.CloseReceivedShutdownMessage || closeType == live.CloseAuthFailed {
            log.Println("WebsocketClient exit")
            return
        }

        // 对于可能的情况下重新连接
        // 注意: 在某些场景下 startResp 会变化, 需要重新获取
        // 此外, 一但 AppHeartbeat 失败, 会导致 startResp.GameInfo.GameID 变化, 需要重新获取
        err := wcs.Reconnection(startResp)
        if err != nil {
            log.Println("Reconnection fail", err)
        }
    }

    // 一键开启websocket
    wsClient, err := basic.StartWebsocket(startResp, dispatcherHandleMap, onCloseCallback, basic.DefaultLoggerGenerator())
    if err != nil {
        panic(err)
    }

    defer wsClient.Close()

    // 监听退出信号
    c := make(chan os.Signal, 1)
    signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
    for {
        s := <-c
        switch s {
        case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
            log.Println("WebsocketClient exit")
            return
        default:
            return
        }
    }
}

```

### H5-API

```go
package main

import (
    "log"
    "net/http"

    "github.com/vtb-link/bianka/live"
)

const (
    mockSecret = "NPRZADNURSKNGYDFMDKJOOTLQMGDHL"
    mockURL    = "https://play-live.bilibili.com/plugins-full/1234567?Timestamp=1650012983&Code=460803&Mid=110000345&Caller=bilibili&CodeSign=8c1fa83955d83960680277122bd31fd6f209a82787d57912c1d3817487bfc2ef"
)

var mockRequest, _ = http.NewRequest(http.MethodGet, mockURL, nil)

func main() {
    appID := 1234567 // 申请的appID
    sdkConfig := live.NewConfig("申请的key", mockSecret, int64(appID))

    // 创建sdk实例
    sdk := live.NewClient(sdkConfig)

    // 如果只是验证请求签名
    ok := sdk.VerifyH5RequestSignature(mockRequest)
    if !ok {
        log.Println("VerifyH5RequestSignature fail")
        return
    }

    // 推荐先解析然后再验证
    h5sp := live.ParseH5SignatureParamsWithRequest(mockRequest)

    // 验证请求签名
    ok = sdk.VerifyH5RequestSignatureWithParams(h5sp)
    // 或者
    ok = h5sp.ValidateSignature(mockSecret)

    // 这样的好处是，可以在验证签名后，直接使用h5sp中的参数
    log.Println(h5sp)
}
```

## 错误处理

`bianka`中的错误处理使用了`github.com/pkg/errors`，所以你可以使用`errors.Cause`来获取原始错误。
同时`bianka`也提供了一些预定义的错误，你可以使用`errors.Is`来判断错误类型。

## 自定义使用

bianka 既提供高级封装，也提供了低级封装，如果你需要自定义使用，可以参考以下方法

**注意：使用这些方法我会默认你对go语言以及bilibili开放平台有一定的了解,所以我将不做太多说明.**

### 仅使用 proto进行数据解析

```go
package main

import (
    "github.com/vtb-link/bianka/proto"
)

func main() {
    // 读取到的长连数据
    raw := []byte("xxxxxx")
    // 解析数据
    msgList, err := proto.UnpackMessage(raw)

    sendRaw := []byte("xxxxxx")
    // pack msg
    msg := proto.PackMessage(proto.HeaderDefaultSequence, proto.OperationMessage, sendRaw)

    // 如果仅仅需要pack header
    header := proto.PackHeader(proto.HeaderDefaultSequence, proto.OperationMessage)

    // 如果只需要unpack header
    header, err := proto.UnpackHeader(raw[:proto.PackageHeaderTotalLength])
}
```

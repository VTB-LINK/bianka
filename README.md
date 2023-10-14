# Bianka (bilibili开放平台go-sdk)

*For Bianka·Ataegina*

## 介绍
**Bianka** 是一个基于go语言的bilibili开放平台sdk，目前支持以下功能：

- live（直播）
  - [x] 项目开启
  - [x] 项目关闭
  - [x] 项目心跳
  - [x] 项目批量心跳
  - 直播间长连接
    - [x] websocket 连接
    - [x] 心跳
    - [x] 鉴权
    - [x] 弹幕
    - [x] 礼物
    - [x] super chat（上下线）
    - [x] 上舰（开通大航海）
    - [x] 点赞
  - H5-API
    - [x] 请求签名解析
    - [x] 请求签名验证


## 先决条件
- go 1.20+
- bilibili开放平台账号
  - 申请方式：[成为开发者并获取开发密钥](https://open-live.bilibili.com/document/849b924b-b421-8586-3e5e-765a72ec3840)


## 安装
```shell
  go get vlink.dev/vtb-live/bianka
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

	"vlink.dev/vtb-live/bianka/live"
	"vlink.dev/vtb-live/bianka/proto"
)

func messageHandle(msg *proto.Message) error {
	// 单条消息raw 如果需要自己解析可以使用
	log.Println(string(msg.Payload()))

	// sdk提供了自动解析消息的方法，可以快速解析为对应的cmd和data
	// 具体的cmd 可以参考 live/cmd.go
	cmd, data, err := live.AutomaticParsingMessageCommand(msg.Payload())
	if err != nil {
		return err
	}

	// 你可以使用cmd进行switch
	switch cmd {
	case live.CmdLiveOpenPlatformDanmu:
		log.Println(cmd, data.(*live.CmdLiveOpenPlatformDanmuData))
	}

	// 也可以使用data进行switch
	switch v := data.(type) {
	case *live.CmdLiveOpenPlatformGuardData:
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

	// 注册事件处理器
	// 如果需要注册其他事件，可以参考 proto/op.go
	// SDK 已经默认处理了心跳和鉴权事件，所以目前为止只需要注册 proto.OperationMessage
	// 注意：注册的事件处理器内不要做耗时操作，如果需要做耗时操作，请创建新的goroutine
	dispatcherHandle := map[uint32]live.DispatcherHandle{
		proto.OperationMessage: messageHandle,
	}

	// 关闭回调事件
	// 此事件会在websocket连接关闭后触发
	// 时序如下：
	// 1. close eventLoop // 不再处理任何消息
	// 2. close websocket // 关闭websocket连接
	// 3. onCloseCallback // 触发关闭回调事件
	onCloseCallback := func(startResp *live.AppStartResponse) {
		log.Println("WebsocketClient onClose", startResp)
	}

	// 一键开启websocket
	wsClient, err := sdk.StartWebsocket(startResp, dispatcherHandle, onCloseCallback)
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

	"vlink.dev/vtb-live/bianka/live"
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

## 自定义使用

bianka 既提供高级封装，也提供了低级封装，如果你需要自定义使用，可以参考以下方法

**注意：使用这些方法我会默认你对go语言以及bilibili开放平台有一定的了解,所以我将不做太多说明.**

### 仅使用 proto进行数据解析
```go
package main

import (
    "vlink.dev/vtb-live/bianka/proto"
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

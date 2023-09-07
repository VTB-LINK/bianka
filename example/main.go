package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"vlink.dev/vtb-live/bianca-danmu/live"
	"vlink.dev/vtb-live/bianca-danmu/proto"
)

var rCfg = live.RoomConfig{
	AccessKey:            "申请的key",
	AccessKeySecret:      "申请的secret",
	OpenPlatformHttpHost: "https://live-open.biliapi.com", //开放平台 (线上环境)
	IdCode:               "开播主播的身份码",
	AppID:                0, // 应用id
}

func main() {
	liveClient := live.NewClient(&rCfg)

	startResp, err := liveClient.StartApp()
	if err != nil {
		panic(err)
	}

	defer liveClient.EndApp(startResp)

	wcs, err := liveClient.StartWebsocket(startResp, map[uint32]live.DispatcherHandle{
		proto.OperationMessage: messageHandle,
	})

	if err != nil {
		panic(err)
	}

	defer wcs.Close()

	// 退出
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			log.Println("WebsocketClient exit")
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

func messageHandle(msg *proto.Message) error {
	// 单条消息raw
	fmt.Println(string(msg.Payload()))

	// 自动解析
	cmd, data, err := live.AutomaticParsingMessageCommand(msg.Payload())
	if err != nil {
		return err
	}

	// Switch cmd
	switch cmd {
	case live.CmdLiveOpenPlatformDanmu:
		fmt.Println(cmd, data.(*live.CmdLiveOpenPlatformDanmuData))
	}

	return nil
}

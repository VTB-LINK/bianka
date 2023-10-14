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

var rCfg = live.NewConfig(
	"申请的key",
	"申请的secret",
	0, // 应用id
)

var code = "主播的code" // 身份码 也叫 idCode

func main() {
	liveClient := live.NewClient(rCfg)

	startResp, err := liveClient.AppStart(code)
	if err != nil {
		panic(err)
	}

	tk := time.NewTicker(time.Second * 20)
	go func() {
		for {
			select {
			case <-tk.C:
				// 心跳
				if err := liveClient.AppHeartbeat(startResp.GameInfo.GameID); err != nil {
					log.Println("Heartbeat fail", err)
				}
			}
		}
	}()

	defer func() {
		tk.Stop()
		liveClient.AppEnd(startResp.GameInfo.GameID)
	}()

	wcs, err := liveClient.StartWebsocket(startResp, map[uint32]live.DispatcherHandle{
		proto.OperationMessage: messageHandle,
	}, func(startResp *live.AppStartResponse) {
		// 注册关闭回调
		log.Println("WebsocketClient onClose", startResp)
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
	log.Println(string(msg.Payload()))

	// 自动解析
	cmd, data, err := live.AutomaticParsingMessageCommand(msg.Payload())
	if err != nil {
		return err
	}

	// Switch cmd
	switch cmd {
	case live.CmdLiveOpenPlatformDanmu:
		log.Println(cmd, data.(*live.CmdLiveOpenPlatformDanmuData))
	}

	// Switch data type
	switch v := data.(type) {
	case *live.CmdLiveOpenPlatformGuardData:
		log.Println(cmd, v)
	}

	return nil
}

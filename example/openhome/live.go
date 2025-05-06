package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vtb-link/bianka/basic"
	"github.com/vtb-link/bianka/openhome"
	"github.com/vtb-link/bianka/proto"
)

var appCfg = &openhome.AppConfig{
	ClientID:     "开放平台的ClientID",
	ClientSecret: "开放平台的ClientSecret",
}

var accessToken = &openhome.AccessToken{
	AccessToken:  "授权后获取的AccessToken",
	RefreshToken: "授权后获取的RefreshToken",
	ExpiresIn:    1723543509,
}

func main() {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	appClient := openhome.NewAppClient(appCfg)

	startResp, err := appClient.Live.WsStart(accessToken.AccessToken)
	if err != nil {
		panic(err)
	}

	tk := time.NewTicker(time.Second * 20)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-tk.C:
				// 心跳
				log.Println("WsHeartbeat")
				if err := appClient.Live.WsHeartbeat(accessToken.AccessToken, startResp.ConnID); err != nil {
					log.Println("Heartbeat fail", err)
				}
			}
		}
	}()

	// close 事件处理
	onCloseHandle := func(wcs *basic.WsClient, startResp basic.StartResp, closeType int) {
		// 注册关闭回调
		log.Println("WebsocketClient onClose", startResp)

		// 注意检查关闭类型, 避免无限重连
		if closeType == basic.CloseActively || closeType == basic.CloseReceivedShutdownMessage || closeType == basic.CloseAuthFailed {
			log.Println("WebsocketClient exit")
			return
		}

		// 对于可能的情况下重新连接
		// 注意: 在某些场景下 startResp 会变化, 需要重新获取
		// 此外, 一但 WsHeartbeat 失败, 会导致 startResp.ConnID 变化, 需要重新获取
		err := wcs.Reconnection(startResp)
		if err != nil {
			log.Println("Reconnection fail", err)
		}
	}

	// 消息处理 Handle
	dispatcherHandleMap := basic.DispatcherHandleMap{
		proto.OperationMessage: messageHandle,
	}

	wcs, err := basic.StartWebsocket(startResp, dispatcherHandleMap, onCloseHandle, basic.DefaultLoggerGenerator())

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

func messageHandle(wcs *basic.WsClient, msg *proto.Message) error {
	// 单条消息raw
	log.Println(string(msg.Payload()))

	//// 自动解析
	//cmd, data, err := proto.AutomaticParsingMessageCommand(msg.Payload())
	//if err != nil {
	//    return err
	//}
	//
	//// Switch cmd
	//// 由于open-home 和open-live 的命令并不一致 所以更推荐使用 switch data.(type) 进行类型判断,不然你就要自己case 2个cmd
	//switch cmd {
	//case proto.CmdLiveOpenPlatformDanmu:
	//    log.Println(cmd, data.(*proto.CmdDanmuData))
	//}
	//
	//// Switch data type
	//switch v := data.(type) {
	//case *proto.CmdDanmuData:
	//    log.Println(cmd, v)
	//}
	//
	return nil
}

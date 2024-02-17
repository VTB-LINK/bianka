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

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vtb-link/bianka/basic"
	"github.com/vtb-link/bianka/live"
	"github.com/vtb-link/bianka/proto"
)

var rCfg = live.NewConfig(
	"申请的key",
	"申请的secret",
	0, // 应用id
)

var code = "主播的code" // 身份码 也叫 idCode

func main() {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	liveClient := live.NewClient(rCfg)

	startResp, err := liveClient.AppStart(code)
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
		// 此外, 一但 AppHeartbeat 失败, 会导致 startResp.GameInfo.GameID 变化, 需要重新获取
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

package live

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"golang.org/x/exp/slog"
	"vlink.dev/vtb-live/bianca-danmu/proto"
)

type RoomConfig struct {
	AccessKey            string //access_key
	AccessKeySecret      string //access_key_secret
	OpenPlatformHttpHost string //开放平台 (线上环境)
	IDCode               string // 主播身份码
	AppID                int64  // 应用id
}

type DispatcherHandle func(msg *proto.Message) error

type WsClient struct {
	conn *websocket.Conn
	rCfg *RoomConfig

	once       sync.Once
	closeChan  chan struct{}
	msgChan    chan *proto.Message
	dispatcher map[uint32]DispatcherHandle

	authed bool

	logger *slog.Logger
}

func NewWsClient(rCfg *RoomConfig, dispatcherHandleMap map[uint32]DispatcherHandle, logger *slog.Logger) *WsClient {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	return (&WsClient{
		rCfg:      rCfg,
		closeChan: make(chan struct{}),
		msgChan:   make(chan *proto.Message, 1024),
		logger:    logger.WithGroup("WebsocketClient"),
	}).initDispatcherHandleMap(dispatcherHandleMap)
}

func (wc *WsClient) initDispatcherHandleMap(dispatcherHandleMap map[uint32]DispatcherHandle) *WsClient {
	if dispatcherHandleMap == nil {
		dispatcherHandleMap = make(map[uint32]DispatcherHandle)
	}

	// 注册分发处理函数
	dispatcherHandleMap[proto.OperationUserAuthenticationReply] = wc.authResp
	dispatcherHandleMap[proto.OperationHeartbeatReply] = wc.heartBeatResp

	wc.dispatcher = dispatcherHandleMap
	return wc
}

func (wc *WsClient) Close() error {
	var err error
	wc.once.Do(func() {
		close(wc.closeChan)
		err = wc.conn.Close()
	})

	return err
}

// Dial 链接
func (wc *WsClient) Dial(links []string) error {
	var err error
	for _, link := range links {
		wc.conn, _, err = websocket.DefaultDialer.Dial(link, nil)
		if err != nil {
			continue
		}
		break
	}

	if err != nil {
		return errors.Wrapf(err, "websocket dial fail. links:%v", links)
	}

	wc.logger.Info("dial success")
	return nil
}

// EventLoop 处理事件
func (wc *WsClient) EventLoop() {
	defer wc.logger.Info("ws event loop stop")

	ticker := time.NewTicker(time.Second * 15)
	tr := time.NewTimer(time.Second * 5)
	for {
		select {
		case <-wc.closeChan:
			return
		case <-tr.C:
			if !wc.authed {
				wc.logger.Error("auth timeout")
				_ = wc.Close()
				return
			}
		case <-ticker.C:
			wc.logger.Info("ws send heartbeat")
			if err := wc.SendHeartbeat(); err != nil {
				wc.logger.Error("send heartbeat fail", slog.String("err", err.Error()))
			}
		case msg := <-wc.msgChan:
			if msg == nil {
				continue
			}

			handle, ok := wc.dispatcher[msg.Operation()]
			if ok {
				if err := handle(msg); err != nil {
					wc.logger.Error("handle msg fail", slog.String("err", err.Error()))
				}
			}
		}
	}
}

// ReadMsg 读取信息
func (wc *WsClient) ReadMsg() {
	defer wc.logger.Info("ws read stop")

	for {
		select {
		case <-wc.closeChan:
			return
		default:
			// 读取err or read close message 会导致关闭链接
			msgType, buf, err := wc.conn.ReadMessage()
			if err != nil {
				wc.logger.Error("read message fail", slog.String("err", err.Error()))
				_ = wc.Close()
				return
			} else if msgType == websocket.CloseMessage {
				wc.logger.Error("read message fail", slog.String("err", "read close message"))
				_ = wc.Close()
				return
			} else if msgType == websocket.PongMessage || msgType == websocket.PingMessage {
				wc.logger.Info("read message", slog.String("msg_type", "ping/pong"))
				continue
			}

			msgList, err := proto.UnpackMessage(buf)
			if err != nil {
				wc.logger.Error("unpack message fail", slog.String("err", err.Error()))
				continue
			}

			for _, msg := range msgList {
				wc.msgChan <- &msg
			}
		}
	}
}

// SendAuth 发送鉴权信息
func (wc *WsClient) SendAuth(authBody string) error {
	return wc.SendMessage(proto.PackMessage(proto.HeaderDefaultSequence, proto.OperationUserAuthentication, []byte(authBody)))
}

// SendMessage 发送消息
func (wc *WsClient) SendMessage(msg proto.Message) error {
	err := wc.conn.WriteMessage(websocket.BinaryMessage, msg.ToBytes())
	if err != nil {
		return errors.Wrapf(err, "send message fail. payload:%s", msg.Payload())
	}

	return nil
}

// SendHeartbeat 发送心跳
func (wc *WsClient) SendHeartbeat() error {
	return wc.SendMessage(proto.PackMessage(proto.HeaderDefaultSequence, proto.OperationHeartbeat, nil))
}

type AuthRespParam struct {
	Code int64 `json:"code,omitempty"`
}

// authResp  认证结果
func (wc *WsClient) authResp(msg *proto.Message) error {
	defer func() {
		// 鉴权失败，关闭链接
		if !wc.authed {
			_ = wc.Close()
		}
	}()

	resp := &AuthRespParam{}

	if err := json.Unmarshal(msg.Payload(), resp); err != nil {
		return errors.Wrapf(err, "json unmarshal fail. payload:%s", msg.Payload())
	}

	if resp.Code != 0 {
		return errors.New(fmt.Sprintf("auth fail. code:%d", resp.Code))
	}

	wc.logger.Info("auth success")
	wc.authed = true
	return nil
}

// heartBeatResp  心跳结果
func (wc *WsClient) heartBeatResp(msg *proto.Message) (err error) {
	wc.logger.Info("heartbeat success", slog.String("payload", string(msg.Payload())))
	return
}

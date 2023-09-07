package live

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"vlink.dev/vtb-live/bianca-danmu/proto"
)

type RoomConfig struct {
	AccessKey            string //access_key
	AccessKeySecret      string //access_key_secret
	OpenPlatformHttpHost string //开放平台 (线上环境)
	IdCode               string // 主播身份码
	AppID                int64  // 应用id
}

type DispatcherHandle func(msg *proto.Message) error

type WsClient struct {
	conn *websocket.Conn
	rCfg *RoomConfig

	closeChan  chan struct{}
	msgChan    chan *proto.Message
	dispatcher map[uint32]DispatcherHandle

	authed bool
}

func NewWsClient(rCfg *RoomConfig, dispatcherHandleMap map[uint32]DispatcherHandle) *WsClient {
	return (&WsClient{
		rCfg:      rCfg,
		closeChan: make(chan struct{}),
		msgChan:   make(chan *proto.Message, 1024),
	}).initDispatcherHandleMap(dispatcherHandleMap)
}

func (wc *WsClient) initDispatcherHandleMap(dispatcherHandleMap map[uint32]DispatcherHandle) *WsClient {
	if dispatcherHandleMap == nil {
		dispatcherHandleMap = make(map[uint32]DispatcherHandle)
	}

	wc.dispatcher = dispatcherHandleMap

	// 注册分发处理函数
	dispatcherHandleMap[proto.OperationUserAuthenticationReply] = wc.authResp
	dispatcherHandleMap[proto.OperationHeartbeatReply] = wc.heartBeatResp

	return wc
}

func (wc *WsClient) Close() error {
	close(wc.closeChan)
	return wc.conn.Close()
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
		return errors.Wrap(err, "websocket dial fail")
	}

	log.Println("[WebsocketClient | Dial] dial success")
	return nil
}

// EventLoop 处理事件
func (wc *WsClient) EventLoop() {
	ticker := time.NewTicker(time.Second * 15)
	for {
		select {
		case <-wc.closeChan:
			log.Println("[WebsocketClient | EventLoop] close")
			return
		case <-ticker.C:
			log.Println("[WebsocketClient | EventLoop] send heartbeat")
			if err := wc.SendHeartbeat(); err != nil {
				log.Println("[WebsocketClient | EventLoop] send heartbeat fail", err)
			}
		case msg := <-wc.msgChan:
			if msg == nil {
				continue
			}

			handle, ok := wc.dispatcher[msg.Operation()]
			if ok {
				if err := handle(msg); err != nil {
					log.Println("[WebsocketClient | EventLoop] handle msg fail", err)
				}
			}
		}
	}
}

// ReadMsg 读取信息
func (wc *WsClient) ReadMsg() {
	for {
		select {
		case <-wc.closeChan:
			log.Println("[WebsocketClient | ReadMsg] close")
			return
		default:
			// TODO: 处理网络关闭事件等
			_, buf, err := wc.conn.ReadMessage()
			if err != nil {
				log.Println("[WebsocketClient | ReadMsg] read message fail", err)
				continue
			}

			msgList, err := proto.UnpackMessage(buf)
			if err != nil {
				log.Println("[WebsocketClient | ReadMsg] unpack message fail", err)
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
		return errors.Wrap(err, "send message fail")
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
	resp := &AuthRespParam{}

	if err := json.Unmarshal(msg.Payload(), resp); err != nil {
		return errors.Wrap(err, "json unmarshal fail")
	}

	if resp.Code != 0 {
		return errors.New("auth fail")
	}

	log.Println("[WebsocketClient | authResp] auth success")
	wc.authed = true
	return nil
}

// heartBeatResp  心跳结果
func (wc *WsClient) heartBeatResp(msg *proto.Message) (err error) {
	log.Println("[WebsocketClient | heartBeatResp] HeartBeat resp", msg.Payload())
	return
}

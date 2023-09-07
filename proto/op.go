package proto

const (
	OperationHeartbeat               = 2 // 心跳
	OperationHeartbeatReply          = 3 // 心跳回包
	OperationMessage                 = 5 // 消息包
	OperationUserAuthentication      = 7 // 鉴权包
	OperationUserAuthenticationReply = 8 // 鉴权回包
)

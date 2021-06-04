package socket

type SocketEventType int

const (
	SocketEventType_Connect    SocketEventType = 0
	SocketEventType_DisConnect SocketEventType = 1
)

type DisConnectType int

const (
	DisConnectType_Active  DisConnectType = 0 //主动断开
	DisConnectType_Passive DisConnectType = 1 //被动断开
)

type SocketEvent struct {
	EventType        SocketEventType
	Ses              *Session
	DisConnectReason DisConnectType
}

const (
	MainCmd_LogonServerIdBegin = 0   // 发送给登陆服务器的主命令开始
	MainCmd_LogonServerIdEnd   = 99  // 发送给登陆服务器的主命令结束
	MainCmd_GateServerIdBegin  = 100 // 发送给代理服务器的主命令开始
	MainCmd_GateServerIdEnd    = 149 // 发送给代理服务器的主命令开始
	MainCmd_GameServerIdBegin  = 150 // 发送给代理服务器的主命令开始
	MainCmd_GameServerIdEnd    = 249 // 发送给代理服务器的主命令开始
)

const (
	MainGateCmd_Service = 100  // 发送给代理服务器的主命令
	MainGateCmd_Network = 1000 // 网络相关主命令
)

//代理服务器子命令
const (
	SubGateCmd_QueryServerInfo = 0 // 查询服务器信息
	SubGateCmd_KeepAlive       = 1 // 连接到服务器
	SubGateCmd_ServerReady     = 2 // 服务器就绪
	SubGateCmd_ClientAlive     = 3 // 客户端
)

package socket

import (
	"deal/pkg/stack"
	"deal/pkg/utils"
	"net"
	"os"
	"strings"
	"sync"
)

const (
	KEEPALIVE_DETECT_TIME = 5 //心跳定时器间隔
)

type NetworkStatus int

const (
	NetworkStatus_Normal    NetworkStatus = 0 // 正常
	NetworkStatus_KeepAlive NetworkStatus = 1 // 心跳检查
)

const (
	allEthsIpv4 = "0.0.0.0"
	allEthsIpv6 = "[::]"
	envPodIp    = "POD_IP"
)

type ConnectStatusType int

//心跳状态
const (
	KeepAlive_Dead ConnectStatusType = 0 // 死亡连接
	KeepAlive_Warn ConnectStatusType = 1 // 危险连接
	KeepAlive_Safe ConnectStatusType = 2 // 安全连接
)

type SocketAcceptor struct {
	listener  net.Listener
	running   bool
	sockMap   map[uint64]*Session
	fdcounter uint64
	liner     *utils.LinerDispatch
	mutex     sync.Mutex
}

func (s *SocketAcceptor) Start(address string, keepAliveStatus NetworkStatus) error {

	ln, err := net.Listen("tcp", address)

	s.sockMap = make(map[uint64]*Session)

	if err != nil {
		return err
	}

	s.listener = ln

	s.running = true
	stack.SafeCallNull(func() {
		for s.running {
			conn, err := ln.Accept()
			if err != nil {
				continue
			}

			//与客户端的连接需要心跳 需要加密
			encrypt := true
			s.fdcounter++

			ses := NewSession(s.fdcounter, conn, encrypt, s.liner)

			s.mutex.Lock()
			s.sockMap[s.fdcounter] = ses
			s.mutex.Unlock()

			// s.iattemper.OnEventTCPNetworkLink(ses)
			event := &SocketEvent{
				EventType: SocketEventType_Connect,
				Ses:       ses,
			}

			s.liner.Add(event)

			ses.OnClose = func(ises *Session) {
				s.mutex.Lock()
				delete(s.sockMap, ises.ID())
				s.mutex.Unlock()
			}

			ses.SetBindData(KeepAlive_Safe)
		}
	})

	return nil
}

func (s *SocketAcceptor) OnTimer(callback int) {
	var deadSesVec []*Session

	s.mutex.Lock()

	//遍历所有session
	for i := range s.sockMap {

		ses := s.sockMap[i]
		if ses == nil {
			continue
		}

		userData := ses.GetBindData()

		switch status := userData.(type) {
		case ConnectStatusType:
			if status == KeepAlive_Dead {

				deadSesVec = append(deadSesVec, ses)
			} else if status == KeepAlive_Safe {
				status--
				ses.SetBindData(status)

				ses.Send(MainGateCmd_Network, 0, nil)
			} else if status == KeepAlive_Warn {
				status--
				ses.SetBindData(status)

				ses.Send(MainGateCmd_Network, 0, nil)
			}
		}
	}

	s.mutex.Unlock()

	for _, ses := range deadSesVec {
		// s.iattemper.OnEventTCPNetworkShut(ses)
		if ses.OnClose != nil {
			ses.OnClose(ses)
			ses.Close(CloseType_UpLayer)
		}
	}

}

func (s *SocketAcceptor) GetSessionByID(id uint64) *Session {
	value, ok := s.sockMap[id]
	if ok {
		return value
	} else {
		return nil
	}
}

func (s *SocketAcceptor) GetListenIp() string {
	ipStr := s.listener.Addr().String()

	lastIdx := strings.LastIndex(ipStr, ":")
	if lastIdx == -1 {
		return ipStr
	}

	host := ipStr[:lastIdx]
	port := ipStr[lastIdx+1:]

	if len(host) > 0 && host != allEthsIpv4 && host != allEthsIpv6 {
		return ipStr
	}

	ip := os.Getenv(envPodIp)
	if len(ip) == 0 {
		ip = InternalIp()
	}
	if len(ip) == 0 {
		return ipStr
	}

	return strings.Join(append([]string{ip}, port), ":")
}

func (s *SocketAcceptor) Stop() {

	if !s.running {
		return
	}

	s.running = false

	s.listener.Close()
}

func (s *SocketAcceptor) GetNetworkInfo() string {
	return s.listener.Addr().String()
}

func NewAcceptor(liner *utils.LinerDispatch) *SocketAcceptor {

	s := &SocketAcceptor{
		liner:     liner,
		fdcounter: 0,
	}

	return s
}

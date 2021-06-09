package socket

import (
	"bytes"
	"encoding/gob"
	"errors"
	"net"
	"sync"

	"github.com/viwii/utils/stack"
	"github.com/viwii/utils/utils"
)

type StopWrite struct{}

type SocketConnector struct {
	conn             net.Conn             // 连接
	autoReconnectSec int                  // 重连间隔时间, 0为不重连
	closeSignal      chan bool            // 关闭信号
	stream           PacketStream         // 流
	useritem         interface{}          // 绑定的用户数据
	writeChan        chan interface{}     // 写入通道
	iskeep           bool                 // 是否保持长连接
	ip               string               // 连接的地址
	endSync          sync.WaitGroup       // 同步锁
	isdone           bool                 // 标志是否断开
	linerDispatch    *utils.LinerDispatch // 线性分配器
}

//普通封包
type TcpPacket struct {
	MainId  uint16           //主命令ID
	SubId   uint16           //子命令ID
	Data    []byte           //数据
	Connect *SocketConnector //连接器
}

func (s *SocketConnector) Start(ip string) error {
	s.ip = ip
	return s.connect()
}

func (s *SocketConnector) Stop() bool {
	// 通知关闭写协程
	s.writeChan <- StopWrite{}

	// 通知接收线程ok
	s.endSync.Done()

	return true
}

func (s *SocketConnector) connect() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", s.ip)
	if err != nil {
		return err
	}

	s.conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err
	}

	//和中心服务器通讯不需要加密
	s.stream = NewPacketStream(s.conn, false)

	// 退出线程
	stack.SafeCallNull(s.existThread)

	// 接收线程
	stack.SafeCallNull(s.recvThread)

	// 发送线程
	stack.SafeCallNull(s.sendThread)

	s.isdone = false

	//self.iattemper.OnEventTCPSocketLink(self)

	return nil
}

func (s *SocketConnector) SendGobMessage(mainId uint16, subId uint16, e interface{}) error {
	var netdata bytes.Buffer
	enc := gob.NewEncoder(&netdata)
	if err := enc.Encode(e); err != nil {
		return err
	}
	return s.Send(mainId, subId, netdata.Bytes())
}

//发包
func (s *SocketConnector) Send(mainId uint16, subId uint16, data []byte) error {
	pkt := MakePacket(mainId, subId, data, nil)

	if s.isdone {
		return errors.New("connector has done!")
	} else {
		s.writeChan <- pkt
	}

	return nil
}

func (s *SocketConnector) sendThread() {
	for {
		switch pkt := (<-s.writeChan).(type) {
		case *Packet:
			if err := s.stream.Write(pkt); err != nil {
				goto exist_loop
			}
		case *StopWrite:
			goto exist_loop
		}
	}

exist_loop:
	s.stream.Close()
	s.endSync.Done()
}

func (s *SocketConnector) recvThread() {
	for {
		// 从Socket读取封包
		pk, err := s.stream.Read()

		if err != nil {
			break
		}

		var pkt TcpPacket
		pkt.MainId = pk.MainId
		pkt.SubId = pk.SubId
		pkt.Data = pk.Data
		pkt.Connect = s
		s.linerDispatch.Add(&pkt)
	}

	if !s.isdone {
		s.writeChan <- &StopWrite{}

		s.endSync.Done()
	}
}

func (s *SocketConnector) existThread() {
	// 布置接收和发送2个任务
	s.endSync.Add(2)
	// 等待2个任务结束
	s.endSync.Wait()

	//self.iattemper.OnEventTCPSocketShut(self)

	if s.iskeep {
		for {
			err := s.connect()
			if err == nil {
				break
			}
		}

	}
}

func NewConnector(liner *utils.LinerDispatch) *SocketConnector {
	self := &SocketConnector{
		writeChan:     make(chan interface{}),
		isdone:        true,
		iskeep:        true,
		linerDispatch: liner,
	}

	return self
}

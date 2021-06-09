package socket

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/viwii/utils/utils"

	"github.com/golang/protobuf/proto"
)

type CloseType int

var (
	CloseType_UpLayer    CloseType = 0 //上层主动
	CloseType_UnderLayer CloseType = 1 //下层主动
)

type Session struct {
	sockid        uint64
	writeChan     chan interface{}
	OnClose       func(*Session)
	stream        PacketStream
	liner         *utils.LinerDispatch
	userItem      interface{}
	endSync       sync.WaitGroup //协程结束
	needStopWrite bool           //是否需要主动断开写协程
	isDone        bool           //session是否已经关闭
	bindData      interface{}
	sendList      *utils.PacketList //将发包改为发送列表
}

// 发包
func (s *Session) Send(mainId uint16, subId uint16, data []byte) error {
	pkt := MakePacket(mainId, subId, data, s)
	if s.isDone {
		return errors.New("session has done")
	}
	s.sendList.Add(pkt)

	return nil
}

// 断开
func (s *Session) Close(ctype CloseType) {
	if !s.isDone {
		s.isDone = true
		s.Send(InvliadPacketMainID, InvliadPacketSubID, nil)
		s.sendList.Close()
	}

}

// 标示ID
func (s *Session) ID() uint64 {
	return s.sockid
}

func (s *Session) BindUserItem(item interface{}) {
	s.userItem = item
}

func (s *Session) GetBindUserItem() interface{} {
	return s.userItem
}

func (s *Session) GetRemoteAddr() string {
	return s.stream.GetRemoteAddr()
}

func (s *Session) sendThread() {
	var sendList []interface{}
	for true {
		sendList = sendList[0:0]

		// 复制出队列
		packetList := s.sendList.BeginPick()

		sendList = append(sendList, packetList...)

		s.sendList.EndPick()
		willExit := false
		// 写队列
		for _, item := range sendList {
			p := item.(*Packet)

			if p.MainId != 1000 {
				fmt.Println("begin send packet --->>", p.MainId, p.SubId)
			}

			if p.MainId == InvliadPacketMainID {
				goto exitSendLoop
			} else if err := s.stream.Write(p); err != nil {
				willExit = true
				break
			}

			if p.MainId != 1000 {
				fmt.Println("end  send packet --->>", p.MainId, p.SubId)
			}
		}

		if willExit {
			goto exitSendLoop
		}
	}

exitSendLoop:
	// 通知关闭写协程
	s.sendList.Close()
	s.isDone = true
	s.needStopWrite = false
	s.endSync.Done()
}

func (s *Session) recvThread() {
	for {

		pk, err := s.stream.Read()
		if err != nil {
			fmt.Println("recv packet <<---", err)
			break
		}

		if pk.MainId != MainGateCmd_Network { //心跳检测
			if pk.MainId != 1000 {
				fmt.Println("recv packet <<---", pk.MainId, pk.SubId)
			}

			pk.Ses = s
			s.liner.Add(pk)
		} else if pk.SubId == 3 {
			s.Send(MainGateCmd_Network, SubGateCmd_ClientAlive, pk.Data)
		}

		s.SetBindData(KeepAlive_Safe)

	}

	if s.needStopWrite {
		s.Send(InvliadPacketMainID, InvliadPacketSubID, nil)
	}
	s.isDone = true

	s.endSync.Done()
}

func (s *Session) Recv(uint16, uint16, []byte) bool {
	return false
}

func (s *Session) existThread() {
	s.endSync.Wait()

	event := &SocketEvent{
		EventType: SocketEventType_DisConnect,
		Ses:       s,
	}

	s.liner.Add(event)

	if s.OnClose != nil {
		s.OnClose(s)
	}

	s.stream.Close()
}

func (s *Session) SetBindData(data interface{}) {
	s.bindData = data
}

func (s *Session) GetBindData() interface{} {
	return s.bindData
}

func (s *Session) SendPbMessage(mainId uint16, subId uint16, pb proto.Message) error {
	data, err := proto.Marshal(pb)
	if err != nil {
		return err
	}
	return s.Send(mainId, subId, data)
}

func (s *Session) SendGobMessage(mainId uint16, subId uint16, e interface{}) error {
	var netdata bytes.Buffer
	enc := gob.NewEncoder(&netdata)
	if err := enc.Encode(e); err != nil {
		return err
	}
	return s.Send(mainId, subId, netdata.Bytes())
}

func NewSession(sockid uint64, tcpcon net.Conn, isEncrypt bool, liner *utils.LinerDispatch) *Session {
	s := &Session{
		writeChan:     make(chan interface{}),
		sockid:        sockid,
		liner:         liner,
		isDone:        false,
		needStopWrite: true,
		bindData:      nil,
		sendList:      utils.NewPacketList(),
	}

	s.stream = NewPacketStream(tcpcon, isEncrypt)

	s.endSync.Add(2)

	go s.existThread()

	go s.recvThread()

	go s.sendThread()

	return s
}

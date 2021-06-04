package socket

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
)

const (
	PacketCommandSize = 4 //MainId(uint16) + SubId(uint16)
	PacketHeaderSize  = 5 //Msgflag(uint16) + MsgType(uint8) + MsgSize(uint16)
	PacketedHeadSize  = 9 //Msgflag(uint16) + MsgType(uint8) + MsgSize(uint16) + MainId(uint16) + SubId(uint16)
	//MaxPacketSize     = 1024 * 8 //包的长度
)

// 封包流
type PacketStream interface {
	Read() (*Packet, error)
	Write(pkt *Packet) error
	Close() error
	GetRemoteAddr() string
}

type Stream struct {
	conn    net.Conn
	encrypt bool
}

var (
	packetTagNotMatch     = errors.New("ReadPacket: packet tag not match")
	packetDataSizeInvalid = errors.New("ReadPacket: packet crack, invalid size")
	packetTooBig          = errors.New("ReadPacket: packet too big")
)

// 从socket读取1个封包,并返回
func (s *Stream) Read() (p *Packet, err error) {

	//读取包头
	headdata := make([]byte, PacketHeaderSize)
	if _, err = io.ReadFull(s.conn, headdata); err != nil {
		return nil, err
	}
	//读取MsgTag
	msgTag := uint16(binary.BigEndian.Uint16(headdata[0:2]))

	//非法包
	if msgTag != MSG_TAG {
		return nil, packetTagNotMatch
	}

	//读取包类型
	msgType := headdata[2]

	//读取整包大小
	fullSize := uint16(binary.BigEndian.Uint16(headdata[3:5]))

	dataSize := fullSize - PacketHeaderSize
	if dataSize < 0 {
		return nil, packetDataSizeInvalid
	}

	//读取加密数据
	packData := make([]byte, dataSize)
	if _, err = io.ReadFull(s.conn, packData); err != nil {
		return nil, err
	}

	//效验接收到的大小
	if packData == nil || dataSize != uint16(len(packData)) {
		return nil, err
	}

	p, err = ToDecryptPacket(packData, msgType)
	if err != nil {
		return nil, err
	}

	return p, nil
}

//将一个封包发送到socket
func (s *Stream) Write(pkt *Packet) (err error) {

	buffer := make([]byte, PacketHeaderSize)

	//写MsgTag
	binary.BigEndian.PutUint16(buffer[0:2], uint16(MSG_TAG))

	var userdata []byte
	userdata = pkt.ToByteArray()

	//写MsgSize
	packetsize := len(userdata) + PacketHeaderSize
	binary.BigEndian.PutUint16(buffer[3:5], uint16(packetsize))

	//写数据
	buffer = append(buffer[:PacketHeaderSize], userdata...)
	if _, err = s.conn.Write(buffer); err != nil {
		return err
	}

	return nil
}

func (s *Stream) GetRemoteAddr() string {
	ipstring, _, _ := net.SplitHostPort(s.conn.RemoteAddr().String())
	return ipstring
}

//关闭
func (s *Stream) Close() error {
	return s.conn.Close()
}

//封包流 relay模式: 在封包头有clientid信息
func NewPacketStream(conn net.Conn, encrypt bool) PacketStream {
	return &Stream{
		conn:    conn,
		encrypt: encrypt,
	}
}

package socket

import (
	"encoding/binary"
	"errors"
)

const (
	MSG_TAG   = 0x6279
	XXTEA_KEY = "6279baiyikejitest"
)

const (
	ENCRYPT_NON   = 0
	ENCRYPT_XXTEA = 1 << 0
)

const (
	InvliadPacketMainID = uint16(0xFFFF)
	InvliadPacketSubID  = uint16(0xFFFF)
)

// 普通封包
type Packet struct {
	MainId uint16   // 主命令ID
	SubId  uint16   // 子命令ID
	Data   []byte   // 数据
	Ses    *Session // Session
}

func (s *Packet) ToByteArray() []byte {

	buffer := make([]byte, PacketCommandSize)

	binary.BigEndian.PutUint16(buffer[0:2], s.MainId)

	binary.BigEndian.PutUint16(buffer[2:4], s.SubId)

	buffer = append(buffer[0:PacketCommandSize], s.Data...)

	return buffer
}

func (s *Packet) ToEncryptByteArray() ([]byte, uint8) {

	buffer := make([]byte, PacketCommandSize)

	binary.BigEndian.PutUint16(buffer[0:2], s.MainId)

	binary.BigEndian.PutUint16(buffer[2:4], s.SubId)

	buffer = append(buffer[0:PacketCommandSize], s.Data...)

	msgType := uint8(ENCRYPT_XXTEA)

	return buffer, msgType
}

func ToDecryptPacket(data []byte, msgType uint8) (p *Packet, err error) {
	if len(data) < 4 {
		return nil, errors.New("Packet length is too short")
	}

	p = &Packet{}

	p.MainId = uint16(binary.BigEndian.Uint16(data[0:2]))

	p.SubId = uint16(binary.BigEndian.Uint16(data[2:4]))

	p.Data = data[4:]

	return p, nil
}

func MakePacket(mainId uint16, subId uint16, data []byte, ses *Session) *Packet {
	return &Packet{
		MainId: mainId,
		SubId:  subId,
		Data:   data,
		Ses:    ses,
	}
}

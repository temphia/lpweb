package wire

import (
	"fmt"

	nanoid "github.com/matoous/go-nanoid/v2"
)

type PacketType = uint8

const (
	PTypeSendHeader PacketType = iota
	PtypeSendBody   PacketType = iota
	PtypeEndBody    PacketType = iota
	PtypeReSendBody PacketType = iota
)

type Packet struct {
	PType  PacketType
	Offset int32 // current offset
	Total  int32 // total body size
	Data   []byte
}

func (p *Packet) String() string {
	return fmt.Sprintf("Packet{\n\tPType: %d,\n Offset: %d,\n Total: %d,\n Data: SIZE<%d>}", p.PType, p.Offset, p.Total, len(p.Data))
}

func GetRequestId() []byte {
	id, err := nanoid.New(16)
	if err != nil {
		panic(err)
	}

	if len(id) != 16 {
		panic("id is not 16 bytes")
	}

	return []byte(id)
}

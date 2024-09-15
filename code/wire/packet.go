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

var (
	PtypeMap = map[PacketType]string{
		PTypeSendHeader: "SendHeader",
		PtypeSendBody:   "SendBody",
		PtypeEndBody:    "EndBody",
		PtypeReSendBody: "ReSendBody",
	}
)

type Packet struct {
	PType  PacketType
	Offset int32 // current offset
	Total  int32 // total body size
	Data   []byte
}

const FragmentSize = 1024 * 256

func (p *Packet) String() string {
	ptype := PtypeMap[p.PType]

	return fmt.Sprintf("Packet{\n\tPType: %s,\n\t Offset: %d,\n\t Total: %d,\n\t Data: SIZE<%d>}", ptype, p.Offset, p.Total, len(p.Data))
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

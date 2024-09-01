package wire

import (
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

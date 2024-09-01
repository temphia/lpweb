package wire

type PacketType = uint8

const (
	PTypeSendHeader PacketType = iota
	PtypeSendBody   PacketType = iota + 1
	PtypeEndBody    PacketType = iota + 2
	PtypeReSendBody PacketType = iota + 3
)

type Packet struct {
	PType  PacketType
	Offset uint32 // current offset
	Total  uint32 // total body size
	Length uint32 // current data length
	Data   []byte
}

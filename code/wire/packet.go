package wire

type PacketType = uint8

const (
	PTypeSendHeader PacketType = iota
	PtypeSendBody   PacketType = iota
	PtypeEndBody    PacketType = iota
	PtypeReSendBody PacketType = iota
)

type Packet struct {
	PType  PacketType
	Offset uint32 // current offset
	Total  uint32 // total body size
	Length uint32 // current data length
	Data   []byte
}

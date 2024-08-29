package streamer

import (
	"context"
	"encoding/binary"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// type doneOffset struct {
// 	from uint32
// 	to   uint32
// }

type Packet struct {
	PType  uint8 // 1:start_send, 2:send, 3:end_send, 4:tally, 5:resend
	Offset uint32
	Length uint32
	Total  uint32
	Data   []byte
}

type Streamer struct {
	Context   context.Context
	RequestId []byte
	LocalNode host.Host

	OutData      []byte
	ActiveStream network.Stream
	TargetPeer   peer.ID

	// DonePacketFrags []doneOffset

	//	DoneInDataChan chan uint32
	InData []byte
	//  CloseChan chan struct{}
}

func (rc *Streamer) SendData() error {
	return writePacket(rc.ActiveStream, &Packet{
		Length: uint32(len(rc.OutData)),
		Data:   rc.OutData,
	})

}

func (rc *Streamer) ReceiveData() error {

	packet, err := readPacket(rc.ActiveStream)
	if err != nil {
		return err
	}

	rc.OutData = packet.Data

	return nil

}

func (rc *Streamer) AddStream(current *Packet, stream network.Stream) {

	panic("implement me")
}

func writePacket(stream network.Stream, packet *Packet) error {

	// write packet type
	stream.Write([]byte{packet.PType})

	// length, offset, total
	intBytes := make([]byte, 4)

	// write length
	binary.BigEndian.PutUint32(intBytes, packet.Length)
	stream.Write(intBytes)

	// write offset
	binary.BigEndian.PutUint32(intBytes, packet.Offset)
	stream.Write(intBytes)

	// write total
	binary.BigEndian.PutUint32(intBytes, packet.Total)
	stream.Write(intBytes)

	// final data
	_, err := stream.Write(packet.Data)

	return err
}

func readPacket(stream network.Stream) (*Packet, error) {
	packet := &Packet{}
	intBytes := make([]byte, 4)

	// read packet type
	_, err := stream.Read(intBytes[:1])
	if err != nil {
		return nil, err
	}

	ptype := uint8(intBytes[0])

	// read length
	_, err = stream.Read(intBytes)
	if err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(intBytes)

	// read offset

	_, err = stream.Read(intBytes)
	if err != nil {
		return nil, err
	}
	offset := binary.BigEndian.Uint32(intBytes)

	// read total
	_, err = stream.Read(intBytes)
	if err != nil {
		return nil, err
	}
	total := binary.BigEndian.Uint32(intBytes)

	// read data

	dataBytes := make([]byte, length)
	_, err = stream.Read(dataBytes)
	if err != nil {
		return nil, err
	}

	packet.PType = ptype
	packet.Length = length
	packet.Offset = offset
	packet.Total = total
	packet.Data = dataBytes

	return packet, nil

}

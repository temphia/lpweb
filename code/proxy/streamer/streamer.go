package streamer

import (
	"context"
	"encoding/binary"
	"log/slog"

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

	slog.Info("readPacket/1")

	packet := &Packet{}
	intBytes := make([]byte, 4)

	// read packet type
	_, err := stream.Read(intBytes[:1])
	if err != nil {
		slog.Info("readPacket/2")

		return nil, err
	}

	slog.Info("readPacket/3")

	ptype := uint8(intBytes[0])

	// read length
	_, err = stream.Read(intBytes)
	if err != nil {
		slog.Info("readPacket/4")
		return nil, err
	}

	length := binary.BigEndian.Uint32(intBytes)

	// read offset

	slog.Info("readPacket/5")

	_, err = stream.Read(intBytes)
	if err != nil {
		slog.Info("readPacket/6")
		return nil, err
	}
	offset := binary.BigEndian.Uint32(intBytes)

	slog.Info("readPacket/7")

	// read total
	_, err = stream.Read(intBytes)
	if err != nil {
		slog.Info("readPacket/8")
		return nil, err
	}
	total := binary.BigEndian.Uint32(intBytes)

	slog.Info("readPacket/9")

	// read data

	dataBytes := make([]byte, length)
	_, err = stream.Read(dataBytes)
	if err != nil {
		slog.Info("readPacket/10")
		return nil, err
	}

	slog.Info("readPacket/11")

	packet.PType = ptype
	packet.Length = length
	packet.Offset = offset
	packet.Total = total
	packet.Data = dataBytes

	return packet, nil

}

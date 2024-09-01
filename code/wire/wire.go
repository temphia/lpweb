package wire

import (
	"encoding/binary"
	"log/slog"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/network"
)

func WritePacket(stream network.Stream, packet *Packet) error {

	slog.Info("writePacket/1")

	// write packet type
	_, err := stream.Write([]byte{packet.PType})
	if err != nil {
		slog.Info("writePacket/2")
		return err
	}

	// length, offset, total
	intBytes := make([]byte, 4)

	// write length
	binary.BigEndian.PutUint32(intBytes, packet.Length)
	_, err = stream.Write(intBytes)
	if err != nil {
		slog.Info("writePacket/3")
		return err
	}

	// write offset
	binary.BigEndian.PutUint32(intBytes, packet.Offset)
	_, err = stream.Write(intBytes)
	if err != nil {
		slog.Info("writePacket/4")
		return err
	}

	// write total
	binary.BigEndian.PutUint32(intBytes, packet.Total)
	_, err = stream.Write(intBytes)
	if err != nil {
		slog.Info("writePacket/5")
		return err
	}

	// final data
	written, err := stream.Write(packet.Data)

	if err != nil {
		slog.Info("writePacket/6")
		return err
	}

	pp.Println("writePacket/7 total/written", len(packet.Data), written)

	return err
}

func ReadPacket(stream network.Stream) (*Packet, error) {

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
	readSize, err := stream.Read(dataBytes)
	if err != nil {
		slog.Info("readPacket/10")
		return nil, err
	}

	pp.Println("@read_data", length, readSize)

	packet.PType = ptype
	packet.Length = length
	packet.Offset = offset
	packet.Total = total
	packet.Data = dataBytes

	slog.Info("readPacket/11")

	pp.Println("@read_data", packet)

	return packet, nil

}

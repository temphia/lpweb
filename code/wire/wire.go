package wire

import (
	"encoding/binary"
	"log"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/network"
)

func WritePacket(stream network.Stream, packet *Packet) error {

	log.Println("writePacket/1")

	// write packet type
	_, err := stream.Write([]byte{packet.PType})
	if err != nil {
		log.Println("writePacket/2")
		return err
	}

	// length, offset, total
	intBytes := make([]byte, 4)

	// write length
	binary.BigEndian.PutUint32(intBytes, uint32(len(packet.Data)))
	_, err = stream.Write(intBytes)
	if err != nil {
		log.Println("writePacket/3")
		return err
	}

	// write offset
	binary.BigEndian.PutUint32(intBytes, uint32(packet.Offset))
	_, err = stream.Write(intBytes)
	if err != nil {
		log.Println("writePacket/4")
		return err
	}

	// write total
	binary.BigEndian.PutUint32(intBytes, uint32(packet.Total))
	_, err = stream.Write(intBytes)
	if err != nil {
		log.Println("writePacket/5")
		return err
	}

	// final data
	written, err := stream.Write(packet.Data)

	if err != nil {
		log.Println("writePacket/6")
		return err
	}

	pp.Println("PACKET", packet)

	pp.Println("writePacket/7 total/written", len(packet.Data), written)

	return err
}

func ReadPacket(stream network.Stream) (*Packet, error) {

	log.Println("readPacket/1")

	packet := &Packet{}
	intBytes := make([]byte, 4)

	// read packet type
	_, err := stream.Read(intBytes[:1])
	if err != nil {
		log.Println("readPacket/2")

		return nil, err
	}

	log.Println("readPacket/3")

	ptype := uint8(intBytes[0])

	// read length
	_, err = stream.Read(intBytes)
	if err != nil {
		log.Println("readPacket/4")
		return nil, err
	}

	length := binary.BigEndian.Uint32(intBytes)

	// read offset

	log.Println("readPacket/5")

	_, err = stream.Read(intBytes)
	if err != nil {
		log.Println("readPacket/6")
		return nil, err
	}
	offset := binary.BigEndian.Uint32(intBytes)

	log.Println("readPacket/7")

	// read total
	_, err = stream.Read(intBytes)
	if err != nil {
		log.Println("readPacket/8")
		return nil, err
	}
	total := binary.BigEndian.Uint32(intBytes)

	log.Println("readPacket/9")

	// read data

	dataBytes := make([]byte, length)
	readSize, err := stream.Read(dataBytes)
	if err != nil {
		log.Println("readPacket/10")
		return nil, err
	}

	pp.Println("@read_data", length, readSize)

	packet.PType = ptype
	packet.Offset = int32(offset)
	packet.Total = int32(total)
	packet.Data = dataBytes

	log.Println("readPacket/11")

	pp.Println("@read_data", packet)

	return packet, nil

}

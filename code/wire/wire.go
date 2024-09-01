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

	log.Println("writePacket/ptype/ok")

	// length, offset, total
	intBytes := make([]byte, 4)

	// write length
	binary.BigEndian.PutUint32(intBytes, uint32(len(packet.Data)))
	_, err = stream.Write(intBytes)
	if err != nil {
		log.Println("writePacket/3")
		return err
	}

	log.Println("writePacket/len/ok")

	// write offset
	binary.BigEndian.PutUint32(intBytes, uint32(packet.Offset))
	_, err = stream.Write(intBytes)
	if err != nil {
		log.Println("writePacket/4")
		return err
	}

	log.Println("writePacket/offset/ok")

	// write total
	binary.BigEndian.PutUint32(intBytes, uint32(packet.Total))
	_, err = stream.Write(intBytes)
	if err != nil {
		log.Println("writePacket/5")
		return err
	}

	log.Println("writePacket/total/ok")

	totalWritten := 0

	for {
		// final data
		pp.Println("@W>>")
		written, err := stream.Write(packet.Data[totalWritten:])
		pp.Println("<<W@")

		if err != nil {
			log.Println("writePacket/6")
			return err
		}
		totalWritten += written
		if totalWritten >= int(packet.Total) {
			break
		}

	}

	pp.Println("PACKET", packet)

	pp.Println("writePacket/7 total/written", len(packet.Data), totalWritten)

	if len(packet.Data) > 10 {
		pp.Println("TAIL_DATA", string(packet.Data[len(packet.Data)-10:]))
	}

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

	log.Println("readPacket/4/ptype", int64(ptype))

	// read length
	_, err = stream.Read(intBytes)
	if err != nil {
		log.Println("readPacket/4")
		return nil, err
	}

	length := binary.BigEndian.Uint32(intBytes)

	// read offset

	log.Println("readPacket/5/length", int64(length))

	_, err = stream.Read(intBytes)
	if err != nil {
		log.Println("readPacket/6")
		return nil, err
	}
	offset := binary.BigEndian.Uint32(intBytes)

	log.Println("readPacket/7/offset", int64(offset))

	// read total
	_, err = stream.Read(intBytes)
	if err != nil {
		log.Println("readPacket/8")
		return nil, err
	}
	total := binary.BigEndian.Uint32(intBytes)

	log.Println("readPacket/9/total", int64(total))

	// read data

	dataBytes := make([]byte, length)
	totalRead := 0

	for {

		pp.Println("@R>>")
		readSize, err := stream.Read(dataBytes[totalRead:])
		pp.Println("<<R@")

		if err != nil {
			log.Println("readPacket/10")
			return nil, err
		}
		totalRead += readSize

		if totalRead >= int(length) {
			break
		}

	}

	pp.Println("@read_data", length, totalRead)

	packet.PType = ptype
	packet.Offset = int32(offset)
	packet.Total = int32(total)
	packet.Data = dataBytes

	pp.Println("readPacket/11", packet)

	if len(packet.Data) > 10 {
		pp.Println("TAIL_DATA", string(packet.Data[len(packet.Data)-10:]))
	}

	return packet, nil

}

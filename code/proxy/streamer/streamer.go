package streamer

import (
	"context"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/temphia/lpweb/code/wire"
)

type Streamer struct {
	Context   context.Context
	RequestId []byte
	LocalNode host.Host

	OutData      []byte
	ActiveStream network.Stream
	TargetAddr   peer.AddrInfo

	// DonePacketFrags []doneOffset

	//	DoneInDataChan chan uint32
	InData []byte
	//  CloseChan chan struct{}
}

func (rc *Streamer) SendData() error {

	return wire.WritePacket(rc.ActiveStream, &wire.Packet{
		Offset: 0,
		Total:  int32(len(rc.OutData)),
		PType:  uint8(1),
		Data:   rc.OutData,
	})

}

func (rc *Streamer) ReceiveData() error {

	packet, err := wire.ReadPacket(rc.ActiveStream)
	if err != nil {
		return err
	}

	rc.InData = packet.Data

	return nil

}

func (rc *Streamer) AddStream(current *wire.Packet, stream network.Stream) {

	panic("implement me")
}

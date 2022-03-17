package p2p

import (
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/temphia/temphia_relay/core"
)

const Protocol = "/temphia_p2p/1.0.0"

type Instance struct {
	localNode   host.Host
	dhtOut      *dht.IpfsDHT
	exitServers map[string]*ExitServer
}

func NewInstance() *Instance {

	h, dth, err := core.NewHost("drwytfvhjq")
	if err != nil {
		panic(err)
	}

	h.SetStreamHandler(Protocol, nil)

	return &Instance{
		localNode:   h,
		dhtOut:      dth,
		exitServers: map[string]*ExitServer{},
	}
}

func (i *Instance) streamHandler(stream network.Stream) {
	defer stream.Close()

}

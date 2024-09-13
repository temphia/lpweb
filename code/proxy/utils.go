package proxy

import (
	"context"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func deny(s network.Stream) {
	pp.Println("proxy won't accept lpweb request, it proxies web reqest to libweb")
	s.Close()
}

func (wp *WebProxy) nodeAddrs(target string) (*peer.AddrInfo, error) {

	id, err := peer.IDFromBytes([]byte(target))
	if err != nil {
		return nil, err
	}

	pp.Println(id.Validate())

	addr, err := wp.Mesh.DHT.FindPeer(context.TODO(), id)
	return &addr, err
}

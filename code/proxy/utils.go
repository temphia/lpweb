package proxy

import (
	"context"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

const proxyKey = "drwytfvhjqui3bdoi32kd832houj3p2i3pj3nkl821oeubj01[eoj;kn.o2penkl1ohpekln2,m12hep1nkl,m1ephil"

func deny(s network.Stream) {
	pp.Println("proxy won't accept lpweb request, it proxies web reqest to libweb")
	s.Close()
}

func (wp *WebProxy) nodeAddrs(target string) (*peer.AddrInfo, error) {

	id, err := peer.IDFromString(target)
	if err != nil {
		return nil, err
	}

	pp.Println(id.Validate())

	addr, err := wp.dhtOut.FindPeer(context.TODO(), id)
	return &addr, err
}

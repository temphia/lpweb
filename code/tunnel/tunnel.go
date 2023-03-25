package tunnel

import (
	"log"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/temphia/lpweb/code/core"
	"github.com/temphia/lpweb/code/seekers"
	"github.com/temphia/lpweb/code/seekers/etcd"
)

type HttpTunnel struct {
	localNode host.Host
	seekers   []seekers.Seeker
}

func NewHttpTunnel(port int) *HttpTunnel {

	h, _, err := core.NewHost(tunKey, 0)
	if err != nil {
		panic(err)
	}

	h.SetStreamHandler(core.Protocol, func(s network.Stream) {
		pp.Println("@new request")
	})

	log.Println("p2p_relay@", h.ID())
	for _, m := range h.Addrs() {
		log.Println("httpd@", m.String())
	}

	seeker := etcd.New()

	instance := &HttpTunnel{
		localNode: h,

		seekers: []seekers.Seeker{
			seeker,
		},
	}

	return instance
}

func (ht *HttpTunnel) Run() error {

	paddr := peer.AddrInfo{
		ID:    ht.localNode.ID(),
		Addrs: ht.localNode.Addrs(),
	}

	out, err := paddr.MarshalJSON()
	if err != nil {
		return err
	}

	for _, s := range ht.seekers {
		s.Set(paddr.ID.String(), string(out))
	}

	ch := make(chan bool)
	ch <- false

	return nil

}

package tunnel

import (
	"log"
	"strings"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/temphia/lpweb/code/core/mesh"
	"github.com/temphia/lpweb/code/core/seekers"
	"github.com/temphia/lpweb/code/core/seekers/etcd"
)

type HttpTunnel struct {
	mesh         *mesh.Mesh
	tunnelToPort int
	localNode    host.Host
	seekers      []seekers.Seeker
}

func NewHttpTunnel(port int) *HttpTunnel {

	m, err := mesh.New(tunKey, 0)
	if err != nil {
		panic(err)
	}

	log.Println("p2p_relay@", m.Host.ID().String())
	for _, m := range m.Host.Addrs() {
		log.Println("httpd@", m.String())
	}

	seeker := etcd.New()

	instance := &HttpTunnel{
		mesh:         m,
		localNode:    m.Host,
		tunnelToPort: port,
		seekers:      []seekers.Seeker{seeker},
	}

	m.Host.SetStreamHandler(mesh.Protocol, instance.streamHandler)

	return instance
}

func (ht *HttpTunnel) Run() error {

	addrs := ht.localNode.Addrs()

	maddr, err := ht.mesh.PublicMultiAddr()
	if err == nil {
		addrs = append(addrs, maddr)
	}

	paddr := peer.AddrInfo{
		ID:    ht.localNode.ID(),
		Addrs: addrs,
	}

	out, err := paddr.MarshalJSON()
	if err != nil {
		return err
	}

	for _, s := range ht.seekers {
		s.Set(strings.ToLower(paddr.ID.String()), string(out))
	}

	ch := make(chan bool)
	ch <- false

	return nil

}

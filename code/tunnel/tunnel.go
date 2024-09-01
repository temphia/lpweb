package tunnel

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/temphia/lpweb/code/core/config"
	"github.com/temphia/lpweb/code/core/mesh"
	"github.com/temphia/lpweb/code/core/seekers"
	"github.com/temphia/lpweb/code/core/seekers/etcd"
	"github.com/temphia/lpweb/code/proxy/streamer"
)

type HttpTunnel struct {
	mesh         *mesh.Mesh
	tunnelToPort int
	localNode    host.Host
	seekers      []seekers.Seeker

	activeStramers map[string]*streamer.Streamer
	rcLock         sync.Mutex
}

func NewHttpTunnel(port int) *HttpTunnel {
	conf := config.Get()

	m, err := mesh.New(conf.TunnelKey, 0)
	if err != nil {
		panic(err)
	}

	log.Println("p2p_relay@", m.Host.ID().String())
	for _, m := range m.Host.Addrs() {
		log.Println("httpd@", m.String())
	}

	seeker := etcd.New(conf.UUID)

	instance := &HttpTunnel{
		mesh:           m,
		localNode:      m.Host,
		tunnelToPort:   port,
		seekers:        []seekers.Seeker{seeker},
		activeStramers: make(map[string]*streamer.Streamer),
		rcLock:         sync.Mutex{},
	}

	m.Host.SetStreamHandler(mesh.ProtocolHttp, instance.streamHandleHttp)
	m.Host.SetStreamHandler(mesh.ProtocolHttp3, instance.streamHandleHttp3)
	m.Host.SetStreamHandler(mesh.ProtocolWS, instance.streamHandleWS)

	servHost := fmt.Sprintf("http://%s.lpweb", strings.ToLower(m.Host.ID().String()))
	pp.Println("@serving_in_libp2p", servHost)

	return instance
}

func (ht *HttpTunnel) Run() error {

	addrs := ht.localNode.Addrs()

	// maddr, err := ht.mesh.PublicMultiAddr()
	// if err == nil {
	// 	addrs = append(addrs, maddr...)
	// }

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

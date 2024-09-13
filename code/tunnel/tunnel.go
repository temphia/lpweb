package tunnel

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/temphia/lpweb/code/core/config"
	"github.com/temphia/lpweb/code/core/mesh"
	"github.com/temphia/lpweb/code/proxy/streamer"
)

type HttpTunnel struct {
	Mesh         *mesh.Mesh
	tunnelToPort int
	localNode    host.Host

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
		log.Println("httpd@tunnel", m.String())
	}

	instance := &HttpTunnel{
		Mesh:           m,
		localNode:      m.Host,
		tunnelToPort:   port,
		activeStramers: make(map[string]*streamer.Streamer),
		rcLock:         sync.Mutex{},
	}

	m.Host.SetStreamHandler(mesh.ProtocolHttp3, instance.streamHandleHttp3)

	servHost := fmt.Sprintf("http://%s.lpweb", strings.ToLower(m.Host.ID().String()))
	pp.Println("@serving_in_libp2p", servHost)

	return instance
}

func (ht *HttpTunnel) Run() error {

	return nil

}

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

func New(port int) *HttpTunnel {
	conf := config.Get()

	m, err := mesh.New(conf.TunnelKey, 0)
	if err != nil {
		panic(err)
	}

	return NewUsingMesh(port, m)
}

func NewUsingMesh(port int, m *mesh.Mesh) *HttpTunnel {
	instance := &HttpTunnel{
		Mesh:           m,
		localNode:      m.Host,
		tunnelToPort:   port,
		activeStramers: make(map[string]*streamer.Streamer),
		rcLock:         sync.Mutex{},
	}

	log.Println("p2p_relay@", m.Host.ID().String())
	for _, m := range m.Host.Addrs() {
		log.Println("httpd@tunnel", m.String())
	}

	m.Host.SetStreamHandler(mesh.ProtocolHttp, instance.streamHandleHttp)
	m.Host.SetStreamHandler(mesh.ProtocolWS, instance.streamHandleWS)

	servHost := fmt.Sprintf("http://%s.lpweb", strings.ToLower(m.Host.ID().String()))
	pp.Println("@serving_in_libp2p", servHost)

	return instance
}

func (ht *HttpTunnel) Run() error {

	return nil
}

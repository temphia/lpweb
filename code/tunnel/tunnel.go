package tunnel

import (
	"sync"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/temphia/lpweb/code/core/config"
	"github.com/temphia/lpweb/code/core/mesh"
	"github.com/temphia/lpweb/code/proxy/streamer"
)

type HttpTunnel struct {
	Mesh          *mesh.Mesh
	tunnelToPort  int
	localNode     host.Host
	tunnelAnyPort bool

	activeStramers map[string]*streamer.Streamer
	rcLock         sync.Mutex
}

func New(port int, anyPort bool) *HttpTunnel {
	conf := config.Get()

	m, err := mesh.New(conf.TunnelKey, 0)
	if err != nil {
		panic(err)
	}

	return NewUsingMesh(port, m, anyPort)
}

func NewUsingMesh(port int, m *mesh.Mesh, anyPort bool) *HttpTunnel {
	instance := &HttpTunnel{
		Mesh:           m,
		localNode:      m.Host,
		tunnelToPort:   port,
		activeStramers: make(map[string]*streamer.Streamer),
		rcLock:         sync.Mutex{},
	}

	m.Host.SetStreamHandler(mesh.ProtocolHttp, instance.streamHandleHttp)
	m.Host.SetStreamHandler(mesh.ProtocolWS, instance.streamHandleWS)

	return instance
}

func (ht *HttpTunnel) Run() error {

	return nil
}

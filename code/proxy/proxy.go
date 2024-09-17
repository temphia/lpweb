package proxy

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sync"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/host"

	"github.com/temphia/lpweb/code/core/config"
	"github.com/temphia/lpweb/code/core/mesh"
	"github.com/temphia/lpweb/code/proxy/streamer"
)

type WebProxy struct {
	Mesh      *mesh.Mesh
	localNode host.Host

	upnodeLock sync.Mutex

	proxyPort int

	requests map[uint32]*streamer.Streamer
	reqMLock sync.Mutex
}

func New(port int) *WebProxy {

	conf := config.Get()

	m, err := mesh.New(conf.ProxyKey, 0)
	if err != nil {
		panic(err)
	}

	return NewUsingMesh(port, m)
}

func NewUsingMesh(port int, m *mesh.Mesh) *WebProxy {
	instance := &WebProxy{
		Mesh:       m,
		localNode:  m.Host,
		proxyPort:  0,
		requests:   make(map[uint32]*streamer.Streamer),
		reqMLock:   sync.Mutex{},
		upnodeLock: sync.Mutex{},
	}

	return instance
}

func (wp *WebProxy) Run() error {

	addr := fmt.Sprintf(":%d", wp.proxyPort)

	log.Println("listening proxy ", addr)
	pp.Println("listening proxy ", addr)
	return http.ListenAndServe(addr, wp)
}

var hostRegex = regexp.MustCompile(`^([a-zA-Z0-9-]+)\.localhost`)

func (wp *WebProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pp.Println("@ALL_INTERCEPT", r.Method)

	if hostRegex.MatchString(r.Host) {
		pp.Println("@IPWEB_INTERCEPT", r.Method)

		isWs := r.Method == "GET" && r.Header.Get("Upgrade") == "websocket"
		if r.Method == "CONNECT" || isWs {
			wp.handleWS(r, w)
		} else {
			wp.HandleHttp3(r, w)
		}

		return
	}

}

package proxy

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"sync"

	"github.com/elazarl/goproxy"
	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"

	"github.com/temphia/lpweb/code/core/config"
	"github.com/temphia/lpweb/code/core/mesh"
	"github.com/temphia/lpweb/code/proxy/streamer"
)

type WebProxy struct {
	Mesh      *mesh.Mesh
	localNode host.Host
	proxy     *goproxy.ProxyHttpServer

	upnodeLock sync.Mutex

	proxyPort int

	requests map[uint32]*streamer.Streamer
	reqMLock sync.Mutex
}

func NewWebProxy(port int) *WebProxy {

	conf := config.Get()

	proxy := goproxy.NewProxyHttpServer()

	m, err := mesh.New(conf.ProxyKey, 0)
	if err != nil {
		panic(err)
	}

	m.Host.SetStreamHandler(mesh.ProtocolWS, deny)

	log.Println("p2p_relay@", m.Host.ID().String())
	for _, m := range m.Host.Addrs() {
		log.Println("httpd@proxy", m.String())
	}

	instance := &WebProxy{
		Mesh:      m,
		localNode: m.Host,
		proxy:     proxy,
		proxyPort: port,
		//upNodes:    make(map[string]*UpNode),
		requests:   make(map[uint32]*streamer.Streamer),
		reqMLock:   sync.Mutex{},
		upnodeLock: sync.Mutex{},
	}

	m.Host.SetStreamHandler(mesh.ProtocolHttp3, func(s network.Stream) {
		panic("Not implemented")
	})

	return instance
}

func (wp *WebProxy) Run() error {

	wp.proxy.Verbose = true

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

	wp.proxy.ServeHTTP(w, r)
}

func deny(s network.Stream) {
	pp.Println("proxy won't accept lpweb request, it proxies web reqest to libweb")
	s.Close()
}

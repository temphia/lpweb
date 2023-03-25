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

	"github.com/temphia/lpweb/code/core/config"
	"github.com/temphia/lpweb/code/core/mesh"
	"github.com/temphia/lpweb/code/core/seekers"
	"github.com/temphia/lpweb/code/core/seekers/etcd"
)

var r = regexp.MustCompile(`\.lpweb`)

type WebProxy struct {
	mesh      *mesh.Mesh
	localNode host.Host
	proxy     *goproxy.ProxyHttpServer
	seekers   []seekers.Seeker

	upNodes    map[string]*UpNode
	upnodeLock sync.Mutex

	proxyPort int
}

func NewWebProxy(port int) *WebProxy {

	conf := config.Get()

	proxy := goproxy.NewProxyHttpServer()

	m, err := mesh.New(conf.ProxyKey, 0)
	if err != nil {
		panic(err)
	}

	m.Host.SetStreamHandler(mesh.ProtocolHttp, deny)
	m.Host.SetStreamHandler(mesh.ProtocolWS, deny)

	log.Println("p2p_relay@", m.Host.ID().String())
	for _, m := range m.Host.Addrs() {
		log.Println("httpd@", m.String())
	}

	seeker := etcd.New(conf.UUID)

	instance := &WebProxy{
		mesh:       m,
		localNode:  m.Host,
		proxy:      proxy,
		proxyPort:  port,
		upNodes:    make(map[string]*UpNode),
		upnodeLock: sync.Mutex{},

		seekers: []seekers.Seeker{
			seeker,
		},
	}

	return instance
}

func (wp *WebProxy) Run() error {

	wp.proxy.OnRequest(goproxy.ReqHostMatches(r)).DoFunc(wp.handle)

	wp.proxy.Verbose = true

	addr := fmt.Sprintf(":%d", wp.proxyPort)

	log.Println("listening proxy ", addr)
	pp.Println("listening proxy ", addr)
	return http.ListenAndServe(addr, wp.proxy)
}

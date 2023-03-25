package proxy

import (
	"log"
	"net/http"
	"regexp"
	"sync"

	"github.com/elazarl/goproxy"
	"github.com/libp2p/go-libp2p-core/host"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/temphia/lpweb/code/core"
	"github.com/temphia/lpweb/code/seekers"
	"github.com/temphia/lpweb/code/seekers/etcd"
)

var r = regexp.MustCompile(`\.lpweb`)

type WebProxy struct {
	localNode host.Host
	dhtOut    *dht.IpfsDHT
	proxy     *goproxy.ProxyHttpServer
	seekers   []seekers.Seeker

	upNodes    map[string]*UpNode
	upnodeLock sync.Mutex
}

func NewWebProxy(port int) *WebProxy {

	proxy := goproxy.NewProxyHttpServer()

	h, dth, err := core.NewHost(proxyKey, 0)
	if err != nil {
		panic(err)
	}

	h.SetStreamHandler(core.Protocol, deny)

	log.Println("p2p_relay@", h.ID())
	for _, m := range h.Addrs() {
		log.Println("httpd@", m.String())
	}

	seeker := etcd.New()

	instance := &WebProxy{
		localNode:  h,
		dhtOut:     dth,
		proxy:      proxy,
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

	go wp.listenTLS()

	log.Println("listening proxy")
	return http.ListenAndServe(":8080", wp.proxy)
}

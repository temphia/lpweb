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
	"github.com/temphia/lpweb/code/core/seekers"
	"github.com/temphia/lpweb/code/core/seekers/etcd"
	"github.com/temphia/lpweb/code/proxy/streamer"
	"github.com/temphia/lpweb/code/wire"

	"github.com/fxamacker/cbor"
)

type WebProxy struct {
	mesh      *mesh.Mesh
	localNode host.Host
	proxy     *goproxy.ProxyHttpServer
	seekers   []seekers.Seeker

	upNodes    map[string]*UpNode
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
		requests:   make(map[uint32]*streamer.Streamer),
		reqMLock:   sync.Mutex{},
		upnodeLock: sync.Mutex{},

		seekers: []seekers.Seeker{
			seeker,
		},
	}

	m.Host.SetStreamHandler(mesh.ProtocolHttp2, func(s network.Stream) {

		m := cbor.NewDecoder(s)
		packet := wire.Packet{}

		err := m.Decode(&packet)

		if err != nil {
			pp.Println("@err_unmarshal", err.Error())
			s.Close()
			return
		}

		instance.reqMLock.Lock()
		req := instance.requests[packet.HttpRequestId]
		instance.reqMLock.Unlock()

		if req == nil {
			pp.Println("@err_no_req_found", packet.HttpRequestId)
			s.Close()
			return
		}

		pp.Println("@NOT_IMPLEMENTED")

		// req.OutsidePacket <- rcycle.SideChannelPacket{
		// 	Packet:     &packet,
		// 	FromStream: s,
		// }

		// err = req.StreamReadLoop(s)
		// if err != nil {
		// 	pp.Println("@err_stream_read", err.Error())
		// 	s.Close()
		// }

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

var hostRegex = regexp.MustCompile(`[A-Za-z0-9]*\.*[A-Za-z0-9]*\.lpweb`)

func (wp *WebProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pp.Println("@ALL_INTERCEPT", r.Method)

	if hostRegex.MatchString(r.Host) {
		pp.Println("@IPWEB_INTERCEPT", r.Method)

		if r.Method == "CONNECT" {
			wp.handleWS(r, w)
		} else {
			wp.handleHttp2(r, w)
		}

		return
	}

	wp.proxy.ServeHTTP(w, r)
}

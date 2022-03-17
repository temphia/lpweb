package p2p

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/temphia/temphia_relay/core"
)

type Instance struct {
	localNode   host.Host
	dhtOut      *dht.IpfsDHT
	exitServers map[string]*ExitServer
	mlock       sync.Mutex
}

func NewInstance() *Instance {

	log.Println("creating instance...")

	h, dth, err := core.NewHost("drwytfvhjqui3bdoi32kd832houj3p2i3pj3nkl821oeubj01[eoj;kn.o2penkl1ohpekln2,m12hep1nkl,m1ephil", 8084)
	if err != nil {
		panic(err)
	}

	h.SetStreamHandler(core.Protocol, func(s network.Stream) {
		pp.Println("I should not be getting this")
	})

	log.Println("p2p_relay@", h.ID())
	for _, m := range h.Addrs() {
		log.Println("httpd@", m.String())
	}

	a, err := peer.AddrInfoFromString("/ip4/127.0.0.1/tcp/8083/p2p/12D3KooWC8Vr1LMXape8KDVWikJ5YTehorc1BTSeN1htwvUgEsUC")
	if err != nil {
		panic(err)
	}

	err = h.Connect(context.TODO(), *a)
	if err != nil {
		log.Println(err)
	}

	instance := &Instance{
		localNode:   h,
		dhtOut:      dth,
		exitServers: map[string]*ExitServer{},
	}

	return instance
}

var r = regexp.MustCompile(`\.temphiap2p`)

func (i *Instance) Run() {
	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest(goproxy.ReqHostMatches(r)).DoFunc(i.proxied)
	proxy.Verbose = true

	log.Println("running p2p instance...")

	log.Fatal(http.ListenAndServe(":8080", proxy))

}

func (i *Instance) proxied(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	pp.Println("@@@@", r.Host)

	host := strings.Split(r.Host, ".")[0]

	enode := i.getExitNode(host)

	stream, err := i.localNode.NewStream(context.TODO(), enode.addr.ID, core.Protocol)
	if err != nil {
		panic(err)
	}

	out, err := httputil.DumpRequest(r, true)
	if err != nil {
		panic(err)
	}

	io.Copy(stream, bytes.NewBuffer(out))

	resp, err := http.ReadResponse(bufio.NewReader(stream), r)
	if err != nil {
		panic(err)
	}

	return nil, resp
}

func (i *Instance) getExitNode(target string) *ExitServer {
	i.mlock.Lock()
	defer i.mlock.Unlock()

	enode, ok := i.exitServers[target]
	if ok {
		return enode
	}

	addr, err := i.ensureConn(target)
	if err != nil {
		panic(err)
	}

	enode = &ExitServer{
		localNode:  i.localNode,
		instance:   i,
		p2pPubId:   target,
		addr:       addr,
		sessionKey: "",
		mLock:      sync.Mutex{},
	}

	i.exitServers[target] = enode
	return enode
}

func (i *Instance) ensureConn(target string) (*peer.AddrInfo, error) {
	a, err := peer.AddrInfoFromString("/ip4/127.0.0.1/tcp/8083/p2p/12D3KooWC8Vr1LMXape8KDVWikJ5YTehorc1BTSeN1htwvUgEsUC")
	return a, err

	//return i.nodeAddrs(target)
	// if err != nil {
	// 	return nil, err
	// }

	// err = i.localNode.Connect(context.TODO(), *nodeAddr)
	// if err != nil {
	// 	return nil, err
	// }

	// return nodeAddr, nil

}

// "QmUPjX5Zmmdb8R4hm1D8mB2QshTydkWgi9Uquip3PshsgQ" => "/ip4/127.0.0.1/tcp/38733/p2p/QmUPjX5Zmmdb8R4hm1D8mB2QshTydkWgi9Uquip3PshsgQ"

func (i *Instance) nodeAddrs(target string) (*peer.AddrInfo, error) {
	// fixme => right now use traget directly but use ipns to
	// to resolve all possible nodes

	pp.Println(target)

	id, err := peer.IDFromString(target)
	if err != nil {
		return nil, err
	}

	pp.Println(id.Validate())

	addr, err := i.dhtOut.FindPeer(context.TODO(), id)
	return &addr, err
}

func (i *Instance) debugLoop() {
	for {

		peers := i.localNode.Network().Peers()

		for _, peer := range peers {
			pp.Println(peer.Pretty())
		}

		time.Sleep(time.Second * 5)

		pp.Println("#################")
	}
}

package p2p

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/elazarl/goproxy"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/temphia/temphia_relay/core"
)

const Protocol = "/temphia_p2p/1.0.0"

type Instance struct {
	localNode   host.Host
	dhtOut      *dht.IpfsDHT
	exitServers map[string]*ExitServer
	mlock       sync.Mutex
}

func NewInstance() *Instance {

	h, dth, err := core.NewHost("drwytfvhjq")
	if err != nil {
		panic(err)
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
	log.Fatal(http.ListenAndServe(":8080", proxy))

}

func (i *Instance) proxied(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	host := strings.Split(r.Host, ".")[0]

	enode := i.getExitNode(host)

	fmt.Println(enode)

	return r, nil
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

	nodeAddr, err := i.nodeAddrs(target)
	if err != nil {
		return nil, err
	}

	err = i.localNode.Connect(context.TODO(), *nodeAddr)
	if err != nil {
		return nil, err
	}

	return nodeAddr, nil

}

// "QmUPjX5Zmmdb8R4hm1D8mB2QshTydkWgi9Uquip3PshsgQ" => "/ip4/127.0.0.1/tcp/38733/p2p/QmUPjX5Zmmdb8R4hm1D8mB2QshTydkWgi9Uquip3PshsgQ"

func (i *Instance) nodeAddrs(traget string) (*peer.AddrInfo, error) {
	// fixme => right now use traget directly but use ipns to
	// to resolve all possible nodes

	id, err := peer.Decode(traget)
	if err != nil {
		return nil, err
	}

	addr, err := i.dhtOut.FindPeer(context.TODO(), id)
	return &addr, err
}

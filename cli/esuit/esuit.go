package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/k0kubun/pp"
	"github.com/temphia/lpweb/code/core"
	"github.com/temphia/lpweb/code/core/mesh"
	"github.com/temphia/lpweb/code/proxy"
	"github.com/temphia/lpweb/code/tunnel"
)

type Esuit struct {
	tunnel *tunnel.HttpTunnel

	proxy *proxy.WebProxy
}

func main() {

	wproxy := proxy.NewWebProxy(0)

	tunnel := tunnel.NewHttpTunnel(0)

	suit := &Esuit{
		tunnel: tunnel,
		proxy:  wproxy,
	}

	go suit.StartHttpServer()

	go suit.StartFileServer()

	peerKey, err := suit.tunnel.Mesh.GetPeerKey().MarshalBinary()
	if err != nil {
		panic(err)
	}

	pp.Println("@SERVER_PEER_KEY", suit.tunnel.Mesh.GetPeerKey().String())
	pp.Println("@CLIENT_PEER_KEY", suit.proxy.Mesh.GetPeerKey().String())

	suit.tunnel.Mesh.SetAltPeers(suit.proxy.Mesh.GetSelfPeerAddr())
	suit.proxy.Mesh.SetAltPeers(suit.tunnel.Mesh.GetSelfPeerAddr())

	tunnel.Mesh.Host.SetStreamHandler(mesh.ProtocolHttp3, suit.Handler)

	pp.Println(string(core.EncodeToSafeString(peerKey)), "@")

	url, err := url.Parse(fmt.Sprintf("http://%s.localhost:8001/", string(core.EncodeToSafeString(peerKey))))
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 10)

	pp.Println("@connecting to", url.String())

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		panic(err)
	}

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	// wait here forever
	select {}

}

package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/k0kubun/pp"
	"github.com/temphia/lpweb/code/core"
	"github.com/temphia/lpweb/code/proxy"
	"github.com/temphia/lpweb/code/tunnel"
)

type Esuit struct {
	tunnel *tunnel.HttpTunnel

	proxy *proxy.WebProxy
}

func main() {

	wproxy := proxy.NewWebProxy(0)

	tunnel := tunnel.NewHttpTunnel(8002)

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

	entryHttpUrl := fmt.Sprintf("http://%s.localhost:8001/", string(core.EncodeToSafeString(peerKey)))

	pp.Println("@serving_in_libp2p", entryHttpUrl)

	url, err := url.Parse(entryHttpUrl)
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 10)

	pp.Println("@connecting to", url.String())

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		panic(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	pp.Println("@RESPONSE_BODY", string(out))

	// wait here forever
	select {}

}

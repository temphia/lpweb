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
)

type Esuit struct {
	sMesh *mesh.Mesh

	proxy *proxy.WebProxy
}

func main() {
	//	fmt.Println("Hello World")

	sMesh, err := mesh.New("1234567890451678scfagvhbjknlmagvshjbkntcaytvsuyghukfcgv", 0)
	if err != nil {
		panic(err)
	}

	wproxy := proxy.NewWebProxy(0)

	suit := &Esuit{
		sMesh: sMesh,
		proxy: wproxy,
	}

	go suit.StartHttpServer()

	go suit.StartFileServer()

	peerKey, err := suit.sMesh.GetPeerKey().MarshalBinary()
	if err != nil {
		panic(err)
	}

	pp.Println("@SERVER_PEER_KEY", suit.sMesh.GetPeerKey().String())
	pp.Println("@CLIENT_PEER_KEY", suit.proxy.GetPeerKey().String())

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

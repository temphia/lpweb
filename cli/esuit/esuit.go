package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"

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

	tunnel := tunnel.NewHttpTunnel(7703)

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

	entryHttpUrl := fmt.Sprintf("http://%s.localhost:7703/", string(core.EncodeToSafeString(peerKey)))

	pp.Println("@serving_in_libp2p", entryHttpUrl)

	err = tryNormalHttp(entryHttpUrl)
	if err != nil {
		panic(err.Error())
	}

	err = tryUpload(entryHttpUrl)
	if err != nil {
		panic(err.Error())
	}

	// wait here forever
	select {}

}

func tryNormalHttp(baseURL string) error {

	url, err := url.Parse(fmt.Sprintf("%s/list", baseURL))
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	pp.Println("@RESPONSE_BODY", string(out))

	return nil

}

func tryUpload(baseURL string) error {

	url, err := url.Parse(fmt.Sprintf("%s/upload", baseURL))
	if err != nil {
		return err
	}

	req2, err := http.NewRequest("POST", url.String(), bytes.NewReader([]byte("hello world"))) //

	if err != nil {
		return err
	}

	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		return err
	}

	defer resp2.Body.Close()

	out2, err := io.ReadAll(resp2.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(out2))

	return nil

}

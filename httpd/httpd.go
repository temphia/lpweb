package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/temphia/temphia_relay/core"
)

type MeshOptions struct {
	HttpPort   int
	MeshKey    string
	MeshPort   int
	debugPrint bool
}

type Libp2pMesh struct {
	host   host.Host
	dht    *dht.IpfsDHT
	target string
	closed bool
}

func New(opts MeshOptions) *Libp2pMesh {

	h, dht, err := core.NewHost(opts.MeshKey, opts.MeshPort)
	if err != nil {
		panic(err)
	}

	log.Println("httpd@", h.ID())

	for _, m := range h.Addrs() {
		log.Println("httpd@", m.String())
	}

	mesh := &Libp2pMesh{
		dht:    dht,
		host:   h,
		target: "localhost",
	}

	h.SetStreamHandler(core.Protocol, mesh.streamHandler)

	log.Println("Serving mesh @", fmt.Sprintf("http://%s.temphiap2p", h.ID()))

	return mesh
}

func (l *Libp2pMesh) streamHandler(stream network.Stream) {
	defer stream.Close()

	req, err := http.ReadRequest(bufio.NewReader(stream))
	if err != nil {
		panic(err)
	}

	req.URL.Host = "localhost:4000"
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		panic(err)
	}

	out, err := httputil.DumpResponse(resp, true)
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer(out)
	io.Copy(stream, buf)
}

func (h *Libp2pMesh) GetAddr() []multiaddr.Multiaddr {
	return h.host.Addrs()
}

func (l *Libp2pMesh) debugLoop() {
	for {

		if l.closed {
			return
		}

		peers := l.host.Network().Peers()
		log.Println("connected nodes:", len(peers))
		for _, peer := range peers {
			log.Println(peer.Pretty())
		}

		time.Sleep(time.Second * 10)

	}
}

func (l *Libp2pMesh) Close() error {
	l.closed = true
	return l.host.Close()
}

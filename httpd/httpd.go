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

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/temphia/temphia_relay/core"
)

type Httpd struct {
	host host.Host
	mux  *http.ServeMux
	dht  *dht.IpfsDHT
}

func New() *Httpd {

	h, dht, err := core.NewHost("ye2uih109ikdgu1ibkj`1l;.,s;gu2e1j[p21je;l1u2k ei2bj1", 8083)
	if err != nil {
		panic(err)
	}

	log.Println("httpd@", h.ID())

	for _, m := range h.Addrs() {
		log.Println("httpd@", m.String())
	}

	return &Httpd{
		dht:  dht,
		mux:  nil,
		host: h,
	}
}

func (h *Httpd) Run() {

	h.host.SetStreamHandler(core.Protocol, h.streamHandler)

	go h.debugLoop()

	h.mux = http.NewServeMux()
	h.mux.HandleFunc("/", h.Handle)

	log.Println("Starting httpd")

	fmt.Println(http.ListenAndServe(":3333", h.mux))
}

func (h *Httpd) Handle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello universe!!!!!!"))
}

func (h *Httpd) streamHandler(stream network.Stream) {
	defer stream.Close()

	req, err := http.ReadRequest(bufio.NewReader(stream))
	if err != nil {
		panic(err)
	}

	req.URL.Host = "localhost:3333"
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

func (h *Httpd) GetAddr() []multiaddr.Multiaddr {
	return h.host.Addrs()
}

const relay = "12D3KooWAAs35ppLtXZyc24KCdio3z3J3VymAwndoQwWn5e6wmoX"

func (h *Httpd) debugLoop() {
	for {

		peers := h.host.Network().Peers()

		for _, peer := range peers {
			if relay == peer.Pretty() {
				pp.Println("Connected =>", relay)
			}

		}

		time.Sleep(time.Second * 5)

		pp.Println("#################")
	}
}

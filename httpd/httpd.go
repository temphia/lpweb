package httpd

import (
	"bufio"
	"fmt"
	"log"
	"net/http"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/temphia/temphia_relay/core"
)

type Httpd struct {
	host host.Host
	mux  *http.ServeMux
	dht  *dht.IpfsDHT
}

func New() *Httpd {

	h, dht, err := core.NewHost("ye2uih109ik")
	if err != nil {
		panic(err)
	}

	return &Httpd{
		dht:  dht,
		mux:  nil,
		host: h,
	}
}

func (h *Httpd) Run() {

	h.mux = http.NewServeMux()
	h.mux.HandleFunc("/", h.Handle)

	fmt.Println(http.ListenAndServe(":3333", h.mux))
}

func (h *Httpd) Handle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello universe!!!!!!"))
}

func (h *Httpd) streamHandler(stream network.Stream) {
	buf := bufio.NewReader(stream)

	req, err := http.ReadRequest(buf)
	if err != nil {
		stream.Reset()
		log.Println(err)
	}
	defer req.Body.Close()

	req.URL.Scheme = "http"
	req.URL.Host = "localhost:3333"

	outreq := new(http.Request)
	*outreq = *req

	log.Printf("Making request to %s\n.", req.URL)

	resp, err := http.DefaultTransport.RoundTrip(outreq)
	if err != nil {
		stream.Reset()
		log.Println(err)
		return
	}

	resp.Write(stream)
}

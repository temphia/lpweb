package tunnel

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/network"
)

func (ht *HttpTunnel) streamHandler(stream network.Stream) {

	pp.Println("@new_conn")

	defer stream.Close()

	req, err := http.ReadRequest(bufio.NewReader(stream))
	if err != nil {
		panic(err)
	}

	req.URL.Host = fmt.Sprintf("localhost:%d", ht.tunnelToPort)
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		pp.Println("@req", req)
		panic(err)
	}

	out, err := httputil.DumpResponse(resp, true)
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer(out)
	io.Copy(stream, buf)
}

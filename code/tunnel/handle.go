package tunnel

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/libp2p/go-libp2p-core/network"
)

func (ht *HttpTunnel) streamHandler(stream network.Stream) {
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

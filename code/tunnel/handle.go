package tunnel

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/network"
)

func (ht *HttpTunnel) streamHandleHttp(stream network.Stream) {

	maddr, _ := stream.Conn().RemoteMultiaddr().MarshalJSON()
	pp.Println("@new_http_from", string(maddr))

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

	defer resp.Body.Close()

	bodyBackup := resp.Body
	resp.Body = nil

	out, err := httputil.DumpResponse(resp, false)
	if err != nil {
		panic(err)
	}

	pp.Println("@resp", string(out))

	pp.Print("@write_head")
	pp.Println(stream.Write(out))

	pp.Print("@write_body")
	pp.Println(io.Copy(stream, bufio.NewReader(bodyBackup)))
}

func (ht *HttpTunnel) streamHandleWS(stream network.Stream) {
	pp.Println("@new_ws")

	defer stream.Close()

}

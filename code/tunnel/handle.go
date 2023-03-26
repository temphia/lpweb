package tunnel

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/websocket"

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

var (
	DefaultUpgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	DefaultDialer = websocket.DefaultDialer
)

func (ht *HttpTunnel) streamHandleWS(stream network.Stream) {
	pp.Println("@new_ws")
	req, err := http.ReadRequest(bufio.NewReader(stream))
	if err != nil {
		panic(err)
	}

	req.URL.Host = fmt.Sprintf("localhost:%d", ht.tunnelToPort)

	pp.Println("@before_dial")

	wconn, resp, err := DefaultDialer.Dial(req.URL.String(), req.Header)
	if err != nil {
		panic(err)
	}

	pp.Println("@after_dial, dump_resp")

	out, err := httputil.DumpResponse(resp, true)
	if err != nil {
		panic(err)
	}

	pp.Println("@copy_resp_to_stream")
	pp.Println(io.Copy(stream, bytes.NewReader(out)))

	nconn := wconn.UnderlyingConn()

	go func() {
		pp.Println("@copy1")
		pp.Println(io.Copy(nconn, stream))
	}()

	pp.Println("@copy2")
	pp.Println(io.Copy(stream, nconn))
}

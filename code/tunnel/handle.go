package tunnel

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/websocket"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/network"
)

func (ht *HttpTunnel) streamHandleHttp(stream network.Stream) {

	maddr, err := stream.Conn().RemoteMultiaddr().MarshalJSON()
	if err != nil {
		panic(err)
	}

	pp.Println("@new_http_from", string(maddr))

	defer stream.Close()

	req, err := http.ReadRequest(bufio.NewReader(stream))
	if err != nil {
		panic(err)
	}

	req.URL.Host = fmt.Sprintf("localhost:%d", ht.tunnelToPort)
	req.URL.Scheme = "http"
	req.RequestURI = ""

	pp.Println("@connecting_to", req.URL.String())

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
	pp.Println(io.Copy(stream, (bodyBackup)))

}

var (
	DefaultUpgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	DefaultDialer = websocket.DefaultDialer
)

func (ht *HttpTunnel) streamHandleWS(stream network.Stream) {
	defer stream.Close()

	tcpServer, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", ht.tunnelToPort))
	if err != nil {
		panic(err)
	}

	tconn, err := net.DialTCP("tcp", nil, tcpServer)
	if err != nil {
		panic(err)
	}

	pp.Println("@after_dial")

	go func() {
		pp.Println("@copy1")
		pp.Println(io.Copy(tconn, stream))
	}()

	pp.Println("@copy2")
	pp.Println(io.Copy(stream, tconn))
}

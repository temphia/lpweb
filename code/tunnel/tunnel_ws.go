package tunnel

import (
	"fmt"
	"io"
	"net"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/network"
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

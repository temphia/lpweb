package tunnel

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/temphia/lpweb/code/proxy/streamer"
)

func (ht *HttpTunnel) streamHandleHttp2(stream network.Stream) {

	defer stream.Close()

	maddr := stream.Conn().RemoteMultiaddr().String()
	pp.Println("@new_http_from", maddr)

	peerId := stream.Conn().RemotePeer()

	requestIdBytes := make([]byte, 16)
	_, err := stream.Read(requestIdBytes)
	if err != nil {
		pp.Println("@err/Read", err.Error())
		return
	}

	request := &streamer.Streamer{
		RequestId:    requestIdBytes,
		LocalNode:    ht.localNode,
		OutData:      nil,
		ActiveStream: stream,
		Context:      context.TODO(),
		TargetPeer:   peerId,
		InData:       nil,
	}

	err = request.ReceiveData()
	if err != nil {
		pp.Println("@err/ReceiveData", err.Error())
		return

	}

	pp.Println("@read_data", string(request.InData))

	reader := bytes.NewBuffer(request.InData)

	req, err := http.ReadRequest(bufio.NewReader(reader))
	if err != nil {
		pp.Println("@err/ReadRequest", err.Error())
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

	out, err := httputil.DumpResponse(resp, true)
	if err != nil {
		panic(err)
	}

	request.OutData = out

	err = request.SendData()
	if err != nil {
		pp.Println("@err/SendData", err.Error())
		return
	}

}

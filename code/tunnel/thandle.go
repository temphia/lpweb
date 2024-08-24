package tunnel

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"sync"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/network"

	"github.com/fxamacker/cbor"

	"github.com/temphia/lpweb/code/proxy/rcycle"
	"github.com/temphia/lpweb/code/wire"
)

func (ht *HttpTunnel) streamHandleHttp2(stream network.Stream) {

	defer stream.Close()

	maddr := stream.Conn().RemoteMultiaddr().String()
	pp.Println("@new_http_from", maddr)

	peerId := stream.Conn().RemotePeer()

	um := cbor.NewDecoder(stream)
	packet := wire.Packet{}

	err := um.Decode(&packet)
	if err != nil {
		panic(err)
	}

	if packet.PacketType != wire.FragmentSend {

		if packet.HttpRequestId == 0 {
			stream.Close()
			return
		}

		ht.rcLock.Lock()
		request := ht.requestCycles[packet.HttpRequestId]
		ht.rcLock.Unlock()

		if request == nil {
			stream.Close()
			return
		}

		request.OutsidePacket <- rcycle.SideChannelPacket{
			Packet:     &packet,
			FromStream: stream,
		}

		return
	}

	request := rcycle.RequestCycle{
		Context:   context.Background(),
		RequestId: packet.HttpRequestId,
		LocalNode: ht.mesh.Host,

		OutsidePacket:  make(chan rcycle.SideChannelPacket),
		OutData:        packet.Data,
		ActiveStream:   stream,
		TotalFragments: packet.TotalFragments,
		TargetPeer:     peerId,

		DonePacketFrags: make(map[uint32]bool),
		DoneInDataChan:  make(chan uint32),
		InData:          nil,

		CloseChan: make(chan struct{}),
	}

	ht.rcLock.Lock()
	ht.requestCycles[packet.HttpRequestId] = &request
	ht.rcLock.Unlock()

	var wg sync.WaitGroup

	wg.Add(1)

	go request.ControlLoop(&wg, false)

	err = request.StreamReadLoop(stream)
	if err != nil {
		pp.Println("@err/StreamReadLoop", err.Error())
		panic(err)
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

	err = request.StreamWriteLoop()
	if err != nil {
		pp.Println("@err/StreamWriteLoop", err.Error())
	}

	wg.Wait()
}

package proxy

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"sync"
	"sync/atomic"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/temphia/lpweb/code/core/mesh"
	"github.com/temphia/lpweb/code/proxy/rcycle"
	"github.com/temphia/lpweb/code/wire"
)

type SideChannelPacket struct {
	Packet     *wire.Packet
	FromStream network.Stream
}

func (wp *WebProxy) handleHttp2(r *http.Request, w http.ResponseWriter) {
	hash := extractHostHash(r.Host)

	pp.Println("@new_normal_conn", r.Host)

	enode := wp.getExitNode(hash)

	out, err := httputil.DumpRequest(r, true)
	if err != nil {
		panic(err)
	}

	// count no of fragments needed bashed on out size

	totalFragments := uint32(len(out) / wire.PacketFragmentationSize)
	if len(out)%wire.PacketFragmentationSize != 0 {
		totalFragments++
	}

	reqId := atomic.AddUint32(&wp.requestIdCounter, 1)

	stream, err := wp.localNode.NewStream(r.Context(), enode.addr.ID, mesh.ProtocolHttp2)
	if err != nil {
		panic(err)
	}

	request := &rcycle.RequestCycle{
		Context:         r.Context(),
		RequestId:       reqId,
		LocalNode:       wp.localNode,
		OutsidePacket:   make(chan rcycle.SideChannelPacket, 1),
		OutData:         out,
		ActiveStream:    stream,
		TotalFragments:  totalFragments,
		TargetPeer:      enode.addr.ID,
		DonePacketFrags: make(map[uint32]bool),
		DoneInDataChan:  make(chan uint32, 1),
		InData:          nil,
		CloseChan:       make(chan struct{}),
	}

	wp.reqMLock.Lock()
	wp.requests[reqId] = request

	defer func() {
		wp.reqMLock.Lock()
		delete(wp.requests, reqId)
		wp.reqMLock.Unlock()

	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go request.ControlLoop(&wg)

	err = request.StreamWriteLoop()
	if err != nil {
		panic(err)
	}

	err = request.StreamReadLoop(stream)
	if err != nil {
		panic(err)
	}

	request.Close()

	wg.Wait()

	io.Copy(w, bytes.NewBuffer(request.InData))

}

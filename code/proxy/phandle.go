package proxy

import (
	"bufio"
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

	go request.ControlLoop(&wg, true)

	err = request.StreamWriteLoop()
	if err != nil {
		pp.Println("@err/StreamWriteLoop", err.Error())
	}

	err = request.StreamReadLoop(stream)
	if err != nil {
		pp.Println("@err/StreamReadLoop", err.Error())
	}

	request.Close()

	wg.Wait()

	reader := bytes.NewReader(request.InData)
	resp, err := http.ReadResponse(bufio.NewReader(reader), r)
	if err != nil {
		pp.Println("@err/ReadResponse", err.Error())
		panic(err)
	}

	defer resp.Body.Close()

	header := w.Header()
	for k, v := range resp.Header {
		header[k] = v
	}

	w.WriteHeader(resp.StatusCode)

	pp.Println("@write_response")
	if resp.Header.Get("Content-Length") == "" && header.Get("Transfer-Encoding") != "chunked" && resp.Header.Get("Content-Type") == "" {
		pp.Println("@forcing_chunked_mode")
		header.Set("Transfer-Encoding", "chunked")
		pp.Println(io.Copy(httputil.NewChunkedWriter(w), reader))

	} else {
		io.Copy(w, reader)
	}

}

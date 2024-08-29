package proxy

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/temphia/lpweb/code/core/mesh"
	"github.com/temphia/lpweb/code/proxy/streamer"
	"github.com/temphia/lpweb/code/wire"

	nanoid "github.com/matoous/go-nanoid/v2"
)

type SideChannelPacket struct {
	Packet     *wire.Packet
	FromStream network.Stream
}

func (wp *WebProxy) handleHttp2(r *http.Request, w http.ResponseWriter) {
	hash := extractHostHash(r.Host)

	pp.Println("@new_normal_conn", r.Host)

	enode := wp.getExitNode(hash)
	if enode == nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	pp.Println("@handleHttp2/dump_request/1")

	out, err := httputil.DumpRequest(r, true)
	if err != nil {
		panic(err)
	}

	// count no of fragments needed bashed on out size

	pp.Println("@handleHttp2/calc_fragments/2")

	totalFragments := uint32(len(out) / wire.PacketFragmentationSize)
	if len(out)%wire.PacketFragmentationSize != 0 {
		totalFragments++
	}

	pp.Println("@handleHttp2/new_stream/3", enode.addr.ID.String())
	pp.Println("addr_len", len(enode.addr.Addrs))

	stream, err := wp.localNode.NewStream(context.TODO(), enode.addr.ID, mesh.ProtocolHttp2)
	if err != nil {
		pp.Println("@err_new_stream", err.Error())
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	id, err := nanoid.New(16)
	if err != nil {
		panic(err)
	}

	if len(id) != 16 {
		panic("id is not 16 bytes")
	}

	ss := streamer.Streamer{
		RequestId:    []byte(id),
		LocalNode:    wp.localNode,
		OutData:      out,
		TargetPeer:   enode.addr.ID,
		InData:       nil,
		Context:      context.TODO(),
		ActiveStream: stream,
	}

	err = ss.SendData()
	if err != nil {
		panic(err)
	}

	err = ss.ReceiveData()
	if err != nil {
		panic(err)
	}

	reader := bytes.NewReader(ss.InData)
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

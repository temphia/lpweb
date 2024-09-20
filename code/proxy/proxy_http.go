package proxy

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/temphia/lpweb/code/core"
	"github.com/temphia/lpweb/code/core/mesh"
	"github.com/temphia/lpweb/code/proxy/streamer"
	"github.com/temphia/lpweb/code/wire"
)

func (wp *WebProxy) HandleHttp3(r *http.Request, w http.ResponseWriter) {
	hash := extractHostHash(r.Host)

	log.Println("@new_normal_conn", r.Host)

	enode := wp.getExitNode(hash.PeerId)
	if enode == nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	log.Println("@handleHttp2/dump_request/1")

	r.URL.Host = fmt.Sprintf("localhost:%d", hash.Port)
	u2, _ := url.Parse(r.URL.String())
	r.URL = u2

	r.Header.Set("Host", fmt.Sprintf("localhost:%d", hash.Port))

	out, err := httputil.DumpRequest(r, false)
	if err != nil {
		panic(err)
	}

	log.Println("@handleHttp2/new_stream/3", enode.TargetAddr.ID.String())
	log.Println("addr_len", len(enode.TargetAddr.Addrs))

	stream, err := wp.localNode.NewStream(context.TODO(), enode.TargetAddr.ID, mesh.ProtocolHttp)
	if err != nil {
		log.Println("@err_new_stream", err.Error())
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	id := wire.GetRequestId()

	enode.RequestId = id
	enode.ActiveStream = stream
	enode.Context = context.TODO()

	_, err = stream.Write(id)
	if err != nil {
		panic(err)
	}

	wire.WritePacket(stream, &wire.Packet{
		PType:  wire.PTypeSendHeader,
		Offset: 0,
		Total:  int32(r.ContentLength),
		Data:   out,
	})

	// check if request has body

	if r.ContentLength > 0 {

		offset := uint32(0)
		fbuf := make([]byte, wire.FragmentSize)

		for {
			pp.Println("@offset", offset)

			last := false
			n, err := r.Body.Read(fbuf)
			if err != nil {
				if err == io.EOF {
					log.Println("EOF")
					last = true
				} else {
					log.Println("@err/Read", err.Error())
					panic(err)
				}
			}

			ptype := wire.PtypeSendBody
			if last {
				ptype = wire.PtypeEndBody
			}

			toSend := fbuf[:n]

			err = wire.WritePacket(stream, &wire.Packet{
				PType:  ptype,
				Offset: int32(offset),
				Total:  int32(r.ContentLength),
				Data:   toSend,
			})

			if err != nil {
				panic(err)
			}

			offset += uint32(len(toSend))

			if offset >= uint32(r.ContentLength) {
				break
			}

			if last {
				log.Println("@last/break")
				break
			}

		}

	}

	wpack, err := wire.ReadPacket(stream)
	if err != nil {
		panic(err)
	}

	if wpack.PType != wire.PTypeSendHeader {
		panic("invalid packet type 1")
	}

	reader := bytes.NewReader(wpack.Data)
	resp, err := http.ReadResponse(bufio.NewReader(reader), r)
	if err != nil {
		log.Println("@err/ReadResponse", err.Error())
		panic(err)
	}

	header := w.Header()
	if resp.ContentLength > -1 {
		header.Del("Content-Length")
	}
	for k, v := range resp.Header {
		header[k] = v
	}

	w.WriteHeader(resp.StatusCode)

	offset := int32(0)

	pp.Println("RESPONSE_BODY@PROXY/1")

	for {
		pp.Println("@offset", offset)

		wpack, err := wire.ReadPacket(stream)
		if err != nil {
			log.Println("@err/ReadResponse", err.Error())
			panic(err)
		}

		if wpack.Offset < offset {
			panic("invalid offset")
		}

		offset = wpack.Offset

		if wpack.PType != wire.PtypeSendBody &&
			wpack.PType != wire.PtypeEndBody &&
			wpack.PType != wire.PtypeReSendBody {
			pp.Println("HH", int64(wpack.PType))
			panic("invalid packet type 2")
		}

		w.Write(wpack.Data)

		if wpack.PType == wire.PtypeEndBody {
			break
		}
	}

}

// utils

func (wp *WebProxy) getExitNode(target string) *streamer.Streamer {

	addr, err := streamer.ResolveAndConnect(wp.Mesh, target)
	if err != nil {
		pp.Println("@err_creating_upnode", err)
		return nil
	}

	enode := &streamer.Streamer{
		LocalNode:    wp.localNode,
		TargetAddr:   *addr,
		Context:      context.Background(),
		ActiveStream: nil,
		RequestId:    nil,
		InData:       nil,
		OutData:      nil,
	}

	return enode
}

type PeerHost struct {
	PeerId string
	Port   int
}

func extractHostHash(host string) *PeerHost {

	pubkeyEncoded := strings.Split(host, ".")[0]
	port := 0

	pp.Println("@extractHostHash", pubkeyEncoded)

	if strings.Contains(pubkeyEncoded, "-") {
		sllited := strings.Split(pubkeyEncoded, "-")
		pubkeyEncoded = sllited[0]
		port, _ = strconv.Atoi(sllited[1])
	}

	pubkeyDecoded, err := core.DecodeToBytes(pubkeyEncoded)
	if err != nil {
		panic(err)
	}

	//added := append([]byte{0}, pubkeyDecoded...)

	peerId, err := peer.IDFromBytes(pubkeyDecoded)
	if err != nil {
		panic(err)
	}

	pp.Println("@FINAL", peerId.String())

	return &PeerHost{
		PeerId: peerId.String(),
		Port:   port,
	}

}

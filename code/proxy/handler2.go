package proxy

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/polydawn/refmt/cbor"
	"github.com/temphia/lpweb/code/core/mesh"
	"github.com/temphia/lpweb/code/wire"
)

type OnGoingRequest struct {
	id            uint32
	resPacket     map[uint32]bool
	outsidePacket chan SideChannelPacket
}

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

	request := &OnGoingRequest{
		id:            reqId,
		resPacket:     make(map[uint32]bool),
		outsidePacket: make(chan SideChannelPacket, 1),
	}

	wp.reqMLock.Lock()
	wp.requests[reqId] = request

	stream, err := wp.localNode.NewStream(r.Context(), enode.addr.ID, mesh.ProtocolHttp2)
	if err != nil {
		panic(err)
	}

	defer func() {
		wp.reqMLock.Lock()
		delete(wp.requests, reqId)
		wp.reqMLock.Unlock()

	}()

	// send loop

	fragmentId := uint32(0)
	counter := uint32(0)
	writeErrorCount := uint32(0)

	for {
		counter++

		if fragmentId == totalFragments {
			break
		}

		if counter > totalFragments*3 {
			w.Write([]byte("timeout 1"))
			stream.Close()
			return
		}

		// rotate stream if we have too many errors
		if writeErrorCount > 2 {
			stream.Close()

			stream, err = wp.localNode.NewStream(r.Context(), enode.addr.ID, mesh.ProtocolHttp2)
			if err != nil {
				panic(err)
			}

			writeErrorCount = 0
		}

		startOffset := fragmentId * wire.PacketFragmentationSize
		endOffset := fragmentId*wire.PacketFragmentationSize + wire.PacketFragmentationSize

		if endOffset > uint32(len(out)) {
			endOffset = uint32(len(out))
		}

		packet := wire.Packet{
			PacketType:     wire.FragmentSend,
			HttpRequestId:  reqId,
			FragmentId:     fragmentId,
			TotalFragments: totalFragments,
			Data:           out[startOffset:endOffset],
		}

		pout, err := cbor.Marshal(packet)
		if err != nil {
			panic(err)
		}

		_, err = io.Copy(stream, bytes.NewBuffer(pout))
		if err != nil {
			writeErrorCount++
			continue
		}

		fragmentId++
	}

	tallyPacket := wire.Packet{
		PacketType:     wire.FragmentSend,
		HttpRequestId:  reqId,
		FragmentId:     0,
		TotalFragments: totalFragments,
		Data:           nil,
	}

	tbyte, err := cbor.Marshal(tallyPacket)
	if err != nil {
		panic(err)
	}

	counter = uint32(0)
	writeErrorCount = uint32(0)

	for {
		counter++

		if counter > 5 {
			w.Write([]byte("timeout 2"))
			stream.Close()
			return
		}

		if writeErrorCount > 2 {
			stream.Close()

			stream, err = wp.localNode.NewStream(r.Context(), enode.addr.ID, mesh.ProtocolHttp2)
			if err != nil {
				panic(err)
			}

			writeErrorCount = 0
		}

		_, err = io.Copy(stream, bytes.NewBuffer(tbyte))
		if err != nil {
			writeErrorCount++
			continue
		}

		break

	}

	// recv loop

	var (
		finalOutut    []byte          = nil
		doneFragments map[uint32]bool = make(map[uint32]bool)
	)

	// why processPacket is saying unreachable code?
	processPacket := func(sstream network.Stream, rPacket wire.Packet) {

		if rPacket.PacketType == wire.FragmentResend {
			ids := make([]uint32, 0)
			for _, id := range strings.Split(string(rPacket.Data), ",") {
				id, _ := strconv.ParseUint(id, 10, 32)
				ids = append(ids, uint32(id))
			}

			for _, id := range ids {

				startOffset := id * wire.PacketFragmentationSize
				endOffset := id*wire.PacketFragmentationSize + wire.PacketFragmentationSize

				if endOffset > uint32(len(out)) {
					endOffset = uint32(len(out))
				}

				packet := &wire.Packet{
					PacketType:     wire.FragmentSend,
					HttpRequestId:  reqId,
					FragmentId:     id,
					TotalFragments: totalFragments,
					Data:           out[startOffset:endOffset],
				}

				pout, err := cbor.Marshal(packet)
				if err != nil {
					panic(err)
				}

				_, err = io.Copy(stream, bytes.NewBuffer(pout))
				if err != nil {
					writeErrorCount++
					continue
				}

			}

			return
		}

		if rPacket.PacketType == wire.FragmentSend {
			if finalOutut == nil {
				finalOutut = make([]byte, rPacket.TotalFragments*wire.PacketFragmentationSize)
				for i := 0; i < int(rPacket.TotalFragments); i++ {
					doneFragments[uint32(i)] = false
				}
			}

			startOffset := rPacket.FragmentId * wire.PacketFragmentationSize
			endOffset := rPacket.FragmentId*wire.PacketFragmentationSize + wire.PacketFragmentationSize

			if endOffset > uint32(len(finalOutut)) {
				endOffset = uint32(len(finalOutut))
			}

			copy(finalOutut[startOffset:endOffset], rPacket.Data)
			doneFragments[rPacket.FragmentId] = true

			return

		}

		if rPacket.PacketType == wire.FragmentTally {
			//check if all fragments are done
			pendingIds := strings.Builder{}
			isFirst := true

			for fid, done := range doneFragments {
				if !done {

					if isFirst {
						isFirst = false
					} else {
						pendingIds.WriteString(",")
					}

					pendingIds.WriteString(strconv.FormatUint(uint64(fid), 10))

				}

			}

			packet := &wire.Packet{
				PacketType:     wire.FragmentResend,
				HttpRequestId:  reqId,
				FragmentId:     0,
				TotalFragments: totalFragments,
				Data:           []byte(pendingIds.String()),
			}

			pout, err := cbor.Marshal(packet)
			if err != nil {
				panic(err)
			}

			io.Copy(sstream, bytes.NewBuffer(pout))
		}

	}

	for {

		rPacket := wire.Packet{}

		m := cbor.NewUnmarshaller(cbor.DecodeOptions{
			CoerceUndefToNull: true,
		}, stream)

		err = m.Unmarshal(&rPacket)
		if err != nil {
			panic(err)
		}

		processPacket(stream, rPacket)

		// all done
		alldone := true
		for _, done := range doneFragments {
			if !done {
				alldone = false
			}
		}

		if alldone {
			break
		}

		select {
		case outreq := <-request.outsidePacket:
			processPacket(stream, *outreq.Packet)
		default:
		}

	}

	resp, err := http.ReadResponse(bufio.NewReader(stream), r)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	header := w.Header()
	for k, v := range resp.Header {
		header[k] = v
	}

	w.WriteHeader(resp.StatusCode)

	pp.Println(io.Copy(w, stream))

}

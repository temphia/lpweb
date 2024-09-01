package proxy

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/k0kubun/pp"
	"github.com/temphia/lpweb/code/core/mesh"
	"github.com/temphia/lpweb/code/wire"
)

const fragmentSize = 1024 * 256

func (wp *WebProxy) handleHttp3(r *http.Request, w http.ResponseWriter) {
	hash := extractHostHash(r.Host)

	log.Println("@new_normal_conn", r.Host)

	enode := wp.getExitNode(hash)
	if enode == nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	log.Println("@handleHttp2/dump_request/1")

	out, err := httputil.DumpRequest(r, false)
	if err != nil {
		panic(err)
	}

	log.Println("@handleHttp2/new_stream/3", enode.addr.ID.String())
	log.Println("addr_len", len(enode.addr.Addrs))

	stream, err := wp.localNode.NewStream(context.TODO(), enode.addr.ID, mesh.ProtocolHttp3)
	if err != nil {
		log.Println("@err_new_stream", err.Error())
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	id := wire.GetRequestId()

	_, err = stream.Write(id)
	if err != nil {
		panic(err)
	}

	wire.WritePacket(stream, &wire.Packet{
		PType:  wire.PTypeSendHeader,
		Offset: 0,
		Total:  uint32(r.ContentLength),
		Data:   out,
	})

	// check if request has body

	if r.ContentLength > 0 {

		offset := uint32(0)
		fbuf := make([]byte, fragmentSize)

		for {
			pp.Println("@offset", offset)

			last := false
			n, err := r.Body.Read(fbuf)
			if err == io.EOF {
				log.Println("EOF")
				last = true
			}

			if err != nil {
				panic(err)
			}

			ptype := wire.PtypeSendBody
			if last {
				ptype = wire.PtypeEndBody
			}

			err = wire.WritePacket(stream, &wire.Packet{
				PType:  ptype,
				Offset: uint32(offset),
				Total:  uint32(r.ContentLength),
				Data:   fbuf[:n],
			})

			if err != nil {
				panic(err)
			}

			offset += uint32(n)

			if offset >= uint32(r.ContentLength) {
				break
			}

			if last {
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
	for k, v := range resp.Header {
		header[k] = v
	}

	w.WriteHeader(resp.StatusCode)

	if resp.ContentLength > 0 {
		offset := uint32(0)

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

}

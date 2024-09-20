package tunnel

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/temphia/lpweb/code/wire"
)

func (ht *HttpTunnel) streamHandleHttp(stream network.Stream) {

	defer stream.Close()

	maddr := stream.Conn().RemoteMultiaddr().String()
	pp.Println("@new_http_from", maddr)

	peerId := stream.Conn().RemotePeer()

	pp.Println("@new_http_from", peerId.String())

	requestIdBytes := make([]byte, 16)
	_, err := stream.Read(requestIdBytes)
	if err != nil {
		pp.Println("@err/Read", err.Error())
		return
	}

	pp.Println("@read_data", string(requestIdBytes))

	wpak, err := wire.ReadPacket(stream)
	if err != nil {
		log.Println("@err/Read", err.Error())
		return
	}

	if wpak.PType != wire.PTypeSendHeader {
		panic("invalid packet type 3")
	}

	reader := bytes.NewBuffer(wpak.Data)

	req, err := http.ReadRequest(bufio.NewReader(reader))
	if err != nil {
		pp.Println("@err/ReadRequest", err.Error())
		panic(err)
	}

	port := ht.GetTunnelPort(req.URL.String())

	req.URL.Host = fmt.Sprintf("localhost:%s", port)
	req.URL.Scheme = "http"
	req.RequestURI = ""
	req.Host = fmt.Sprintf("localhost:%d", ht.tunnelToPort)

	if req.ContentLength > 0 {
		req.Body = &proxyReader{
			stream:           stream,
			lastPendingBytes: make([]byte, 0),
			isEOF:            false,
		}

	}

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		pp.Println("@req", req)
		panic(err)
	}

	defer resp.Body.Close()

	out, err := httputil.DumpResponse(resp, false)
	if err != nil {
		panic(err)
	}

	err = wire.WritePacket(stream, &wire.Packet{
		PType:  wire.PTypeSendHeader,
		Offset: 0,
		Total:  int32(resp.ContentLength),
		Data:   out,
	})
	if err != nil {
		log.Println("@err/Write", err.Error())
		return
	}

	pp.Println("@write_data", string(out))

	offset := uint32(0)
	fbuf := make([]byte, wire.FragmentSize)

	pp.Println("RESPONSE_BODY@TUNNEL/1", resp.ContentLength)

	for {

		pp.Println("@offset", offset)

		last := false

		n, err := resp.Body.Read(fbuf)
		if err == io.EOF {
			last = true
		}

		ptype := wire.PtypeSendBody
		if last {
			ptype = wire.PtypeEndBody
		}

		toSend := fbuf[:n]

		err = wire.WritePacket(stream, &wire.Packet{
			PType:  ptype,
			Offset: int32(offset),
			Total:  int32(resp.ContentLength),
			Data:   toSend,
		})

		if err != nil {
			log.Println("@err/Write", err.Error())
			return
		}

		offset += uint32(len(toSend))

		if resp.ContentLength != 0 && offset >= uint32(resp.ContentLength) {
			break
		}

		if last {
			break
		}

	}

}

func (ht *HttpTunnel) GetTunnelPort(u string) string {
	port := strconv.Itoa(ht.tunnelToPort)

	if ht.tunnelAnyPort {
		u, err := url.Parse(u)
		if err != nil {
			panic(err)
		}

		if u.Port() != "0" || u.Port() == "" {
			port = u.Port()
		}
	}

	return port
}

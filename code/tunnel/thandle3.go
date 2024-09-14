package tunnel

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/temphia/lpweb/code/wire"
)

func (ht *HttpTunnel) streamHandleHttp3(stream network.Stream) {

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

	req.URL.Host = fmt.Sprintf("localhost:%d", ht.tunnelToPort)
	req.URL.Scheme = "http"
	req.RequestURI = ""
	req.Host = fmt.Sprintf("localhost:%d", ht.tunnelToPort)

	// fixme => request body proxy / post upload  support

	if req.ContentLength > 0 {
		req.Body = &channelReader{
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

	pp.Println("RESPONSE_BODY")

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

type channelReader struct {
	stream           network.Stream
	lastPendingBytes []byte
	isEOF            bool
}

func (c *channelReader) Close() error {
	return c.stream.Close()
}

func (c *channelReader) Read(p []byte) (int, error) {

	bufSize := len(p)
	copied := 0

	pp.Println("@channelReader/1", bufSize)

	if len(c.lastPendingBytes) > 0 {

		pp.Println("@channelReader/2")

		bytesToCopy := len(c.lastPendingBytes)
		if bytesToCopy > len(p) {
			pp.Println("@channelReader/3")
			bytesToCopy = len(p)
		}

		pp.Println("@channelReader/4", bytesToCopy)

		n := copy(p, c.lastPendingBytes[:bytesToCopy])

		pp.Println("@channelReader/5", n)

		if bytesToCopy != n {
			panic("bytesToCopy != n -> copy should not happen")
		}

		c.lastPendingBytes = c.lastPendingBytes[bytesToCopy:]

		pp.Println("@channelReader/6", len(c.lastPendingBytes))

		copied = n

		if bufSize == copied {
			if len(c.lastPendingBytes) == 0 && c.isEOF {
				return copied, io.EOF
			}

			return copied, nil
		}

		p = p[bytesToCopy:]
	}

	pp.Println("@channelReader/7")

	if len(c.lastPendingBytes) != 0 {
		panic("lastPendingBytes should be empty")
	}

	loopCount := 0

	for {
		pp.Println("@channelReader/8", loopCount)

		packet, err := wire.ReadPacket(c.stream)
		if err != nil {
			pp.Println("@channelReader/9", err.Error())
			return 0, err
		}

		pp.Println("@channelReader/10")

		if packet.PType != wire.PtypeSendBody &&
			packet.PType != wire.PtypeEndBody &&
			packet.PType != wire.PtypeReSendBody {
			panic("invalid packet type")
		}

		pp.Println("@channelReader/11")

		if packet.PType != wire.PtypeEndBody {
			pp.Println("@channelReader/12")

			c.isEOF = true
		}

		pp.Println("@channelReader/13")

		n := copy(p, packet.Data)
		copied += n
		c.lastPendingBytes = packet.Data[n:]

		pp.Println("@channelReader/14", copied)

		if bufSize == copied {
			pp.Println("@channelReader/15")

			if len(c.lastPendingBytes) == 0 && c.isEOF {
				pp.Println("@channelReader/16")

				return copied, io.EOF
			}

			pp.Println("@channelReader/17")

			return copied, nil
		}

		pp.Println("@channelReader/18")

		p = p[n:]
		loopCount++
	}

}

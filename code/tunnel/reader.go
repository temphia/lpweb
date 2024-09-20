package tunnel

import (
	"io"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/temphia/lpweb/code/wire"
)

type proxyReader struct {
	stream           network.Stream
	lastPendingBytes []byte
	isEOF            bool
}

func (c *proxyReader) Close() error {
	return c.stream.Close()
}

func (c *proxyReader) Read(p []byte) (int, error) {

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

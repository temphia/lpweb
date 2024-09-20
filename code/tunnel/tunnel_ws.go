package tunnel

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/temphia/lpweb/code/wire"
)

func (ht *HttpTunnel) streamHandleWS(stream network.Stream) {
	defer stream.Close()

	reqId := make([]byte, 16)

	_, err := stream.Read(reqId)
	if err != nil {
		pp.Println("@streamHandleWS/1", err.Error())
		return
	}

	packet, err := wire.ReadPacket(stream)
	if err != nil {
		panic(err.Error())
	}

	reader := bytes.NewBuffer(packet.Data)

	req, err := http.ReadRequest(bufio.NewReader(reader))
	if err != nil {
		pp.Println("@streamHandleWS/2", err.Error())
		return
	}

	wsUrl := fmt.Sprintf("ws://localhost:%d%s", ht.tunnelToPort, req.URL.Path)

	pp.Println("@streamHandleWS/3", wsUrl)

	c, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		pp.Println("streamHandleWS/dial2:", err)
		return
	}

	defer c.Close()

	go func() {

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				pp.Println("streamHandleWS/read1:", err)
				return
			}

			for {
				if len(message) == 0 {
					break
				}

				n, err := stream.Write(message)
				if err != nil {
					pp.Println("streamHandleWS/write1:", err)
					return
				}

				message = message[n:]
			}

		}
	}()

	streamBuf := make([]byte, 1024)

	for {
		n, err := stream.Read(streamBuf)
		if err != nil {
			pp.Println("streamHandleWS/read2:", err)
			break
		}
		err = c.WriteMessage(websocket.TextMessage, streamBuf[:n])
		if err != nil {
			pp.Println("streamHandleWS/write2:", err)
			break
		}
	}

}

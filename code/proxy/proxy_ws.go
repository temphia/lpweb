package proxy

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/websocket"
	"github.com/k0kubun/pp"
	"github.com/temphia/lpweb/code/core/mesh"
	"github.com/temphia/lpweb/code/wire"
)

var upgrader = websocket.Upgrader{}

func (wp *WebProxy) HandleWS(r *http.Request, w http.ResponseWriter) {
	hash := extractHostHash(r.Host)
	pp.Println("@new_ws_conn", r.URL)

	pp.Println(hash)

	enode := wp.getExitNode(hash.PeerId)

	stream, err := wp.localNode.NewStream(context.TODO(), enode.TargetAddr.ID, mesh.ProtocolWS)
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	out, err := httputil.DumpRequest(r, false)
	if err != nil {
		panic(err)
	}

	reqid := wire.GetRequestId()

	_, err = stream.Write(reqid)
	if err != nil {
		panic(err)
	}

	err = wire.WritePacket(stream, &wire.Packet{
		PType:  wire.PTypeSendHeader,
		Offset: 0,
		Total:  int32(r.ContentLength),
		Data:   out,
	})
	if err != nil {
		panic(err)
	}

	pp.Println("@opened_new_stream")

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	defer c.Close()

	go func() {

		readBuf := make([]byte, 1024)

		for {
			message, err := stream.Read(readBuf)
			if err != nil {
				log.Println("read:", err)
				break
			}

			err = c.WriteMessage(websocket.TextMessage, readBuf[:message])
			if err != nil {
				log.Println("write:", err)
				break
			}
		}

	}()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		for {
			if len(message) == 0 {
				break
			}

			n, err := stream.Write(message)
			if err != nil {
				log.Println("write:", err)
				break
			}
			message = message[n:]
		}

	}

}

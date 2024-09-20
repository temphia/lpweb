package proxy

import (
	"context"
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

	pp.Println("@HandleWS/1", hash)

	streamer := wp.getExitNode(hash.PeerId)

	stream, err := wp.localNode.NewStream(context.TODO(), streamer.TargetAddr.ID, mesh.ProtocolWS)
	if err != nil {
		pp.Println("@HandleWS/2", err.Error())
		return
	}
	defer stream.Close()

	out, err := httputil.DumpRequest(r, false)
	if err != nil {
		pp.Println("@HandleWS/3", err.Error())
		return
	}

	reqid := wire.GetRequestId()

	_, err = stream.Write(reqid)
	if err != nil {
		pp.Println("@HandleWS/4", err.Error())
		return
	}

	err = wire.WritePacket(stream, &wire.Packet{
		PType:  wire.PTypeSendHeader,
		Offset: 0,
		Total:  int32(r.ContentLength),
		Data:   out,
	})
	if err != nil {
		pp.Println("@HandleWS/5", err.Error())
		return
	}

	pp.Println("@HandleWS/6")

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		pp.Println("HandleWS/upgrade/7", err)
		return
	}

	defer c.Close()

	go func() {

		pp.Println("HandleWS/go/8")

		readBuf := make([]byte, 1024)

		for {
			message, err := stream.Read(readBuf)
			if err != nil {
				pp.Println("HandleWS/go/9", err.Error())
				break
			}

			pp.Println("HandleWS/go/10")

			err = c.WriteMessage(websocket.TextMessage, readBuf[:message])
			if err != nil {
				pp.Println("HandleWS/go/11", err.Error())
				break
			}
			pp.Println("HandleWS/go/12")
		}

		pp.Println("HandleWS/go/13")

	}()

	pp.Println("HandleWS/go/14")

	for {

		pp.Println("HandleWS/go/15")

		_, message, err := c.ReadMessage()
		if err != nil {
			pp.Println("HandleWS/go/16", err.Error())
			break
		}

		for {
			pp.Println("HandleWS/go/17")
			if len(message) == 0 {
				pp.Println("HandleWS/go/18")
				break
			}
			pp.Println("HandleWS/go/19")

			n, err := stream.Write(message)
			if err != nil {
				pp.Println("HandleWS/go/20", err.Error())
				break
			}

			pp.Println("HandleWS/go/21")

			message = message[n:]
		}

	}

}

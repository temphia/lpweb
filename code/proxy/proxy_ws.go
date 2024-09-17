package proxy

import (
	"context"
	"io"
	"net/http"

	"github.com/k0kubun/pp"
	"github.com/temphia/lpweb/code/core/mesh"
)

func (wp *WebProxy) handleWS(r *http.Request, w http.ResponseWriter) {
	hash := extractHostHash(r.Host)
	pp.Println("@new_ws_conn", r.URL)

	pp.Println(hash)

	w.Write([]byte("HTTP/1.0 200 Connection established\r\n\r\n"))
	pp.Println("@accepted_connect")

	enode := wp.getExitNode(hash)

	stream, err := wp.localNode.NewStream(context.TODO(), enode.TargetAddr.ID, mesh.ProtocolWS)
	if err != nil {
		panic(err)
	}
	defer stream.Close()

	pp.Println("@opened_new_stream")

	pp.Println("@hijack_success")
	hjconn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		pp.Println("@err_while_hijacking", err.Error())
		return
	}

	defer hjconn.Close()

	go func() {
		pp.Println("@copy_stream/req2stream")
		pp.Println(io.Copy(stream, hjconn))
	}()

	pp.Println("@copy_stream/stream2req")
	pp.Println(io.Copy(hjconn, stream))

}

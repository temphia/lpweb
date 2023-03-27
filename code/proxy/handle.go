package proxy

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/temphia/lpweb/code/core/mesh"
)

func (wp *WebProxy) handleHttp(r *http.Request, w http.ResponseWriter) {
	hash := extractHostHash(r.Host)

	pp.Println("@new_normal_conn", r.Host)

	enode := wp.getExitNode(hash)

	stream, err := wp.localNode.NewStream(context.TODO(), enode.addr.ID, mesh.ProtocolHttp)
	if err != nil {
		panic(err)
	}

	out, err := httputil.DumpRequest(r, true)
	if err != nil {
		panic(err)
	}

	pp.Println("@copy_request")
	pp.Println(io.Copy(stream, bytes.NewBuffer(out)))

	resp, err := http.ReadResponse(bufio.NewReader(stream), r)
	if err != nil {
		panic(err)
	}

	header := w.Header()
	for k, v := range resp.Header {
		header[k] = v
	}

	w.WriteHeader(resp.StatusCode)

	pp.Println("@write_response")
	pp.Println(io.Copy(w, resp.Body))
	resp.Body.Close()
}

func (wp *WebProxy) handleWS(r *http.Request, w http.ResponseWriter) {
	hash := extractHostHash(r.Host)
	pp.Println("@new_ws_conn", r.Host)
	pp.Println(hash)

	enode := wp.getExitNode(hash)

	stream, err := wp.localNode.NewStream(context.TODO(), enode.addr.ID, mesh.ProtocolWS)
	if err != nil {
		panic(err)
	}

	pp.Println("@opened_new_stream")

	out, err := httputil.DumpRequest(r, true)
	if err != nil {
		panic(err)
	}
	pp.Println("@dump_request")

	pp.Println("@copy_req_to_stream")
	pp.Println(io.Copy(stream, bytes.NewBuffer(out)))

	hjconn, rw, err := w.(http.Hijacker).Hijack()
	if err != nil {
		pp.Println("@err_while_hijacking", err.Error())
		return
	}
	pp.Println("@hijack_success")

	pp.Println("@write_handlshake")
	io.Copy(hjconn, bufio.NewReader(stream))

	go func() {
		pp.Println("@copy_stream1")
		pp.Println(io.Copy(rw, stream))
	}()

	pp.Println(io.Copy(stream, rw))

}

func extractHostHash(host string) string {
	return strings.Split(host, ".")[0]
}

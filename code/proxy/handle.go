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
	if enode == nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	stream, err := wp.localNode.NewStream(context.TODO(), enode.addr.ID, mesh.ProtocolHttp)
	if err != nil {
		panic(err)
	}

	defer stream.Close()

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
	defer resp.Body.Close()

	header := w.Header()
	for k, v := range resp.Header {
		header[k] = v
	}

	w.WriteHeader(resp.StatusCode)

	pp.Println("@write_response")
	if resp.Header.Get("Content-Length") == "" && header.Get("Transfer-Encoding") != "chunked" && resp.Header.Get("Content-Type") == "" {
		pp.Println("@forcing_chunked_mode")
		header.Set("Transfer-Encoding", "chunked")
		pp.Println(io.Copy(httputil.NewChunkedWriter(w), stream))
		return
	}

	pp.Println(io.Copy(w, stream))

}

func (wp *WebProxy) handleWS(r *http.Request, w http.ResponseWriter) {
	hash := extractHostHash(r.Host)
	pp.Println("@new_ws_conn", r.URL)

	pp.Println(hash)

	w.Write([]byte("HTTP/1.0 200 Connection established\r\n\r\n"))
	pp.Println("@accepted_connect")

	enode := wp.getExitNode(hash)
	if enode == nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	stream, err := wp.localNode.NewStream(context.TODO(), enode.addr.ID, mesh.ProtocolWS)
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

func extractHostHash(host string) string {
	return strings.Split(host, ".")[0]
}

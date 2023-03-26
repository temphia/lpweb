package proxy

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/elazarl/goproxy"
	"github.com/k0kubun/pp"
	"github.com/temphia/lpweb/code/core/mesh"
)

func (wp *WebProxy) handle(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {

	pp.Println(r.Header)

	if r.Header.Get("Upgrade") == "websocket" {
		return wp.handleWS(r, ctx)
	}

	return wp.handleHttp(r, ctx)
}

func (wp *WebProxy) handleHttp(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
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

	return nil, resp
}

func (wp *WebProxy) handleWS(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	hash := extractHostHash(r.Host)
	pp.Println("@new_ws_conn", r.Host)

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

	pp.Println(io.Copy(stream, bytes.NewBuffer(out)))

	pp.Println("@opened_new_stream")

	resp, err := http.ReadResponse(bufio.NewReader(stream), r)
	if err != nil {
		panic(err)
	}

	pp.Println("@read_response")

	_, rw, err := ctx.RespWriter.(http.Hijacker).Hijack()
	if err != nil {
		pp.Println("@err_while_hijacking", err.Error())
		return nil, resp
	}

	pp.Println("@hijack_success")

	go func() {
		pp.Println("@copy_stream1")
		pp.Println(io.Copy(rw, stream))
	}()

	go func() {
		pp.Println("@copy_stream2")
		pp.Println(io.Copy(stream, rw))
	}()

	return nil, resp
}

func extractHostHash(host string) string {
	return strings.Split(host, ".")[0]
}

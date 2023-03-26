package proxy

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/elazarl/goproxy"
	"github.com/temphia/lpweb/code/core/mesh"
)

func (wp *WebProxy) handle(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {

	if r.Header.Get("Upgrade") == "websocket" {
		return wp.handleWS(r, ctx)
	}

	return wp.handleHttp(r, ctx)
}

func (wp *WebProxy) handleHttp(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	hash := extractHostHash(r.Host)

	log.Println("@new_conn", r.Host)

	enode := wp.getExitNode(hash)

	stream, err := wp.localNode.NewStream(context.TODO(), enode.addr.ID, mesh.ProtocolHttp)
	if err != nil {
		panic(err)
	}

	out, err := httputil.DumpRequest(r, true)
	if err != nil {
		panic(err)
	}

	io.Copy(stream, bytes.NewBuffer(out))

	resp, err := http.ReadResponse(bufio.NewReader(stream), r)
	if err != nil {
		panic(err)
	}

	return nil, resp
}

func (wp *WebProxy) handleWS(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	hash := extractHostHash(r.Host)
	log.Println("@new_conn", r.Host)

	enode := wp.getExitNode(hash)

	stream, err := wp.localNode.NewStream(context.TODO(), enode.addr.ID, mesh.ProtocolWS)
	if err != nil {
		panic(err)
	}

	out, err := httputil.DumpRequest(r, true)
	if err != nil {
		panic(err)
	}

	io.Copy(stream, bytes.NewBuffer(out))

	resp, err := http.ReadResponse(bufio.NewReader(stream), r)
	if err != nil {
		panic(err)
	}

	// http.Hijacker here

	return nil, resp
}

func extractHostHash(host string) string {
	return strings.Split(host, ".")[0]
}

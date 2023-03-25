package proxy

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/elazarl/goproxy"
	"github.com/inconshreveable/go-vhost"
	"github.com/temphia/lpweb/code/core/mesh"
)

func (wp *WebProxy) handle(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {

	host := strings.Split(r.Host, ".")[0]

	enode := wp.getExitNode(host)

	stream, err := wp.localNode.NewStream(context.TODO(), enode.addr.ID, mesh.Protocol)
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

func (wp *WebProxy) listenTLS() {
	ln, err := net.Listen("tcp", "localhost:8043")
	if err != nil {
		log.Fatalf("Error listening for https connections - %v", err)
	}
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting new connection - %v", err)
			continue
		}
		go func(c net.Conn) {
			tlsConn, err := vhost.TLS(c)
			if err != nil {
				log.Printf("Error accepting new connection - %v", err)
			}
			if tlsConn.Host() == "" {
				log.Printf("Cannot support non-SNI enabled clients")
				return
			}
			connectReq := &http.Request{
				Method: "CONNECT",
				URL: &url.URL{
					Opaque: tlsConn.Host(),
					Host:   net.JoinHostPort(tlsConn.Host(), "443"),
				},
				Host:       tlsConn.Host(),
				Header:     make(http.Header),
				RemoteAddr: c.RemoteAddr().String(),
			}
			resp := dumbResponseWriter{tlsConn}

			wp.proxy.ServeHTTP(resp, connectReq)
		}(c)
	}

}

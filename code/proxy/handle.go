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
	"github.com/k0kubun/pp"
	"github.com/temphia/lpweb/code/core/mesh"
)

func (wp *WebProxy) handle(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {

	pp.Println("@new_connection")

	host := strings.Split(r.Host, ".")[0]

	log.Println("@new_conn", r.Host)

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

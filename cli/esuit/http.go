package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/k0kubun/pp"
)

func (e *Esuit) StartHttpServer() {

	server := &http.Server{
		Addr: ":" + strconv.Itoa(proxyPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// check if its favicon
			if strings.Contains(r.URL.Path, "favicon") {
				w.WriteHeader(http.StatusOK)
				return
			}

			pp.Println("@ALL_INTERCEPT", r.URL.String())
			pp.Println("@ALL_HOST", r.Host)

			// make <pubkey>.localhost to xyz.lpweb

			hostParts := strings.Split(r.Host, ".")
			newHostName := hostParts[0] + ".lpweb"

			pp.Println("@new_host_name", hostParts, newHostName)

			r.URL.Host = newHostName
			r.Host = newHostName

			if r.Header.Get("Upgrade") == "websocket" {
				e.proxy.HandleWS(r, w)
			} else {
				e.proxy.HandleHttp3(r, w)
			}
		}),
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}

}

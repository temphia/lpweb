package main

import (
	"net/http"
	"strings"

	"github.com/k0kubun/pp"
)

func (e *Esuit) StartHttpServer() {

	server := &http.Server{
		Addr: ":8001",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			pp.Println("@ALL_INTERCEPT", r.URL.String())

			// make xyz.localhost.com to xyz.lpweb

			hostParts := strings.Split(r.Host, ".")
			r.URL.Host = hostParts[0] + ".lpweb"
			e.proxy.HandleHttp3(r, w)
		}),
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}

}

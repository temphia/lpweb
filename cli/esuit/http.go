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

			pp.Println("@ALL_INTERCEPT", r.URL.String())
			pp.Println("@ALL_HOST", r.Host)

			// make <pubkey>.localhost to xyz.lpweb

			hostParts := strings.Split(r.Host, ".")
			newHostName := hostParts[0] + ".lpweb"

			pp.Println("@new_host_name", hostParts, newHostName)

			r.URL.Host = newHostName
			r.Host = newHostName

			e.proxy.HandleHttp3(r, w)
		}),
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}

}

package main

import (
	"io"
	"net/http"

	"github.com/k0kubun/pp"
)

func (e *Esuit) StartFileServer() {

	// static file server

	fserver := http.FileServer(http.Dir("./"))

	server := &http.Server{
		Addr: ":7704",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// dump req body

			out, err := io.ReadAll(r.Body)
			if err != nil {
				panic(err)
			}

			pp.Println("@DUMP_REQ_BODY", string(out))

			fserver.ServeHTTP(w, r)
		}),
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}

}

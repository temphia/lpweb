package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/k0kubun/pp"
)

func (e *Esuit) StartFileServer() {

	// static file server

	fserver := http.FileServer(http.Dir("./"))

	server := http.NewServeMux()

	server.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		fserver.ServeHTTP(w, r)
	})

	server.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			panic("invalid method")
		}

		out, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		pp.Println("@UPLOAD_REQ_BODY", string(out))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("uploaded %d bytes \n\n%s", len(out), string(out))))
	})

	err := http.ListenAndServe(":"+strconv.Itoa(tunnelPort), server)
	if err != nil {
		panic(err)
	}

}

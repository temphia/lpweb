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
		pp.Println("@list")
		fserver.ServeHTTP(w, r)
	})

	server.HandleFunc("/text_file", func(w http.ResponseWriter, r *http.Request) {

		w.WriteHeader(http.StatusOK)
		w.Write(TestUploadData2)
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

	port := tunnelPort

	err := http.ListenAndServe(":"+strconv.Itoa(port), server)
	if err != nil {
		panic(err)
	}

}

package main

import "net/http"

func (e *Esuit) StartFileServer() {

	// static file server

	fserver := http.FileServer(http.Dir("./"))

	server := &http.Server{
		Addr: ":7704",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fserver.ServeHTTP(w, r)
		}),
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}

}

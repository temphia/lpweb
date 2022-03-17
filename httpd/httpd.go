package httpd

import (
	"fmt"
	"net/http"

	"github.com/libp2p/go-libp2p-core/host"
)

type Httpd struct {
	host host.Host
}

func New() *Httpd {
	return &Httpd{}
}

func (h *Httpd) Run() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.Handle)

	fmt.Println(http.ListenAndServe(":3333", mux))
}

func (h *Httpd) Handle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello universe!!!!!!"))
}

package etcd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/tidwall/pretty"
)

var uuid = "9e7caedb576b1ab1a4d209da31f96389"
var URL = "https://discovery.etcd.io/%s"

func init() {
	// InitilizeUUID()
}

type PubEtcd struct {
	baseURL string
	uuid    string
}

func New() *PubEtcd {
	burl := fmt.Sprintf(URL, uuid)

	return &PubEtcd{
		uuid:    uuid,
		baseURL: burl,
	}
}

func (p *PubEtcd) Set(hash, addr string) error {

	vals := make(url.Values)
	vals.Set("value", addr)

	ev := vals.Encode()

	url := urlkey(hash, p.baseURL)

	pp.Println("@pulishing_discovery", url)

	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(ev))
	if err != nil {
		return err
	}

	req.Header.Set("Content-type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Print("@resp", string((pretty.Color(pretty.Pretty(out), nil))))
	fmt.Print("\n")

	return nil
}

func (p *PubEtcd) Get(hash string) (string, error) {
	url := urlkey(hash, p.baseURL)

	pp.Println("@getting_discovery", url)

	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return "", err
	}

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func urlkey(hash, url string) string {
	return fmt.Sprintf("%s/%s-servers", url, hashIt(hash))
}

func hashIt(bv string) string {
	return strings.ToLower(bv)
}

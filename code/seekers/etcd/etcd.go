package etcd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/k0kubun/pp"
)

var uuid = "f116f98c258214878c4a1fa08563efbf"
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

	req, err := http.NewRequest(http.MethodPut, urlkey(hash, p.baseURL), strings.NewReader(ev))
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

	pp.Println("@resp", string(out))

	return nil
}

func (p *PubEtcd) Get(hash string) (string, error) {
	resp, err := http.DefaultClient.Get(urlkey(hash, p.baseURL))
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
	return fmt.Sprintf("%s-%s/servers", hash, url)
}

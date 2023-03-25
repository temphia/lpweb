package etcd

import "fmt"

var uuid = "f116f98c258214878c4a1fa08563efbf"
var URL = "https://discovery.etcd.io/%s"

func init() {

}

type PubEtcd struct {
	baseURL string
	uuid    string
}

func New() *PubEtcd {

	return &PubEtcd{
		uuid:    uuid,
		baseURL: fmt.Sprintf(URL, uuid),
	}
}

func (p *PubEtcd) Set(hash, addr string) error {

	return nil
}

func (p *PubEtcd) Get(hash string) (string, error) {
	return "", nil
}

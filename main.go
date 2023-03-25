package main

import (
	"github.com/k0kubun/pp"
	"github.com/temphia/lpweb/code/seekers/etcd"
)

func main() {

	es := etcd.New()

	pp.Println(es.Get(""))

	// cli.RunCLI()

}

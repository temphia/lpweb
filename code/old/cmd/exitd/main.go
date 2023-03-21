package main

import (
	"github.com/temphia/lpweb/cmd/exitd/exitserver"
)

func main() {
	instance := exitserver.NewInstance()
	instance.Run()
}

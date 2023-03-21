package main

import "github.com/temphia/lpweb/code/core/old/cmd/exitd/exitserver"

func main() {
	instance := exitserver.NewInstance()
	instance.Run()
}

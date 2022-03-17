package main

import (
	"github.com/temphia/temphia_relay/p2p"
)

func main() {
	instance := p2p.NewInstance()
	instance.Run()
}

package core

import (
	"fmt"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/tidwall/pretty"
)

func PrintPeerAddr(pa peer.AddrInfo) {
	paddr, _ := pa.MarshalJSON()
	PrintBytes(paddr)
}

func PrintBytes(out []byte) {
	fmt.Print(string(pretty.Color(pretty.Pretty(out), nil)))
	pp.Printf("\n")
}

func FromRelay(stream network.Stream) bool {
	return strings.Contains(stream.Conn().RemoteMultiaddr().String(), "p2p-circuit")
}

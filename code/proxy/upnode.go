package proxy

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/tidwall/gjson"
)

type UpNode struct {
	webProxy   *WebProxy
	localNode  host.Host
	p2pPubId   string
	addr       *peer.AddrInfo
	sessionKey string
	mLock      sync.Mutex
}

func (wp *WebProxy) getExitNode(target string) *UpNode {

	wp.upnodeLock.Lock()
	enode, ok := wp.upNodes[target]
	wp.upnodeLock.Unlock()

	if ok {
		return enode
	}

	addr, err := wp.resolveAndConnect(target)
	if err != nil {
		pp.Println("@err_creating_upnode", err)
		return nil
	}

	enode = &UpNode{
		localNode:  wp.localNode,
		p2pPubId:   target,
		addr:       addr,
		sessionKey: "",
		mLock:      sync.Mutex{},
	}

	wp.upnodeLock.Lock()
	wp.upNodes[target] = enode
	wp.upnodeLock.Unlock()

	return enode
}

// convert pubkeyhash like 12D3KooWQbUAAEbYha8TxxsKrsxqbpY5dxPdGwcTYgSaTHAFcngE to actual connectable
// address /ip4/127.0.0.1/tcp/8083/p2p/12D3KooWQbUAAEbYha8TxxsKrsxqbpY5dxPdGwcTYgSaTHAFcngE like this and connect
func (wp *WebProxy) resolveAndConnect(target string) (*peer.AddrInfo, error) {

	addr := peer.AddrInfo{
		ID:    "",
		Addrs: make([]multiaddr.Multiaddr, 0),
	}

	for _, s := range wp.seekers {
		out, err := s.Get(target)
		if err != nil {
			continue
		}

		pp.Println("seeker_resp", out)

		maybeAddr := peer.AddrInfo{}
		err = maybeAddr.UnmarshalJSON([]byte(gjson.Get(out, "node.value").String()))
		if err != nil {
			pp.Println("@err_skipping", err)
			continue
		}

		if addr.ID == "" && strings.ToLower(maybeAddr.ID.String()) == target {
			pp.Println("@assigning_addr")
			addr.ID = maybeAddr.ID
		}

		for _, m := range maybeAddr.Addrs {
			addr.Addrs = append(addr.Addrs, m)
		}

	}

	fmt.Println("@address is |>", addr)

	err := wp.localNode.Connect(context.Background(), addr)
	if err != nil {
		pp.Println(err)
		return nil, err
	}

	return &addr, nil
}

package proxy

import (
	"context"
	"sync"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
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

	pp.Println("@FIXME")
	if enode == nil {
		panic("impl")
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
	pid, err := peer.IDFromBytes([]byte(target))
	if err != nil {
		return nil, err
	}

	addr := peer.AddrInfo{
		ID:    pid,
		Addrs: make([]multiaddr.Multiaddr, 0),
	}

	for _, s := range wp.seekers {
		out, err := s.Get(pid.String())
		if err != nil {
			continue
		}
		maybeAddr := peer.AddrInfo{}
		maybeAddr.UnmarshalJSON([]byte(out))

		if addr.ID != maybeAddr.ID {
			pp.Println("different addr", addr.ID, maybeAddr.ID)
			continue
		}

		for _, m := range maybeAddr.Addrs {
			addr.Addrs = append(addr.Addrs, m)
		}

	}

	err = wp.localNode.Connect(context.Background(), addr)
	if err != nil {
		pp.Println(err)
	}

	return &addr, nil
}

package proxy

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/k0kubun/pp"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/temphia/lpweb/code/core"
)

type UpNode struct {
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
	newnewnode, ok := wp.upNodes[target]
	if ok {
		wp.upnodeLock.Unlock()
		return newnewnode
	}

	wp.upNodes[target] = enode
	wp.upnodeLock.Unlock()

	return enode
}

// convert pubkeyhash like 12D3KooWQbUAAEbYha8TxxsKrsxqbpY5dxPdGwcTYgSaTHAFcngE to actual connectable
// address /ip4/127.0.0.1/tcp/8083/p2p/12D3KooWQbUAAEbYha8TxxsKrsxqbpY5dxPdGwcTYgSaTHAFcngE like this and connect
func (wp *WebProxy) resolveAndConnect(target string) (*peer.AddrInfo, error) {

	addr := peer.AddrInfo{
		ID:    peer.ID(target),
		Addrs: make([]multiaddr.Multiaddr, 0),
	}

	daddr, err := wp.getDHTAddrs(addr.ID)
	if err == nil {
		addr.Addrs = append(addr.Addrs, daddr.Addrs...)
	}

	addr.Addrs = removeDuplicateAddrs(daddr.Addrs)
	addr.Addrs = append(addr.Addrs, wp.constructCircuitAddr(wp.mesh.GetPossiblePeers(), target)...)

	//	pp.Println("@final_address/len", len(wp.mesh.Host.Network().ConnsToPeer(addr.ID)))
	pp.Println("@FINAL_ADDRESS")
	core.PrintPeerAddr(addr)

	err = wp.localNode.Connect(context.Background(), addr)
	if err == nil {
		curcuit := true
		for cid, rconn := range wp.mesh.Host.Network().ConnsToPeer(addr.ID) {
			pp.Println("@conn", cid, rconn.RemoteMultiaddr().String())
			if !strings.Contains(rconn.RemoteMultiaddr().String(), "p2p-circuit") {
				pp.Println("@found_direct_connection")
				curcuit = false
				break
			}
		}

		if curcuit {
			pp.Println("@trying_direct_connect")

			wticker := time.NewTimer(time.Second * 10)
			doneChan := make(chan struct{}, 1)

			go func() {
				err := wp.mesh.HolePunchService.DirectConnect(addr.ID)
				if err != nil {
					pp.Println("@err_while_direct_connect", err.Error())
				} else {
					pp.Println("@direct_connect_success ?")
				}

				doneChan <- struct{}{}
			}()

			select {
			case <-doneChan:
				wticker.Stop()

			case <-wticker.C:
				pp.Println("@connect_timeout")
			}

		} else {
			pp.Println("@skipping_direct_connection")
		}

		return &addr, nil
	}

	pp.Println("@could_not_connect", err.Error())

	return nil, err
}

func (wp *WebProxy) getDHTAddrs(pi peer.ID) (peer.AddrInfo, error) {
	return wp.mesh.DHT.FindPeer(context.Background(), pi)
}

func removeDuplicateAddrs(strSlice []multiaddr.Multiaddr) []multiaddr.Multiaddr {
	// map to store unique keys
	keys := make(map[string]bool)
	returnSlice := []multiaddr.Multiaddr{}
	for _, item := range strSlice {
		if _, value := keys[item.String()]; !value {
			keys[item.String()] = true
			returnSlice = append(returnSlice, item)
		}
	}

	return returnSlice
}

// /p2p/QmRelay/p2p-circuit/p2p/QmAlice

func (wp *WebProxy) constructCircuitAddr(relays []peer.ID, target string) []multiaddr.Multiaddr {
	// return fmt.Sprintf("/p2p/%s/p2p-circuit/p2p/%s", target, target)

	addrs := make([]multiaddr.Multiaddr, 0)

	for _, relay := range relays {
		pp.Println("@processing_relay", relay.String())

		addrInfo, err := wp.getDHTAddrs(relay)
		if err != nil || len(addrInfo.Addrs) == 0 {
			continue
		}

		for _, addr := range addrInfo.Addrs {
			ma, err := multiaddr.NewMultiaddr(fmt.Sprintf("%s/p2p/%s/p2p-circuit/p2p/%s", addr.String(), relay.String(), target))
			if err != nil {
				continue
			}
			addrs = append(addrs, ma)
		}

		if len(addrs) > 64 {
			break
		}

	}

	return addrs
}

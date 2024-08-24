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
	"github.com/tidwall/gjson"
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

		addr.Addrs = append(addr.Addrs, maybeAddr.Addrs...)
	}

	if len(addr.Addrs) == 0 {
		return nil, fmt.Errorf("no addresses found for %s", target)
	}

	pp.Println("@found_sth_from_seekers")
	core.PrintPeerAddr(addr)

	daddr, err := wp.getDHTAddrs(addr.ID)
	if err == nil {
		core.PrintPeerAddr(addr)
		// combine and deduplicate
		daddr.Addrs = append(daddr.Addrs, addr.Addrs...)
		addr.Addrs = removeDuplicateAddrs(daddr.Addrs)
	}

	pp.Println("@final_address/len", len(wp.mesh.Host.Network().ConnsToPeer(addr.ID)))
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

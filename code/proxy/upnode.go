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
	pid, err := peer.Decode(target)
	if err != nil {
		pp.Println("@decode/err", err.Error())
		return nil, err
	}

	addr := peer.AddrInfo{
		ID:    pid,
		Addrs: make([]multiaddr.Multiaddr, 0),
	}

	pp.Println("@resolveAndConnect/1", string(peer.ID(target)))

	daddr, err := wp.getDHTAddrs(addr.ID)
	if err == nil {
		addr.Addrs = append(addr.Addrs, daddr.Addrs...)
	} else {
		pp.Println("@err_getting_dht_addrs", err.Error())
	}

	addr.Addrs = removeDuplicateAddrs(daddr.Addrs)

	if len(addr.Addrs) == 0 {
		addr.Addrs = append(addr.Addrs, wp.constructCircuitAddr(wp.Mesh.GetPossiblePeers(), target)...)
	}

	//	pp.Println("@final_address/len", len(wp.mesh.Host.Network().ConnsToPeer(addr.ID)))
	pp.Println("@FINAL_ADDRESS")
	core.PrintPeerAddr(addr)

	err = wp.localNode.Connect(context.Background(), addr)
	if err == nil {
		curcuit := true
		for cid, rconn := range wp.Mesh.Host.Network().ConnsToPeer(addr.ID) {
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
				err := wp.Mesh.HolePunchService.DirectConnect(addr.ID)
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
	pp.Println(err)

	return nil, err
}

func (wp *WebProxy) getDHTAddrs(pi peer.ID) (peer.AddrInfo, error) {
	pp.Println("@getDHTAddrs/1", pi.String())
	pp.Println("@getDHTAddrs/2", pi)

	addr, err := wp.Mesh.DHT.FindPeer(context.Background(), pi)
	if err == nil {
		return addr, nil
	}

	pp.Println("@getDHTAddrs/3")

	addr.ID = pi

	// size, err := wp.Mesh.DHT.NetworkSize()
	// if err != nil {
	// 	pp.Println("@err_getting_network_size", err.Error())
	// }

	// wp.Mesh.DHT.RoutingTable().Print()

	// pp.Println("@MODE", wp.Mesh.DHT.Mode())

	// pp.Println("@D_STAT", wp.Mesh.DHT.GetRoutingTableDiversityStats())

	// pp.Println("@DHT_SIZE", size)

	pp.Println("@getDHTAddrs/4")

	altAddr := wp.Mesh.GetAltPeer(string(pi))
	if altAddr != nil {
		pp.Println("@using_alt_addr", altAddr.String())
		addr.Addrs = append(addr.Addrs, altAddr.Addrs...)
		return addr, nil
	}

	pp.Println("@getDHTAddrs/5")

	return addr, err

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

	ma, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip6/2604:a880:1:20::204:4001/udp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ/p2p-circuit/p2p/%s", target))
	if err == nil {
		addrs = append(addrs, ma)
	}

	pp.Println("err", err)

	return addrs

	// for _, relay := range relays {
	// 	pp.Println("@processing_relay", relay.String())

	// 	addrInfo, err := wp.getDHTAddrs(relay)
	// 	if err != nil || len(addrInfo.Addrs) == 0 {
	// 		continue
	// 	}

	// 	for _, addr := range addrInfo.Addrs {
	// 		ma, err := multiaddr.NewMultiaddr(fmt.Sprintf("%s/p2p/%s/p2p-circuit/p2p/%s", addr.String(), relay.String(), target))
	// 		if err != nil {
	// 			continue
	// 		}
	// 		addrs = append(addrs, ma)
	// 	}

	// 	if len(addrs) > 64 {
	// 		break
	// 	}

	// }

	// return addrs
}

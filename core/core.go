package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-tcp-transport"
	ma "github.com/multiformats/go-multiaddr"
)

var (
	BootStrapPeers = []string{
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
		"/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
		"/ip4/104.131.131.82/udp/4001/quic/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
	}
)

func NewHost(keystr string) (host.Host, *dht.IpfsDHT, error) {
	privateKey, _, err := crypto.GenerateKeyPairWithReader(1, 2048, bytes.NewReader([]byte(keystr)))
	if err != nil {
		panic(err)
	}
	return NewHostWithKey(privateKey, 0)
}

func NewHostWithKey(privateKey crypto.PrivKey, port int) (node host.Host, dhtOut *dht.IpfsDHT, err error) {
	ctx := context.Background()

	if port == 0 {
		port, err = getFreePort()
		if err != nil {
			panic(err)
		}

	}

	ip6tcp := fmt.Sprintf("/ip6/::/tcp/%d", port)
	ip4tcp := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)

	// Create libp2p node
	node, err = libp2p.New(
		libp2p.ListenAddrStrings(ip6tcp, ip4tcp),
		libp2p.Identity(privateKey),
		libp2p.DefaultSecurity,
		libp2p.NATPortMap(),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.FallbackDefaults,
		libp2p.PrivateNetwork(nil),
	)
	if err != nil {
		return
	}

	// Create DHT Subsystem
	dhtOut = dht.NewDHTClient(ctx, node, datastore.NewMapDatastore())

	// Convert Bootstap Nodes into usable addresses.
	addrs := make(map[peer.ID]*peer.AddrInfo, len(BootStrapPeers))
	for _, addrStr := range BootStrapPeers {
		addr, err := ma.NewMultiaddr(addrStr)
		if err != nil {
			return node, dhtOut, err
		}
		pii, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return node, dhtOut, err
		}
		pi, ok := addrs[pii.ID]
		if !ok {
			pi = &peer.AddrInfo{ID: pii.ID}
			addrs[pi.ID] = pi
		}
		pi.Addrs = append(pi.Addrs, pii.Addrs...)
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the
	// other nodes in the network.
	var wg sync.WaitGroup
	lock := sync.Mutex{}
	count := 0
	wg.Add(len(addrs))
	for _, peerInfo := range addrs {
		go func(peerInfo *peer.AddrInfo) {
			defer wg.Done()
			err := node.Connect(ctx, *peerInfo)
			if err == nil {
				lock.Lock()
				count++
				lock.Unlock()

			}
		}(peerInfo)
	}
	wg.Wait()

	if count < 1 {
		return node, dhtOut, errors.New("unable to bootstrap libp2p node")
	}

	return node, dhtOut, nil
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

package mesh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/k0kubun/pp"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"

	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"

	ma "github.com/multiformats/go-multiaddr"
)

const (
	ProtocolHttp  = "/lpweb/http/1.0.0"
	ProtocolHttp2 = "/lpweb/http/1.2.0"
	ProtocolWS    = "/lpweb/ws/1.0.0"
)

var (
	BootStrapPeers = []string{
		"/ip4/139.178.91.71/udp/4001/quic/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
		"/ip4/147.75.87.27/udp/4001/quic/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
		"/ip4/145.40.118.135/udp/4001/quic/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
		"/ip4/104.131.131.82/udp/4001/quic/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
		"/ip6/2604:1380:4602:5c00::3/tcp/4001/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
		"/ip4/139.178.91.71/tcp/4001/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
		"/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",

		"/dnsaddr/bootstrap.libp2p.io/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	}
)

type Mesh struct {
	Host     host.Host
	DHT      *dht.IpfsDHT
	Port     int
	PublicIp string

	ResourceManager *ResourceManager

	HolePunchService *holepunch.Service
}

func (m *Mesh) PublicMultiAddr() ([]ma.Multiaddr, error) {
	tcpaddr := fmt.Sprintf("/ip4/%s/tcp/%d", m.PublicIp, m.Port)
	qaddr := fmt.Sprintf("/ip4/%s/udp/%d/quic", m.PublicIp, m.Port)

	if strings.Contains(m.PublicIp, ":") {
		tcpaddr = fmt.Sprintf("/ip6/%s/tcp/%d", m.PublicIp, m.Port)
		qaddr = fmt.Sprintf("/ip6/%s/udp/%d/quic", m.PublicIp, m.Port)
	}

	m1, err := ma.NewMultiaddr(tcpaddr)
	if err != nil {
		return nil, err
	}

	m2, err := ma.NewMultiaddr(qaddr)
	if err != nil {
		return nil, err
	}

	return []ma.Multiaddr{m1, m2}, nil
}

func New(keystr string, port int) (*Mesh, error) {
	privateKey, _, err := crypto.GenerateKeyPairWithReader(1, 2048, bytes.NewReader([]byte(keystr)))
	if err != nil {
		panic(err)
	}

	if port == 0 {
		port, err = getFreePort()
		if err != nil {
			panic(err)
		}
	}

	baseAddrs := []string{
		fmt.Sprintf("/ip6/::/tcp/%d", port),
		fmt.Sprintf("/ip6/::/udp/%d/quic", port),
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port),
		fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", port),
	}

	// pubIp, err := findPublicIpAddr()
	// if err == nil {
	// 	pp.Println("@listening_to_public_addrs", pubIp)

	// 	baseAddrs = append(baseAddrs, fmt.Sprintf("/ip4/%s/tcp/%d", pubIp, port))
	// 	baseAddrs = append(baseAddrs, fmt.Sprintf("/ip4/%s/udp/%d/quic", pubIp, port))
	// }

	hps, dh, err := NewHostWithKey(privateKey, port, baseAddrs)
	if err != nil {
		return nil, err
	}

	host := dh.Host()

	mesh := &Mesh{
		Host: host,
		DHT:  dh,
		Port: port,
		//		PublicIp:         pubIp,
		ResourceManager:  host.Network().ResourceManager().(*ResourceManager),
		HolePunchService: hps,
	}

	return mesh, nil

}

func NewHostWithKey(privateKey crypto.PrivKey, port int, baseAddrs []string) (hps *holepunch.Service, dhtOut *dht.IpfsDHT, err error) {
	ctx := context.Background()

	// Convert Bootstap Nodes into usable addresses.
	addrs := make(map[peer.ID]*peer.AddrInfo, len(BootStrapPeers))
	for _, addrStr := range BootStrapPeers {
		addr, err := ma.NewMultiaddr(addrStr)
		if err != nil {
			pp.Println("@parsing_bootstrap_node1", addrStr)
			continue
		}

		pii, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			pp.Println("@parsing_bootstrap_node2", addrStr)

			continue
		}
		pi, ok := addrs[pii.ID]
		if !ok {
			pi = &peer.AddrInfo{ID: pii.ID}
			addrs[pi.ID] = pi
		}
		pi.Addrs = append(pi.Addrs, pii.Addrs...)
	}

	finalAddrs := make([]peer.AddrInfo, 0, len(BootStrapPeers))

	for _, addr := range addrs {
		finalAddrs = append(finalAddrs, *addr)
	}

	rm, err := NewResourceManager()
	if err != nil {
		panic(err)
	}

	// Create libp2p node
	node, err := libp2p.New(
		libp2p.EnableAutoNATv2(),
		libp2p.UserAgent("lpweb"),
		libp2p.EnableNATService(),
		libp2p.ListenAddrStrings(baseAddrs...),
		libp2p.Identity(privateKey),
		libp2p.DefaultSecurity,
		libp2p.NATPortMap(),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(quic.NewTransport),
		libp2p.EnableRelay(),
		libp2p.ResourceManager(rm),
		libp2p.ForceReachabilityPublic(),

		libp2p.PrivateNetwork(nil),

		libp2p.EnableHolePunching(holepunch.WithTracer(&tracer{}), func(s *holepunch.Service) error {
			hps = s
			return nil
		}),

		libp2p.EnableAutoRelayWithStaticRelays(finalAddrs),
		// libp2p.FallbackDefaults,
	)

	if err != nil {
		return
	}

	// Create DHT Subsystem
	dhtOut = dht.NewDHTClient(ctx, node, datastore.NewMapDatastore())

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

			if err != nil {
				pp.Println("@error_connecting_bootstrap_nodes", peerInfo.String(), err.Error())
			}

		}(peerInfo)
	}
	wg.Wait()

	if count < 1 {
		return hps, dhtOut, errors.New("unable to bootstrap libp2p node")
	}

	return hps, dhtOut, nil
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

func findPublicIpAddr() (string, error) {
	resp, err := http.Get("https://api.ipify.org/")
	if err != nil {
		return "", err
	}

	out, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(out), err
}

type tracer struct{}

func (t *tracer) Trace(evt *holepunch.Event) {
	// pp.Println("TRACER|>", evt.Peer.Loggable())
	// pp.Println("TRACER|>", evt)
}

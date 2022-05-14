package exitserver

import (
	"sync"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
)

type ExitServer struct {
	localNode  host.Host
	instance   *Instance
	p2pPubId   string
	addr       *peer.AddrInfo
	sessionKey string
	mLock      sync.Mutex
}

func (e *ExitServer) Handle() {

}

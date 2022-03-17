package p2p

import (
	"sync"

	"github.com/libp2p/go-libp2p-core/host"
)

type ExitServer struct {
	localNode     host.Host
	instance      *Instance
	ipnsHostKey   string
	p2pNodes      map[string]string
	p2pActiveNode string
	sessionKey    string
	mLock         sync.Mutex
}

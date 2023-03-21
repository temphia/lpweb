package proxy

import "github.com/temphia/lpweb/code/core"

type Proxy struct {
	mesh *core.Libp2pMesh
}

func New() *Proxy {

	mesh := core.New(core.MeshOptions{
		MeshKey:    "",
		MeshPort:   0,
		DebugPrint: true,
	})

	return &Proxy{
		mesh: mesh,
	}
}

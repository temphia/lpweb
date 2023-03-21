package proxy

type Proxy struct {
	mesh *core.Libp2pMesh
}

func New() *Proxy {

	

	mesh := core.New(core.MeshOptions{})
	


	return &Proxy{
		mesh: mesh
	}
}

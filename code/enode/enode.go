package enode

import (
	"github.com/temphia/lpweb/code/core/mesh"
	"github.com/temphia/lpweb/code/proxy"
	"github.com/temphia/lpweb/code/tunnel"
)

type EmbedLPWebNode struct {
	mesh *mesh.Mesh

	tunnel *tunnel.HttpTunnel
	proxy  *proxy.WebProxy
}

type Options struct {
	Port               int
	AllowTunnelAnyPort bool
	Key                string
}

func New(opts *Options) *EmbedLPWebNode {

	m, err := mesh.New(opts.Key, opts.Port)
	if err != nil {
		panic(err)
	}

	tun := tunnel.New(opts.Port, opts.AllowTunnelAnyPort)
	pxy := proxy.New(opts.Port)

	return &EmbedLPWebNode{
		mesh:   m,
		tunnel: tun,
		proxy:  pxy,
	}
}

// control loop

func (e *EmbedLPWebNode) Run() error {

	err := e.tunnel.Run()
	if err != nil {
		return err
	}

	err = e.proxy.Run()
	if err != nil {
		return err
	}

	// e.mesh.RunControlLoop()

	return nil
}

package main

import "github.com/temphia/lpweb/core"

func main() {
	mesh := core.New(core.MeshOptions{
		DebugPrint: true,
		HttpPort:   4000,
		MeshKey:    "aaaaaasaudiwhjduibquqbwwdinlkwqdhq wuqwbd dvwuqhd qowbdjlinkl",
		MeshPort:   8083,
	})

	mesh.DebugLoop()

}

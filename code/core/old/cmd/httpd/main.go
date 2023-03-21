package main

import "github.com/temphia/lpweb/code/core"

func main() {
	mesh := core.New(core.MeshOptions{
		DebugPrint: true,
		MeshKey:    "aaaaaasaudiwhjduibquqbwwdinlkwqdhq wuqwbd dvwuqhd qowbdjlinkl",
		MeshPort:   8083,
	})

	mesh.DebugLoop()

}

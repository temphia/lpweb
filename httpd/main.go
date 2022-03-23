package main

func main() {
	mesh := New(MeshOptions{
		debugPrint: true,
		HttpPort:   4000,
		MeshKey:    "aaaaaasaudiwhjduibquqbwwdinlkwqdhq wuqwbd dvwuqhd qowbdjlinkl",
		MeshPort:   8083,
	})

	mesh.debugLoop()

}

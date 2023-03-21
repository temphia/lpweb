package cli

import (
	"github.com/alecthomas/kong"
	"github.com/k0kubun/pp"
)

type CLI struct {
	Proxy struct {
		port string `arg:"" help:"port run proxy on"`
	} `cmd:"" help:"tunnel http port"`

	Tunnel struct {
		port string `arg:"" help:"port to tunnel"`
	} `cmd:"" help:"tunnel http port"`

	Key string
}

func RunCLI() {
	cli := &CLI{}
	ctx := kong.Parse(cli)

	switch ctx.Command() {
	case "proxy":
		pp.Println("@proxy")
	case "tunnel":
		pp.Println("@tunnel")
	default:
		panic("Not implemented" + ctx.Command())
	}

}

package cli

import (
	"github.com/alecthomas/kong"
	"github.com/k0kubun/pp"
)

type CLI struct {
	WebProxy struct {
		port string `arg:"" help:"port run proxy on"`
	} `cmd:"" help:"web proxy for forwrading local requests to target libweb server"`

	HttpTunnel struct {
		port string `arg:"" help:"port to tunnel"`
	} `cmd:"" help:"http tunnel to a http service running in local port"`

	Key string
}

func RunCLI() {
	cli := &CLI{}
	ctx := kong.Parse(cli)

	switch ctx.Command() {
	case "web-proxy":
		pp.Println("@proxy")
	case "http-tunnel":
		pp.Println("@tunnel")
	default:
		panic("Not implemented" + ctx.Command())
	}

}

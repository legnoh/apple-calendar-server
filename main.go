package main

import (
	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"
	"github.com/legnoh/apple-calendar-server/cmd"
)

var (
	version = "0.1.0"
)

type CLI struct {
	Init    cmd.InitCmd      `cmd:"" help:"Generate sample configuration file"`
	Serve   cmd.ServeCmd     `cmd:"" help:"Start the server"`
	Version kong.VersionFlag `help:"Show version" name:"version" short:"v"`
}

func main() {
	cli := &CLI{}
	configPath := kong.ExpandPath(cmd.DefaultConfigPath)

	ctx := kong.Parse(
		cli,
		kong.Vars{"version": version},
		kong.Description("apple-calendar-server"),
		kong.Configuration(kongyaml.Loader, configPath),
	)

	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

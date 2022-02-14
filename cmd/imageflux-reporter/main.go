package main

import (
	"context"
	"flag"
	"os"

	"github.com/bloom42/rz-go"
	"github.com/bloom42/rz-go/log"
	"github.com/google/subcommands"
	reporter "github.com/handlename/imageflux-reporter"
)

func main() {

	var (
		configPath string
	)

	flag.StringVar(&configPath, "config", "", "path to config yaml file")
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&reporter.RateCmd{}, "")
	flag.Parse()

	config, err := reporter.LoadConfig(configPath)
	if err != nil {
		log.Error("failed to load config", rz.Err(err))
		os.Exit(1)
	}

	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx, config)))
}

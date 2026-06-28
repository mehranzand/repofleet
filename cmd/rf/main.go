package main

import (
	"os"

	"github.com/mehranzand/repofleet/commands/root"
)

var version = "dev"

func main() {
	cmd := root.NewRootCmd(version)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

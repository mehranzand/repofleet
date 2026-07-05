package main

import (
	"fmt"
	"os"

	"github.com/mehranzand/repofleet/commands/root"
	"github.com/mehranzand/repofleet/internal/iostreams"
	"github.com/mehranzand/repofleet/internal/shell"
)

var version = "dev"

func main() {
	ensureShellIntegration()

	cmd := root.NewRootCmd(version)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func ensureShellIntegration() {
	sh := shell.Detect()
	installed, rcPath, err := shell.Install(sh)
	if err != nil || !installed {
		return
	}
	fmt.Fprintln(os.Stderr, iostreams.Success("One-time setup: shell integration added to "+rcPath)+" — run: source "+rcPath)
}

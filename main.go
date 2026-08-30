package main

import (
	"fmt"
	"os"

	"github.com/ryoukaii/spkg/internal/cli"
)

// version is overridden at build time via -ldflags "-X main.version=v1.4.0".
// The default "dev" is used for local builds that don't pass ldflags.
var version = "dev"

func main() {
	cli.SetVersion(version)
	cmd := cli.NewRootCmd()
	if err := cmd.Execute(); err != nil {
		// Cobra's usage/error output is silenced in root.go, so print a clean
		// one-line message here and exit non-zero.
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

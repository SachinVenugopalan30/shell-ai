package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/SachinVenugopalan30/shell-ai/cmd"
)

// version is injected at build time via -ldflags "-X main.version=vX.Y.Z".
// Falls back to module version from build info (set by `go install pkg@vX.Y.Z`).
var version = "dev"

func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

func printVersion() {
	fmt.Printf("shellai %s (%s, %s/%s)\n", resolveVersion(), runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

func main() {
	// handle --version / -v before cobra parses (root requires 1 arg)
	for _, a := range os.Args[1:] {
		if a == "--version" || a == "-v" {
			printVersion()
			return
		}
	}

	// version subcommand
	cmd.Root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(_ *cobra.Command, _ []string) {
			printVersion()
		},
	})

	if err := cmd.Root.Execute(); err != nil {
		os.Exit(1)
	}
}

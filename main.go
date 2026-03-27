package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/SachinVenugopalan30/shellai/cmd"
)

// version is injected at build time via -ldflags "-X main.version=vX.Y.Z"
var version = "dev"

func main() {
	// version subcommand
	cmd.Root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("shellai %s (%s, %s/%s)\n", version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		},
	})

	if err := cmd.Root.Execute(); err != nil {
		os.Exit(1)
	}
}

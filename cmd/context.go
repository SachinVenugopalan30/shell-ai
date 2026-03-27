package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	ctx "github.com/SachinVenugopalan30/shellai/internal/context"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Show detected environment (OS, shell, package manager, etc.)",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		env, err := ctx.Detect()
		if err != nil {
			return err
		}
		fmt.Print(env.String())
		return nil
	},
}

func init() {
	Root.AddCommand(contextCmd)
}

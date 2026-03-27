package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SachinVenugopalan30/shell-ai/internal/config"
	"github.com/SachinVenugopalan30/shell-ai/internal/prompt"
	"github.com/SachinVenugopalan30/shell-ai/internal/provider"
)

var explainCmd = &cobra.Command{
	Use:   "explain <command>",
	Short: "Explain what a shell command does in plain English",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		cfg, err := config.Load(config.DefaultPath())
		if err != nil {
			return err
		}
		if flagProvider != "" {
			cfg.Provider = flagProvider
		}
		if flagEndpoint != "" {
			cfg.Endpoint = flagEndpoint
		}
		if flagModel != "" {
			cfg.Model = flagModel
		}

		p, err := provider.New(cfg)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
		defer cancel()

		if err := p.Ping(ctx); err != nil {
			return err
		}

		msgs := prompt.BuildExplainMessages(args[0])
		res, err := p.Complete(ctx, msgs)
		if err != nil {
			return err
		}

		fmt.Printf("\n  %s\n  %s\n  %s\n\n", args[0], dashes(len(args[0])), res)
		return nil
	},
}

func init() {
	Root.AddCommand(explainCmd)
}

package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SachinVenugopalan30/shellai/internal/config"
	"github.com/SachinVenugopalan30/shellai/internal/provider"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available models from the current provider",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
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

		p, err := provider.New(cfg)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
		defer cancel()

		models, err := p.ListModels(ctx)
		if err != nil {
			return err
		}

		fmt.Printf("Provider: %s (%s)\n", cfg.Provider, cfg.Endpoint)
		fmt.Println(dashes(40))
		if len(models) == 0 {
			fmt.Println("No models found.")
			return nil
		}
		for i, m := range models {
			if i == 0 && cfg.Model == "" {
				fmt.Printf("• %s (default)\n", m)
			} else {
				fmt.Printf("• %s\n", m)
			}
		}
		return nil
	},
}

func init() {
	Root.AddCommand(modelsCmd)
}

func dashes(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += "─"
	}
	return s
}

package cmd

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/SachinVenugopalan30/shell-ai/internal/config"
	"github.com/SachinVenugopalan30/shell-ai/internal/history"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Reset config and wipe command history",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		cfgPath := config.DefaultPath()
		hisPath := history.DefaultPath()

		fmt.Println("\n  ⚠  This will reset all settings and clear command history.")
		fmt.Printf("    • %s\n", cfgPath)
		fmt.Printf("    • %s\n", hisPath)

		sel := promptui.Select{
			Label: "Are you sure?",
			Items: []string{"Yes, reset everything", "Abort"},
		}
		i, _, err := sel.Run()
		if err != nil || i != 0 {
			fmt.Println("Aborted.")
			return nil
		}

		os.Remove(cfgPath)
		history.Clear(hisPath)

		fmt.Println("\n  ✓ Config reset to defaults")
		fmt.Println("  ✓ History cleared")
		fmt.Println("  Done — run 'shellai set' to reconfigure.")
		return nil
	},
}

func init() {
	Root.AddCommand(clearCmd)
}

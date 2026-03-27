package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SachinVenugopalan30/shellai/internal/config"
)

var setCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Configure shellai settings",
	Long:  "Run with no args to launch the setup wizard, or pass a key and value to set a single field.",
	Args:  cobra.MaximumNArgs(2),
	RunE:  runSet,
}

func init() {
	Root.AddCommand(setCmd)
}

func runSet(_ *cobra.Command, args []string) error {
	path := config.DefaultPath()

	// direct key=value
	if len(args) == 2 {
		if err := config.SetField(path, args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("✓ %s set to %q\n", args[0], args[1])
		return nil
	}

	// interactive wizard
	cfg, err := config.Load(path)
	if err != nil {
		return err
	}

	sc := bufio.NewScanner(os.Stdin)

	fmt.Println("\n  shellai setup")
	fmt.Println("  ─────────────")

	cfg.Provider = ask(sc, fmt.Sprintf("Provider [%s]: ", cfg.Provider), cfg.Provider)
	cfg.Endpoint = ask(sc, fmt.Sprintf("Endpoint [%s]: ", cfg.Endpoint), cfg.Endpoint)
	cfg.Model = ask(sc, fmt.Sprintf("Model [%s]: ", cfg.Model), cfg.Model)

	if err := config.Save(cfg, path); err != nil {
		return err
	}
	fmt.Printf("\n  ✓ Config saved to %s\n", path)
	return nil
}

// ask prints a prompt, reads input, and returns the input (or def if empty)
func ask(sc *bufio.Scanner, prompt, def string) string {
	fmt.Print("  " + prompt)
	sc.Scan()
	v := strings.TrimSpace(sc.Text())
	if v == "" {
		return def
	}
	return v
}

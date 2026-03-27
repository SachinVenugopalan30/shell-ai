package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/SachinVenugopalan30/shell-ai/internal/config"
	ctx "github.com/SachinVenugopalan30/shell-ai/internal/context"
	"github.com/SachinVenugopalan30/shell-ai/internal/executor"
	"github.com/SachinVenugopalan30/shell-ai/internal/history"
	"github.com/SachinVenugopalan30/shell-ai/internal/prompt"
	"github.com/SachinVenugopalan30/shell-ai/internal/provider"
	"github.com/SachinVenugopalan30/shell-ai/internal/safety"
	"github.com/SachinVenugopalan30/shell-ai/internal/spinner"
)

var (
	flagModel    string
	flagProvider string
	flagEndpoint string
	flagYes      bool
)

var Root = &cobra.Command{
	Use:   "shellai",
	Short: "Translate natural language into shell commands using a local LLM",
	Args:  cobra.ExactArgs(1),
	RunE:  run,
}

func init() {
	Root.PersistentFlags().StringVar(&flagModel, "model", "", "LLM model to use")
	Root.PersistentFlags().StringVar(&flagProvider, "provider", "", "Provider: ollama or llamacpp")
	Root.PersistentFlags().StringVar(&flagEndpoint, "endpoint", "", "Provider API endpoint URL")
	Root.PersistentFlags().BoolVarP(&flagYes, "yes", "y", false, "Auto-confirm non-destructive commands")
}

func run(cmd *cobra.Command, args []string) error {
	intent := args[0]
	if len(intent) > 500 {
		return fmt.Errorf("intent too long (%d chars, max 500)", len(intent))
	}

	// load config and apply flag overrides
	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return fmt.Errorf("Error loading config: %w", err)
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

	// connect to provider
	p, err := provider.New(cfg)
	if err != nil {
		return err
	}

	bgCtx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := p.Ping(bgCtx); err != nil {
		return err
	}

	// collect environment
	env, err := ctx.Detect()
	if err != nil {
		return err
	}

	// call LLM with a spinner
	msgs := prompt.BuildMessages(env, intent)
	stop := spinner.Start("Thinking...")
	raw, err := p.Complete(bgCtx, msgs)
	stop()
	if err != nil {
		return fmt.Errorf("LLM error: %w", err)
	}

	// parse response
	res := prompt.ParseResponse(raw)
	if res.Command == "" {
		fmt.Println("Could not parse a command from the response:")
		fmt.Println(res.Raw)
		return nil
	}

	// safety checks
	check := safety.Check(res.Command, env.PackageManager)

	// print warnings
	if check.PkgManagerMismatch {
		color.Yellow("⚠  Note: %s", check.MismatchDetail)
	}
	if check.IsDestructive {
		color.Red("WARNING: potentially destructive command (%s)", check.DestructiveReason)
	}

	// display command + reason
	fmt.Printf("\n  Command:  %s\n", res.Command)
	if res.Reason != "" {
		fmt.Printf("  Reason:   %s\n\n", res.Reason)
	}

	// confirm via arrow-key selector
	if !flagYes || check.IsDestructive || isRiskyForAutoConfirm(res.Command) {
		items := []string{"Run it", "Abort"}
		if check.IsDestructive {
			items = []string{"Yes, run it (destructive)", "Abort"}
		}
		sel := promptui.Select{
			Label: "What would you like to do?",
			Items: items,
		}
		i, _, err := sel.Run()
		if err != nil || i != 0 {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// execute with a 30s timeout to prevent runaway commands
	execCtx, execCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer execCancel()
	result, _ := executor.Run(execCtx, res.Command, env.Shell)
	if execCtx.Err() == context.DeadlineExceeded {
		fmt.Fprintln(os.Stderr, "Command timed out after 30s and was stopped.")
	}

	// log to history
	history.Append(history.DefaultPath(), history.Entry{
		Timestamp: time.Now().UTC(),
		Intent:    intent,
		Command:   res.Command,
		Reason:    res.Reason,
		ExitCode:  result.ExitCode,
		Executed:  true,
	})

	// exit code 1 often means "no results" (grep, lsof, etc.) — only warn for real errors
	if result.ExitCode > 1 {
		fmt.Fprintf(os.Stderr, "Command exited with code %d\n", result.ExitCode)
	}

	return nil
}

// isRiskyForAutoConfirm returns true for commands that should always prompt even with -y
func isRiskyForAutoConfirm(cmd string) bool {
	risky := []string{"| bash", "| sh", "| zsh", "eval ", "exec ", "`", "$("}
	for _, p := range risky {
		if strings.Contains(cmd, p) {
			return true
		}
	}
	return false
}


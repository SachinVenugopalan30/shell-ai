package executor

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
)

type Result struct {
	ExitCode int
	Output   string
}

func Run(ctx context.Context, command, shell string) (*Result, error) {
	if shell == "" {
		shell = "/bin/sh"
	}

	var buf bytes.Buffer
	cmd := exec.CommandContext(ctx, shell, "-c", command)

	// stream output to terminal and capture it
	cmd.Stdout = io.MultiWriter(os.Stdout, &buf)
	cmd.Stderr = io.MultiWriter(os.Stderr, &buf)

	err := cmd.Run()
	code := 0
	if cmd.ProcessState != nil {
		code = cmd.ProcessState.ExitCode()
	}

	return &Result{ExitCode: code, Output: buf.String()}, err
}

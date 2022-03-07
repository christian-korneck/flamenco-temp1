package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"os/exec"
)

// CLIRunner is a wrapper around exec.CommandContext() to allow mocking.
type CLIRunner struct {
}

func NewCLIRunner() *CLIRunner {
	return &CLIRunner{}
}

func (cli *CLIRunner) CommandContext(ctx context.Context, name string, arg ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, arg...)
}

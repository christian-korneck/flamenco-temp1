package cli_runner

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/rs/zerolog"
)

// The buffer size used to read stdout/stderr output from subprocesses.
// Effectively this determines the maximum line length that can be handled.
const StdoutBufferSize = 40 * 1024

// CLIRunner is a wrapper around exec.CommandContext() to allow mocking.
type CLIRunner struct {
}

func NewCLIRunner() *CLIRunner {
	return &CLIRunner{}
}

func (cli *CLIRunner) CommandContext(ctx context.Context, name string, arg ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, arg...)
}

// RunWithTextOutput runs a command and sends its output line-by-line to the
// lineChannel. Stdout and stderr are combined.
func (cli *CLIRunner) RunWithTextOutput(
	ctx context.Context,
	logger zerolog.Logger,
	execCmd *exec.Cmd,
	logChunker LogChunker,
	lineChannel chan<- string,
) error {
	outPipe, err := execCmd.StdoutPipe()
	if err != nil {
		return err
	}
	execCmd.Stderr = execCmd.Stdout // Redirect stderr to stdout.

	if err := execCmd.Start(); err != nil {
		logger.Error().Err(err).Msg("error starting CLI execution")
		return err
	}

	blenderPID := execCmd.Process.Pid
	logger = logger.With().Int("pid", blenderPID).Logger()

	reader := bufio.NewReaderSize(outPipe, StdoutBufferSize)

	for {
		lineBytes, isPrefix, readErr := reader.ReadLine()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			logger.Error().Err(err).Msg("error reading stdout/err")
			return err
		}

		line := string(lineBytes)
		if isPrefix {
			logger.Warn().
				Str("line", fmt.Sprintf("%s...", line[:256])).
				Int("lineLength", len(line)).
				Msg("unexpectedly long line read, truncating")
		}

		logger.Debug().Msg(line)
		if lineChannel != nil {
			lineChannel <- line
		}

		if err := logChunker.Append(ctx, fmt.Sprintf("pid=%d > %s", blenderPID, line)); err != nil {
			return fmt.Errorf("appending log entry to log chunker: %w", err)
		}
	}

	if err := logChunker.Flush(ctx); err != nil {
		return fmt.Errorf("flushing log chunker: %w", err)
	}

	if err := execCmd.Wait(); err != nil {
		logger.Error().Err(err).Msg("error in CLI execution")
		return err
	}

	if execCmd.ProcessState.Success() {
		logger.Info().Msg("command exited succesfully")
	} else {
		logger.Error().
			Int("exitCode", execCmd.ProcessState.ExitCode()).
			Msg("command exited abnormally")
		return fmt.Errorf("command exited abnormally with code %d", execCmd.ProcessState.ExitCode())
	}

	return nil
}

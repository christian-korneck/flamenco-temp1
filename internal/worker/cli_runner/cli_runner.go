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
// Before returning. RunWithTextOutput() waits for the subprocess, to ensure it
// doesn't become defunct.
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

	// returnErr determines which error is returned to the caller. More important
	// errors overwrite less important ones. This is done via a variable instead
	// of simply returning, because the function must be run to completion in
	// order to wait for processes (and not create defunct ones).
	var returnErr error = nil
readloop:
	for {
		lineBytes, isPrefix, readErr := reader.ReadLine()

		switch {
		case readErr == io.EOF:
			break readloop
		case readErr != nil:
			logger.Error().Err(err).Msg("error reading stdout/err")
			returnErr = readErr
			break readloop
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
			returnErr = fmt.Errorf("appending log entry to log chunker: %w", err)
			break readloop
		}
	}

	if err := logChunker.Flush(ctx); err != nil {
		// any readErr is less important, as these are likely caused by other
		// issues, which will surface on the Wait() and Success() calls.
		returnErr = fmt.Errorf("flushing log chunker: %w", err)
	}

	if err := execCmd.Wait(); err != nil {
		logger.Error().
			Int("exitCode", execCmd.ProcessState.ExitCode()).
			Msg("command exited abnormally")
		returnErr = fmt.Errorf("command exited abnormally with code %d", execCmd.ProcessState.ExitCode())
	}

	if returnErr != nil {
		logger.Error().Err(err).
			Int("exitCode", execCmd.ProcessState.ExitCode()).
			Msg("command exited abnormally")
		return returnErr
	}

	logger.Info().Msg("command exited succesfully")
	return nil
}

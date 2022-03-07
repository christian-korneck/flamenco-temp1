package worker

// SPDX-License-Identifier: GPL-3.0-or-later

/* This file contains the commands in the "blender" type group. */

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/google/shlex"
	"github.com/rs/zerolog"

	"git.blender.org/flamenco/pkg/api"
)

// The buffer size used to read stdout/stderr output from Blender.
// Effectively this determines the maximum line length that can be handled.
const StdoutBufferSize = 40 * 1024

const timeFormat = time.RFC3339Nano

type BlenderParameters struct {
	exe        string   // Expansion of `{blender}`: executable path + its CLI parameters defined by the Manager.
	argsBefore []string // Additional CLI arguments defined by the job compiler script, to go before the blend file name.
	blendfile  string   // Path of the file to open.
	args       []string // Additional CLI arguments defined by the job compiler script, to go after the blend file name.
}

// cmdBlender executes the "blender-render" command.
func (ce *CommandExecutor) cmdBlenderRender(ctx context.Context, logger zerolog.Logger, taskID string, cmd api.Command) error {
	cmdCtx, cmdCtxCancel := context.WithCancel(ctx)
	defer cmdCtxCancel()

	execCmd, err := ce.cmdBlenderRenderCommand(cmdCtx, logger, taskID, cmd)
	if err != nil {
		return err
	}

	execCmd.Stderr = execCmd.Stdout // Redirect stderr to stdout.
	outPipe, err := execCmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := execCmd.Start(); err != nil {
		logger.Error().Err(err).Msg("error starting CLI execution")
		return err
	}

	blenderPID := execCmd.Process.Pid
	logger = logger.With().Int("pid", blenderPID).Logger()

	reader := bufio.NewReaderSize(outPipe, StdoutBufferSize)
	logChunker := NewLogChunker(taskID, ce.listener, ce.timeService)

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
		logChunker.Append(ctx, fmt.Sprintf("pid=%d > %s", blenderPID, line))
	}
	logChunker.Flush(ctx)

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

func (ce *CommandExecutor) cmdBlenderRenderCommand(
	ctx context.Context,
	logger zerolog.Logger,
	taskID string,
	cmd api.Command,
) (*exec.Cmd, error) {
	parameters, err := cmdBlenderRenderParams(logger, cmd)
	if err != nil {
		return nil, err
	}

	cliArgs := make([]string, 0)
	cliArgs = append(cliArgs, parameters.argsBefore...)
	cliArgs = append(cliArgs, parameters.blendfile)
	cliArgs = append(cliArgs, parameters.args...)
	execCmd := ce.cli.CommandContext(ctx, parameters.exe, cliArgs...)
	if execCmd == nil {
		logger.Error().
			Str("cmdName", cmd.Name).
			Msg("unable to create command executor")
		return nil, ErrNoExecCmd
	}
	logger.Info().
		Str("cmdName", cmd.Name).
		Str("execCmd", execCmd.String()).
		Msg("going to execute Blender")

	if err := ce.listener.LogProduced(ctx, taskID, fmt.Sprintf("going to run: %s %q", parameters.exe, cliArgs)); err != nil {
		return nil, err
	}

	return execCmd, nil
}

func cmdBlenderRenderParams(logger zerolog.Logger, cmd api.Command) (BlenderParameters, error) {
	var (
		parameters BlenderParameters
		ok         bool
	)

	if parameters.exe, ok = cmdParameter[string](cmd, "exe"); !ok || parameters.exe == "" {
		logger.Warn().Interface("command", cmd).Msg("missing 'exe' parameter")
		return parameters, fmt.Errorf("missing 'exe' parameter: %+v", cmd.Parameters)
	}
	if parameters.argsBefore, ok = cmdParameterAsStrings(cmd, "argsBefore"); !ok {
		logger.Warn().Interface("command", cmd).Msg("invalid 'argsBefore' parameter")
		return parameters, fmt.Errorf("invalid 'argsBefore' parameter: %+v", cmd.Parameters)
	}
	if parameters.blendfile, ok = cmdParameter[string](cmd, "blendfile"); !ok || parameters.blendfile == "" {
		logger.Warn().Interface("command", cmd).Msg("missing 'blendfile' parameter")
		return parameters, fmt.Errorf("missing 'blendfile' parameter: %+v", cmd.Parameters)
	}
	if parameters.args, ok = cmdParameterAsStrings(cmd, "args"); !ok {
		logger.Warn().Interface("command", cmd).Msg("invalid 'args' parameter")
		return parameters, fmt.Errorf("invalid 'args' parameter: %+v", cmd.Parameters)
	}

	// Move any CLI args from 'exe' to 'argsBefore'.
	exeArgs, err := shlex.Split(parameters.exe)
	if err != nil {
		logger.Warn().Err(err).Interface("command", cmd).Msg("error parsing 'exe' parameter with shlex")
		return parameters, fmt.Errorf("parsing 'exe' parameter %q: %w", parameters.exe, err)
	}
	if len(exeArgs) > 1 {
		allArgsBefore := []string{}
		allArgsBefore = append(allArgsBefore, exeArgs[1:]...)
		allArgsBefore = append(allArgsBefore, parameters.argsBefore...)
		parameters.exe = exeArgs[0]
		parameters.argsBefore = allArgsBefore
	}

	return parameters, nil
}

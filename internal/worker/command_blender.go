package worker

/* ***** BEGIN GPL LICENSE BLOCK *****
 *
 * Original Code Copyright (C) 2022 Blender Foundation.
 *
 * This file is part of Flamenco.
 *
 * Flamenco is free software: you can redistribute it and/or modify it under
 * the terms of the GNU General Public License as published by the Free Software
 * Foundation, either version 3 of the License, or (at your option) any later
 * version.
 *
 * Flamenco is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
 * A PARTICULAR PURPOSE.  See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * Flamenco.  If not, see <https://www.gnu.org/licenses/>.
 *
 * ***** END GPL LICENSE BLOCK ***** */

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
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
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
	logChunker := NewLogChunker(taskID, ce.listener)

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

		timestamp := time.Now().Format(timeFormat)
		// %35s because trailing zeroes in the nanoseconds aren't output by the
		// formatted timestamp, and thus it has a variable length. Using a fixed
		// width in this Sprintf() call ensures the rest of the line aligns visually
		// with the preceeding ones.
		logLine := fmt.Sprintf("%35s: pid=%d > %s", timestamp, blenderPID, line)
		logChunker.Append(ctx, logLine)
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
		return parameters, fmt.Errorf("error parsing 'exe' parameter %q: %w", parameters.exe, err)
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

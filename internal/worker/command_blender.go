package worker

// SPDX-License-Identifier: GPL-3.0-or-later

/* This file contains the commands in the "blender" type group. */

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"sync"

	"github.com/google/shlex"
	"github.com/rs/zerolog"

	"git.blender.org/flamenco/internal/find_blender"
	"git.blender.org/flamenco/pkg/api"
	"git.blender.org/flamenco/pkg/crosspath"
)

var regexpFileSaved = regexp.MustCompile("Saved: '(.*)'")

type BlenderParameters struct {
	exe        string   // Expansion of `{blender}`: executable path + its CLI parameters defined by the Manager.
	argsBefore []string // Additional CLI arguments defined by the job compiler script, to go before the blend file name.
	blendfile  string   // Path of the file to open.
	args       []string // Additional CLI arguments defined by the job compiler script, to go after the blend file name.
}

// cmdBlender executes the "blender-render" command.
func (ce *CommandExecutor) cmdBlenderRender(ctx context.Context, logger zerolog.Logger, taskID string, cmd api.Command) error {
	cmdCtx, cmdCtxCancel := context.WithCancel(ctx)
	defer cmdCtxCancel() // Ensure the subprocess exits whenever this function returns.

	execCmd, err := ce.cmdBlenderRenderCommand(cmdCtx, logger, taskID, cmd)
	if err != nil {
		return err
	}

	logChunker := NewLogChunker(taskID, ce.listener, ce.timeService)
	lineChannel := make(chan string)

	// Process the output of Blender.
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for line := range lineChannel {
			ce.processLineBlender(ctx, logger, taskID, line)
		}
	}()

	// Run the subprocess.
	subprocessErr := ce.cli.RunWithTextOutput(ctx,
		logger,
		execCmd,
		logChunker,
		lineChannel,
	)

	// Wait for the processing to stop.
	close(lineChannel)
	wg.Wait()

	if subprocessErr != nil {
		logger.Error().Err(subprocessErr).
			Int("exitCode", execCmd.ProcessState.ExitCode()).
			Msg("command exited abnormally")
		return subprocessErr
	}

	logger.Info().Msg("command exited succesfully")
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

	if crosspath.Dir(parameters.exe) == "." {
		// No directory path given. Check that the executable can be found on the
		// path.
		if _, err := exec.LookPath(parameters.exe); err != nil {
			// Attempt a platform-specific way to find which Blender executable to
			// use. If Blender cannot not be found, just use the configured command
			// and let the OS produce the errors.
			path, err := find_blender.FileAssociation()
			if err == nil {
				logger.Info().Str("path", path).Msg("found Blender")
				parameters.exe = path
			}
		}
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
		return parameters, NewParameterMissingError("exe", cmd)
	}
	if parameters.argsBefore, ok = cmdParameterAsStrings(cmd, "argsBefore"); !ok {
		logger.Warn().Interface("command", cmd).Msg("invalid 'argsBefore' parameter")
		return parameters, NewParameterInvalidError("argsBefore", cmd, "cannot convert to list of strings")
	}
	if parameters.blendfile, ok = cmdParameter[string](cmd, "blendfile"); !ok || parameters.blendfile == "" {
		logger.Warn().Interface("command", cmd).Msg("missing 'blendfile' parameter")
		return parameters, NewParameterMissingError("blendfile", cmd)
	}
	if parameters.args, ok = cmdParameterAsStrings(cmd, "args"); !ok {
		logger.Warn().Interface("command", cmd).Msg("invalid 'args' parameter")
		return parameters, NewParameterInvalidError("args", cmd, "cannot convert to list of strings")
	}

	// Move any CLI args from 'exe' to 'argsBefore'.
	exeArgs, err := shlex.Split(parameters.exe)
	if err != nil {
		logger.Warn().Err(err).Interface("command", cmd).Msg("error parsing 'exe' parameter with shlex")
		return parameters, NewParameterInvalidError("exe", cmd, err.Error())
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

func (ce *CommandExecutor) processLineBlender(ctx context.Context, logger zerolog.Logger, taskID string, line string) {
	// TODO: check for "Warning: Unable to open" and other indicators of missing
	// files. Flamenco v2 updated the task.Activity field for such situations.

	match := regexpFileSaved.FindStringSubmatch(line)
	if len(match) < 2 {
		return
	}
	filename := match[1]

	logger = logger.With().Str("outputFile", filename).Logger()
	logger.Info().Msg("output produced")

	err := ce.listener.OutputProduced(ctx, taskID, filename)
	if err != nil {
		logger.Warn().Err(err).Msg("error submitting produced output to listener")
	}
}

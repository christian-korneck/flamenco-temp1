package worker

// SPDX-License-Identifier: GPL-3.0-or-later

/* This file contains the commands in the "ffmpeg" type group. */

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"

	"github.com/google/shlex"
	"github.com/rs/zerolog"

	"git.blender.org/flamenco/pkg/api"
	"git.blender.org/flamenco/pkg/crosspath"
)

type CreateVideoParams struct {
	exe        string   // Expansion of `{ffmpeg}`: executable path + its CLI parameters defined by the Manager.
	inputGlob  string   // Glob of input files.
	outputFile string   // File to save the video to.
	argsBefore []string // Additional CLI arguments from `exe`.
	args       []string // Additional CLI arguments defined by the job compiler script, to between the input and output filenames.
}

// cmdFramesToVideo uses ffmpeg to concatenate image frames to a video file.
func (ce *CommandExecutor) cmdFramesToVideo(ctx context.Context, logger zerolog.Logger, taskID string, cmd api.Command) error {
	cmdCtx, cmdCtxCancel := context.WithCancel(ctx)
	defer cmdCtxCancel()

	execCmd, cleanup, err := ce.cmdFramesToVideoExeCommand(cmdCtx, logger, taskID, cmd)
	if err != nil {
		return err
	}
	defer cleanup()

	outPipe, err := execCmd.StdoutPipe()
	if err != nil {
		return err
	}
	execCmd.Stderr = execCmd.Stdout // Redirect stderr to stdout.

	if err := execCmd.Start(); err != nil {
		logger.Error().Err(err).Msg("error starting CLI execution")
		return err
	}

	ffmpegPID := execCmd.Process.Pid
	logger = logger.With().Int("pid", ffmpegPID).Logger()

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
		if err := logChunker.Append(ctx, fmt.Sprintf("pid=%d > %s", ffmpegPID, line)); err != nil {
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

func (ce *CommandExecutor) cmdFramesToVideoExeCommand(
	ctx context.Context,
	logger zerolog.Logger,
	taskID string,
	cmd api.Command,
) (*exec.Cmd, func(), error) {
	parameters, err := cmdFramesToVideoParams(logger, cmd)
	if err != nil {
		return nil, nil, err
	}

	inputGlobArgs, cleanup := parameters.getInputGlob()

	cliArgs := make([]string, 0)
	cliArgs = append(cliArgs, parameters.argsBefore...)
	cliArgs = append(cliArgs, inputGlobArgs...)
	cliArgs = append(cliArgs, parameters.args...)
	cliArgs = append(cliArgs, parameters.outputFile)

	execCmd := ce.cli.CommandContext(ctx, parameters.exe, cliArgs...)
	if execCmd == nil {
		logger.Error().
			Str("cmdName", cmd.Name).
			Msg("unable to create command executor")
		return nil, nil, ErrNoExecCmd
	}
	logger.Info().
		Str("cmdName", cmd.Name).
		Str("execCmd", execCmd.String()).
		Msg("going to execute FFmpeg")

	if err := ce.listener.LogProduced(ctx, taskID, fmt.Sprintf("going to run: %s %q", parameters.exe, cliArgs)); err != nil {
		return nil, nil, err
	}

	return execCmd, cleanup, nil
}

func cmdFramesToVideoParams(logger zerolog.Logger, cmd api.Command) (CreateVideoParams, error) {
	var (
		parameters CreateVideoParams
		ok         bool
	)

	if parameters.exe, ok = cmdParameter[string](cmd, "exe"); !ok || parameters.exe == "" {
		logger.Warn().Interface("command", cmd).Msg("missing 'exe' parameter")
		return parameters, fmt.Errorf("missing 'exe' parameter: %+v", cmd.Parameters)
	}
	if parameters.inputGlob, ok = cmdParameter[string](cmd, "inputGlob"); !ok || parameters.inputGlob == "" {
		logger.Warn().Interface("command", cmd).Msg("missing 'inputGlob' parameter")
		return parameters, fmt.Errorf("missing 'inputGlob' parameter: %+v", cmd.Parameters)
	}
	if parameters.outputFile, ok = cmdParameter[string](cmd, "outputFile"); !ok || parameters.outputFile == "" {
		logger.Warn().Interface("command", cmd).Msg("missing 'outputFile' parameter")
		return parameters, fmt.Errorf("missing 'outputFile' parameter: %+v", cmd.Parameters)
	}
	if parameters.argsBefore, ok = cmdParameterAsStrings(cmd, "argsBefore"); !ok {
		logger.Warn().Interface("command", cmd).Msg("invalid 'argsBefore' parameter")
		return parameters, fmt.Errorf("invalid 'argsBefore' parameter: %+v", cmd.Parameters)
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

// getInputGlob constructs CLI arguments for FFmpeg input file globbing.
// The 2nd return value is a cleanup function.
func (p *CreateVideoParams) getInputGlob() ([]string, func()) {
	if runtime.GOOS == "windows" {
		// FFMpeg on Windows doesn't support globbing, so we have to do that in Go
		// instead.
		// TODO: implement this!
		// 	index_file = self.create_index_file(input_files)
		// 	args += [
		// 			"-f",
		// 			"concat",
		// 			"-i",
		// 			index_file.as_posix(),
		// 	]
		// TODO: the returned cleanup function should delete the index file.
		panic("not implemented yet")
	}

	cliArgs := []string{
		"-pattern_type",
		"glob",
		"-i",
		crosspath.ToSlash(p.inputGlob),
	}
	cleanup := func() {}
	return cliArgs, cleanup
}

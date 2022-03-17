package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/pkg/api"
)

type CommandExecutor struct {
	cli         CommandLineRunner
	listener    CommandListener
	timeService TimeService

	// registry maps a command name to a function that runs that command.
	registry map[string]commandCallable
}

var _ CommandRunner = (*CommandExecutor)(nil)

type commandCallable func(ctx context.Context, logger zerolog.Logger, taskID string, cmd api.Command) error

// Generate mock implementation of this interface.
//go:generate go run github.com/golang/mock/mockgen -destination mocks/command_listener.gen.go -package mocks git.blender.org/flamenco/internal/worker CommandListener

// CommandListener sends the result of commands (log, output files) to the Manager.
type CommandListener interface {
	// LogProduced sends any logging to whatever service for storing logging.
	// logLines are concatenated.
	LogProduced(ctx context.Context, taskID string, logLines ...string) error
	// OutputProduced tells the Manager there has been some output (most commonly a rendered frame or video).
	OutputProduced(ctx context.Context, taskID string, outputLocation string) error
}

// TimeService is a service that operates on time.
type TimeService interface {
	After(duration time.Duration) <-chan time.Time
	Now() time.Time
}

//go:generate go run github.com/golang/mock/mockgen -destination mocks/cli_runner.gen.go -package mocks git.blender.org/flamenco/internal/worker CommandLineRunner
// CommandLineRunner is an interface around exec.CommandContext().
type CommandLineRunner interface {
	CommandContext(ctx context.Context, name string, arg ...string) *exec.Cmd
}

// ErrNoExecCmd means CommandLineRunner.CommandContext() returned nil.
// This shouldn't happen in production, but can happen in unit tests when the
// test just wants to check the CLI arguments that are supposed to be executed,
// without actually executing anything.
var ErrNoExecCmd = errors.New("no exec.Cmd could be created")

func NewCommandExecutor(cli CommandLineRunner, listener CommandListener, timeService TimeService) *CommandExecutor {
	ce := &CommandExecutor{
		cli:         cli,
		listener:    listener,
		timeService: timeService,
	}

	// Registry of supported commands. Having this as a map (instead of a big
	// switch statement) makes it possible to do things like reporting the list of
	// supported commands.
	ce.registry = map[string]commandCallable{
		// misc
		"echo":  ce.cmdEcho,
		"sleep": ce.cmdSleep,

		// blender
		"blender-render": ce.cmdBlenderRender,

		// ffmpeg
		"frames-to-video": ce.cmdFramesToVideo,

		// file-management
		"move-directory": ce.cmdMoveDirectory,
	}

	return ce
}

func (ce *CommandExecutor) Run(ctx context.Context, taskID string, cmd api.Command) error {
	logger := log.With().Str("task", string(taskID)).Str("command", cmd.Name).Logger()
	logger.Info().Interface("parameters", cmd.Parameters).Msg("running command")

	runner, ok := ce.registry[cmd.Name]
	if !ok {
		return fmt.Errorf("unknown command: %q", cmd.Name)
	}

	return runner(ctx, logger, taskID, cmd)
}

// cmdParameterAsStrings converts an array parameter ([]interface{}) to a []string slice.
// A missing parameter is ok and returned as empty slice.
func cmdParameterAsStrings(cmd api.Command, key string) ([]string, bool) {
	parameter, found := cmd.Parameters[key]
	if !found {
		return []string{}, true
	}

	if asStrSlice, ok := parameter.([]string); ok {
		return asStrSlice, true
	}

	interfSlice, ok := parameter.([]interface{})
	if !ok {
		return []string{}, false
	}

	strSlice := make([]string, len(interfSlice))
	for idx := range interfSlice {
		switch v := interfSlice[idx].(type) {
		case string:
			strSlice[idx] = v
		case fmt.Stringer:
			strSlice[idx] = v.String()
		default:
			strSlice[idx] = fmt.Sprintf("%v", v)
		}
	}
	return strSlice, true
}

// cmdParameter retrieves a single parameter of a certain type.
func cmdParameter[T any](cmd api.Command, key string) (T, bool) {
	setting, found := cmd.Parameters[key]
	if !found {
		var zeroValue T
		return zeroValue, false
	}

	value, ok := setting.(T)
	return value, ok
}

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
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

type BlenderParameters struct {
	exe        string   // Expansion of `{blender}`: executable path + its CLI parameters defined by the Manager.
	argsBefore []string // Additional CLI arguments defined by the job compiler script, to go before the blend file name.
	blendfile  string   // Path of the file to open.
	args       []string // Additional CLI arguments defined by the job compiler script, to go after the blend file name.
}

// cmdBlender executes the "blender-render" command.
func (ce *CommandExecutor) cmdBlenderRender(ctx context.Context, logger zerolog.Logger, taskID string, cmd api.Command) error {
	var (
		parameters BlenderParameters
		ok         bool
	)

	if parameters.exe, ok = cmdParameter[string](cmd, "exe"); !ok || parameters.exe == "" {
		logger.Warn().Interface("command", cmd).Msg("missing 'exe' parameter")
		return fmt.Errorf("missing 'exe' parameter: %+v", cmd.Parameters)
	}
	if parameters.argsBefore, ok = cmdParameterAsStrings(cmd, "argsBefore"); !ok {
		logger.Warn().Interface("command", cmd).Msg("invalid 'argsBefore' parameter")
		return fmt.Errorf("invalid 'argsBefore' parameter: %+v", cmd.Parameters)
	}
	if parameters.blendfile, ok = cmdParameter[string](cmd, "blendfile"); !ok || parameters.blendfile == "" {
		logger.Warn().Interface("command", cmd).Msg("missing 'blendfile' parameter")
		return fmt.Errorf("missing 'blendfile' parameter: %+v", cmd.Parameters)
	}
	if parameters.args, ok = cmdParameterAsStrings(cmd, "args"); !ok {
		logger.Warn().Interface("command", cmd).Msg("invalid 'args' parameter")
		return fmt.Errorf("invalid 'args' parameter: %+v", cmd.Parameters)
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
		return ErrNoExecCmd
	}
	logger.Info().
		Str("cmdName", cmd.Name).
		Str("execCmd", execCmd.String()).
		Msg("going to execute Blender")

	if err := ce.listener.LogProduced(ctx, taskID, fmt.Sprintf("going to run: %s", cliArgs)); err != nil {
		return err
	}
	return nil
}

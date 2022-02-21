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
	"os/exec"

	"github.com/rs/zerolog"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

type BlenderSettings struct {
	exe        string   // Expansion of `{blender}`: executable path + its CLI parameters defined by the Manager.
	argsBefore []string // Additional CLI arguments defined by the job compiler script, to go before the blend file name.
	blendfile  string   // Path of the file to open.
	args       []string // Additional CLI arguments defined by the job compiler script, to go after the blend file name.
}

// cmdBlender executes the "blender-render" command.
func (ce *CommandExecutor) cmdBlenderRender(ctx context.Context, logger zerolog.Logger, taskID string, cmd api.Command) error {
	settings := BlenderSettings{
		exe:       cmd.Parameters["exe"].(string),
		blendfile: cmd.Parameters["blendfile"].(string),
	}
	var ok bool
	if settings.exe == "" {
		logger.Warn().Interface("command", cmd).Msg("missing 'exe' setting")
		return fmt.Errorf("missing 'exe' setting: %+v", cmd)
	}
	if settings.argsBefore, ok = cmdSettingAsStrings(cmd, "argsBefore"); !ok {
		logger.Warn().Interface("command", cmd).Msg("invalid 'argsBefore' setting")
		return fmt.Errorf("invalid 'argsBefore' setting: %+v", cmd)
	}
	if settings.blendfile == "" {
		logger.Warn().Interface("command", cmd).Msg("missing 'blendfile' setting")
		return fmt.Errorf("missing 'blendfile' setting: %+v", cmd)
	}
	if settings.args, ok = cmdSettingAsStrings(cmd, "args"); !ok {
		logger.Warn().Interface("command", cmd).Msg("invalid 'args' setting")
		return fmt.Errorf("invalid 'args' setting: %+v", cmd)
	}

	cliArgs := make([]string, 0)
	cliArgs = append(cliArgs, settings.argsBefore...)
	cliArgs = append(cliArgs, settings.blendfile)
	cliArgs = append(cliArgs, settings.args...)
	execCmd := exec.CommandContext(ctx, settings.exe, cliArgs...)
	logger.Info().
		Str("cmdName", cmd.Name).
		Str("execCmd", execCmd.String()).
		Msg("going to execute Blender")

	// if err := ce.listener.LogProduced(ctx, taskID, fmt.Sprintf("echo: %q", messageStr)); err != nil {
	// 	return err
	// }
	return nil
}

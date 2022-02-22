package worker

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

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

func TestCmdBlenderSimpleCliArgs(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ce, mocks := testCommandExecutor(t, mockCtrl)

	taskID := "1d54c6fe-1242-4c8f-bd63-5a09e358d7b6"
	cmd := api.Command{
		Name: "blender",
		Parameters: map[string]interface{}{
			"exe":        "/path/to/blender",
			"argsBefore": []string{"--background"},
			"blendfile":  "file.blend",
			"args":       []string{"--render-output", "/frames"},
		},
	}

	cliArgs := []string{"--background", "file.blend", "--render-output", "/frames"}
	mocks.cli.EXPECT().CommandContext(gomock.Any(), "/path/to/blender", cliArgs).Return(nil)

	err := ce.cmdBlenderRender(context.Background(), zerolog.Nop(), taskID, cmd)
	assert.Equal(t, ErrNoExecCmd, err, "nil *exec.Cmd should result in ErrNoExecCmd")
}

func TestCmdBlenderCliArgsInExeParameter(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ce, mocks := testCommandExecutor(t, mockCtrl)

	taskID := "1d54c6fe-1242-4c8f-bd63-5a09e358d7b6"
	cmd := api.Command{
		Name: "blender",
		Parameters: map[string]interface{}{
			"exe":        "/path/to/blender --factory-startup --python-expr \"import bpy; print('hello world')\"",
			"argsBefore": []string{"-no-audio"},
			"blendfile":  "file with spaces.blend",
			"args":       []string{"--debug"},
		},
	}

	mocks.cli.EXPECT().CommandContext(gomock.Any(),
		"/path/to/blender",                 // from 'exe'
		"--factory-startup",                // from 'exe'
		"--python-expr",                    // from 'exe'
		"import bpy; print('hello world')", // from 'exe'
		"-no-audio",                        // from 'argsBefore'
		"file with spaces.blend",           // from 'blendfile'
		"--debug",                          // from 'args'
	).Return(nil)

	err := ce.cmdBlenderRender(context.Background(), zerolog.Nop(), taskID, cmd)
	assert.Equal(t, ErrNoExecCmd, err, "nil *exec.Cmd should result in ErrNoExecCmd")
}

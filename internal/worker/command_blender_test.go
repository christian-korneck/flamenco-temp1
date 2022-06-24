package worker

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/pkg/api"
)

// SPDX-License-Identifier: GPL-3.0-or-later

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

func TestProcessLineBlender(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ce, mocks := testCommandExecutor(t, mockCtrl)
	taskID := "c194ea21-1fda-46f6-bc9a-34bd302cfb19"

	// This shouldn't call anything on the mocks.
	ce.processLineBlender(ctx, log.Logger, taskID, "starting Blender")

	// This should be recognised as produced output.
	mocks.listener.EXPECT().OutputProduced(ctx, taskID, "/path/to/file.exr")
	ce.processLineBlender(ctx, log.Logger, taskID, "Saved: '/path/to/file.exr'")
}

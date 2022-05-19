// SPDX-License-Identifier: GPL-3.0-or-later
package worker

import (
	"context"
	"runtime"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/pkg/api"
)

func TestCmdFramesToVideoSimplePosix(t *testing.T) {
	// Windows and non-Windows platforms differ in how they communicate globs to FFmpeg.
	if runtime.GOOS == "windows" {
		t.Skipf("skipping POSIX test on %s", runtime.GOOS)
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ce, mocks := testCommandExecutor(t, mockCtrl)

	taskID := "1d54c6fe-1242-4c8f-bd63-5a09e358d7b6"
	cmd := api.Command{
		Name: "blender",
		Parameters: map[string]interface{}{
			"exe":        "/path/to/ffmpeg -v quiet",
			"argsBefore": []string{"-report"},
			"inputGlob":  "path/to/renders/*.png",
			"fps":        10.0,
			"args": []string{
				"-c:v", "hevc",
				"-crf", "31",
				"-vf", "pad=ceil(iw/2)*2:ceil(ih/2)*2",
			},
			"outputFile": "path/to/renders/preview.mkv",
		},
	}

	cliArgs := []string{
		"-v", "quiet", // exe
		"-report",                                              // argsBefore
		"-pattern_type", "glob", "-i", "path/to/renders/*.png", // inputGlob
		"-c:v", "hevc", "-crf", "31", "-vf", "pad=ceil(iw/2)*2:ceil(ih/2)*2", // args
		"-r", "10", // fps
		"path/to/renders/preview.mkv", // outputFile
	}
	mocks.cli.EXPECT().CommandContext(gomock.Any(), "/path/to/ffmpeg", cliArgs).Return(nil)

	err := ce.cmdFramesToVideo(context.Background(), zerolog.Nop(), taskID, cmd)
	assert.Equal(t, ErrNoExecCmd, err, "nil *exec.Cmd should result in ErrNoExecCmd")
}

func TestCmdFramesToVideoSimpleWindows(t *testing.T) {
	// Windows and non-Windows platforms differ in how they communicate globs to FFmpeg.
	if runtime.GOOS != "windows" {
		t.Skipf("skipping Windows test on %s", runtime.GOOS)
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ce, mocks := testCommandExecutor(t, mockCtrl)

	taskID := "1d54c6fe-1242-4c8f-bd63-5a09e358d7b6"
	cmd := api.Command{
		Name: "blender",
		Parameters: map[string]interface{}{
			"exe":        "/path/to/ffmpeg -v quiet",
			"argsBefore": []string{"-report"},
			// NOTE: these files MUST exist, otherwise the glob index file creation will fail.
			"inputGlob": "command_ffmpeg_test_files/*.png",
			"fps":       10.0,
			"args": []string{
				"-c:v", "hevc",
				"-crf", "31",
				"-vf", "pad=ceil(iw/2)*2:ceil(ih/2)*2",
			},
			"outputFile": "path/to/renders/preview.mkv",
		},
	}

	cliArgs := []string{
		"-v", "quiet", // exe
		"-report",                                                                              // argsBefore
		"-f", "concat", "-safe", "0", "-i", "command_ffmpeg_test_files\\ffmpeg-file-index.txt", // input glob
		"-c:v", "hevc", "-crf", "31", "-vf", "pad=ceil(iw/2)*2:ceil(ih/2)*2", // args
		"-r", "10", // fps
		"path/to/renders/preview.mkv", // outputFile
	}
	mocks.cli.EXPECT().
		CommandContext(gomock.Any(), "/path/to/ffmpeg", gomock.Any()).
		DoAndReturn(func(ctx context.Context, name string, arg ...string) error {
			assert.EqualValues(t, cliArgs, arg)
			return nil // this is the nil *exec.Cmd referenced below
		})

	err := ce.cmdFramesToVideo(context.Background(), zerolog.Nop(), taskID, cmd)
	assert.Equal(t, ErrNoExecCmd, err, "nil *exec.Cmd should result in ErrNoExecCmd")

	assert.NoFileExists(t, "command_ffmpeg_test_files/ffmpeg-file-index.txt",
		"the glob index file should have been removed")
}

package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/internal/worker/mocks"
	"git.blender.org/flamenco/pkg/api"
)

type CommandExecutorMocks struct {
	cli      *mocks.MockCommandLineRunner
	listener *mocks.MockCommandListener
	clock    *clock.Mock
}

func testCommandExecutor(t *testing.T, mockCtrl *gomock.Controller) (*CommandExecutor, *CommandExecutorMocks) {
	cli := mocks.NewMockCommandLineRunner(mockCtrl)
	listener := mocks.NewMockCommandListener(mockCtrl)
	clock := mockedClock(t)

	ce := NewCommandExecutor(cli, listener, clock)
	mocks := CommandExecutorMocks{
		cli:      cli,
		listener: listener,
		clock:    clock,
	}

	return ce, &mocks
}

func TestCmdSettingAsStrings(t *testing.T) {
	cmd := api.Command{
		Name: "test",
		Parameters: map[string]interface{}{
			"strings": []string{"a", "b"},
			"ints":    []int{3, 4},
			"floats":  []float64{0.47, 0.327},
			"mixed":   []interface{}{"a", 47, 0.327},
		},
	}

	{
		slice, ok := cmdParameterAsStrings(cmd, "strings")
		if ok {
			assert.Equal(t, []string{"a", "b"}, slice)
		} else {
			t.Error("not ok")
		}
	}
	{
		_, ok := cmdParameterAsStrings(cmd, "ints")
		assert.False(t, ok, "only []string or []interface{} are expected to work")
	}
	{
		_, ok := cmdParameterAsStrings(cmd, "floats")
		assert.False(t, ok, "only []string or []interface{} are expected to work")
	}
	{
		slice, ok := cmdParameterAsStrings(cmd, "mixed")
		if ok {
			assert.Equal(t, []string{"a", "47", "0.327"}, slice)
		} else {
			t.Error("not ok")
		}
	}
}

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

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gitlab.com/blender/flamenco-ng-poc/internal/worker/mocks"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

type MockCommandExecutor struct {
	ce       *CommandExecutor
	cli      *mocks.MockCommandLineRunner
	listener *mocks.MockCommandListener
	clock    *clock.Mock
}

func testCommandExecutor(t *testing.T, mockCtrl *gomock.Controller) *MockCommandExecutor {
	cli := mocks.NewMockCommandLineRunner(mockCtrl)
	listener := mocks.NewMockCommandListener(mockCtrl)
	clock := mockedClock(t)
	ce := NewCommandExecutor(cli, listener, clock)

	return &MockCommandExecutor{
		ce:       ce,
		cli:      cli,
		listener: listener,
		clock:    clock,
	}
}

func TestCommandEcho(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tce := testCommandExecutor(t, mockCtrl)

	ctx := context.Background()
	message := "понављај за мном"
	taskID := "90e9d656-e201-4ef0-b6b0-c80684fafa27"
	cmd := api.Command{
		Name:       "echo",
		Parameters: map[string]interface{}{"message": message},
	}

	tce.listener.EXPECT().LogProduced(gomock.Any(), taskID, "echo: \"понављај за мном\"")

	err := tce.ce.Run(ctx, taskID, cmd)
	assert.NoError(t, err)
}

func TestCommandSleep(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	tce := testCommandExecutor(t, mockCtrl)

	ctx := context.Background()
	taskID := "90e9d656-e201-4ef0-b6b0-c80684fafa27"
	cmd := api.Command{
		Name:       "sleep",
		Parameters: map[string]interface{}{"duration_in_seconds": 47},
	}

	tce.listener.EXPECT().LogProduced(gomock.Any(), taskID, "slept 47s")

	timeBefore := tce.clock.Now()

	// Run the test in a goroutine, as we also need to actually increase the
	// mocked clock at the same time; without that, the command will sleep
	// indefinitely.
	runDone := make(chan struct{})
	var err error
	go func() {
		err = tce.ce.Run(ctx, taskID, cmd)
		close(runDone)
	}()

	timeStepSize := 100 * time.Millisecond
loop:
	for {
		select {
		case <-runDone:
			break loop
		default:
			tce.clock.Add(timeStepSize)
		}
	}

	assert.NoError(t, err)
	timeAfter := tce.clock.Now()
	// Within the step size is precise enough. We're testing our implementation, not the precision of `time.After()`.
	assert.WithinDuration(t, timeBefore.Add(47*time.Second), timeAfter, timeStepSize)
}

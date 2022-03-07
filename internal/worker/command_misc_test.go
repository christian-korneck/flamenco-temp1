package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/pkg/api"
)

func TestCommandEcho(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ce, mocks := testCommandExecutor(t, mockCtrl)

	ctx := context.Background()
	message := "понављај за мном"
	taskID := "90e9d656-e201-4ef0-b6b0-c80684fafa27"
	cmd := api.Command{
		Name:       "echo",
		Parameters: map[string]interface{}{"message": message},
	}

	mocks.listener.EXPECT().LogProduced(gomock.Any(), taskID, "echo: \"понављај за мном\"")

	err := ce.Run(ctx, taskID, cmd)
	assert.NoError(t, err)
}

func TestCommandSleep(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	ce, mocks := testCommandExecutor(t, mockCtrl)

	ctx := context.Background()
	taskID := "90e9d656-e201-4ef0-b6b0-c80684fafa27"
	cmd := api.Command{
		Name:       "sleep",
		Parameters: map[string]interface{}{"duration_in_seconds": 47},
	}

	mocks.listener.EXPECT().LogProduced(gomock.Any(), taskID, "slept 47s")

	timeBefore := mocks.clock.Now()

	// Run the test in a goroutine, as we also need to actually increase the
	// mocked clock at the same time; without that, the command will sleep
	// indefinitely.
	runDone := make(chan struct{})
	var err error
	go func() {
		err = ce.Run(ctx, taskID, cmd)
		close(runDone)
	}()

	timeStepSize := 1 * time.Second
loop:
	for {
		select {
		case <-runDone:
			break loop
		default:
			mocks.clock.Add(timeStepSize)
		}
	}

	assert.NoError(t, err)
	timeAfter := mocks.clock.Now()
	// Within the step size is precise enough. We're testing our implementation, not the precision of `time.After()`.
	assert.WithinDuration(t, timeBefore.Add(47*time.Second), timeAfter, timeStepSize)
}

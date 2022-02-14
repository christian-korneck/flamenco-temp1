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
	"github.com/stretchr/testify/assert"
	"gitlab.com/blender/flamenco-ng-poc/pkg/api"
)

type mockCommandListener struct {
	log    []loggedLines
	output []producedOutput
}
type loggedLines struct {
	taskID   TaskID
	logLines []string
}
type producedOutput struct {
	taskID         TaskID
	outputLocation string
}

// LogProduced sends any logging to whatever service for storing logging.
func (ml *mockCommandListener) LogProduced(taskID TaskID, logLines ...string) error {
	ml.log = append(ml.log, loggedLines{taskID, logLines})
	return nil
}

// OutputProduced tells the Manager there has been some output (most commonly a rendered frame or video).
func (ml *mockCommandListener) OutputProduced(taskID TaskID, outputLocation string) error {
	ml.output = append(ml.output, producedOutput{taskID, outputLocation})
	return nil
}

func mockedClock(t *testing.T) *clock.Mock {
	c := clock.NewMock()
	now, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05+07:00")
	assert.NoError(t, err)
	c.Set(now)
	return c
}

func TestCommandEcho(t *testing.T) {
	l := mockCommandListener{}
	clock := mockedClock(t)
	ce := NewCommandExecutor(&l, clock)

	ctx := context.Background()
	message := "понављај за мном"
	taskID := TaskID("90e9d656-e201-4ef0-b6b0-c80684fafa27")
	cmd := api.Command{
		Name:     "echo",
		Settings: map[string]interface{}{"message": message},
	}

	err := ce.Run(ctx, taskID, cmd)
	assert.NoError(t, err)

	assert.Len(t, l.log, 1)
	assert.Equal(t, taskID, l.log[0].taskID)
	assert.Equal(t, "echo: \"понављај за мном\"", l.log[0].logLines[0])
	assert.Len(t, l.output, 0)
}

func TestCommandSleep(t *testing.T) {
	l := mockCommandListener{}
	clock := mockedClock(t)
	ce := NewCommandExecutor(&l, clock)

	ctx := context.Background()
	taskID := TaskID("90e9d656-e201-4ef0-b6b0-c80684fafa27")
	cmd := api.Command{
		Name:     "sleep",
		Settings: map[string]interface{}{"time_in_seconds": 47},
	}

	timeBefore := clock.Now()

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
			clock.Add(timeStepSize)
		}
	}

	assert.NoError(t, err)
	timeAfter := clock.Now()
	// Within the step size is precise enough. We're testing our implementation, not the precision of `time.After()`.
	assert.WithinDuration(t, timeBefore.Add(47*time.Second), timeAfter, timeStepSize)

	assert.Len(t, l.log, 1)
	assert.Equal(t, taskID, l.log[0].taskID)
	assert.Equal(t, "slept 47s", l.log[0].logLines[0])
	assert.Len(t, l.output, 0)
}

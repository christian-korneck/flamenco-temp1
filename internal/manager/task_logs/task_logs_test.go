package task_logs

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"git.blender.org/flamenco/internal/manager/task_logs/mocks"
	"github.com/benbjohnson/clock"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestLogWriting(t *testing.T) {
	s, finish, mocks := taskLogsTestFixtures(t)
	defer finish()

	// Expect broadcastst for each call to s.Write()
	mocks.broadcaster.EXPECT().BroadcastTaskLogUpdate(gomock.Any()).Times(2)

	err := s.Write(zerolog.Nop(),
		"25c5a51c-e0dd-44f7-9f87-74f3d1fbbd8c",
		"20ff9d06-53ec-4019-9e2e-1774f05f170a",
		"Ovo je priča")
	assert.NoError(t, err)

	err = s.Write(zerolog.Nop(),
		"25c5a51c-e0dd-44f7-9f87-74f3d1fbbd8c",
		"20ff9d06-53ec-4019-9e2e-1774f05f170a",
		"Ima dvije linije")
	assert.NoError(t, err)

	filename := filepath.Join(
		s.BasePath,
		"job-25c5",
		"25c5a51c-e0dd-44f7-9f87-74f3d1fbbd8c",
		"task-20ff9d06-53ec-4019-9e2e-1774f05f170a.txt")

	contents, err := ioutil.ReadFile(filename)
	assert.NoError(t, err, "the log file should exist")
	assert.Equal(t, "Ovo je priča\nIma dvije linije\n", string(contents))
}

func TestLogRotation(t *testing.T) {
	s, finish, mocks := taskLogsTestFixtures(t)
	defer finish()

	mocks.broadcaster.EXPECT().BroadcastTaskLogUpdate(gomock.Any())

	err := s.Write(zerolog.Nop(),
		"25c5a51c-e0dd-44f7-9f87-74f3d1fbbd8c",
		"20ff9d06-53ec-4019-9e2e-1774f05f170a",
		"Ovo je priča")
	assert.NoError(t, err)

	s.RotateFile(zerolog.Nop(),
		"25c5a51c-e0dd-44f7-9f87-74f3d1fbbd8c",
		"20ff9d06-53ec-4019-9e2e-1774f05f170a")

	filename := filepath.Join(
		s.BasePath,
		"job-25c5",
		"25c5a51c-e0dd-44f7-9f87-74f3d1fbbd8c",
		"task-20ff9d06-53ec-4019-9e2e-1774f05f170a.txt")
	rotatedFilename := filename + ".1"

	contents, err := ioutil.ReadFile(rotatedFilename)
	assert.NoError(t, err, "the rotated log file should exist")
	assert.Equal(t, "Ovo je priča\n", string(contents))

	_, err = os.Stat(filename)
	assert.True(t, os.IsNotExist(err))
}

func TestLogTail(t *testing.T) {
	s, finish, mocks := taskLogsTestFixtures(t)
	defer finish()

	jobID := "25c5a51c-e0dd-44f7-9f87-74f3d1fbbd8c"
	taskID := "20ff9d06-53ec-4019-9e2e-1774f05f170a"

	// Expect broadcastst for each call to s.Write()
	mocks.broadcaster.EXPECT().BroadcastTaskLogUpdate(gomock.Any()).Times(3)

	contents, err := s.Tail(jobID, taskID)
	assert.ErrorIs(t, err, os.ErrNotExist)
	assert.Equal(t, "", contents)

	err = s.Write(zerolog.Nop(), jobID, taskID, "Just a single line")
	assert.NoError(t, err)
	contents, err = s.Tail(jobID, taskID)
	assert.NoError(t, err)
	assert.Equal(t, "Just a single line\n", string(contents))

	// A short file shouldn't do any line stripping.
	err = s.Write(zerolog.Nop(), jobID, taskID, "And another line!")
	assert.NoError(t, err)
	contents, err = s.Tail(jobID, taskID)
	assert.NoError(t, err)
	assert.Equal(t, "Just a single line\nAnd another line!\n", string(contents))

	bigString := ""
	for lineNum := 1; lineNum < 1000; lineNum++ {
		bigString += fmt.Sprintf("This is line #%d\n", lineNum)
	}
	err = s.Write(zerolog.Nop(), jobID, taskID, bigString)
	assert.NoError(t, err)

	contents, err = s.Tail(jobID, taskID)
	assert.NoError(t, err)
	assert.Equal(t,
		"This is line #887\nThis is line #888\nThis is line #889\nThis is line #890\nThis is line #891\n"+
			"This is line #892\nThis is line #893\nThis is line #894\nThis is line #895\nThis is line #896\n"+
			"This is line #897\nThis is line #898\nThis is line #899\nThis is line #900\nThis is line #901\n"+
			"This is line #902\nThis is line #903\nThis is line #904\nThis is line #905\nThis is line #906\n"+
			"This is line #907\nThis is line #908\nThis is line #909\nThis is line #910\nThis is line #911\n"+
			"This is line #912\nThis is line #913\nThis is line #914\nThis is line #915\nThis is line #916\n"+
			"This is line #917\nThis is line #918\nThis is line #919\nThis is line #920\nThis is line #921\n"+
			"This is line #922\nThis is line #923\nThis is line #924\nThis is line #925\nThis is line #926\n"+
			"This is line #927\nThis is line #928\nThis is line #929\nThis is line #930\nThis is line #931\n"+
			"This is line #932\nThis is line #933\nThis is line #934\nThis is line #935\nThis is line #936\n"+
			"This is line #937\nThis is line #938\nThis is line #939\nThis is line #940\nThis is line #941\n"+
			"This is line #942\nThis is line #943\nThis is line #944\nThis is line #945\nThis is line #946\n"+
			"This is line #947\nThis is line #948\nThis is line #949\nThis is line #950\nThis is line #951\n"+
			"This is line #952\nThis is line #953\nThis is line #954\nThis is line #955\nThis is line #956\n"+
			"This is line #957\nThis is line #958\nThis is line #959\nThis is line #960\nThis is line #961\n"+
			"This is line #962\nThis is line #963\nThis is line #964\nThis is line #965\nThis is line #966\n"+
			"This is line #967\nThis is line #968\nThis is line #969\nThis is line #970\nThis is line #971\n"+
			"This is line #972\nThis is line #973\nThis is line #974\nThis is line #975\nThis is line #976\n"+
			"This is line #977\nThis is line #978\nThis is line #979\nThis is line #980\nThis is line #981\n"+
			"This is line #982\nThis is line #983\nThis is line #984\nThis is line #985\nThis is line #986\n"+
			"This is line #987\nThis is line #988\nThis is line #989\nThis is line #990\nThis is line #991\n"+
			"This is line #992\nThis is line #993\nThis is line #994\nThis is line #995\nThis is line #996\n"+
			"This is line #997\nThis is line #998\nThis is line #999\n",
		string(contents),
	)
}

func TestLogWritingParallel(t *testing.T) {
	s, finish, mocks := taskLogsTestFixtures(t)
	defer finish()

	numGoroutines := 1000 // How many goroutines run in parallel.
	runLength := 100      // How many characters are logged, per goroutine.
	wg := sync.WaitGroup{}
	wg.Add(numGoroutines)

	jobID := "6d9a05a1-261e-4f6f-93b0-8c4f6b6d500d"
	taskID := "d19888cc-c389-4a24-aebf-8458ababdb02"

	mocks.broadcaster.EXPECT().BroadcastTaskLogUpdate(gomock.Any()).Times(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		// Write lines of 100 characters to the task log. Each goroutine writes a
		// different character, starting at 'A'.
		go func(i int32) {
			defer wg.Done()

			logger := log.With().Int32("goroutine", i).Logger()
			letter := rune(int32('A') + (i % 26))
			if len(string(letter)) > 1 {
				panic("this test assumes only single-byte runes are used")
			}
			logText := strings.Repeat(string(letter), runLength)

			assert.NoError(t, s.Write(logger, jobID, taskID, logText))
		}(int32(i))
	}
	wg.Wait()

	// Test that the final log contains 1000 lines of of 100 characters, without
	// any run getting interrupted by another one.
	contents, err := os.ReadFile(s.filepath(jobID, taskID))
	assert.NoError(t, err)
	lines := strings.Split(string(contents), "\n")
	assert.Equal(t, numGoroutines+1, len(lines),
		"each goroutine should have written a single line, and the file should have a newline at the end")

	for lineIndex, line := range lines {
		if lineIndex == numGoroutines {
			assert.Empty(t, line, "the last line should be empty")
		} else {
			assert.Lenf(t, line, runLength, "each line should be %d runes long; line #%d is not", line, lineIndex)
		}
	}
}

type TaskLogsMocks struct {
	clock       *clock.Mock
	broadcaster *mocks.MockChangeBroadcaster
}

func taskLogsTestFixtures(t *testing.T) (*Storage, func(), *TaskLogsMocks) {
	mockCtrl := gomock.NewController(t)

	mocks := &TaskLogsMocks{
		clock:       clock.NewMock(),
		broadcaster: mocks.NewMockChangeBroadcaster(mockCtrl),
	}

	mockedNow, err := time.Parse(time.RFC3339, "2022-06-09T16:52:04+02:00")
	if err != nil {
		panic(err)
	}
	mocks.clock.Set(mockedNow)

	temppath, err := ioutil.TempDir("", "testlogs")
	if err != nil {
		panic(err)
	}

	// This should be called at the end of each unit test.
	finish := func() {
		os.RemoveAll(temppath)
		mockCtrl.Finish()
	}

	sm := NewStorage(temppath, mocks.clock, mocks.broadcaster)
	return sm, finish, mocks
}

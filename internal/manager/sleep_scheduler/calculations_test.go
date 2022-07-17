package sleep_scheduler

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
)

func TestCalculateNextCheck(t *testing.T) {
	_, mocks, _ := testFixtures(t)

	var sched persistence.SleepSchedule
	empty := persistence.EmptyTimeOfDay()

	// Below, S, N, and E respectively mean Start, Now, and End times.
	// Their order shows their relation to "Now". Lower-case letters mean "no value".
	// Note that N can never be before 's' or after 'e'.

	// S N E -> E
	sched = persistence.SleepSchedule{StartTime: mkToD(9, 0), EndTime: mkToD(18, 0)}
	assert.Equal(t, mocks.todayAt(18, 0), calculateNextCheck(mocks.todayAt(11, 16), &sched))

	// S E N -> end of day
	sched = persistence.SleepSchedule{StartTime: mkToD(9, 0), EndTime: mkToD(18, 0)}
	assert.Equal(t, mocks.endOfDay(), calculateNextCheck(mocks.todayAt(19, 16), &sched))

	// N S E -> S
	sched = persistence.SleepSchedule{StartTime: mkToD(9, 0), EndTime: mkToD(18, 0)}
	assert.Equal(t, mocks.todayAt(9, 0), calculateNextCheck(mocks.todayAt(8, 47), &sched))

	// s N e -> end of day
	sched = persistence.SleepSchedule{StartTime: empty, EndTime: empty}
	assert.Equal(t, mocks.endOfDay(), calculateNextCheck(mocks.todayAt(7, 47), &sched))

	// S N e -> end of day
	sched = persistence.SleepSchedule{StartTime: mkToD(9, 0), EndTime: empty}
	assert.Equal(t, mocks.endOfDay(), calculateNextCheck(mocks.todayAt(10, 47), &sched))

	// s N E -> E
	sched = persistence.SleepSchedule{StartTime: empty, EndTime: mkToD(18, 0)}
	assert.Equal(t, mocks.todayAt(18, 0), calculateNextCheck(mocks.todayAt(7, 47), &sched))
}

func TestScheduledWorkerStatus(t *testing.T) {
	_, mocks, _ := testFixtures(t)

	var sched persistence.SleepSchedule
	empty := persistence.EmptyTimeOfDay()

	// Below, S, N, and E respectively mean Start, Now, and End times.
	// Their order shows their relation to "Now". Lower-case letters mean "no value".
	// Note that N can never be before 's' or after 'e'.

	// Test time logic without any DaysOfWeek set, i.e. the scheduled times apply
	// to each day.

	// S N E -> asleep
	sched = persistence.SleepSchedule{StartTime: mkToD(9, 0), EndTime: mkToD(18, 0), IsActive: true}
	assert.Equal(t, api.WorkerStatusAsleep, scheduledWorkerStatus(mocks.todayAt(11, 16), &sched))

	// S E N -> awake
	assert.Equal(t, api.WorkerStatusAwake, scheduledWorkerStatus(mocks.todayAt(19, 16), &sched))

	// N S E -> awake
	assert.Equal(t, api.WorkerStatusAwake, scheduledWorkerStatus(mocks.todayAt(8, 47), &sched))

	// s N e -> asleep
	sched = persistence.SleepSchedule{StartTime: empty, EndTime: empty, IsActive: true}
	assert.Equal(t, api.WorkerStatusAsleep, scheduledWorkerStatus(mocks.todayAt(7, 47), &sched))

	// S N e -> asleep
	sched = persistence.SleepSchedule{StartTime: mkToD(9, 0), EndTime: empty, IsActive: true}
	assert.Equal(t, api.WorkerStatusAsleep, scheduledWorkerStatus(mocks.todayAt(10, 47), &sched))

	// s N E -> asleep
	sched = persistence.SleepSchedule{StartTime: empty, EndTime: mkToD(18, 0), IsActive: true}
	assert.Equal(t, api.WorkerStatusAsleep, scheduledWorkerStatus(mocks.todayAt(7, 47), &sched))

	// Test DaysOfWeek logic, but only with explicit start & end times. The logic
	// for missing start/end is already covered above.
	// The mocked "today" is a Tuesday.

	// S N E unmentioned day -> awake
	sched = persistence.SleepSchedule{DaysOfWeek: "mo we", StartTime: mkToD(9, 0), EndTime: mkToD(18, 0), IsActive: true}
	assert.Equal(t, api.WorkerStatusAwake, scheduledWorkerStatus(mocks.todayAt(11, 16), &sched))

	// S E N unmentioned day -> awake
	assert.Equal(t, api.WorkerStatusAwake, scheduledWorkerStatus(mocks.todayAt(19, 16), &sched))

	// N S E unmentioned day -> awake
	assert.Equal(t, api.WorkerStatusAwake, scheduledWorkerStatus(mocks.todayAt(8, 47), &sched))

	// S N E mentioned day -> asleep
	sched = persistence.SleepSchedule{DaysOfWeek: "tu th fr", StartTime: mkToD(9, 0), EndTime: mkToD(18, 0), IsActive: true}
	assert.Equal(t, api.WorkerStatusAsleep, scheduledWorkerStatus(mocks.todayAt(11, 16), &sched))

	// S E N mentioned day -> awake
	assert.Equal(t, api.WorkerStatusAwake, scheduledWorkerStatus(mocks.todayAt(19, 16), &sched))

	// N S E mentioned day -> awake
	assert.Equal(t, api.WorkerStatusAwake, scheduledWorkerStatus(mocks.todayAt(8, 47), &sched))
}

func TestCleanupDaysOfWeek(t *testing.T) {
	assert.Equal(t, "", cleanupDaysOfWeek(""))
	assert.Equal(t, "mo tu we", cleanupDaysOfWeek("mo tu we"))
	assert.Equal(t, "mo tu we", cleanupDaysOfWeek("    mo   tu we \n"))
	assert.Equal(t, "mo tu we", cleanupDaysOfWeek("monday tuesday wed"))
}

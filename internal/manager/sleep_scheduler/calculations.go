package sleep_scheduler

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"strings"
	"time"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
)

// scheduledWorkerStatus returns the expected worker status at the given date/time.
func scheduledWorkerStatus(now time.Time, sched *persistence.SleepSchedule) api.WorkerStatus {
	if sched == nil {
		// If there is no schedule at all, the worker should be awake.
		return api.WorkerStatusAwake
	}

	tod := persistence.MakeTimeOfDay(now)

	if !sched.IsActive {
		return api.WorkerStatusAwake
	}

	if sched.DaysOfWeek != "" {
		weekdayName := strings.ToLower(now.Weekday().String()[:2])
		if !strings.Contains(sched.DaysOfWeek, weekdayName) {
			// There are days configured, and today is not a sleeping day.
			return api.WorkerStatusAwake
		}
	}

	beforeStart := sched.StartTime.HasValue() && tod.IsBefore(sched.StartTime)
	afterEnd := sched.EndTime.HasValue() && !tod.IsBefore(sched.EndTime)

	if beforeStart || afterEnd {
		// Outside sleeping time.
		return api.WorkerStatusAwake
	}

	return api.WorkerStatusAsleep
}

func cleanupDaysOfWeek(daysOfWeek string) string {
	trimmed := strings.TrimSpace(daysOfWeek)
	if trimmed == "" {
		return ""
	}

	daynames := strings.Fields(trimmed)
	for idx, name := range daynames {
		daynames[idx] = strings.ToLower(strings.TrimSpace(name))[:2]
	}
	return strings.Join(daynames, " ")
}

// Return a timestamp when the next scheck for this schedule is due.
func calculateNextCheck(now time.Time, schedule *persistence.SleepSchedule) time.Time {
	// calcNext returns the given time of day on "today" if that hasn't passed
	// yet, otherwise on "tomorrow".
	calcNext := func(tod persistence.TimeOfDay) time.Time {
		nextCheck := tod.OnDate(now)
		if nextCheck.Before(now) {
			nextCheck = nextCheck.AddDate(0, 0, 1)
		}
		return nextCheck
	}

	nextChecks := []time.Time{
		// Always check at the end of the day.
		calcNext(persistence.TimeOfDay{Hour: 24, Minute: 0}),
	}

	// No start time means "start of the day", which is already covered by
	// yesterday's "end of the day" check.
	if schedule.StartTime.HasValue() {
		nextChecks = append(nextChecks, calcNext(schedule.StartTime))
	}
	// No end time means "end of the day", which is already covered by today's
	// "end of the day" check.
	if schedule.EndTime.HasValue() {
		nextChecks = append(nextChecks, calcNext(schedule.EndTime))
	}

	next := earliestTime(nextChecks)
	return next
}

func earliestTime(timestamps []time.Time) time.Time {
	earliest := timestamps[0]
	for _, timestamp := range timestamps[1:] {
		if timestamp.Before(earliest) {
			earliest = timestamp
		}
	}
	return earliest
}

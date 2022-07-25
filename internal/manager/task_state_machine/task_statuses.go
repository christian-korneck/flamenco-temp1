package task_state_machine

// SPDX-License-Identifier: GPL-3.0-or-later

import "git.blender.org/flamenco/pkg/api"

var (
	// Task statuses that always get requeued when the job is requeueing.
	nonCompletedStatuses = []api.TaskStatus{
		api.TaskStatusCanceled,
		api.TaskStatusFailed,
		api.TaskStatusPaused,
		api.TaskStatusSoftFailed,
	}

	// Workers are allowed to keep running tasks when they are in this status.
	// 'queued', 'claimed-by-manager', and 'soft-failed' aren't considered runnable,
	// as those statuses indicate the task wasn't assigned to a Worker by the scheduler.
	runnableStatuses = map[api.TaskStatus]bool{
		api.TaskStatusActive: true,
	}
)

// IsRunnableTaskStatus returns whether the given status is considered "runnable".
// In other words, workers are allowed to keep running such tasks.
func IsRunnableTaskStatus(status api.TaskStatus) bool {
	return runnableStatuses[status]
}

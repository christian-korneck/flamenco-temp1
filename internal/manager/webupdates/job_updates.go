// SPDX-License-Identifier: GPL-3.0-or-later
package webupdates

import (
	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/internal/manager/persistence"
	"git.blender.org/flamenco/pkg/api"
)

// NewJobUpdate returns a partial SocketIOJobUpdate struct for the given job.
// It only fills in the fields that represent the current state of the job. For
// example, it omits `PreviousStatus`. The ommitted fields can be filled in by
// the caller.
func NewJobUpdate(job *persistence.Job) api.SocketIOJobUpdate {
	jobUpdate := api.SocketIOJobUpdate{
		Id:       job.UUID,
		Name:     &job.Name,
		Updated:  job.UpdatedAt,
		Status:   job.Status,
		Type:     job.JobType,
		Priority: job.Priority,
	}
	return jobUpdate
}

// NewTaskUpdate returns a partial TaskUpdate struct for the given task. It only
// fills in the fields that represent the current state of the task. For
// example, it omits `PreviousStatus`. The omitted fields can be filled in by
// the caller.
//
// Assumes task.Job is not nil.
func NewTaskUpdate(task *persistence.Task) api.SocketIOTaskUpdate {
	taskUpdate := api.SocketIOTaskUpdate{
		Id:       task.UUID,
		JobId:    task.Job.UUID,
		Name:     task.Name,
		Updated:  task.UpdatedAt,
		Status:   task.Status,
		Activity: task.Activity,
	}
	return taskUpdate
}

// BroadcastJobUpdate sends the job update to clients.
func (b *BiDirComms) BroadcastJobUpdate(jobUpdate api.SocketIOJobUpdate) {
	log.Debug().Interface("jobUpdate", jobUpdate).Msg("socketIO: broadcasting job update")
	b.BroadcastTo(SocketIORoomJobs, SIOEventJobUpdate, jobUpdate)
}

// BroadcastNewJob sends a "new job" notification to clients.
// This function should be called when the job has been completely created, so
// including its tasks.
func (b *BiDirComms) BroadcastNewJob(jobUpdate api.SocketIOJobUpdate) {
	if jobUpdate.PreviousStatus != nil {
		log.Warn().Interface("jobUpdate", jobUpdate).Msg("socketIO: new jobs should not have a previous state")
		jobUpdate.PreviousStatus = nil
	}

	log.Debug().Interface("jobUpdate", jobUpdate).Msg("socketIO: broadcasting new job")
	b.BroadcastTo(SocketIORoomJobs, SIOEventJobUpdate, jobUpdate)
}

// BroadcastTaskUpdate sends the task update to clients.
func (b *BiDirComms) BroadcastTaskUpdate(taskUpdate api.SocketIOTaskUpdate) {
	log.Debug().Interface("taskUpdate", taskUpdate).Msg("socketIO: broadcasting task update")
	room := roomForJob(taskUpdate.JobId)
	b.BroadcastTo(room, SIOEventTaskUpdate, taskUpdate)
}

// roomForJob will return the SocketIO room name for the given job. Clients in
// this room will receive info scoped to this job, so for example updates to all
// tasks of this job.
//
// Note that `api.SocketIOJobUpdate`s themselves are sent to all SocketIO clients, and
// not to this room.
func roomForJob(jobUUID string) SocketIORoomName {
	return SocketIORoomName("job-" + jobUUID)
}

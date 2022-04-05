// SPDX-License-Identifier: GPL-3.0-or-later
package webupdates

import (
	"github.com/rs/zerolog/log"

	"git.blender.org/flamenco/pkg/api"
)

// BroadcastJobUpdate sends the job update to clients.
func (b *BiDirComms) BroadcastJobUpdate(jobUpdate api.JobUpdate) {
	log.Debug().Interface("jobUpdate", jobUpdate).Msg("socketIO: broadcasting job update")
	b.sockserv.BroadcastTo(string(SocketIORoomJobs), "/jobs", jobUpdate)
}

// BroadcastNewJob sends a "new job" notification to clients.
func (b *BiDirComms) BroadcastNewJob(jobUpdate api.JobUpdate) {
	if jobUpdate.PreviousStatus != nil {
		log.Warn().Interface("jobUpdate", jobUpdate).Msg("socketIO: new jobs should not have a previous state")
		jobUpdate.PreviousStatus = nil
	}

	log.Debug().Interface("jobUpdate", jobUpdate).Msg("socketIO: broadcasting new job")
	b.sockserv.BroadcastTo(string(SocketIORoomJobs), "/jobs", jobUpdate)
}

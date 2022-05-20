// SPDX-License-Identifier: GPL-3.0-or-later
package webupdates

import (
	"fmt"

	"git.blender.org/flamenco/pkg/api"
	"github.com/google/uuid"
	gosocketio "github.com/graarh/golang-socketio"
)

// Separate type aliases for room names and event types; it's otherwise too easy
// to confuse the two.
type (
	SocketIORoomName  string
	SocketIOEventType string
)

const (
	// Predefined SocketIO rooms. There will be others, but those will have a
	// dynamic name like `job-fa48930a-105c-4125-a7f7-0aa1651dcd57` and cannot be
	// listed here as constants. See `roomXXX()` functions for those.
	SocketIORoomChat SocketIORoomName = "Chat" // For chat messages.
	SocketIORoomJobs SocketIORoomName = "Jobs" // For job updates.
)

const (
	// Predefined SocketIO event types.
	SIOEventChatMessageRcv  SocketIOEventType = "/chat"         // clients send chat messages here
	SIOEventChatMessageSend SocketIOEventType = "/message"      // chat messages are broadcasted here
	SIOEventJobUpdate       SocketIOEventType = "/jobs"         // sends api.SocketIOJobUpdate
	SIOEventTaskUpdate      SocketIOEventType = "/task"         // sends api.SocketIOTaskUpdate
	SIOEventSubscription    SocketIOEventType = "/subscription" // clients send api.SocketIOSubscription
)

func (b *BiDirComms) BroadcastTo(room SocketIORoomName, eventType SocketIOEventType, payload interface{}) {
	b.sockserv.BroadcastTo(string(room), string(eventType), payload)
}

func (b *BiDirComms) registerRoomEventHandlers() {
	_ = b.sockserv.On(string(SIOEventSubscription), b.handleRoomSubscription)
}

func (b *BiDirComms) handleRoomSubscription(c *gosocketio.Channel, subs api.SocketIOSubscription) string {
	logger := sioLogger(c)
	logger = logger.With().
		Str("op", string(subs.Op)).
		Str("type", string(subs.Type)).
		Str("uuid", string(subs.Uuid)).
		Logger()

	// Make sure the UUID is actually a valid one.
	uuid, err := uuid.Parse(subs.Uuid)
	if err != nil {
		logger.Warn().Msg("socketIO: invalid UUID, ignoring subscription request")
		return "invalid UUID, ignoring request"
	}

	var sioRoom SocketIORoomName
	switch subs.Type {
	case api.SocketIOSubscriptionTypeJob:
		sioRoom = roomForJob(uuid.String())
	default:
		logger.Warn().Msg("socketIO: invalid subscription type, ignoring")
		return "invalid subscription type, ignoring request"
	}

	switch subs.Op {
	case api.SocketIOSubscriptionOperationSubscribe:
		err = c.Join(string(sioRoom))
	case api.SocketIOSubscriptionOperationUnsubscribe:
		err = c.Leave(string(sioRoom))
	default:
		logger.Warn().Msg("socketIO: invalid subscription operation, ignoring")
		return "invalid subscription operation, ignoring request"
	}

	if err != nil {
		logger.Warn().Err(err).Msg("socketIO: performing subscription operation")
		return fmt.Sprintf("unable to perform subscription operation: %v", err)
	}

	logger.Debug().Msg("socketIO: subscription")
	return "ok"
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

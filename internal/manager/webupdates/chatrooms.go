// SPDX-License-Identifier: GPL-3.0-or-later
package webupdates

type SocketIORoomName string

const (
	// Predefined SocketIO rooms.
	SocketIORoomChat SocketIORoomName = "Chat" // For chat messages.
	SocketIORoomJobs SocketIORoomName = "Jobs" // For job updates.
)

type SocketIOEventType string

const (
	// Predefined SocketIO event types.
	SIOEventChatMessageRcv  SocketIOEventType = "/chat"    // clients send messages here
	SIOEventChatMessageSend SocketIOEventType = "/message" // messages are broadcasted here
	SIOEventJobUpdate       SocketIOEventType = "/jobs"    // sends api.JobUpdate
)

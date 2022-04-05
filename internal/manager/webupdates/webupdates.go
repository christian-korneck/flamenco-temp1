// package webupdates uses SocketIO to send updates to a web client.
// SPDX-License-Identifier: GPL-3.0-or-later
package webupdates

import (
	gosocketio "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type BiDirComms struct {
	sockserv *gosocketio.Server
}

type Message struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

func New() *BiDirComms {
	return &BiDirComms{
		sockserv: socketIOServer(),
	}
}

func (b *BiDirComms) RegisterHandlers(router *echo.Echo) {
	router.Any("/socket.io/", echo.WrapHandler(b.sockserv))
}

func socketIOServer() *gosocketio.Server {
	sio := gosocketio.NewServer(transport.GetDefaultWebsocketTransport())
	log.Info().Msg("initialising SocketIO")

	// the sio.On() and c.Join() calls only return an error when there is no
	// server connected to them, but that's not possible with our setup.
	// Errors are explicitly silenced (by assigning to _) to reduce clutter.

	// socket connection
	_ = sio.On(gosocketio.OnConnection, func(c *gosocketio.Channel) {
		log.Debug().Str("clientID", c.Id()).Msg("socketIO: connected")
		_ = c.Join(string(SocketIORoomChat)) // All clients connect to the chat room.
		_ = c.Join(string(SocketIORoomJobs)) // All clients subscribe to job updates.
	})

	// socket disconnection
	_ = sio.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel) {
		log.Debug().Str("clientID", c.Id()).Msg("socketIO: disconnected")
	})

	_ = sio.On(gosocketio.OnError, func(c *gosocketio.Channel) {
		log.Warn().Interface("c", c).Msg("socketIO: socketio error")
	})

	// chat socket
	_ = sio.On(string(SIOEventChatMessageRcv), func(c *gosocketio.Channel, message Message) string {
		log.Info().Str("clientID", c.Id()).
			Str("text", message.Text).
			Str("name", message.Name).
			Msg("socketIO: message received")
		c.BroadcastTo(string(SocketIORoomChat), string(SIOEventChatMessageSend), message)
		return "message sent successfully."
	})

	return sio
}

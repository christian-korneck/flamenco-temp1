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

	// socket connection
	sio.On(gosocketio.OnConnection, func(c *gosocketio.Channel) {
		log.Info().Str("clientID", c.Id()).Msg("connected")
		c.Join("Room")
	})

	// socket disconnection
	sio.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel) {
		log.Info().Str("clientID", c.Id()).Msg("disconnected")
		c.Leave("Room")
	})

	sio.On(gosocketio.OnError, func(c *gosocketio.Channel) {
		log.Warn().Interface("c", c).Msg("socketio error")
	})

	// chat socket
	sio.On("/chat", func(c *gosocketio.Channel, message Message) string {
		log.Info().Str("clientID", c.Id()).
			Str("text", message.Text).
			Str("name", message.Name).
			Msg("message received")
		c.BroadcastTo("Room", "/message", message.Text)
		return "message sent successfully."
	})

	return sio
}

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

	var err error

	// socket connection
	err = sio.On(gosocketio.OnConnection, func(c *gosocketio.Channel) {
		log.Debug().Str("clientID", c.Id()).Msg("socketIO: connected")
		if err := c.Join(string(SocketIORoomChat)); err != nil {
			log.Warn().Err(err).Str("clientID", c.Id()).Msg("socketIO: unable to make client join broadcast message room")
		}
	})
	if err != nil {
		log.Error().Err(err).Msg("socketIO: unable to register OnConnection handler")
	}

	// socket disconnection
	err = sio.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel) {
		log.Debug().Str("clientID", c.Id()).Msg("socketIO: disconnected")
		if err := c.Leave(string(SocketIORoomChat)); err != nil {
			log.Warn().Err(err).Str("clientID", c.Id()).Msg("socketIO: unable to make client leave broadcast message room")
		}
	})
	if err != nil {
		log.Error().Err(err).Msg("socketIO: unable to register OnDisconnection handler")
	}

	err = sio.On(gosocketio.OnError, func(c *gosocketio.Channel) {
		log.Warn().Interface("c", c).Msg("socketIO: socketio error")
	})
	if err != nil {
		log.Error().Err(err).Msg("socketIO: unable to register OnError handler")
	}

	// chat socket
	err = sio.On(string(SIOEventChatMessageRcv), func(c *gosocketio.Channel, message Message) string {
		log.Info().Str("clientID", c.Id()).
			Str("text", message.Text).
			Str("name", message.Name).
			Msg("socketIO: message received")
		c.BroadcastTo(string(SocketIORoomChat), string(SIOEventChatMessageSend), message)
		return "message sent successfully."
	})
	if err != nil {
		log.Error().Err(err).Msg("socketIO: unable to register /chat handler")
	}

	return sio
}

// SPDX-License-Identifier: GPL-3.0-or-later
package webupdates

import gosocketio "github.com/graarh/golang-socketio"

func (b *BiDirComms) registerChatEventHandlers() {
	_ = b.sockserv.On(string(SIOEventChatMessageRcv),
		func(c *gosocketio.Channel, message Message) string {
			logger := sioLogger(c)
			logger.Info().
				Str("text", message.Text).
				Str("name", message.Name).
				Msg("socketIO: message received")
			b.BroadcastTo(SocketIORoomChat, SIOEventChatMessageSend, message)
			return "message sent successfully."
		})
}

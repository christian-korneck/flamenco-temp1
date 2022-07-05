---
title: SocketIO
weight: 50
---

[SocketIO v2](https://socket.io/docs/v2/) is used for sending updates from
Flamenco Manager to the web frontend. Version 2 of the protocol was chosen,
because that has a mature Go server implementation readily available.

SocketIO messages have an *event name* and *room name*.

- **Web interface clients** send messages to the server with just an *event
  name*. These are received in handlers set up by
  `internal/manager/webupdates/webupdates.go`, function
  `registerSIOEventHandlers()`.
- **Manager** typically sends to all clients in a specific *room*. Which client
  has joined which room is determined by the Manager as well. By default every
  client joins the "job updates" and "chat" rooms. This is done in the
  `OnConnection` handler defined in `registerSIOEventHandlers()`.
- Received messages (regardless of by whom) are handled based only on their
  *event name*. The *room name* only determines *which* client receives those
  messages.

package main

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/ziflex/lecho/v3"

	gosocketio "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"

	"git.blender.org/flamenco/internal/appinfo"
)

type Message struct {
	Name string `json:"name"`
	Text string `json:"text"`
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

func addRoutes(router *echo.Echo, server *gosocketio.Server) {
	cors := middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:8080/", "http://localhost:8081/"},

		// List taken from https://www.bacancytechnology.com/blog/real-time-chat-application-using-socketio-golang-vuejs/
		AllowHeaders: []string{
			echo.HeaderAccept,
			echo.HeaderAcceptEncoding,
			echo.HeaderAccessControlAllowOrigin,
			echo.HeaderAccessControlRequestHeaders,
			echo.HeaderAccessControlRequestMethod,
			echo.HeaderAuthorization,
			echo.HeaderContentLength,
			echo.HeaderContentType,
			echo.HeaderOrigin,
			echo.HeaderXCSRFToken,
			echo.HeaderXRequestedWith,
			"Cache-Control",
			"Connection",
			"Host",
			"Referer",
			"User-Agent",
			"X-header",
		},
		AllowMethods: []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
	})

	router.Any("/socket.io/", echo.WrapHandler(server), cors)

}

func main() {
	output := zerolog.ConsoleWriter{Out: colorable.NewColorableStdout(), TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)
	log.Info().Str("version", appinfo.ApplicationVersion).Msgf("starting Socket.IO PoC %v", appinfo.ApplicationName)

	socketio := socketIOServer()

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Hook Zerolog onto Echo:
	e.Use(lecho.Middleware(lecho.Config{
		Logger: lecho.From(log.Logger),
	}))

	// Ensure panics when serving a web request won't bring down the server.
	e.Use(middleware.Recover())

	addRoutes(e, socketio)

	listen := ":8081"
	log.Info().Str("listen", listen).Msg("server starting")
	log.Info().Msg("Run `yarn serve` from the 'web' dir to run the frontend server")

	// Start the web server.
	finalErr := e.Start(listen)
	log.Warn().Err(finalErr).Msg("shutting down")
}

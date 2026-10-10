package config

import (
	"fmt"
	"net/http"
	"time"

	applogger "github.com/damolaoladipo/pricebase/internal/common/logger"
	"github.com/damolaoladipo/pricebase/internal/utils"
	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	engineio "github.com/zishang520/socket.io/servers/engine/v3"
	socketio "github.com/zishang520/socket.io/servers/socket/v3"
	"github.com/zishang520/socket.io/v3/pkg/types"
	"go.uber.org/zap"
)

// SocketServer is the process-wide Socket.IO handle, set by StartServer.
// Services emit through this to push events to connected clients.
var SocketServer *socketio.Server

// StartServer wraps a fully-built Gin engine in a WebServer that Socket.IO can
// own, mounts Socket.IO at /socket/, and binds the port.
//
// The engine must already have all middleware and routes registered, since
// NewWebServer captures it as the fallback handler. Listen does not block; it
// returns the underlying *http.Server.
func StartServer(engine *gin.Engine, env *utils.ENV) *http.Server {
	defer func() {
		if recovered := recover(); recovered != nil {
			utils.Log.Error("Server StartServer() method error:")
			applogger.SystemLogger.Error("server StartServer() panic", zap.Any("panic", recovered))
			panic(recovered)
		}
	}()

	// Gin becomes the fallback handler; Socket.IO claims /socket/.
	server := types.NewWebServer(engine)

	opts := socketio.DefaultServerOptions()
	opts.SetPath("/socket/")
	opts.SetAllowEIO3(true)
	opts.SetPingTimeout(20 * time.Second)
	opts.SetTransports(types.NewSet(engineio.WebSocket))

	// Socket.IO validates the handshake origin itself, independent of the Gin
	// CORS middleware. Tighten this alongside middlewares.CORS.
	opts.SetCors(&types.Cors{
		Origin:  "*",
		Methods: []string{"GET", "POST", "OPTIONS", "PUT", "PATCH", "DELETE"},
	})

	ioServer := socketio.NewServer(server, opts)
	ioServer.On("connection", func(clients ...any) {
		so := clients[0].(*socketio.Socket)
		so.On("disconnect", func(...any) {
			so.Emit("disconnected")
		})
	})
	SocketServer = ioServer

	addr := fmt.Sprintf(":%d", env.APP_PORT)

	return server.Listen(addr, func() {
		applogger.SystemLogger.Info("server listening",
			zap.String("env", env.APP_ENV),
			zap.Int("port", env.APP_PORT),
		)
		color.New(color.FgYellow, color.Bold).
			Printf("Server running in %s on port %d\n", env.APP_ENV, env.APP_PORT)
	})
}

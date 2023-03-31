package http

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/logger/accesslog"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	hzero "github.com/hertz-contrib/logger/zerolog"
	"github.com/weimob-tech/go-project-base/pkg/config"
	"github.com/weimob-tech/go-project-boot/pkg/wlog"
)

type hertzServer struct {
	*server.Hertz
}

func (s *hertzServer) Start() {
	s.Spin()
}

func (s *hertzServer) GetServer() any {
	return s.Hertz
}

func NewHertzServer() *hertzServer {
	// server default
	config.SetDefault("server.readTimeout", 3*time.Minute)
	config.SetDefault("server.writeTimeout", 3*time.Minute)
	config.SetDefault("server.address", ":8888")
	h := server.New(
		server.WithReadTimeout(config.GetDuration("server.readTimeout")),
		server.WithWriteTimeout(config.GetDuration("server.writeTimeout")),
		server.WithHostPorts(config.GetString("server.address")),
	)

	// setup hertz logger
	lvl := hlog.LevelInfo
	debug := config.Debug("log")
	if debug {
		lvl = hlog.LevelDebug
	}
	logger := wlog.Logger.With().CallerWithSkipFrameCount(5).Logger()
	hlog.SetLogger(hzero.From(logger, hzero.WithLevel(lvl)))

	// inject zero-logger contextual
	h.Use(func(c context.Context, ctx *app.RequestContext) {
		ctx.Next(logger.WithContext(c))
	})

	// setup access logger
	if config.GetBool("server.access-log") {
		h.Use(accesslog.New(
			accesslog.WithAccessLogFunc(func(c context.Context, f string, v ...interface{}) {
				hlog.Infof(f, v...)
			}),
		))
	}

	// default with recovery
	h.Use(recovery.Recovery())

	h.GET("ping", pong)
	h.GET("health", pong)
	h.GET("healthcheck", pong)
	h.GET("health_check", pong)

	return &hertzServer{h}
}

const pongResult = "pong"

func pong(c context.Context, ctx *app.RequestContext) {
	ctx.String(consts.StatusOK, pongResult)
}

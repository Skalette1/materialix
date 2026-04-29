package root

import (
	"context"
	"net/http"
	"time"

	"github.com/DeanThompson/ginpprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	apmgin "go.elastic.co/apm/module/apmgin/v2"
	"go.elastic.co/apm/v2"
)

type Config struct {
	Addr              string
	ReleaseID         string
	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
}

type Dependencies struct {
	Tracer *apm.Tracer

	RequesterLogger gin.HandlerFunc
	Recovery        gin.HandlerFunc

	DBPing func(context.Context) error
}

const apiV1Group = "/api/v1"

func NewHTTPServer(deps *Dependencies) *http.Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	if deps.Tracer != nil {
		engine.Use(apmgin.Middleware(engine, apmginWithTracer(deps.Tracer)))
	}

	if deps.RequestLogger != nil {
		engine.Use(deps.RequestLogger)
	}
	if deps.Recovery != nil {
		engine.Use(deps.Recovery)
	} else {
		engine.Use(gin.Recovery())
	}

	initServiceRoutes(engine, cfg, deps)
	initAPIRoutes(engine, deps)

	return &http.Server{
		Addr:        cfg.Addr,
		Handler:     engine,
		IdleTimeout: cfg.IdleTimeout,
		ReadTimeout: cfg.ReadHeaderTimeout}
}

func initServiceRoutes(engine *gin.Engine, cfg Config, deps Dependencies) {
	ginpprof.Wrap(engine)

	engine.Handle(http.MethodGet, "/metrics", func(c *gin.Context) {
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	})

	engine.Handle(http.MethodGet, "/health", func(c *gin.Context) {
		options := []healthcheck.Option{
			healthcheck.WithReleaseID(cfg.ReleaseID),
		}
		if deps.DBPing != nil {
			options = append(options,
				healthcheck.WithChecker("postgres:connections", healthcheck.CheckerFunc(
					func(ctx context.Context) error { return deps.DBPing(ctx) },
				)))
		}
		healthcheck.Handler(options...).ServeHTTP(c.Writer, c.Request)
	})
}

func InitApiRoutes(engine *gin.Engine, deps Dependencies) {
	api := engine.Group(apiV1Group)
	api.Use(deps.Auth)
}

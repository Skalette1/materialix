package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"volt/config"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

const serviceName = "materialix"

var releaseID = "dev"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	appLogger, err := logging.NewLogger("info", serviceName, releaseID)
	if err != nil {
		zl, _ := zap.NewProduction()
		zl.Fatal("failed to init logger", zap.Error(err))
	}
	log := appLogger.GetZapLogger()
	defer log.Sync()

	cfg, err := config.NewConfigFromEnv()
	if err != nil {
		log.Fatal("config load failed", zap.Error(err))
	}

	cfg.ReleaseID = releaseID
	cfg.ServiceName = serviceName

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		log.Fatal("config validation failed", zap.Error(err))
	}

	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	db, err := postgres.Connect(dbCtx, cfg.Database.DSN)
	if err != nil {
		log.Fatal("db connection failed", zap.Error(err))
	}

	server := NewHTTPServer(cfg, db, log)

	go func() {
		if err := server.Start(); err != nil {
			log.Fatal("server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	log.Info("shutting down...")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := server.Stop(shutdownCtx); err != nil {
		log.Error("server shutdown failed", zap.Error(err))
	}

	db.Close()

	log.Info("service stopped")
}

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	server "github.com/mateusfdl/go-api/adapters/http"
	"github.com/mateusfdl/go-api/adapters/logger"
	"github.com/mateusfdl/go-api/adapters/mongo"
	"github.com/mateusfdl/go-api/config"
	"github.com/mateusfdl/go-api/internal/crops"
	"github.com/mateusfdl/go-api/internal/farms"
	"github.com/mateusfdl/go-api/internal/health"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	ctx := context.Background()
	c, err := config.NewAppConfig()
	if err != nil {
		panic(err)
	}

	// Adapter modules
	l := logger.New(c.Logger)
	db := mongo.New(ctx, l, c.Mongo)
	s := server.New(l, c.HTTP)

	healthModule := health.New(s, l)
	cropsModule := crops.New(db.DB)
	farmsModule := farms.New(l, &cropsModule.Repository, s, db.DB)

	// Bootstrapping
	mongo.HookOnStart(ctx, db, l)

	server.RegisterRoutes(
		healthModule.Controller,
		farmsModule.Controller,
	)

	go s.Listen()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	shutdown(signalChan, ctx, s, db, l)
}

func shutdown(
	signalChan chan os.Signal,
	ctx context.Context,
	s *server.HTTP,
	db *mongo.Mongo,
	l *logger.Logger,
) {
	<-signalChan
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	l.Warn("Gracefully shutting down...")
	s.GracefulShutdown(shutdownCtx)
	mongo.GracefulShutdown(shutdownCtx, db, l)
	l.Info("Shutdown complete.")
	os.Exit(0)
}

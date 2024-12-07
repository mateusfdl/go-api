package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	api "github.com/mateusfdl/go-api/adapters/http"
	"github.com/mateusfdl/go-api/adapters/logger"
	"github.com/mateusfdl/go-api/adapters/mongo"
	"github.com/mateusfdl/go-api/config"
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
	logger := logger.New(&c.Logger)
	mongodb := mongo.New(ctx, logger, &c.Mongo)
	httpServer := api.New(logger, &c.HTTP)

	healthModule := health.New(httpServer, logger)

	// Bootstrapping
	mongo.HealthCheckConnection(ctx, mongodb, logger)

	api.RegisterRoutes(healthModule.Controller)

	go httpServer.Listen()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	shutdown(signalChan, ctx, httpServer, mongodb, logger)
}

func shutdown(
	signalChan chan os.Signal,
	ctx context.Context,
	httpServer *api.HTTP,
	mongodb *mongo.Mongo,
	logger *logger.Logger,
) {
	<-signalChan
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	logger.Warn("Gracefully shutting down...")
	httpServer.GracefulShutdown(shutdownCtx)
	mongo.GracefulShutdown(shutdownCtx, mongodb, logger)
	logger.Info("Shutdown complete.")
	os.Exit(0)
}

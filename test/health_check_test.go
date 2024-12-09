package test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/joho/godotenv"
	"github.com/mateusfdl/go-api/adapters/http"
	"github.com/mateusfdl/go-api/adapters/logger"
	"github.com/mateusfdl/go-api/adapters/mongo"
	"github.com/mateusfdl/go-api/config"
	"github.com/mateusfdl/go-api/internal/health"
)

func TestHealthCheck(t *testing.T) {
	if err := godotenv.Load(".env-test"); err != nil {
		panic(err)
	}

	ctx := context.Background()
	c, err := config.NewAppConfig()
	if err != nil {
		panic(err)
	}

	l := logger.New(&c.Logger)
	mongodb := mongo.New(ctx, l, &c.Mongo)
	httpServer := http.New(l, &c.HTTP)

	healthModule := health.New(httpServer, l)

	mongo.HookOnStart(ctx, mongodb, l)

	http.RegisterRoutes(healthModule.Controller)

	defer func() {
		mongo.GracefulShutdown(ctx, mongodb, l)
		httpServer.GracefulShutdown(ctx)
	}()

	go httpServer.Listen()
	t.Run("HealthCheck", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)

		w := httptest.NewRecorder()

		httpServer.Router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status code 200, but got %d", w.Code)
		}

		if w.Body.String() != "OK" {
			t.Errorf("Expected body 'OK', but got '%s'", w.Body.String())
		}
	})
}

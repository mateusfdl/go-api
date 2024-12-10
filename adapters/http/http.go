package http

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/mateusfdl/go-api/adapters/logger"
)

type HTTP struct {
	Port    int
	Timeout int
	Router  *mux.Router
	l       *logger.Logger
	Server  *http.Server
}

func New(l *logger.Logger, cfg Config) *HTTP {
	router := mux.NewRouter()
	return &HTTP{
		Port:    cfg.Port,
		Timeout: cfg.Timeout,
		Router:  router,
		Server: &http.Server{
			Addr:         ":" + strconv.Itoa(cfg.Port),
			Handler:      router,
			ReadTimeout:  time.Duration(cfg.Timeout) * time.Second,
			WriteTimeout: time.Duration(cfg.Timeout) * time.Second,
			IdleTimeout:  time.Duration(cfg.Timeout) * time.Second,
		},
		l: l,
	}
}

// Starts the HTTP server
func (h *HTTP) Listen() {
	h.Router.Use(h.defaultMiddleware)
	h.l.Info("Starting server on port " + strconv.Itoa(h.Port))

	if err := h.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		h.l.Error("Server failed to start: ", err)
	}
}

// stops the HTTP server gracefully
func (h *HTTP) GracefulShutdown(ctx context.Context) {
	if err := h.Server.Shutdown(ctx); err != nil {
		h.l.Error("Error during server shutdown: ", err)
	} else {
		h.l.Info("Server gracefully stopped")
	}
}

// DefaultMiddleware logs all incoming requests
func (h *HTTP) defaultMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.l.Info("Request received", "method", r.Method, "path", r.URL.Path, "query", r.URL.Query())
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")

		next.ServeHTTP(w, r)
	})
}

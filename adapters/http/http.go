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
	router  *mux.Router
	logger  *logger.Logger
	server  *http.Server
}

func New(logger *logger.Logger, cfg *Config) *HTTP {
	router := mux.NewRouter()
	return &HTTP{
		Port:    cfg.Port,
		Timeout: cfg.Timeout,
		router:  router,
		logger:  logger,
		server: &http.Server{
			Addr:         ":" + strconv.Itoa(cfg.Port),
			Handler:      router,
			ReadTimeout:  time.Duration(cfg.Timeout) * time.Second,
			WriteTimeout: time.Duration(cfg.Timeout) * time.Second,
			IdleTimeout:  time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

func (h *HTTP) Listen() {
	h.router.Use(h.defaultMiddleware)
	h.logger.Info("Starting server on port " + strconv.Itoa(h.Port))

	if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		h.logger.Error("Server failed to start: ", err)
	}
}

func (h *HTTP) GracefulShutdown(ctx context.Context) {
	if err := h.server.Shutdown(ctx); err != nil {
		h.logger.Error("Error during server shutdown: ", err)
	} else {
		h.logger.Info("Server gracefully stopped")
	}
}

// RegisterHandler registers a route handler with an optional middleware chain
func (h *HTTP) RegisterHandler(path string, handler http.HandlerFunc, middlewares ...mux.MiddlewareFunc) {
	h.logger.Info("Registering handler for path " + path)
	route := h.router.HandleFunc(path, handler)
	for _, middleware := range middlewares {
		route.Handler(middleware(route.GetHandler()))
	}
}

// DefaultMiddleware logs all incoming requests
func (h *HTTP) defaultMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.logger.Info("Request received: " + r.Method + " " + r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")

		next.ServeHTTP(w, r)
	})
}

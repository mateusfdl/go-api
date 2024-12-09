package health

import (
	"net/http"

	http_adapter "github.com/mateusfdl/go-api/adapters/http"
	"github.com/mateusfdl/go-api/adapters/logger"
)

type Controller struct {
	h *http_adapter.HTTP
	l *logger.Logger
}

func NewController(h *http_adapter.HTTP, l *logger.Logger) *Controller {
	return &Controller{h: h, l: l}
}

func (c *Controller) RegisterRoutes() {
	c.l.Info("Registering health routes")
	c.h.Router.HandleFunc("/health", c.HealthCheck).Methods("GET").Name("HealthCheck")
}

func (c *Controller) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		c.l.Error("Failed to write response")
	}
}

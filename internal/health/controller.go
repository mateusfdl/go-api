package health

import (
	"net/http"

	http_adapter "github.com/mateusfdl/go-api/adapters/http"
	"github.com/mateusfdl/go-api/adapters/logger"
)

type Controller struct {
	httpServer *http_adapter.HTTP
	logger     *logger.Logger
}

func NewController(httpServer *http_adapter.HTTP, logger *logger.Logger) *Controller {
	return &Controller{
		httpServer: httpServer,
		logger:     logger,
	}
}

func (c *Controller) RegisterRoutes() {
	c.logger.Info("Registering health routes")
	c.httpServer.Router.HandleFunc("/health", c.HealthCheck).Methods("GET").Name("HealthCheck")
}

func (c *Controller) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		c.logger.Error("Failed to write response")
	}
}

package health

import (
	"github.com/mateusfdl/go-api/adapters/http"
	"github.com/mateusfdl/go-api/adapters/logger"
)

type HealthCheckModule struct {
	Controller *Controller
	logger     *logger.Logger
}

func New(h *http.HTTP, l *logger.Logger) *HealthCheckModule {
	return &HealthCheckModule{Controller: NewController(h, l), logger: l}
}

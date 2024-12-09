package farms

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	http_adapter "github.com/mateusfdl/go-api/adapters/http"
	"github.com/mateusfdl/go-api/adapters/logger"
)

type Controller struct {
	farmService *Service
	l           *logger.Logger
	h           *http_adapter.HTTP
}

func NewController(h *http_adapter.HTTP, farmService *Service, logger *logger.Logger) *Controller {
	return &Controller{farmService: farmService, l: logger, h: h}
}

func (c *Controller) RegisterRoutes() {
	c.l.Info("Registering farm routes")
	c.h.Router.HandleFunc("/farms", c.CreateFarm).Methods("POST").Name("CreateFarm")
	c.h.Router.HandleFunc("/farms", c.ListFarms).Methods("GET").Name("ListFarms").Queries("skip", "{skip}", "limit", "{limit}")
}

func (c *Controller) CreateFarm(w http.ResponseWriter, r *http.Request) {
	var dto CreateFarmDTO
	err := json.NewDecoder(r.Body).Decode(&dto)
	if err != nil {
		c.l.Error("Failed to decode request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := c.farmService.CreateFarm(r.Context(), &dto)
	if errors.Is(err, ErrInvalidFarmFields) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err != nil {
		c.l.Error("Failed to create farm", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(`{"id": "` + id + `"}`))
	if err != nil {
		c.l.Error("Failed to write response")
	}
}

func (c *Controller) ListFarms(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()
	skip := query.Get("skip")
	limit := query.Get("limit")

	if skip == "" || limit == "" {
		c.l.Error("Invalid query parameters")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	skipInt, err := strconv.Atoi(skip)
	if err != nil {
		c.l.Error("Failed to parse skip", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.l.Error("Failed to parse limit", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dto := ListFarmQuery{
		Skip:  skipInt,
		Limit: limitInt,
	}

	farms, err := c.farmService.ListFarms(r.Context(), &dto)
	if err != nil {
		c.l.Error("Failed to list farms", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(farms)
	if err != nil {
		c.l.Error("Failed to marshal response", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(response)
	if err != nil {
		c.l.Error("Failed to write response", err)
	}

	w.WriteHeader(http.StatusOK)
}

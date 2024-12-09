package farms

import (
	"github.com/mateusfdl/go-api/adapters/http"
	"github.com/mateusfdl/go-api/adapters/logger"
	"github.com/mateusfdl/go-api/internal/crops"
	"go.mongodb.org/mongo-driver/mongo"
)

type FarmModule struct {
	Repo       Repository
	Service    *Service
	Controller *Controller
}

func New(
	l *logger.Logger,
	cropRepo *crops.Repository,
	h *http.HTTP,
	db *mongo.Database,

) *FarmModule {
	r := NewMongoRepository(db, l)
	s := NewService(l, r, cropRepo)
	c := NewController(h, s, l)
	return &FarmModule{Repo: r, Service: s, Controller: c}
}

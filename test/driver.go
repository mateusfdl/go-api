package test

import (
	"context"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/joho/godotenv"
	http_adapter "github.com/mateusfdl/go-api/adapters/http"
	"github.com/mateusfdl/go-api/adapters/logger"
	"github.com/mateusfdl/go-api/adapters/mongo"
	"github.com/mateusfdl/go-api/config"
	"github.com/mateusfdl/go-api/internal/crops"
	"github.com/mateusfdl/go-api/internal/farms"
	"go.mongodb.org/mongo-driver/bson"
)

type Driver struct {
	Server *http_adapter.HTTP
	Mongo  *mongo.Mongo
	Logger *logger.Logger
	ctx    context.Context
}

func NewDriver() *Driver {
	if err := godotenv.Load(".env-test"); err != nil {
		panic(err)
	}

	ctx := context.Background()
	c, err := config.NewAppConfig()
	if err != nil {
		panic(err)
	}
	l := logger.New(c.Logger)
	db := mongo.New(ctx, l, c.Mongo)
	h := http_adapter.New(l, c.HTTP)

	return &Driver{Server: h, Mongo: db, Logger: l, ctx: ctx}
}

func (s *Driver) Start() {
	cropsModule := crops.New(s.Mongo.DB)
	farmsModule := farms.New(s.Logger, &cropsModule.Repository, s.Server, s.Mongo.DB)

	mongo.HookOnStart(s.ctx, s.Mongo, s.Logger)

	http_adapter.RegisterRoutes(farmsModule.Controller)
	go s.Server.Listen()
}

func (s *Driver) Close() {
	mongo.GracefulShutdown(s.ctx, s.Mongo, s.Logger)
	s.Server.GracefulShutdown(s.ctx)
}

func (s *Driver) PerformRequest(method, path string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	w := httptest.NewRecorder()
	s.Server.Router.ServeHTTP(w, req)
	return w
}

func (s *Driver) WipeCollections(t *testing.T, collectionNames ...string) {
	for _, collectionName := range collectionNames {
		_, err := s.Mongo.DB.Collection(collectionName).DeleteMany(context.Background(), bson.M{})
		if err != nil {
			t.Errorf("Failed to wipe collection %s", collectionName)
		}
	}
}

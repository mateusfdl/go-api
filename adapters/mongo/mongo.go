package mongo

import (
	"context"

	"github.com/mateusfdl/go-api/adapters/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	DB *mongo.Database
}

func New(ctx context.Context, logger *logger.Logger, cfg *Config) *Mongo {
	// "mongodb://127.0.0.1:27017/rewards-poc?authSource=admin&retryWrites=true&w=majority"
	clientOptions := options.Client().ApplyURI(cfg.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.Error("failed to connect to mongo")
	}

	db := client.Database("rewards-poc")
	return &Mongo{DB: db}
}

func HealthCheckConnection(ctx context.Context, client *Mongo, logger *logger.Logger) {
	logger.Info("Health checking mongo connection")
	err := client.DB.Client().Ping(ctx, nil)
	if err != nil {
		logger.Error("Mongo is dead", err)
		panic(err)
	}

	logger.Info("Mongo is alive")
}

func GracefulShutdown(ctx context.Context, client *Mongo, logger *logger.Logger) {
	if err := client.DB.Client().Disconnect(ctx); err != nil {
		logger.Error("Failed to disconnect from mongo")
		return
	}

	logger.Info("Mongo gracefully disconnected")
}

package mongo

import (
	"context"

	"github.com/mateusfdl/go-api/adapters/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	DB *mongo.Database
}

func New(ctx context.Context, l *logger.Logger, cfg Config) *Mongo {
	clientOptions := options.Client().ApplyURI(cfg.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		l.Error("failed to connect to mongo")
	}

	db := client.Database(cfg.DBName)
	return &Mongo{DB: db}
}

func GracefulShutdown(ctx context.Context, client *Mongo, l *logger.Logger) {
	if err := client.DB.Client().Disconnect(ctx); err != nil {
		l.Error("Failed to disconnect from mongo")
		return
	}

	l.Info("Mongo gracefully disconnected")
}

func HookOnStart(ctx context.Context, c *Mongo, l *logger.Logger) {
	err := healthCheckConnection(ctx, c, l)
	if err != nil {
		panic(err)
	}
	err = syncIndexes(ctx, c, l)
	if err != nil {
		panic(err)
	}
}

func healthCheckConnection(ctx context.Context, c *Mongo, l *logger.Logger) error {
	l.Info("Health checking mongo connection")
	err := c.DB.Client().Ping(ctx, nil)
	if err != nil {
		l.Error("Mongo is dead", err)
		return err
	}

	err = syncIndexes(ctx, c, l)
	if err != nil {
		return err
	}

	l.Info("Mongo is alive")
	return nil
}

func syncIndexes(ctx context.Context, c *Mongo, l *logger.Logger) error {
	l.Info("Syncing indexes")
	_, err := c.DB.Collection("farms").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "name", Value: "text"}}, Options: options.Index()},
		{Keys: bson.D{{Key: "name", Value: 1}}, Options: options.Index()},
	})
	if err != nil {
		l.Error("Failed to create farms index", err)
		return err
	}

	_, err = c.DB.Collection("crops").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "type", Value: 1}}, Options: options.Index()},
		{Keys: bson.D{{Key: "farmId", Value: 1}}, Options: options.Index()},
	})

	if err != nil {
		l.Error("Failed to create crop index", err)
		return err
	}

	return nil
}

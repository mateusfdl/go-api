package crops

import (
	"context"

	"github.com/mateusfdl/go-api/adapters/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoRepository struct {
	db     *mongo.Database
	logger *logger.Logger
}

func NewMongoRepository(db *mongo.Database, l *logger.Logger) *MongoRepository {
	return &MongoRepository{db: db}
}

func (r *MongoRepository) CreateMany(
	ctx context.Context,
	farmId string,
	dto *[]CreateCropDTO,
) error {
	oid, err := primitive.ObjectIDFromHex(farmId)
	if err != nil {
		return err
	}
	docs := make([]interface{}, len(*dto))
	for i, d := range *dto {
		d.FarmID = oid
		docs[i] = d.ToMap()
	}

	_, err = r.db.Collection("crops").InsertMany(ctx, docs)
	if err != nil {
		return err
	}

	return nil
}

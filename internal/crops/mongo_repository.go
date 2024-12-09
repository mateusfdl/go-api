package crops

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoRepository struct {
	db *mongo.Database
}

func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{db: db}
}

// Bulk insert crops
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

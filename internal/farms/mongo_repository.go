package farms

import (
	"context"
	"time"

	"github.com/mateusfdl/go-api/adapters/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoRepository struct {
	db *mongo.Database
	l  *logger.Logger
}

func NewMongoRepository(db *mongo.Database, l *logger.Logger) *MongoRepository {
	return &MongoRepository{db: db, l: l}
}

func (r *MongoRepository) Create(
	ctx context.Context,
	dto *CreateFarmDTO,
) (string, error) {
	fields := dto.ToMap()

	doc, err := r.db.Collection("farms").InsertOne(ctx, fields)
	if err != nil {
		if ok := mongo.IsDuplicateKeyError(err); ok {
			return "", ErrFarmAlreadyExists
		}

		return "", err
	}

	oid, ok := doc.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", ErrOnConvertObjectID
	}

	return oid.Hex(), nil
}

func (r *MongoRepository) List(
	ctx context.Context,
	filter *ListFarmQuery,
) ([]Farm, error) {
	pipeline := mongo.Pipeline{}

	pipeline = append(pipeline, bson.D{
		{Key: "$lookup", Value: bson.M{
			"from":         "crops",
			"localField":   "_id",
			"foreignField": "farmId",
			"as":           "crops",
		}},
	})

	if filter.CropType != nil || filter.LandArea != nil {
		matchStage := bson.M{}

		if filter.CropType != nil {
			matchStage["crops.type"] = bson.M{"$in": filter.CropType}
		}

		if filter.LandArea != nil {
			matchStage["landArea"] = bson.M{"$gte": filter.LandArea}
		}

		pipeline = append(pipeline, bson.D{{Key: "$match", Value: matchStage}})
	}

	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: bson.M{"createdAt": -1}}})
	pipeline = append(pipeline, bson.D{{Key: "$skip", Value: filter.Skip}})
	pipeline = append(pipeline, bson.D{{Key: "$limit", Value: filter.Limit}})

	cursor, err := r.db.Collection("farms").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	var farms []Farm
	if err := cursor.All(ctx, &farms); err != nil {
		r.l.Error("error on listing farms", err)
		return nil, err
	}

	return farms, nil
}

func (r *MongoRepository) GetByID(
	ctx context.Context,
	farmId string,
) (*Farm, error) {

	oid, err := primitive.ObjectIDFromHex(farmId)

	if err != nil {
		r.l.Error("error on convert object id", err)
		return nil, ErrOnConvertObjectID
	}

	var farm Farm

	err = r.db.Collection("farms").FindOne(ctx, bson.M{"_id": oid}).Decode(&farm)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrFarmNotFound
		}

		r.l.Error("error on get farm by id", err)
		return nil, err
	}

	return &farm, nil
}

func (r *MongoRepository) Update(ctx context.Context, farmId string, dto *UpdateFarmDTO) (string, error) {
	fields := dto.ToMap()
	fields["updatedAt"] = time.Now()

	oid, err := primitive.ObjectIDFromHex(farmId)
	if err != nil {
		r.l.Error("error on convert object id", err)
		return "", ErrOnConvertObjectID
	}

	filter := bson.M{"_id": oid}
	update := bson.M{"$set": fields}

	_, err = r.db.Collection("farms").UpdateOne(ctx, filter, update)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			r.l.Error("farm not found", err)
			return "", ErrFarmNotFound
		}

		return "", err
	}

	return oid.Hex(), nil
}

func (r *MongoRepository) Delete(
	ctx context.Context,
	farmId string,
) error {
	oid, err := primitive.ObjectIDFromHex(farmId)
	if err != nil {
		r.l.Error("error on convert object id", err)
		return ErrOnConvertObjectID
	}

	_, err = r.db.Collection("farms").DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		r.l.Error("error on delete farm", err)
		return err
	}

	return nil
}

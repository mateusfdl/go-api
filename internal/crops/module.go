package crops

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type CropsModule struct {
	Repository Repository
}

func New(db *mongo.Database) *CropsModule {
	return &CropsModule{Repository: NewMongoRepository(db)}
}

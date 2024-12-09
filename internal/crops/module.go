package crops

import (
	"github.com/mateusfdl/go-api/adapters/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

type CropsModule struct {
	Repository Repository
}

func New(
	db *mongo.Database,
	l *logger.Logger,
) *CropsModule {
	return &CropsModule{Repository: NewMongoRepository(db, l)}
}

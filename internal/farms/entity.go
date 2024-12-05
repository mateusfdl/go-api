package farms

import (
	"time"
)

type Farm struct {
	ID                int       `bson:"_id"`
	Name              string    `bson:"name"`
	Address           string    `bson:"address"`
	LandArea          int64     `bson:"landArea"`
	UnitOfMeasurement string    `bson:"unitOfMeasurement"`
	CreatedAt         time.Time `bson:"createdAt"`
	UpdatedAt         time.Time `bson:"updatedAt"`
}

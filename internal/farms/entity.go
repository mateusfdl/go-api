package farms

import (
	"time"

	"github.com/mateusfdl/go-api/internal/crops"
)

type Farm struct {
	ID                string       `bson:"_id"`
	Name              string       `bson:"name"`
	Address           string       `bson:"address"`
	LandArea          int64        `bson:"landArea"`
	UnitOfMeasurement string       `bson:"unitOfMeasurement"`
	Crops             []crops.Crop `bson:"crops"`
	CreatedAt         time.Time    `bson:"createdAt"`
	UpdatedAt         time.Time    `bson:"updatedAt"`
}

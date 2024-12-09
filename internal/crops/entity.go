package crops

import "time"

const (
	CropTypeCorn    = "CORN"
	CropTypeSoybean = "SOYBEANS"
	CropTypeCoffee  = "COFFEE"
	CropTypeRice    = "RICE"
	CropTypeBeans   = "BEANS"
)

type CropType string

var CropTypes = []CropType{
	CropTypeCorn,
	CropTypeSoybean,
	CropTypeCoffee,
	CropTypeRice,
	CropTypeBeans,
}

type Crop struct {
	ID          string    `bson:"_id"`
	FarmID      string    `bson:"farmId"`
	Type        CropType  `bson:"type"`
	IsIrrigated bool      `bson:"isIrrigated"`
	IsInsured   bool      `bson:"isInsured"`
	CreatedAt   time.Time `bson:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt"`
}

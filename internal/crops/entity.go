package crops

import "time"

type Crop struct {
	ID          int       `bson:"_id"`
	Type        string    `bson:"type"`
	IsIrrigated bool      `bson:"is_irrigated"`
	IsInsured   bool      `bson:"is_insured"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

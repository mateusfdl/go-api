package crops

import "go.mongodb.org/mongo-driver/bson/primitive"

type CreateCropDTO struct {
	Type        CropType `json:"type"`
	IsIrrigated bool     `json:"isIrrigated"`
	IsInsured   bool     `json:"isInsured"`
	FarmID      primitive.ObjectID
}

func (d *CreateCropDTO) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"type":        d.Type,
		"isIrrigated": d.IsIrrigated,
		"isInsured":   d.IsInsured,
		"farmId":      d.FarmID,
	}
}

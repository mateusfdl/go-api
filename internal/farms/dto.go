package farms

import "github.com/mateusfdl/go-api/internal/crops"

type UpdateFarmDTO struct {
	Name              string `json:"name"`
	Address           string `json:"address"`
	LandArea          int64  `json:"landArea"`
	UnitOfMeasurement string `json:"unitOfMeasurement"`
}

type CreateFarmDTO struct {
	Name              string                 `json:"name"`
	Address           string                 `json:"address"`
	LandArea          int64                  `json:"landArea"`
	UnitOfMeasurement string                 `json:"unitOfMeasurement"`
	Crops             *[]crops.CreateCropDTO `json:"crops"`
}

type ListFarmQuery struct {
	Skip     int             `json:"skip"`
	Limit    int             `json:"limit"`
	LandArea *int64          `json:"landArea"`
	CropType *crops.CropType `json:"cropType"`
}

func (dto *UpdateFarmDTO) ToMap() map[string]interface{} {
	m := make(map[string]interface{})

	if dto.Name != "" {
		m["name"] = dto.Name
	}

	if dto.Address != "" {
		m["address"] = dto.Address
	}

	if dto.LandArea != 0 {
		m["landArea"] = dto.LandArea
	}

	if dto.UnitOfMeasurement != "" {
		m["unitOfMeasurement"] = dto.UnitOfMeasurement
	}

	return m
}

func (dto *CreateFarmDTO) ToMap() map[string]interface{} {
	m := make(map[string]interface{})

	m["name"] = dto.Name
	m["address"] = dto.Address
	m["landArea"] = dto.LandArea
	m["unitOfMeasurement"] = dto.UnitOfMeasurement

	return m
}

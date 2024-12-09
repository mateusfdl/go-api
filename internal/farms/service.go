package farms

import (
	"context"
	"errors"

	"github.com/mateusfdl/go-api/adapters/logger"
	"github.com/mateusfdl/go-api/internal/crops"
)

type Service struct {
	logger         *logger.Logger
	farmRepository Repository
	cropRepository crops.Repository
}

func NewService(logger *logger.Logger, farmRepo Repository, cropRepo *crops.Repository) *Service {
	return &Service{logger: logger, farmRepository: farmRepo, cropRepository: *cropRepo}
}

func (s *Service) CreateFarm(ctx context.Context, dto *CreateFarmDTO) (string, error) {
	if err := validateFields(dto); err != nil {
		return "", ErrInvalidFarmFields
	}

	id, err := s.farmRepository.Create(ctx, dto)
	if err != nil {
		return "", err
	}

	if len(*dto.Crops) == 0 {
		return id, nil
	}

	err = s.cropRepository.CreateMany(ctx, id, dto.Crops)
	if err != nil {
		return id, errors.New("failed to bulk persist crops")
	}

	return id, nil
}

func validateFields(dto *CreateFarmDTO) error {
	if dto.Name == "" {
		return errors.New("name is required")
	}

	if dto.LandArea == 0 {
		return errors.New("land area is required")
	}

	if dto.UnitOfMeasurement == "" {
		return errors.New("unit of measurement is required")
	}

	if dto.Address == "" {
		return errors.New("address is required")
	}

	if dto.Crops != nil && len(*dto.Crops) > 0 {
		for _, crop := range *dto.Crops {
			if crop.Type == "" {
				return errors.New("crop type is required")
			}

			for _, cType := range crops.CropTypes {
				if crop.Type == cType {
					return nil
				}
			}

			return errors.New("invalid crop type")
		}
	}

	return nil
}

func (s *Service) ListFarms(ctx context.Context, f *ListFarmQuery) ([]Farm, error) {
	return s.farmRepository.List(ctx, f)
}

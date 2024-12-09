package crops

import "context"

type Repository interface {
	CreateMany(ctx context.Context, farmId string, dtos *[]CreateCropDTO) error
}

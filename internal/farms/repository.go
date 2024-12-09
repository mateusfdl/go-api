package farms

import "context"

type Repository interface {
	Create(ctx context.Context, dto *CreateFarmDTO) (string, error)
	List(ctx context.Context, filter *ListFarmQuery) ([]Farm, error)
	GetByID(ctx context.Context, id string) (*Farm, error)
	Update(ctx context.Context, id string, dto *UpdateFarmDTO) (string, error)
	Delete(ctx context.Context, id string) error
}

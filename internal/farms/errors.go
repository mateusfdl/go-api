package farms

import "errors"

var (
	ErrFarmNotFound      = errors.New("Farm not found")
	ErrFarmAlreadyExists = errors.New("Farm already exists")
	ErrOnConvertObjectID = errors.New("failed to convert to ObjectID")
	ErrInvalidFarmFields = errors.New("invalid farm fields")
)

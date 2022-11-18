package entities

import "github.com/pkg/errors"

var (
	ErrInvalidParam         = errors.New("invalid param")
	ErrReserveAlreadyExists = errors.New("order already exists")
	ErrReserveInvalidValue  = errors.New("reserve invalid value")
	ErrCommitInvalidValue   = errors.New("commit invalid value")
	ErrNotFound             = errors.New("not found")
	ErrInternal             = errors.New("internal")
)

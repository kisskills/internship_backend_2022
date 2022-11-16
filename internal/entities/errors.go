package entities

import "github.com/pkg/errors"

var (
	ErrInvalidParam = errors.New("invalid param")
	ErrNotFound     = errors.New("not found")
	ErrInternal     = errors.New("internal")
)

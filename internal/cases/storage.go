package cases

import (
	"context"
	"service/internal/entities"
)

type Storage interface {
	GetBalance(ctx context.Context, userID string) (*entities.Balance, error)
	CreateOrUpdateBalance(ctx context.Context, balance *entities.Balance, operation *entities.Operation) error
	GetOperation(ctx context.Context, orderID string) (*entities.Operation, error)
	Reserve(ctx context.Context, operation *entities.Operation) error
	Commit(ctx context.Context, operation *entities.Operation) error
	Rollback(ctx context.Context, operation *entities.Operation) error
}

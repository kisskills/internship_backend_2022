package cases

import (
	"context"
	"service/internal/entities"
)

type Storage interface {
	CreateOrUpdateBalance(ctx context.Context, operation *entities.Operation) error
	GetBalance(ctx context.Context, userID string) (*entities.Balance, error)

	CreateOperation(ctx context.Context, operation *entities.Operation) error
	GetOperation(ctx context.Context, userID, orderID, serviceID string) (*entities.Operation, error)
	UpdateOperationReserve(ctx context.Context, operation *entities.Operation) error
	ListOperations(
		ctx context.Context,
		userID string,
		limit, offset int,
		sortBy string, desc bool,
	) ([]*entities.Operation, error)
}

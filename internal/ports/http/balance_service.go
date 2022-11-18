package http

import (
	"context"
	"service/internal/entities"
)

type BalanceService interface {
	GetUserBalance(ctx context.Context, userID string) (*entities.Balance, error)
	CreditBalance(ctx context.Context, userID string, value entities.Currency) error
	ReserveFromBalance(
		ctx context.Context, userID string, serviceID string, orderID string, value entities.Currency) error
	CommitReserve(ctx context.Context, userID string, serviceID string, orderID string, value entities.Currency) error
	ListOperations(
		ctx context.Context,
		userID string,
		limit, offset int,
		sortBy string, desc bool,
	) ([]*entities.Operation, error)
}

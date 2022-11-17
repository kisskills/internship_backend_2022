package http

import (
	"context"
	"service/internal/entities"
)

type BalanceService interface {
	GetUserBalance(ctx context.Context, userID string) (*entities.Balance, error)
	CreditBalance(ctx context.Context, balance *entities.Balance) error
	ReserveFromBalance(ctx context.Context, operation *entities.Operation) error
	CommitReserve(ctx context.Context, operation *entities.Operation) error
	RollbackReserve(ctx context.Context, operation *entities.Operation) error
}

package cases

import (
	"context"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"service/internal/entities"
)

type BalanceService struct {
	log     *zap.SugaredLogger
	storage Storage
}

func NewBalanceService(log *zap.SugaredLogger, storage Storage) (*BalanceService, error) {
	if log == nil {
		return nil, errors.WithMessage(entities.ErrInvalidParam, "empty logger")
	}

	if storage == nil || storage == Storage(nil) {
		return nil, errors.WithMessage(entities.ErrInvalidParam, "empty storage")
	}

	return &BalanceService{
		log:     log,
		storage: storage,
	}, nil
}

func (s *BalanceService) GetUserBalance(ctx context.Context, userID string) (*entities.Balance, error) {
	log := s.log.With("user_id", userID)

	balance, err := s.storage.GetBalance(ctx, userID)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return balance, nil
}

func (s *BalanceService) CreditBalance(ctx context.Context, balance *entities.Balance) error {
	operation := entities.NewOperation(
		balance.UserID(),
		"",
		"",
		entities.Credit,
		balance.Currency(),
	)

	err := s.storage.CreateOrUpdateBalance(ctx, balance, operation)
	if err != nil {
		s.log.Error(err)
		return err
	}

	return nil
}

func (s *BalanceService) ReserveFromBalance(ctx context.Context, operation *entities.Operation) error {
	err := s.storage.Reserve(ctx, operation)
	if err != nil {
		s.log.Error(err)
		return err
	}

	return nil
}

func (s *BalanceService) CommitReserve(ctx context.Context, operation *entities.Operation) error {
	err := s.storage.Commit(ctx, operation)
	if err != nil {
		s.log.Error(err)
		return err
	}

	return nil
}

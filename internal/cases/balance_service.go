package cases

import (
	"context"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"service/internal/entities"
	"time"
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

func (s *BalanceService) CreditBalance(ctx context.Context, userID string, value entities.Currency) error {
	operation := entities.NewOperation(
		userID,
		entities.DefaultCreditServiceID,
		uuid.New().String(),
		entities.Credit,
		value,
		time.Time{},
	)

	err := s.storage.CreateOrUpdateBalance(ctx, operation)
	if err != nil {
		s.log.Error(err)
		return err
	}

	return nil
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

func (s *BalanceService) ReserveFromBalance(
	ctx context.Context,
	userID string,
	serviceID string,
	orderID string,
	value entities.Currency,
) error {
	operation := entities.NewOperation(userID, serviceID, orderID, entities.Debit, value, time.Time{})

	err := s.storage.CreateOperation(ctx, operation)
	if err != nil {
		s.log.Error(err)
		return err
	}

	return nil
}

func (s *BalanceService) CommitReserve(
	ctx context.Context,
	userID string,
	serviceID string,
	orderID string,
	value entities.Currency,
) error {
	op, err := s.storage.GetOperation(ctx, userID, orderID, serviceID)
	if err != nil {
		s.log.Error(err)
		return err
	}
	if err == entities.ErrNotFound {
		s.log.Error(err)
		return errors.WithMessage(err, "operation not found")
	}
	if op.Value() == 0 {
		s.log.Error(err)
		return errors.WithMessage(entities.ErrInvalidParam, "operation already committed")
	}

	operation := entities.NewOperation(userID, serviceID, orderID, entities.Debit, value, time.Time{})

	err = s.storage.UpdateOperationReserve(ctx, operation)
	if err != nil {
		s.log.Error(err)
		return err
	}

	return nil
}

func (s *BalanceService) ListOperations(
	ctx context.Context,
	userID string,
	limit, offset int,
	sortBy string, desc bool,
) ([]*entities.Operation, error) {
	operations, err := s.storage.ListOperations(ctx, userID, limit, offset, sortBy, desc)
	if err != nil {
		s.log.Error(err)
		return nil, err
	}

	return operations, nil
}

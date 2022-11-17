package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"service/internal/cases"
	"service/internal/entities"
	"time"
)

const (
	connTimeout = 5 * time.Second
)

var (
	_ cases.Storage = (*Storage)(nil)
)

var (
	_ cases.Storage = (*Storage)(nil)
)

type Storage struct {
	log    *zap.SugaredLogger
	cancel context.CancelFunc
	db     *pgxpool.Pool
}

// создаю контекст, кторый отменится через 5 секунд...
// cancel ctx.Done() -> lifeline of context и он завершен

func NewStorage(log *zap.SugaredLogger, dsn string) (*Storage, error) {
	if log == nil {
		return nil, errors.WithMessage(entities.ErrInvalidParam, "empty logger")
	}

	if dsn == "" {
		return nil, errors.WithMessage(entities.ErrInvalidParam, "empty dsn")
	}

	st := &Storage{
		log: log,
	}

	ctx, cancel := context.WithCancel(context.Background())
	st.cancel = cancel

	conn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	err = conn.Ping(ctx)
	if err != nil {
		return nil, err
	}

	st.db = conn

	return st, nil
}

func (s *Storage) GetBalance(ctx context.Context, userID string) (*entities.Balance, error) {
	query := `SELECT currency from avito.balances 
	WHERE user_id=$1`

	var (
		currency int
	)

	row := s.db.QueryRow(ctx, query, &userID)
	err := row.Scan(&currency)
	if errors.Is(err, pgx.ErrNoRows) {
		err = errors.WithMessage(entities.ErrNotFound, "balance not found")
		s.log.Error(err)
		return nil, err
	}
	if err != nil {
		s.log.Error(err)
		return nil, errors.WithMessage(entities.ErrInternal, err.Error())
	}

	balance := entities.NewBalance(userID, entities.Currency(currency))

	return balance, nil
}

func (s *Storage) GetOperation(ctx context.Context, orderID string) (*entities.Operation, error) {
	query := `SELECT user_id, service_id, operation_type, 
       operations_status, value 
	from avito.operations 
	WHERE order_id=$1`

	var (
		userID          string
		serviceID       string
		operationType   int
		operationStatus int
		value           int
	)

	row := s.db.QueryRow(ctx, query, &orderID)
	err := row.Scan(&userID, &serviceID, &operationType, &operationStatus, &value)
	if errors.Is(err, pgx.ErrNoRows) {
		err = errors.WithMessage(entities.ErrNotFound, "operation not found")
		s.log.Error(err)
		return nil, err
	}
	if err != nil {
		s.log.Error(err)
		return nil, errors.WithMessage(entities.ErrInternal, err.Error())
	}

	operation := entities.NewOperation(
		userID,
		serviceID,
		orderID,
		entities.OperationType(operationType),
		entities.Status(operationStatus),
		entities.Currency(value),
	)

	return operation, nil
}

func (s *Storage) CreateOrUpdateBalance(
	ctx context.Context,
	balance *entities.Balance,
	operation *entities.Operation,
) error {
	if err := s.tx(ctx, func(tx pgx.Tx) error {
		err := s.createOrUpdateBalance(ctx, tx, balance)
		if err != nil {
			s.log.Error(err)
			return err
		}

		err = s.createOrUpdateOperation(ctx, tx, operation)
		if err != nil {
			s.log.Error(err)
			return err
		}

		return nil
	}); err != nil {
		return errors.WithMessage(entities.ErrInternal, err.Error())
	}

	return nil
}

func (s *Storage) createOrUpdateBalance(ctx context.Context, tx pgx.Tx, balance *entities.Balance) error {
	query := `INSERT INTO avito.balances (user_id, currency, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE 
		SET currency   = balances.currency + EXCLUDED.currency,
			updated_at = EXCLUDED.updated_at`

	_, err := tx.Exec(ctx, query, balance.UserID(), balance.Currency())
	if err != nil {
		s.log.Error(err)
		return err
	}

	return nil
}

func (s *Storage) createOrUpdateOperation(ctx context.Context, tx pgx.Tx, operation *entities.Operation) error {
	query := `INSERT INTO avito.operations 
    (user_id, service_id, order_id, operation_type, operations_status, value, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) 
	ON CONFLICT (order_id) DO UPDATE 
	SET operations_status = EXCLUDED.operations_status,
		updated_at = EXCLUDED.updated_at`

	params := []interface{}{
		operation.UserID(),
		operation.ServiceID(),
		operation.OrderID(),
		operation.OperationType(),
		operation.OperationStatus(),
		operation.Value(),
	}

	_, err := tx.Exec(ctx, query, params...)
	if err != nil {
		s.log.Error(err)
		return err
	}

	return nil
}

func (s *Storage) Reserve(ctx context.Context, operation *entities.Operation) error {
	queryBalance := `UPDATE avito.balances
			SET currency = currency - $1,
				reserve  = reserve + $1
				WHERE user_id = $2`

	if err := s.tx(ctx, func(tx pgx.Tx) error {
		res, err := tx.Exec(ctx, queryBalance, operation.Value(), operation.UserID())
		if err != nil {
			s.log.Error(err)
			return err
		}

		if res.RowsAffected() == 0 {
			err = errors.WithMessage(entities.ErrNotFound, "balance not found")
			s.log.Error(err)
			return err
		}

		err = s.createOrUpdateOperation(ctx, tx, operation)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *Storage) Commit(ctx context.Context, operation *entities.Operation) error {
	queryBalance := `UPDATE avito.balances
			SET reserve = reserve - $1
				WHERE user_id = $2`

	if err := s.tx(ctx, func(tx pgx.Tx) error {
		res, err := tx.Exec(ctx, queryBalance, operation.Value(), operation.UserID())
		if err != nil {
			return err
		}

		if res.RowsAffected() == 0 {
			err = errors.WithMessage(entities.ErrNotFound, "balance not found")
			s.log.Error(err)
			return err
		}

		err = s.createOrUpdateOperation(ctx, tx, operation)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *Storage) Rollback(ctx context.Context, operation *entities.Operation) error {
	queryBalance := `UPDATE avito.balances
			SET currency = currency + $1,
			    reserve  = reserve - $1
				WHERE user_id = $2`

	if err := s.tx(ctx, func(tx pgx.Tx) error {
		res, err := tx.Exec(ctx, queryBalance, operation.Value(), operation.UserID())
		if err != nil {
			return err
		}

		if res.RowsAffected() == 0 {
			err = errors.WithMessage(entities.ErrNotFound, "balance not found")
			s.log.Error(err)
			return err
		}

		err = s.createOrUpdateOperation(ctx, tx, operation)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *Storage) Close() {
	s.db.Close()
	s.cancel()
}

// мы прокидываем функцию, принимающую обьект транзакции
// функция обертки в  одну транзакцию
func (s *Storage) tx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = fn(tx)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

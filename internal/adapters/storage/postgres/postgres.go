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

type Storage struct {
	log *zap.SugaredLogger
	db  *pgxpool.Pool
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

	ctx, cancel := context.WithTimeout(context.Background(), connTimeout)
	defer cancel()

	conn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	err = conn.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return &Storage{
		log: log,
		db:  conn,
	}, nil
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
		return nil, entities.ErrNotFound
	}
	if err != nil {
		return nil, errors.WithMessage(entities.ErrInternal, err.Error())
	}

	balance := entities.NewBalance(userID, entities.Currency(currency))

	return balance, nil
}

func (s *Storage) CreateOrUpdateBalance(
	ctx context.Context,
	balance *entities.Balance,
	operation *entities.Operation,
) error {
	if err := s.tx(ctx, func(tx pgx.Tx) error {
		err := s.createOrUpdateBalance(ctx, tx, balance)
		if err != nil {
			return err
		}

		err = s.createOperation(ctx, tx, operation)
		if err != nil {
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
		return err
	}

	return nil
}

func (s *Storage) createOperation(ctx context.Context, tx pgx.Tx, operation *entities.Operation) error {
	query := `INSERT INTO avito.operations (id, user_id, service_id, order_id, operation_type, value, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`

	params := []interface{}{
		operation.ID(),
		operation.UserID(),
		operation.ServiceID(),
		operation.OrderID(),
		operation.OperationType(),
		operation.Value(),
	}

	_, err := tx.Exec(ctx, query, params...)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) Reserve(ctx context.Context, operation *entities.Operation) error {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) Commit(ctx context.Context, operation *entities.Operation) error {
	//TODO implement me
	panic("implement me")
}

func (s *Storage) Close() {
	s.db.Close()
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

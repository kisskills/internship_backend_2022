package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"service/internal/cases"
	"service/internal/entities"
	"time"
)

var (
	_ cases.Storage = (*Storage)(nil)
)

type Storage struct {
	log    *zap.SugaredLogger
	cancel context.CancelFunc
	db     *pgxpool.Pool
}

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

func (s *Storage) CreateOrUpdateBalance(
	ctx context.Context,
	operation *entities.Operation,
) error {
	if err := s.tx(ctx, func(tx pgx.Tx) error {
		err := s.createOrUpdateBalance(ctx, tx, operation.UserID(), operation.Value())
		if err != nil {
			return err
		}

		err = s.createOperation(ctx, tx, operation, 0)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetBalance(ctx context.Context, userID string) (*entities.Balance, error) {
	query := `SELECT "value" from avito.balances 
		WHERE user_id=$1`

	var (
		value int
	)

	row := s.db.QueryRow(ctx, query, &userID)
	err := row.Scan(&value)
	if errors.Is(err, pgx.ErrNoRows) {
		err = errors.WithMessage(entities.ErrNotFound, "balance not found")
		s.log.Error(err)
		return nil, err
	}
	if err != nil {
		s.log.Error(err)
		return nil, errors.WithMessage(entities.ErrInternal, err.Error())
	}

	balance := entities.NewBalance(userID, entities.Currency(value))

	return balance, nil
}

func (s *Storage) CreateOperation(ctx context.Context, operation *entities.Operation) error {
	if err := s.tx(ctx, func(tx pgx.Tx) error {
		err := s.decreaseBalance(ctx, tx, operation.UserID(), operation.Value())
		if err != nil {
			return err
		}

		if err := s.createOperation(ctx, s.db, operation, operation.Value()); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetOperation(
	ctx context.Context,
	userID string,
	orderID string,
	serviceID string,
) (*entities.Operation, error) {
	query := `SELECT operation_type, "value", created_at
		from avito.operations 
		WHERE order_id=$1 AND service_id=$2 AND user_id=$3`

	var (
		operationType int
		value         int
		createdAt     time.Time
	)

	row := s.db.QueryRow(ctx, query, &orderID, &serviceID, &userID)
	err := row.Scan(&operationType, &value, &createdAt)
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
		entities.Currency(value),
		createdAt,
	)

	return operation, nil
}

func (s *Storage) UpdateOperationReserve(ctx context.Context, operation *entities.Operation) error {
	query := `UPDATE avito.operations
				SET reserve            = reserve - $1
				WHERE order_id = $2 AND service_id=$3 AND user_id = $4`

	params := []interface{}{
		operation.Value(),
		operation.OrderID(),
		operation.ServiceID(),
		operation.UserID(),
	}

	res, err := s.db.Exec(ctx, query, params...)
	var pge *pgconn.PgError
	if errors.As(err, &pge) {
		s.log.Error(err)
		if pge.Code == pgerrcode.CheckViolation {
			return entities.ErrCommitInvalidValue
		}
	}
	if err != nil {
		s.log.Error(err)
		return err
	}

	if res.RowsAffected() == 0 {
		err = errors.WithMessage(entities.ErrNotFound, "operation not found")
		s.log.Error(err)
		return err
	}

	return nil
}

func (s *Storage) ListOperations(
	ctx context.Context,
	userID string,
	limit, offset int,
	sortBy string, desc bool,
) ([]*entities.Operation, error) {

	var orderBy string

	switch sortBy {
	case entities.Date:
		orderBy = "created_at"
	case entities.Value:
		orderBy = "value"
	}

	queryParams := orderBy

	if desc {
		queryParams += " DESC"
	} else {
		queryParams += " ASC"
	}

	if limit != 0 {
		queryParams += fmt.Sprintf(" LIMIT %d", limit)
	}

	if offset != 0 {
		queryParams += fmt.Sprintf(" OFFSET %d", offset)
	}

	query := fmt.Sprintf(`SELECT service_id, order_id, operation_type, value, created_at
				FROM avito.operations
				WHERE user_id = $1
				ORDER BY %s`, queryParams)

	params := []interface{}{
		userID,
	}

	rows, err := s.db.Query(ctx, query, params...)
	if err != nil {
		err = errors.WithMessage(entities.ErrInternal, err.Error())
		s.log.Error(err)
		return nil, err
	}
	defer rows.Close()

	operations := make([]*entities.Operation, 0)

	for rows.Next() {
		var (
			serviceID     string
			orderID       string
			operationType int
			value         int
			createdAt     time.Time
		)

		err = rows.Scan(&serviceID, &orderID, &operationType, &value, &createdAt)
		if err != nil {
			err = errors.WithMessage(entities.ErrInternal, err.Error())
			s.log.Error(err)
			return nil, err
		}

		operations = append(operations, entities.NewOperation(
			userID,
			serviceID,
			orderID,
			entities.OperationType(operationType),
			entities.Currency(value),
			createdAt,
		))
	}

	return operations, nil
}

func (s *Storage) Close() {
	s.db.Close()
	s.cancel()
}

func (s *Storage) createOrUpdateBalance(ctx context.Context, db db, userID string, value entities.Currency) error {
	query := `INSERT INTO 
    avito.balances (user_id, "value", created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (user_id) DO
		UPDATE SET 
		    	"value"   = balances.value + EXCLUDED.value,
				updated_at = EXCLUDED.updated_at`

	_, err := db.Exec(ctx, query, userID, value)
	if err != nil {
		s.log.Error(err)
		return errors.WithMessage(entities.ErrInternal, err.Error())
	}

	return nil
}

func (s *Storage) decreaseBalance(ctx context.Context, db db, userID string, value entities.Currency) error {
	query := `UPDATE avito.balances
			SET "value" = "value" - $1,
				updated_at = NOW()
			WHERE user_id=$2`

	res, err := db.Exec(ctx, query, value, userID)
	var pge *pgconn.PgError
	if errors.As(err, &pge) {
		s.log.Error(err)
		if pge.Code == pgerrcode.CheckViolation {
			return entities.ErrReserveInvalidValue
		}
	}
	if err != nil {
		s.log.Error(err)
		return errors.WithMessage(entities.ErrInternal, err.Error())
	}

	if res.RowsAffected() == 0 {
		err = errors.WithMessage(entities.ErrNotFound, "balance not found")
		s.log.Error(err)
		return err
	}

	return nil
}

func (s *Storage) createOperation(
	ctx context.Context,
	db db,
	operation *entities.Operation,
	reserve entities.Currency,
) error {
	query := `INSERT INTO avito.operations 
    (user_id, service_id, order_id, operation_type, "value", reserve, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`

	params := []interface{}{
		operation.UserID(),
		operation.ServiceID(),
		operation.OrderID(),
		operation.OperationType(),
		operation.Value(),
		reserve,
	}

	_, err := db.Exec(ctx, query, params...)
	var pge *pgconn.PgError
	if errors.As(err, &pge) {
		s.log.Error(err)
		if pge.Code == pgerrcode.UniqueViolation {
			return entities.ErrReserveAlreadyExists
		}
		if pge.Code == pgerrcode.CheckViolation {
			return entities.ErrReserveInvalidValue
		}
	}
	if err != nil {
		s.log.Error(err)
		return err
	}

	return nil
}

// nolint:errcheck // safety in library
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

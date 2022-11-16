package entities

import (
	"github.com/google/uuid"
)

type OperationType int

const (
	Credit OperationType = iota + 1
	Reserve
	Commit
)

func NewOperation(userID string, serviceID string, orderID string, operationType OperationType, value Currency) *Operation {
	// id генерим не бдшкой, потому что это бизнес-логика и мы не отдаем ее на откуп бд
	return &Operation{
		id:            uuid.New().String(),
		userID:        userID,
		serviceID:     serviceID,
		operationType: operationType,
		orderID:       orderID,
		value:         value,
	}
}

type Operation struct {
	id            string
	userID        string
	serviceID     string
	operationType OperationType
	orderID       string
	value         Currency
}

func (o *Operation) ID() string {
	return o.id
}

func (o *Operation) UserID() string {
	return o.userID
}

func (o *Operation) ServiceID() string {
	return o.serviceID
}

func (o *Operation) OrderID() string {
	return o.orderID
}

func (o *Operation) OperationType() OperationType {
	return o.operationType
}

func (o *Operation) Value() Currency {
	return o.value
}

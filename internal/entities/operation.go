package entities

import "time"

const (
	DefaultCreditServiceID = "credit"
)

type OperationType int

const (
	Credit OperationType = iota + 1
	Debit
)

func NewOperation(
	userID string,
	serviceID string,
	orderID string,
	operationType OperationType,
	value Currency,
	createAt time.Time,
) *Operation {
	return &Operation{
		userID:        userID,
		serviceID:     serviceID,
		operationType: operationType,
		orderID:       orderID,
		value:         value,
		createdAt:     createAt,
	}
}

type Operation struct {
	userID        string
	serviceID     string
	orderID       string
	operationType OperationType
	value         Currency
	createdAt     time.Time
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

func (o *Operation) CreatedAt() time.Time {
	return o.createdAt
}

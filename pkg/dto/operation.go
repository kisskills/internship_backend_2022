package dto

import (
	"service/internal/entities"
	"time"
)

// Operation
//
// swagger:model
type Operation struct {
	ServiceID     string    `json:"service_id"`
	OrderID       string    `json:"order_id"`
	OperationType string    `json:"operation_type"`
	Value         int       `json:"value"`
	CreatedAt     time.Time `json:"created_at"`
}

func ToOperation(operation *entities.Operation) Operation {
	var opType string
	switch operation.OperationType() {
	case entities.Debit:
		opType = "Списание"
	case entities.Credit:
		opType = "Начисление"
	}

	dto := Operation{
		ServiceID:     operation.ServiceID(),
		OrderID:       operation.OrderID(),
		OperationType: opType,
		Value:         int(operation.Value()),
		CreatedAt:     operation.CreatedAt(),
	}

	return dto
}

package entities

type OperationType int
type Status int

const (
	Credit OperationType = iota + 1
	Debit
)

const (
	Reserve Status = iota + 1
	Rollback
	Commit
)

func NewOperation(
	userID string,
	serviceID string,
	orderID string,
	operationType OperationType,
	status Status,
	value Currency,
) *Operation {

	return &Operation{
		userID:        userID,
		serviceID:     serviceID,
		operationType: operationType,
		status:        status,
		orderID:       orderID,
		value:         value,
	}
}

type Operation struct {
	userID        string
	serviceID     string
	operationType OperationType
	status        Status
	orderID       string
	value         Currency
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

func (o *Operation) OperationStatus() Status {
	return o.status
}

func (o *Operation) Value() Currency {
	return o.value
}

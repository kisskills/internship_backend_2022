package dto

// CommitReserveRequest
//
// swagger:model
type CommitReserveRequest struct {
	ServiceID string `json:"service_id"`
	OrderID   string `json:"order_id"`
	Currency  int    `json:"currency"`
}

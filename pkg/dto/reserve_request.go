package dto

// ReserveRequest
//
// swagger:model
type ReserveRequest struct {
	ServiceID string `json:"service_id"`
	OrderID   string `json:"order_id"`
	Currency  int    `json:"currency"`
}

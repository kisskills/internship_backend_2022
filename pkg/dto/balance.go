package dto

// Balance info
//
// swagger:model
type Balance struct {
	UserID   string `json:"user_id"`
	Currency int    `json:"currency"`
}

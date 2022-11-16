package entities

func NewBalance(userID string, currency Currency) *Balance {
	return &Balance{
		userID:   userID,
		currency: currency,
	}
}

type Balance struct {
	userID   string
	currency Currency
}

func (b *Balance) UserID() string {
	return b.userID
}

func (b *Balance) Currency() Currency {
	return b.currency
}

type Currency int

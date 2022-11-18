package entities

func NewBalance(userID string, currency Currency) *Balance {
	return &Balance{
		userID: userID,
		value:  currency,
	}
}

type Balance struct {
	userID string
	value  Currency
}

func (b *Balance) UserID() string {
	return b.userID
}

func (b *Balance) Value() Currency {
	return b.value
}

type Currency int

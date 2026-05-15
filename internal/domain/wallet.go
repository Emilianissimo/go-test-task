package domain

import (
	"errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Wallet struct {
	ID      int64           `json:"id"`
	UUID    uuid.UUID       `json:"uuid" swaggertype:"string" format:"uuid"`
	UserID  int64           `json:"user_id" example:"12345"`
	Balance decimal.Decimal `json:"balance" swaggertype:"string" example:"1500.50"`
}

var (
	ErrWalletNotFound = errors.New("wallet not found")
)

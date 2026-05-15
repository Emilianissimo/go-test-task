package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	PayoutStatusCreated   = 0
	PayoutStatusPending   = 1
	PayoutStatusConfirmed = 2
	PayoutStatusRejected  = -1
)

type PayoutRequest struct {
	TargetID       string          `json:"target_id" validate:"required"`
	WalletID       int64           `json:"wallet_id" validate:"required"`
	Amount         decimal.Decimal `json:"amount" validate:"required,gt=0"`
	IdempotencyKey uuid.UUID       `json:"idempotency_key" validate:"required"`
}

type Payout struct {
	ID            int64           `json:"id"`
	UUID          uuid.UUID       `json:"uuid"`
	TxID          string          `json:"tx_id"`
	TargetID      string          `json:"target_id"`
	WalletFrom    int64           `json:"wallet_from"`
	Amount        decimal.Decimal `json:"amount"`
	BalanceBefore decimal.Decimal `json:"balance_before"`
	BalanceAfter  decimal.Decimal `json:"balance_after"`
	Status        int8            `json:"status"`
	CreatedAt     time.Time       `json:"created_at"`
}

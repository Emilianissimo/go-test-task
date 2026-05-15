package domain

import "errors"

var (
	ErrInsufficientFunds      = errors.New("insufficient funds")
	ErrIdempotencyKeyConflict = errors.New("operation with this idempotency key already exists")
)

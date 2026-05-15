package postgres

import (
	"context"
	"errors"
	"fmt"
	"go-test-system/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type WalletRepository struct {
	db *pgxpool.Pool
}

func NewWalletRepository(db *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) GetByID(ctx context.Context, id int64) (*domain.Wallet, error) {
	query := `
		SELECT id, uuid, user_id, balance
		FROM wallets
		WHERE id = $1
		LIMIT 1
	`

	wal := &domain.Wallet{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&wal.ID,
		&wal.UUID,
		&wal.UserID,
		&wal.Balance,
	)

	if err != nil {
		return nil, fmt.Errorf("scan wallet: %w", err)
	}

	return wal, nil
}

func (r *WalletRepository) GetBalanceForUpdate(ctx context.Context, walletID int64) (decimal.Decimal, error) {
	db := extractDB(ctx, r.db)
	var balance decimal.Decimal
	err := db.QueryRow(ctx, "SELECT balance FROM wallets WHERE id = $1 FOR UPDATE", walletID).Scan(&balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return decimal.Zero, domain.ErrWalletNotFound
	}
	return balance, err
}

func (r *WalletRepository) UpdateBalance(ctx context.Context, walletID int64, newBalance decimal.Decimal) error {
	db := extractDB(ctx, r.db)
	_, err := db.Exec(ctx, "UPDATE wallets SET balance = $1 WHERE id = $2", newBalance, walletID)
	return err
}

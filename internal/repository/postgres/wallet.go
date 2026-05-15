package postgres

import (
	"context"
	"fmt"
	"go-test-system/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
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

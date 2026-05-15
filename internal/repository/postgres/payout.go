package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"go-test-system/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type PayoutRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewPayoutRepository(db *pgxpool.Pool, logger *slog.Logger) *PayoutRepository {
	return &PayoutRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PayoutRepository) CreateRecords(
	ctx context.Context,
	req domain.PayoutRequest,
	txID string,
	balanceBefore, balanceAfter decimal.Decimal,
) (*domain.Payout, error) {
	db := extractDB(ctx, r.db)
	var payout domain.Payout

	err := db.QueryRow(ctx, `
		INSERT INTO payouts (target_id, wallet_from, amount, tx_id, balance_before, balance_after, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, uuid, target_id, wallet_from, amount, tx_id, status, created_at`,
		req.TargetID, req.WalletID, req.Amount, txID, balanceBefore, balanceAfter, domain.PayoutStatusConfirmed,
	).Scan(
		&payout.ID, &payout.UUID, &payout.TargetID,
		&payout.WalletFrom, &payout.Amount, &payout.TxID, &payout.Status, &payout.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert payout: %w", err)
	}

	_, err = db.Exec(ctx, `
		INSERT INTO transactions (op_type, op_id, amount, status, tx_id)
		VALUES ($1, $2, $3, $4, $5)`,
		domain.PayoutOpType, payout.ID, req.Amount, domain.TransactionStatusConfirmed, txID,
	)
	if err != nil {
		return nil, fmt.Errorf("insert audit: %w", err)
	}

	return &payout, nil
}

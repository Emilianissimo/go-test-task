package service

import (
	"context"
	"fmt"
	"log/slog"

	"go-test-system/internal/domain"
	"go-test-system/internal/helpers"

	"github.com/shopspring/decimal"
)

type TxManager interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type PayoutWalletRepo interface {
	GetBalanceForUpdate(ctx context.Context, walletID int64) (decimal.Decimal, error)
	UpdateBalance(ctx context.Context, walletID int64, newBalance decimal.Decimal) error
}

type PayoutRepo interface {
	CreateRecords(ctx context.Context, req domain.PayoutRequest, txID string, before, after decimal.Decimal) (*domain.Payout, error)
}

type PayoutService struct {
	txManager  TxManager
	walletRepo PayoutWalletRepo
	payoutRepo PayoutRepo
	logger     *slog.Logger
}

func NewPayoutService(tm TxManager, wr PayoutWalletRepo, pr PayoutRepo, logger *slog.Logger) *PayoutService {
	return &PayoutService{txManager: tm, walletRepo: wr, payoutRepo: pr, logger: logger}
}

func (s *PayoutService) ProcessPayout(ctx context.Context, req domain.PayoutRequest) (*domain.Payout, error) {
	var result *domain.Payout
	txID := helpers.GenerateFastTxID(req.WalletID)

	err := s.txManager.RunInTx(ctx, func(txCtx context.Context) error {

		currentBalance, err := s.walletRepo.GetBalanceForUpdate(txCtx, req.WalletID)
		if err != nil {
			return err
		}

		if currentBalance.LessThan(req.Amount) {
			s.logger.Warn("insufficient funds",
				slog.Int64("wallet_id", req.WalletID),
				slog.String("balance", currentBalance.String()),
				slog.String("amount", req.Amount.String()),
			)
			return domain.ErrInsufficientFunds
		}

		newBalance := currentBalance.Sub(req.Amount)

		if err := s.walletRepo.UpdateBalance(txCtx, req.WalletID, newBalance); err != nil {
			return err
		}

		result, err = s.payoutRepo.CreateRecords(txCtx, req, txID, currentBalance, newBalance)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("payout failed: %w", err)
	}

	s.logger.Info("payout processed successfully", slog.String("tx_id", txID))
	return result, nil
}

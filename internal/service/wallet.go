package service

import (
	"context"
	"errors"
	"fmt"
	"go-test-system/internal/domain"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

type WalletProvider interface {
	GetByID(ctx context.Context, id int64) (*domain.Wallet, error)
}

type WalletService struct {
	repo   WalletProvider
	logger *slog.Logger
}

func NewWalletService(repo WalletProvider, logger *slog.Logger) *WalletService {
	return &WalletService{repo: repo, logger: logger}
}

func (s *WalletService) GetWallet(ctx context.Context, id int64) (*domain.Wallet, error) {
	s.logger.Info("fetching wallet", "id", id)

	wallet, err := s.repo.GetByID(ctx, id)

	s.logger.Debug("trying to get wallet", "wallet", wallet, "err", err)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrWalletNotFound
		}
		return nil, fmt.Errorf("service failed to get wallet: %w", err)
	}

	return wallet, nil
}

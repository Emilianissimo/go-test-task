package service

import (
	"context"
	"go-test-system/internal/domain"
	"log/slog"
)

type PayoutRepository interface {
	CreatePayout(ctx context.Context, req domain.PayoutRequest) (*domain.Payout, error)
}

type PayoutService struct {
	repo   PayoutRepository
	logger *slog.Logger
}

func NewPayoutService(repo PayoutRepository, logger *slog.Logger) *PayoutService {
	return &PayoutService{
		repo:   repo,
		logger: logger,
	}
}

func (s *PayoutService) ProcessPayout(ctx context.Context, req domain.PayoutRequest) (*domain.Payout, error) {
	// This will be the case in the future:
	// Check user limits (Daily/Monthly)
	// Call the anti-fraud system
	// Send an event to Kafka/RabbitMQ about the start of a payout

	return s.repo.CreatePayout(ctx, req)
}

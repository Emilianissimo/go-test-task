package service

import (
	"context"
	"go-test-system/internal/domain"
	"go-test-system/internal/provider/skinport"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"
)

type ItemService struct {
	client *skinport.Client
	logger *slog.Logger
}

func NewItemService(client *skinport.Client, logger *slog.Logger) *ItemService {
	return &ItemService{client: client, logger: logger}
}

func (s *ItemService) GetProcessedItems(ctx context.Context) ([]domain.MergedItem, error) {
	var tradableRaw, nonTradableRaw []domain.ItemResponse
	g, ctx := errgroup.WithContext(ctx)

	s.logger.Info("starting parallel fetch from skinport")
	startFetch := time.Now()

	g.Go(func() error {
		var err error
		tradableRaw, err = s.client.FetchItems(ctx, true)
		if err != nil {
			s.logger.Error("failed to fetch tradable items", slog.Any("error", err))
		}
		return err
	})

	g.Go(func() error {
		var err error
		nonTradableRaw, err = s.client.FetchItems(ctx, false)
		if err != nil {
			s.logger.Error("failed to fetch non-tradable items", slog.Any("error", err))
		}
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	s.logger.Info("fetch completed successfully",
		slog.Duration("duration", time.Since(startFetch)),
		slog.Int("tradable_count", len(tradableRaw)),
		slog.Int("non_tradable_count", len(nonTradableRaw)),
	)

	startMerge := time.Now()

	resMap := make(map[string]*domain.MergedItem)

	for _, r := range tradableRaw {
		resMap[r.MarketHashName] = &domain.MergedItem{
			MarketHashName:   r.MarketHashName,
			MinPriceTradable: r.MinPrice,
		}
	}

	for _, r := range nonTradableRaw {
		if item, ok := resMap[r.MarketHashName]; ok {
			item.MinPriceNonTradable = r.MinPrice
		} else {
			resMap[r.MarketHashName] = &domain.MergedItem{
				MarketHashName:      r.MarketHashName,
				MinPriceNonTradable: r.MinPrice,
			}
		}
	}

	result := make([]domain.MergedItem, 0, len(resMap))
	for _, v := range resMap {
		result = append(result, *v)
	}

	s.logger.Info("items merged successfully",
		slog.Int("total_merged", len(result)),
		slog.Duration("merge_duration", time.Since(startMerge)),
	)

	return result, nil
}

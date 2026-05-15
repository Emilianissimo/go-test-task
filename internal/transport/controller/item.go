package controller

import (
	"context"
	"encoding/json"
	"errors"
	"go-test-system/internal/config"
	_ "go-test-system/internal/domain"
	"go-test-system/internal/provider/skinport"
	"go-test-system/internal/service"
	"log/slog"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type ItemHandler struct {
	service *service.ItemService
	cfg     *config.Config
	redis   *redis.Client
	logger  *slog.Logger
}

func (h *ItemHandler) respondJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *ItemHandler) respondError(w http.ResponseWriter, code int, msg string) {
	h.respondJSON(w, code, map[string]string{"error": msg})
}

func NewItemHandler(service *service.ItemService, cfg *config.Config, redis *redis.Client, logger *slog.Logger) *ItemHandler {
	return &ItemHandler{
		service: service,
		cfg:     cfg,
		redis:   redis,
		logger:  logger,
	}
}

// FetchItems GoDoc
// @Summary      Get merged items from Skinport
// @Description  Fetches tradable and non-tradable items from Skinport API, merges them, and returns with caching.
// @Tags         external
// @Produce      json
// @Success      200  {array}   domain.MergedItem "Successfully merged items"
// @Failure      502  {object}  map[string]string "Upstream error (Skinport unreachable)"
// @Router       /v1/external/items/ [get]
func (h *ItemHandler) FetchItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cacheKey := "skinport:items:merged:" + h.cfg.SkinportAppID + ":" + h.cfg.SkinportCurrency

	startCache := time.Now()
	cached, err := h.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		h.logger.Info("cache hit",
			slog.String("key", cacheKey),
			slog.Duration("latency", time.Since(startCache)),
		)
		h.respondJSON(w, http.StatusOK, json.RawMessage(cached))
		return
	}

	h.logger.Info("cache miss or redis error", slog.String("key", cacheKey), slog.Any("err", err))

	items, err := h.service.GetProcessedItems(ctx)
	if err != nil {
		if errors.Is(err, skinport.ErrRateLimit) {
			h.respondError(w, http.StatusTooManyRequests, "slow down, skinport is angry")
			return
		}
		h.respondError(w, http.StatusBadGateway, "upstream is down")
		return
	}

	go func() {
		// Go redis using set and can't use io.Reader, so we will set data in the other go routine
		data, marshalErr := json.Marshal(items)
		if marshalErr != nil {
			h.logger.Error("failed to marshal items for cache", slog.Any("error", marshalErr))
			return
		}

		setErr := h.redis.Set(context.Background(), cacheKey, data, h.cfg.SkinportCacheTTL).Err()
		if setErr != nil {
			h.logger.Warn("failed to update redis cache", slog.Any("error", setErr))
		} else {
			h.logger.Info("cache updated", slog.String("key", cacheKey))
		}
	}()

	h.respondJSON(w, http.StatusOK, items)
}

package controller

import (
	"context"
	"encoding/json"
	"errors"
	"go-test-system/internal/domain"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type WalletService interface {
	GetWallet(ctx context.Context, id int64) (*domain.Wallet, error)
}

type WalletHandler struct {
	service WalletService
	logger  *slog.Logger
}

func NewWalletHandler(service WalletService, logger *slog.Logger) *WalletHandler {
	return &WalletHandler{
		service: service,
		logger:  logger,
	}
}

func (h *WalletHandler) sendError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(map[string]string{"error": msg})
	if err != nil {
		return
	}
}

// GetByID возвращает данные кошелька по ID
// @Summary Получить кошелек
// @Description Возвращает баланс и UUID кошелька по внутреннему ID
// @Tags Wallets
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} domain.Wallet
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 404 {object} map[string]string "Not Found"
// @Router /v1/wallets/{id} [get]
func (h *WalletHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		h.logger.Warn("invalid id format", "input", idStr)
		h.sendError(w, http.StatusBadRequest, "invalid id format")
		return
	}

	wallet, err := h.service.GetWallet(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrWalletNotFound) {
			h.sendError(w, http.StatusNotFound, "wallet not found")
			return
		}

		h.logger.Warn("failed to get wallet", "error", err, "id", id)
		h.sendError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(wallet)
}

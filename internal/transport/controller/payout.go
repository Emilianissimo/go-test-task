package controller

import (
	"context"
	"encoding/json"
	"errors"
	"go-test-system/internal/domain"
	"io"
	"log/slog"
	"net/http"
)

type PayoutService interface {
	ProcessPayout(ctx context.Context, req domain.PayoutRequest) (*domain.Payout, error)
}

type PayoutHandler struct {
	service PayoutService
	logger  *slog.Logger
}

func NewPayoutHandler(service PayoutService, logger *slog.Logger) *PayoutHandler {
	return &PayoutHandler{
		service: service,
		logger:  logger,
	}
}

// CreatePayout GoDoc
// @Summary      CreatePayout a new payout
// @Description  Initiates a payout from a wallet. Protected by idempotency middleware.
// @Tags         payouts
// @Accept       json
// @Produce      json
// @Param        X-Idempotency-Key header string true "Idempotency Key (UUID/Hash)"
// @Param        request body domain.PayoutRequest true "Payout Data"
// @Success      201  {object}  domain.Payout
// @Failure      400  {object}  map[string]string "Invalid input or negative amount"
// @Failure      402  {object}  map[string]string "Insufficient funds"
// @Failure      404  {object}  map[string]string "Wallet not found"
// @Failure      409  {object}  map[string]string "Idempotency conflict"
// @Router       /v1/payouts [post]
func (h *PayoutHandler) CreatePayout(w http.ResponseWriter, r *http.Request) {
	var req domain.PayoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode payout request", slog.Any("error", err))
		h.respondError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)

	if req.WalletID <= 0 {
		h.respondError(w, http.StatusBadRequest, "invalid wallet_id")
		return
	}
	if req.Amount.IsNegative() || req.Amount.IsZero() {
		h.logger.Warn("attempt to process negative or zero payout", slog.String("amount", req.Amount.String()))
		h.respondError(w, http.StatusBadRequest, "amount must be strictly positive")
		return
	}
	if req.TargetID == "" {
		h.respondError(w, http.StatusBadRequest, "target_id is required")
		return
	}

	payout, err := h.service.ProcessPayout(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrWalletNotFound):
			h.respondError(w, http.StatusNotFound, "wallet not found")
		case errors.Is(err, domain.ErrInsufficientFunds):
			h.respondError(w, http.StatusBadRequest, "insufficient funds")
		case errors.Is(err, domain.ErrIdempotencyKeyConflict):
			h.respondError(w, http.StatusConflict, "idempotency key conflict")
		default:
			h.logger.Error("failed to process payout", slog.Any("error", err))
			h.respondError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	h.respondJSON(w, http.StatusCreated, payout)
}

func (h *PayoutHandler) respondJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func (h *PayoutHandler) respondError(w http.ResponseWriter, code int, msg string) {
	h.respondJSON(w, code, map[string]string{"error": msg})
}

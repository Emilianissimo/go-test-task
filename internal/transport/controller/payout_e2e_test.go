package controller_test

// Gemini generated

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"go-test-system/internal/domain"
	"go-test-system/internal/repository/postgres"
	"go-test-system/internal/service"
	"go-test-system/internal/transport/controller"
	"go-test-system/internal/transport/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestEnv(t *testing.T) (http.Handler, *pgxpool.Pool, *redis.Client, int64) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/test_db?sslmode=disable"
	}
	redisAddr := os.Getenv("TEST_REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err, "failed to connect to test db")

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	require.NoError(t, rdb.Ping(ctx).Err(), "failed to connect to test redis")

	rdb.FlushAll(ctx)
	_, err = pool.Exec(ctx, "TRUNCATE TABLE transactions, payouts, wallets, users RESTART IDENTITY CASCADE")
	require.NoError(t, err)

	walletID := int64(1)
	initialBalance := decimal.NewFromInt(1000)
	_, err = pool.Exec(ctx, "INSERT INTO users (id, name) VALUES (1, 'Test Builder')")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "INSERT INTO wallets (id, user_id, balance) VALUES ($1, 1, $2)", walletID, initialBalance)
	require.NoError(t, err)

	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	txManager := postgres.NewTxManager(pool)
	walletRepo := postgres.NewWalletRepository(pool)
	payoutRepo := postgres.NewPayoutRepository(pool)
	payoutSvc := service.NewPayoutService(txManager, walletRepo, payoutRepo, discardLogger)
	payoutHandler := controller.NewPayoutHandler(payoutSvc, discardLogger)

	r := chi.NewRouter()
	r.With(middleware.Idempotency(rdb, 1*time.Minute)).
		Post("/v1/payouts", payoutHandler.CreatePayout)

	return r, pool, rdb, walletID
}

func makeRequest(t *testing.T, router http.Handler, idempotencyKey string, reqBody domain.PayoutRequest) *httptest.ResponseRecorder {
	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequest(http.MethodPost, "/v1/payouts", bytes.NewReader(body))
	require.NoError(t, err)

	if idempotencyKey != "" {
		req.Header.Set("X-Idempotency-Key", idempotencyKey)
	}
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func checkBalance(t *testing.T, pool *pgxpool.Pool, walletID int64) decimal.Decimal {
	var balance decimal.Decimal
	err := pool.QueryRow(context.Background(), "SELECT balance FROM wallets WHERE id = $1", walletID).Scan(&balance)
	require.NoError(t, err)
	return balance
}

func TestPayoutFlow_EndToEnd(t *testing.T) {
	router, pool, _, walletID := setupTestEnv(t)
	defer pool.Close()

	idempotencyKey := "key-e2e-123"
	payoutAmount := decimal.NewFromFloat(150.50)

	t.Run("1. Success Payout (Happy Path)", func(t *testing.T) {
		reqBody := domain.PayoutRequest{
			WalletID: walletID,
			Amount:   payoutAmount,
			TargetID: "bank-777",
		}

		rec := makeRequest(t, router, idempotencyKey, reqBody)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp domain.Payout
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.EqualValues(t, domain.PayoutStatusConfirmed, resp.Status)

		balance := checkBalance(t, pool, walletID)
		assert.True(t, balance.Equal(decimal.NewFromFloat(849.50)), "Balance should be deducted")
	})

	t.Run("2. Idempotency Check (Same Key)", func(t *testing.T) {
		reqBody := domain.PayoutRequest{
			WalletID: walletID,
			Amount:   payoutAmount,
			TargetID: "bank-777",
		}

		rec := makeRequest(t, router, idempotencyKey, reqBody)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "HIT", rec.Header().Get("X-Cache"), "Should hit idempotency cache")

		balance := checkBalance(t, pool, walletID)
		assert.True(t, balance.Equal(decimal.NewFromFloat(849.50)), "Balance MUST NOT change on retry")
	})

	t.Run("3. Success Payout (New Key)", func(t *testing.T) {
		newKey := "key-e2e-456"
		newAmount := decimal.NewFromInt(49)
		reqBody := domain.PayoutRequest{
			WalletID: walletID,
			Amount:   newAmount,
			TargetID: "bank-888",
		}

		rec := makeRequest(t, router, newKey, reqBody)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Empty(t, rec.Header().Get("X-Cache"), "Cache should be empty for new key")

		balance := checkBalance(t, pool, walletID)
		assert.True(t, balance.Equal(decimal.NewFromFloat(800.50)))
	})

	t.Run("4. Insufficient Funds", func(t *testing.T) {
		failKey := "key-e2e-fail"
		reqBody := domain.PayoutRequest{
			WalletID: walletID,
			Amount:   decimal.NewFromInt(10000),
			TargetID: "greedy-bank",
		}

		rec := makeRequest(t, router, failKey, reqBody)

		assert.Equal(t, http.StatusBadRequest, rec.Code) // Замени на StatusPaymentRequired если используешь 402

		balance := checkBalance(t, pool, walletID)
		assert.True(t, balance.Equal(decimal.NewFromFloat(800.50)), "Balance should rollback to 800.50")
	})

	t.Run("5. Missing Idempotency Key", func(t *testing.T) {
		reqBody := domain.PayoutRequest{
			WalletID: walletID,
			Amount:   decimal.NewFromInt(10),
			TargetID: "bank-000",
		}

		rec := makeRequest(t, router, "", reqBody)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "strictly required")
	})
}

package router

import (
	"go-test-system/internal/config"
	"go-test-system/internal/provider/skinport"
	"go-test-system/internal/repository/postgres"
	"go-test-system/internal/service"
	"go-test-system/internal/transport/controller"
	internalMiddleware "go-test-system/internal/transport/middleware"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Deps struct {
	DB         *pgxpool.Pool
	Redis      *redis.Client
	Logger     *slog.Logger
	HttpClient *http.Client
	Cfg        *config.Config
}

func NewRouter(deps Deps) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// External client and services & controllers
	spClient := skinport.NewClient(deps.Cfg, deps.HttpClient, deps.Logger)
	itemSvc := service.NewItemService(spClient, deps.Logger)
	itemHandler := controller.NewItemHandler(itemSvc, deps.Cfg, deps.Redis, deps.Logger)

	// Infra
	txManager := postgres.NewTxManager(deps.DB)

	// Internal repos and services & controllers
	walletRepo := postgres.NewWalletRepository(deps.DB)
	walletSvc := service.NewWalletService(walletRepo, deps.Logger)
	walletHandler := controller.NewWalletHandler(walletSvc, deps.Logger)

	payoutRepo := postgres.NewPayoutRepository(deps.DB)
	payoutSvc := service.NewPayoutService(txManager, walletRepo, payoutRepo, deps.Logger)
	payoutHandler := controller.NewPayoutHandler(payoutSvc, deps.Logger)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/external/items/", itemHandler.FetchItems)
		r.Get("/wallets/{id}", walletHandler.GetByID)
		r.With(internalMiddleware.Idempotency(deps.Redis, 24*time.Hour)).Post("/payouts", payoutHandler.Create)
	})

	return r
}

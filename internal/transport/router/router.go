package router

import (
	"go-test-system/internal/repository/postgres"
	"go-test-system/internal/service"
	"go-test-system/internal/transport/controller"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Deps struct {
	DB     *pgxpool.Pool
	Redis  *redis.Client
	Logger *slog.Logger
}

func NewRouter(deps Deps) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	walletRepo := postgres.NewWalletRepository(deps.DB)
	walletSvc := service.NewWalletService(walletRepo, deps.Logger)
	walletHandler := controller.NewWalletHandler(walletSvc, deps.Logger)

	r.Route("/v1", func(r chi.Router) {
		r.Get("/wallets/{id}", walletHandler.GetByID)
	})

	return r
}

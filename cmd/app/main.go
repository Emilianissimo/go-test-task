package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"go-test-system/internal/helpers"
	"go-test-system/internal/transport/router"
)

type Config struct {
	DBURL        string
	RedisHost    string
	RedisPort    string
	RedisPass    string
	AppPort      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// @title Go test task API
// @version 1.0
// @description Go test task
// @termsOfService http://swagger.io/terms/

// @contact.name Emilian
// @contact.email emilerofeevskij@gmail.com

// @host localhost:8080
// @BasePath /
func main() {
	// Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := godotenv.Load(".env"); err != nil {
		logger.Warn("No .env file provided, use `make set-env`")
	}

	cfg := &Config{
		DBURL:        os.Getenv("DATABASE_URL"),
		RedisHost:    os.Getenv("REDIS_HOST"),
		RedisPort:    os.Getenv("REDIS_PORT"),
		RedisPass:    os.Getenv("REDIS_PASSWORD"),
		AppPort:      os.Getenv("PORT"),
		ReadTimeout:  helpers.GetEnvDuration("SERVER_READ_TIMEOUT", 5*time.Second, logger),
		WriteTimeout: helpers.GetEnvDuration("SERVER_WRITE_TIMEOUT", 10*time.Second, logger),
		IdleTimeout:  helpers.GetEnvDuration("SERVER_IDLE_TIMEOUT", 120*time.Second, logger),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// PSQL
	dbPool, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		logger.Error("unable to create db pool", "err", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		logger.Error("db ping failed", "err", err)
		os.Exit(1)
	}

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPass,
		DB:       0,
	})
	defer func(rdb *redis.Client) {
		err := rdb.Close()
		if err != nil {
			logger.Error("redis close failed", "err", err)
		}
	}(rdb)

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("redis ping failed", "err", err)
		os.Exit(1)
	}

	logger.Info("infrastructure is ready", "port", cfg.AppPort)

	// Routing & Server
	deps := router.Deps{
		DB:     dbPool,
		Redis:  rdb,
		Logger: logger,
	}
	newRouter := router.NewRouter(deps)

	appPort := cfg.AppPort
	if appPort == "" {
		appPort = "8080"
	}
	srv := &http.Server{
		Addr:         ":" + appPort,
		Handler:      newRouter,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful Shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen failed", "err", err)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", "err", err)
	}

	logger.Info("server exited cleanly")
}

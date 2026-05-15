package helpers

import (
	"log/slog"
	"os"
	"time"
)

func GetEnvDuration(key string, fallback time.Duration, log *slog.Logger) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		log.Warn("invalid duration in env, using fallback", "key", key, "val", val, "fallback", fallback)
		return fallback
	}
	return d
}

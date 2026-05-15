package middleware

import (
	"bytes"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

func Idempotency(rds *redis.Client, ttl time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-Idempotency-Key")
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			resultKey := "idempotency:res:" + key
			lockKey := "idempotency:lock:" + key

			locked, _ := rds.SetNX(ctx, lockKey, "1", 10*time.Second).Result()
			if !locked {
				http.Error(w, "Request already in progress", http.StatusConflict)
				return
			}
			defer rds.Del(ctx, lockKey)

			cached, err := rds.Get(ctx, resultKey).Bytes()
			if err == nil {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Cache", "HIT")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write(cached)
				if err != nil {
					return
				}
				return
			}

			rec := &responseRecorder{
				ResponseWriter: w,
				body:           &bytes.Buffer{},
			}

			next.ServeHTTP(rec, r)

			if rec.status >= 200 && rec.status < 300 {
				rds.Set(ctx, resultKey, rec.body.Bytes(), ttl)
			}
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type CachedResponse struct {
	Status int    `json:"status"`
	Body   []byte `json:"body"`
}

func Idempotency(rds *redis.Client, ttl time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-Idempotency-Key")
			if key == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte(`{"error": "X-Idempotency-Key header is strictly required"}`))
				if err != nil {
					return
				}
				return
			}

			reqCtx := r.Context()
			resultKey := "idempotency:res:" + key
			lockKey := "idempotency:lock:" + key

			locked, _ := rds.SetNX(reqCtx, lockKey, "1", 10*time.Second).Result()
			if !locked {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				_, err := w.Write([]byte(`{"error": "request already in progress"}`))
				if err != nil {
					return
				}
				return
			}

			defer rds.Del(context.Background(), lockKey)

			cachedBytes, err := rds.Get(reqCtx, resultKey).Bytes()
			if err == nil {
				var cachedResp CachedResponse
				if err := json.Unmarshal(cachedBytes, &cachedResp); err == nil {
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("X-Cache", "HIT")
					w.WriteHeader(cachedResp.Status)
					_, err := w.Write(cachedResp.Body)
					if err != nil {
						return
					}
					return
				}
			}

			rec := &responseRecorder{
				ResponseWriter: w,
				body:           &bytes.Buffer{},
				status:         http.StatusOK,
			}

			next.ServeHTTP(rec, r)

			if rec.status >= 200 && rec.status < 300 {
				respToCache := CachedResponse{
					Status: rec.status,
					Body:   rec.body.Bytes(),
				}

				if data, err := json.Marshal(respToCache); err == nil {
					rds.Set(reqCtx, resultKey, data, ttl)
				}
			}
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	body        *bytes.Buffer
	status      int
	wroteHeader bool
}

func (r *responseRecorder) WriteHeader(status int) {
	if !r.wroteHeader {
		r.status = status
		r.wroteHeader = true
		r.ResponseWriter.WriteHeader(status)
	}
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

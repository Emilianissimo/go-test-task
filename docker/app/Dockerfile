# Stage 1: Build
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-test-app ./cmd/app/main.go

# Stage 2: Run
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go-test-app .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./go-test-app"]

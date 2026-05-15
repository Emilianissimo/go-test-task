FROM golang:1.22-alpine AS builder

RUN go install github.com/swaggo/swag/cmd/swag@latest

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN swag init -g cmd/main.go

RUN go build -o go-test-app ./cmd/main.go

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/go-test-app .
COPY --from=builder /app/docs ./docs

EXPOSE 8080
CMD ["./go-test-app"]

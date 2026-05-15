FROM golang:1.26-alpine AS builder

RUN apk add --no-cache curl tar
RUN go install github.com/swaggo/swag/cmd/swag@latest

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN swag init -g cmd/main.go

RUN curl -L https://github.com/swagger-api/swagger-ui/archive/refs/tags/v5.17.14.tar.gz | tar -xz && \
    cp -r swagger-ui-5.17.14/dist/* ./docs/ && \
    sed -i 's|https://petstore.swagger.io/v2/swagger.json|./swagger.json|g' ./docs/swagger-initializer.js

RUN go build -o go-test-app ./cmd/main.go

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/go-test-app .
COPY --from=builder /app/docs ./docs

EXPOSE 8080
CMD ["./go-test-app"]

$DB_PORT = 5433
$REDIS_PORT = 6380
$Env:TEST_DATABASE_URL = "postgres://postgres:postgres@localhost:$DB_PORT/test_db?sslmode=disable"
$Env:TEST_REDIS_ADDR = "localhost:$REDIS_PORT"

try {
    Write-Host "--> Starting isolated infrastructure..." -ForegroundColor Cyan

    $existing = docker ps -aq -f name=test-postgres -f name=test-redis
    if ($existing) {
        docker rm -f $existing | Out-Null
    }

    docker run --name test-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_USER=postgres -e POSTGRES_DB=test_db -p "$($DB_PORT):5432" -d postgres:15-alpine | Out-Null
    docker run --name test-redis -p "$($REDIS_PORT):6379" -d redis:7-alpine | Out-Null

    Write-Host "--> Waiting for Postgres (healthcheck)..." -ForegroundColor Cyan
    while ($true) {
        docker exec test-postgres pg_isready -U postgres 2>$null | Out-Null
        if ($LASTEXITCODE -eq 0) { break }
        Start-Sleep -Seconds 1
    }
    Write-Host "Postgres is ready." -ForegroundColor Green

    Write-Host "--> Running migrations via Docker..." -ForegroundColor Cyan
    docker run --rm -v "$PWD/migrations:/migrations" --network container:test-postgres migrate/migrate -path=/migrations/ -database "postgres://postgres:postgres@localhost:5432/test_db?sslmode=disable" up
    if ($LASTEXITCODE -ne 0) { throw "Migrations failed with exit code $LASTEXITCODE" }

    Write-Host "--> Running integration tests with race detector..." -ForegroundColor Cyan
    go test -v -race -count=1 ./internal/transport/controller/...
    if ($LASTEXITCODE -ne 0) { throw "Tests failed with exit code $LASTEXITCODE" }

    Write-Host "--> Success. Pipeline finished." -ForegroundColor Green
}
finally {
    Write-Host "Cleanup: removing test containers..." -ForegroundColor Yellow
    $existing = docker ps -aq -f name=test-postgres -f name=test-redis
    if ($existing) {
        docker rm -f $existing | Out-Null
    }
}
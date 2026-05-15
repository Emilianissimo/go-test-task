set-env:
	test -f .env || cp .env.example .env

build:
	docker compose build

start:
	docker compose up -d

up:
	docker compose up

stop:
	docker compose stop

down:
	docker compose down

migrate:
	docker compose run --rm migrate up

rollback:
	docker compose run --rm migrate migrate down -all

migrate-clean-all:
	docker compose run --rm migrate drop -f

e2e-windows:
	powershell -ExecutionPolicy Bypass -File .\scripts\test_e2e.ps1

set-env:
	test -f .env || cp .env.example .env

build-app:
	docker compose build app

start-app:
	docker compose up -d app

rebuild-app: build-app start-app

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

enter_db:
	docker exec -it cpp_api_db_container psql --username=core_db_user --dbname=core_db

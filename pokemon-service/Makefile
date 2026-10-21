-include .env
export

DB_USER     ?= postgres
DB_PASSWORD ?= postgres
DB_NAME     ?= pokemon
DB_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@postgres:5432/$(DB_NAME)?sslmode=disable

.PHONY: tidy swag run build test cover db-up migrate-up migrate-down seed up down logs

tidy:
	go mod tidy

SWAG_VERSION ?= v1.16.4

# Generate dokumentasi OpenAPI ke folder docs/ dari anotasi di handler.
# Jalankan ulang setiap kali anotasi/DTO berubah, lalu commit hasilnya.
swag:
	go run github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION) fmt -g cmd/main.go
	go run github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION) init -g cmd/main.go -o docs --parseInternal

run:
	go run ./cmd

build:
	CGO_ENABLED=0 go build -o bin/pokemon-service ./cmd

test:
	go test ./... -race -count=1

cover:
	go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out

# Jalankan Postgres saja (untuk `make run` dari host)
db-up:
	docker compose up -d postgres

migrate-up:
	docker compose run --rm migrate

migrate-down:
	docker compose run --rm migrate -path=/migrations -database="$(DB_URL)" down 1

seed:
	docker compose exec -T postgres psql -U $(DB_USER) -d $(DB_NAME) < scripts/seed.sql

# Jalankan semua: postgres + migrate + app
up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f app

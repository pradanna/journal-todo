.PHONY: run build migrate-up migrate-down tidy

run:
	go run ./cmd/api

build:
	go build -o bin/jurnal-api ./cmd/api

tidy:
	go mod tidy

migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down 1

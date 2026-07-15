.PHONY: run build migrate-up test fmt vet check up down logs

run:
	go run ./cmd/api

build:
	mkdir -p bin
	go build -o bin/api ./cmd/api

migrate-up:
	docker compose run --rm migrate

test:
	go test ./...

fmt:
	gofmt -w $$(find . -name '*.go' -type f)

vet:
	go vet ./...

check: fmt vet test

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f api

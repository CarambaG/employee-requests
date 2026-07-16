.PHONY: run build migrate-up seed seed-small seed-verify benchmark-baseline benchmark-clean test fmt vet check up down logs

run:
	go run ./cmd/api

build:
	mkdir -p bin
	go build -o bin/api ./cmd/api

migrate-up:
	docker compose run --rm migrate

seed:
	docker compose run --rm seed

seed-small:
	EMPLOYEE_COUNT=100 REQUEST_COUNT=10000 docker compose run --rm seed

seed-verify:
	docker compose exec -T db sh -c 'psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" -c "SELECT (SELECT COUNT(*) FROM employees) AS employees, (SELECT COUNT(*) FROM requests) AS requests;"'

benchmark-baseline:
	docker compose run --rm benchmark

benchmark-clean:
	rm -f performance-results/baseline.txt performance-results/baseline_rows.csv

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

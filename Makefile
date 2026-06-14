.PHONY: run test lint migrate-up migrate-down swagger docker-up docker-down

run:
	go run cmd/server/main.go

test:
	go test -v ./...

lint:
	golangci-lint run

migrate-up:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/taskflow?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/taskflow?sslmode=disable" down

swagger:
	swag init -g cmd/server/main.go

docker-up:
	docker compose -f deploy/docker-compose.yml up --build

docker-down:
	docker compose -f deploy/docker-compose.yml down

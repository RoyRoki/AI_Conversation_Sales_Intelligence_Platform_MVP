.PHONY: run migrate test build clean docker-build docker-up docker-down docker-logs docker-stop

run:
	go run cmd/api/main.go

migrate:
	go run cmd/migrate/main.go -direction=up

cleanup:
	go run cmd/cleanup/main.go

cleanup-dry-run:
	go run cmd/cleanup/main.go -dry-run

cleanup-all:
	go run cmd/cleanup/main.go -all

cleanup-all-dry-run:
	go run cmd/cleanup/main.go -all -dry-run

test:
	go test ./...

build:
	go build -o bin/api cmd/api/main.go
	go build -o bin/migrate cmd/migrate/main.go
	go build -o bin/cleanup cmd/cleanup/main.go

clean:
	rm -rf bin/
	rm -f *.db
	rm -rf data/

deps:
	go mod download
	go mod tidy

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-stop:
	docker-compose stop

docker-logs:
	docker-compose logs -f

docker-restart:
	docker-compose restart


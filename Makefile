.PHONY: build run dev clean test swagger

# Generate Swagger documentation
swagger:
	swag init -g cmd/api/main.go -o docs

# Build the application
build: swagger
	go build -o bin/tp-api cmd/api/main.go

# Run the application
run: build
	./bin/tp-api

# Run in development mode with hot reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	air

# Clean build artifacts
clean:
	rm -rf bin/

# Run tests
test:
	go test -v ./...

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run with docker-compose (if you have docker-compose.yml)
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

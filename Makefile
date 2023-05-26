.PHONY: dev-up dev-down start-server install-modules

dev-up:
	@echo "Starting development environment..."
	docker-compose up -d

dev-down:
	@echo "Stopping development environment..."
	docker-compose down

start-server:
	@echo "Starting server with hot-reloading..."
	air

install-modules:
	@echo "Installing Go modules..."
	go get github.com/gofiber/fiber/v2
	go get github.com/google/uuid
	go get github.com/go-playground/validator/v10
	go get -u gorm.io/gorm
	go get gorm.io/driver/postgres
	go get github.com/spf13/viper
	go get github.com/golang-jwt/jwt
	# Hot-reload server
	go install github.com/cosmtrek/air@latest

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  dev-up            Start the development environment (Docker containers)"
	@echo "  dev-down          Stop the development environment (Docker containers)"
	@echo "  start-server      Start the server with hot-reloading"
	@echo "  install-modules   Install required Go modules"
	@echo "  help              Show this help message"

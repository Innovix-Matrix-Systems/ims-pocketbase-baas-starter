# Makefile for IMS PocketBase BaaS Starter
.PHONY: help build start stop restart down logs clean clean-data dev dev-build dev-logs dev-clean test lint format

# Default target
help:
	@echo "Available commands:"
	@echo "  build        - Build the production Docker image"
	@echo "  start        - Start the production containers"
	@echo "  stop         - Stop the containers"
	@echo "  restart      - Restart the containers"
	@echo "  down         - Stop and remove containers"
	@echo "  logs         - Show container logs"
	@echo "  clean        - Remove containers, networks, images, and volumes"
	@echo "  clean-data   - Remove only volumes"
	@echo ""
	@echo "Development commands:"
	@echo "  dev          - Start development environment with hot reload"
	@echo "  dev-build    - Build development image"
	@echo "  dev-logs     - Show development container logs"
	@echo "  dev-clean    - Clean development environment"
	@echo "  dev-data-clean - Clean development data"
	@echo ""
	@echo "Code quality:"
	@echo "  test         - Run tests"
	@echo "  lint         - Run linter"
	@echo "  format       - Format Go code"

# Production commands
build:
	@echo "Building production Docker image..."
	docker-compose build

start:
	@echo "Starting production containers..."
	docker-compose up -d

stop:
	@echo "Stopping containers..."
	docker-compose stop

restart:
	@echo "Restarting containers..."
	docker-compose restart

down:
	@echo "Stopping and removing containers..."
	docker-compose down

logs:
	@echo "Showing container logs..."
	docker-compose logs -f

clean:
	@echo "Removing containers, networks, images, and volumes..."
	docker-compose down --volumes --rmi all

clean-data:
	@echo "Removing only volumes..."
	docker-compose down --volumes

# Development commands
dev:
	@echo "Starting development environment with hot reload..."
	docker-compose -f docker-compose.dev.yml up

dev-build:
	@echo "Building development Docker image..."
	docker-compose -f docker-compose.dev.yml build

dev-logs:
	@echo "Showing development container logs..."
	docker-compose -f docker-compose.dev.yml logs -f

dev-clean:
	@echo "Cleaning development environment..."
	docker-compose -f docker-compose.dev.yml down --volumes --rmi all

dev-data-clean:
	@echo "Cleaning development environment..."
	docker-compose -f docker-compose.dev.yml down --volumes

# Code quality commands
test:
	@echo "Running tests..."
	go test ./...

test-cov:
	@echo "Running tests with coverage report..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

test-short:
	@echo "Running short tests..."
	go test -v -short ./...

lint:
	@echo "Running linter..."
	golangci-lint run

format:
	@echo "Formatting Go code..."
	go fmt ./...

# Utility commands
generate-key:
	@echo "Generating encryption key..."
	@openssl rand -base64 24

setup-env:
	@echo "Setting up environment file..."
	@if [ ! -f .env ]; then \
		cp env.example .env; \
		echo "Created .env file from env.example"; \
		echo "Please update the values in .env file"; \
	else \
		echo ".env file already exists"; \
	fi

# Quick development start (alias for dev)
dev-start: dev

# Quick production start (alias for start)
prod-start: start

# Show status
status:
	@echo "Container status:"
	docker-compose ps

dev-status:
	@echo "Development container status:"
	docker-compose -f docker-compose.dev.yml ps

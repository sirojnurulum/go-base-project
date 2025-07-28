# ==============================================================================
# Makefile for beresin-backend
# ==============================================================================

# --- Variables ---
# Nama binary yang akan dihasilkan
BINARY_NAME=beresin-backend-binary
# Path ke file main.go
CMD_PATH=cmd/api/main.go

# Muat variabel dari .env agar dapat digunakan oleh perintah make (cth: DATABASE_URL untuk goose)
ifneq (,$(wildcard ./.env))
	include .env
	export
endif

# --- Default Goal ---
.PHONY: help
help: ## Tampilkan pesan bantuan ini
	@echo "Usage: make <target>"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

# ==============================================================================
# DEVELOPMENT
# ==============================================================================

.PHONY: run
run: ## Jalankan aplikasi secara lokal untuk pengembangan
	@echo ">> Running application..."
	@go run $(CMD_PATH)

.PHONY: dev
dev: ## Jalankan aplikasi dengan hot-reload untuk pengembangan (menggunakan air)
	@command -v air >/dev/null 2>&1 || { echo ">> 'air' is not installed. Please run 'make air-install' first."; exit 1; }
	@echo ">> Starting application with hot-reload..."
	@air

.PHONY: air-install
air-install: ## Install air, alat hot-reload untuk Go
	@echo ">> Installing/Updating air..."
	@go install github.com/cosmtrek/air@latest


.PHONY: test
test: ## Jalankan semua unit test
	@echo ">> Running tests..."
	@go test -v ./...

.PHONY: tidy
tidy: ## Rapikan dependensi Go (go mod tidy)
	@echo ">> Tidying Go modules..."
	@go mod tidy

# ==============================================================================
# BUILD & CLEANUP
# ==============================================================================

.PHONY: build
build: ## Kompilasi aplikasi menjadi satu file binary
	@echo ">> Building binary..."
	@CGO_ENABLED=0 go build -o $(BINARY_NAME) -v $(CMD_PATH)

.PHONY: clean
clean: ## Hapus file binary yang sudah di-build
	@echo ">> Cleaning up..."
	@rm -f $(BINARY_NAME)

# ==============================================================================
# DATABASE (Goose)
# ==============================================================================
# Membutuhkan goose terinstall (`go install github.com/pressly/goose/v3/cmd/goose@latest`)

.PHONY: migrate-create
migrate-create: ## Buat file migrasi baru. Penggunaan: make migrate-create name=nama_migrasi
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create name=<migration_name>"; \
		exit 1; \
	fi
	@echo ">> Creating migration: $(name)"
	@goose -dir "migrations" create $(name) sql

.PHONY: migrate-up
migrate-up: ## Jalankan semua migrasi yang tertunda (up)
	@echo ">> Applying all up migrations..."
	@goose -dir "migrations" postgres "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down: ## Batalkan migrasi terakhir (down)
	@echo ">> Rolling back the last migration..."
	@goose -dir "migrations" postgres "$(DATABASE_URL)" down

.PHONY: migrate-status
migrate-status: ## Tampilkan status dari semua migrasi
	@echo ">> Checking migration status..."
	@goose -dir "migrations" postgres "$(DATABASE_URL)" status -v

# ==============================================================================
# DOCUMENTATION (Swagger)
# ==============================================================================
# Membutuhkan swag terinstall (`go install github.com/swaggo/swag/cmd/swag@latest`)

.PHONY: swag
swag: ## Generate ulang dokumentasi Swagger
	@echo ">> Generating Swagger docs..."
	@swag init -g cmd/api/main.go


.PHONY: db-reset
db-reset: ## Hapus semua data (DROP SCHEMA public CASCADE) dan buat ulang skema. Membutuhkan konfirmasi.
	@read -p "ARE YOU SURE you want to drop the entire public schema? This is irreversible. (y/n) " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo ">> Dropping and recreating public schema..."; \
		psql "$(DATABASE_URL)" -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"; \
		echo ">> Database has been reset. Run 'make run' or 'make migrate-up' to re-apply migrations."; \
	else \
		echo ">> Reset cancelled."; \
	fi


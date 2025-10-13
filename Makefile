APP_NAME := tinylink-backend
MAIN := ./cmd/server
MIGRATIONS_DIR := internal/db/migrations

.PHONY: run build fmt lint test clean

run:
	@echo "🚀 Starting $(APP_NAME)..."
	go run $(MAIN)

build: 
	@echo "🔧 Building $(APP_NAME)..."
	go build -o bin/$(APP_NAME) $(MAIN)

test:
	go test ./... -v

clean:
	rm -rf bin


.PHONY: swagger

swagger:
	@echo "Generating swagger docs"
	swag init --dir ./cmd/server,./internal/handler/tinylink,./internal/domain/tinylink,./pkg/jsonutil --output ./internal/docs

# 	swag init -g ./cmd/server --parseDependency --parseInternal --output ./internal/docs

# .PHONY: migrate-up migrate-down
# migrate-up:
# 	@echo "⬆️  Applying migrations..."
# 	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up
# migrate-down:
# 	@echo "⬇️  Rollbacking migrations..."
# 	migrate -path $(MIGRATIONS_DIR) -databse "$(DB_URL)" down 1
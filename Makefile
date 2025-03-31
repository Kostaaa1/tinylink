BIN=./bin/tinylink

build: 
	@echo 'Building...'
	@go build -o $(BIN)

server: build
	@echo 'Starting built server...'
	@$(BIN) $(ARGS)

run: 
	@echo "🚀 Tinylink server started! 🚀"
	@go run ./cmd/api $(ARGS)

start-dev:
	@echo 'Starting dev server...'
	@$(MAKE) --no-print-directory run ARGS="--port=8000 --env=development --limiter-rps=2 --limiter-burst=4 --limiter-enabled=true --redis-addr=localhost:6379 --redis-password=lagaosiprovidnokopas --redis-db=0 --redis-pool-size=10"

start-prod:
	@echo 'Starting prod server...'
	@$(MAKE) --no-print-directory run ARGS="--port=8080 --env=production --limiter-rps=5 --limiter-burst=10 --limiter-enabled=true"

gen-auth-secret:
	openssl ecparam -name prime256v1 -genkey -noout -out private.pem
	openssl ec -in private.pem -pubout -out public.pem
.PHONY: up down build logs tidy migrate-user migrate-order

# ─── Docker ────────────────────────────────────────────────────────────────

up:
	docker compose up -d --build

down:
	docker compose down

build:
	docker compose build

logs:
	docker compose logs -f

logs-%:
	docker compose logs -f $*

restart-%:
	docker compose restart $*

# ─── Development ────────────────────────────────────────────────────────────

tidy:
	cd shared && go mod tidy
	cd api-gateway && go mod tidy
	cd user-service && go mod tidy
	cd order-service && go mod tidy
	cd notification-service && go mod tidy

run-gateway:
	cd api-gateway && go run ./cmd

run-users:
	cd user-service && go run ./cmd

run-orders:
	cd order-service && go run ./cmd

run-notifications:
	cd notification-service && go run ./cmd

# ─── Database ───────────────────────────────────────────────────────────────

migrate-user:
	docker exec -i microservices-sqlserver-1 /opt/mssql-tools18/bin/sqlcmd \
		-S localhost -U sa -P "Senha@123" -d microservices \
		-i /dev/stdin < user-service/migrations/001_create_users.sql

migrate-order:
	docker exec -i microservices-sqlserver-1 /opt/mssql-tools18/bin/sqlcmd \
		-S localhost -U sa -P "Senha@123" -d microservices \
		-i /dev/stdin < order-service/migrations/001_create_orders.sql

migrate-create-db:
	docker exec microservices-sqlserver-1 /opt/mssql-tools18/bin/sqlcmd \
		-S localhost -U sa -P "Senha@123" \
		-Q "IF NOT EXISTS (SELECT name FROM sys.databases WHERE name='microservices') CREATE DATABASE microservices"

migrate: migrate-create-db migrate-user migrate-order

# ─── Testing ────────────────────────────────────────────────────────────────

test:
	cd shared && go test ./...
	cd user-service && go test ./...
	cd order-service && go test ./...

# ─── Quick smoke test ───────────────────────────────────────────────────────

smoke:
	@echo "==> Health checks"
	curl -sf http://localhost:8080/health | jq .
	@echo ""
	@echo "==> Register user"
	curl -sf -X POST http://localhost:8080/api/v1/users \
		-H "Content-Type: application/json" \
		-d '{"name":"John Doe","email":"john@example.com","password":"secret123"}' | jq .

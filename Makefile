.PHONY: dev dev-build down logs logs-api logs-web logs-firebase \
        test test-api test-web lint lint-api lint-web \
        clean seed prod-build help run-api run-web

# ─── Desenvolvimento (Docker) ─────────────────────

## Sobe todo o ambiente local (API + Web + Firebase Emulators)
dev:
	docker compose up

## Rebuilda imagens e sobe o ambiente
dev-build:
	docker compose up --build

## Para todos os containers
down:
	docker compose down

## Para e remove volumes (reset completo)
clean:
	docker compose down -v

# ─── Desenvolvimento (local, sem Docker) ──────────

## Roda a API Go localmente
run-api:
	cd api && go run cmd/server/main.go

## Roda o frontend React localmente
run-web:
	cd web && npm run dev

# ─── Logs ─────────────────────────────────────────

logs:
	docker compose logs -f

logs-api:
	docker compose logs -f api

logs-web:
	docker compose logs -f web

logs-firebase:
	docker compose logs -f firebase

# ─── Testes ───────────────────────────────────────

test-api:
	cd api && go test ./... -v -count=1

test-web:
	cd web && npm test

test: test-api

# ─── Linting ──────────────────────────────────────

lint-api:
	cd api && golangci-lint run ./...

lint-web:
	cd web && npm run lint

lint: lint-api lint-web

# ─── Utilitários ──────────────────────────────────

## Abre shell no container da API
shell-api:
	docker compose exec api sh

## Abre shell no container do frontend
shell-web:
	docker compose exec web sh

## Build de produção do Go (teste local)
prod-build:
	docker build -t traceo-api:local ./api

## Roda imagem de produção localmente
prod-run:
	docker run --rm -p 8080:8080 --env-file .env traceo-api:local

# ─── Help ─────────────────────────────────────────

help:
	@echo ""
	@echo "  make dev          Sobe o ambiente completo (Docker)"
	@echo "  make dev-build    Rebuilda e sobe"
	@echo "  make down         Para tudo"
	@echo "  make clean        Para e apaga dados (reset)"
	@echo ""
	@echo "  make run-api      Roda API Go local (sem Docker)"
	@echo "  make run-web      Roda React local (sem Docker)"
	@echo ""
	@echo "  make logs         Logs de todos os serviços"
	@echo "  make test-api     Testes do Go"
	@echo "  make lint         Lint de tudo"
	@echo ""

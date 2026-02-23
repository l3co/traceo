# Fase 0B — Ambiente Local Dockerizado

> **Duração estimada**: 1–2 dias (após Fase 0 concluída)
> **Pré-requisitos**: Docker Desktop instalado, Fase 0 com código inicial criado

---

## Objetivo

Ter **toda a stack rodando localmente com um único comando**. Ao final desta fase:

- `make dev` sobe Go API + React + Firebase Emulators
- Hot-reload no Go (Air) — salva o arquivo, API reinicia automaticamente
- Hot-reload no React (Vite HMR) — mudanças no browser em milissegundos
- Firebase Emulators simulando Firestore, Auth e Storage localmente
- Zero dependência de serviços externos para desenvolver
- `.env.example` documentando todas as variáveis necessárias

---

## Por que dockerizar o ambiente local?

### O problema do "funciona na minha máquina"

Sem Docker, para rodar o projeto você precisa:

1. Instalar Go na versão correta
2. Instalar Node.js na versão correta
3. Instalar Firebase CLI
4. Configurar variáveis de ambiente manualmente
5. Iniciar cada serviço em um terminal separado
6. Lembrar a ordem certa de inicialização
7. Rezar para que as versões de tudo sejam compatíveis

Com Docker:

```bash
make dev
# ☕ Pronto. Tudo rodando.
```

### Benefícios concretos

- **Reprodutibilidade** — qualquer pessoa clona o repo e roda. Não importa se usa macOS, Linux ou Windows.
- **Isolamento** — versões do Go, Node, Firebase CLI estão fixadas nos Dockerfiles. Não poluem o sistema host.
- **Paridade dev/prod** — o mesmo Dockerfile (com variações) roda local e no Cloud Run.
- **Onboarding** — um novo desenvolvedor está produtivo em minutos, não horas.

---

## Arquitetura do Ambiente Local

```
┌─────────────────────────────────────────────────────────────┐
│                    docker-compose.yml                        │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────────┐  │
│  │  api          │  │  web          │  │  firebase          │  │
│  │  (Go + Air)   │  │  (Vite dev)   │  │  (Emulators)       │  │
│  │  :8080        │  │  :5173        │  │  :4000 (UI)        │  │
│  │               │  │               │  │  :8081 (Firestore) │  │
│  │  hot-reload   │  │  HMR          │  │  :9099 (Auth)      │  │
│  │  via Air      │  │  via Vite     │  │  :9199 (Storage)   │  │
│  └──────┬───────┘  └───────┬──────┘  └─────────┬─────────┘  │
│         │                  │                    │             │
│         └──────────────────┼────────────────────┘             │
│                            │                                  │
│                     rede: traceo-net                   │
└─────────────────────────────────────────────────────────────┘
                             │
                        host machine
                     http://localhost:5173  (frontend)
                     http://localhost:8080  (API)
                     http://localhost:4000  (Firebase Emulator UI)
```

### Comunicação entre serviços

Dentro da rede Docker, os serviços se encontram por **hostname** (nome do serviço):

| De | Para | URL interna |
|---|---|---|
| `web` (React) | `api` (Go) | `http://api:8080` |
| `api` (Go) | `firebase` (Emulators) | `firebase:8081` (Firestore) |
| Host (browser) | `web` | `http://localhost:5173` |
| Host (browser) | `api` | `http://localhost:8080` |
| Host (browser) | Firebase UI | `http://localhost:4000` |

---

## Estrutura de Arquivos

```
desaparecidos/
├── api/
│   ├── Dockerfile              ← produção (multi-stage, scratch)
│   ├── Dockerfile.dev          ← desenvolvimento (Air + hot-reload)
│   └── .air.toml               ← configuração do Air (hot-reload Go)
├── web/
│   ├── Dockerfile.dev          ← desenvolvimento (Vite dev server)
│   └── ...
├── firebase/
│   ├── firebase.json           ← configuração dos emulators
│   └── .firebaserc             ← projeto associado
├── docker-compose.yml          ← orquestra tudo
├── docker-compose.prod.yml     ← override para testes de build de produção
├── .env.example                ← template de variáveis de ambiente
├── .env                        ← variáveis reais (gitignored)
└── Makefile                    ← comandos padronizados
```

---

## Arquivos Detalhados

### 1. `docker-compose.yml` — Orquestrador principal

```yaml
services:
  # ─── Go API com hot-reload ─────────────────────────
  api:
    build:
      context: ./api
      dockerfile: Dockerfile.dev
    ports:
      - "8080:8080"
    volumes:
      - ./api:/app
      - go-modules:/go/pkg/mod       # cache de dependências Go
    env_file:
      - .env
    environment:
      - FIRESTORE_EMULATOR_HOST=firebase:8081
      - FIREBASE_AUTH_EMULATOR_HOST=firebase:9099
      - STORAGE_EMULATOR_HOST=firebase:9199
      - ENVIRONMENT=development
    depends_on:
      firebase:
        condition: service_healthy
    networks:
      - traceo-net

  # ─── React com Vite HMR ────────────────────────────
  web:
    build:
      context: ./web
      dockerfile: Dockerfile.dev
    ports:
      - "5173:5173"
    volumes:
      - ./web:/app
      - web-node-modules:/app/node_modules   # evita conflito host vs container
    environment:
      - VITE_API_URL=http://localhost:8080
      - VITE_FIREBASE_PROJECT_ID=traceo-dev
    depends_on:
      - api
    networks:
      - traceo-net

  # ─── Firebase Emulators ────────────────────────────
  firebase:
    image: node:20-alpine
    working_dir: /firebase
    ports:
      - "4000:4000"     # Emulator Suite UI
      - "8081:8081"     # Firestore
      - "9099:9099"     # Authentication
      - "9199:9199"     # Storage
    volumes:
      - ./firebase:/firebase
      - firebase-data:/firebase/data    # persiste dados entre restarts
    command: >
      sh -c "
        npm install -g firebase-tools &&
        firebase emulators:start
          --project traceo-dev
          --import=/firebase/data
          --export-on-exit=/firebase/data
      "
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:4000"]
      interval: 5s
      timeout: 3s
      retries: 10
      start_period: 15s
    networks:
      - traceo-net

networks:
  traceo-net:
    driver: bridge

volumes:
  go-modules:           # cache de módulos Go (evita re-download)
  web-node-modules:     # node_modules isolado do host
  firebase-data:        # dados dos emulators persistidos
```

#### Decisões importantes explicadas

**`volumes: ./api:/app`** — monta o código-fonte local dentro do container. Quando você salva um arquivo no seu editor, o Air (dentro do container) detecta a mudança e recompila automaticamente.

**`go-modules` como named volume** — módulos Go são cacheados em um volume Docker. Sem isso, `go mod download` rodaria a cada restart do container (~30s). Com cache, restart é instantâneo.

**`web-node-modules` como named volume** — `node_modules` é mantido **dentro do container**, não no host. Isso evita conflito entre arquiteturas (macOS ARM vs Linux x86 dentro do container). O diretório `node_modules` do host (se existir) não interfere.

**`depends_on` com `condition: service_healthy`** — a API só inicia **depois** que os Firebase Emulators estão prontos. Sem isso, a API tentaria conectar no Firestore antes dele existir.

**`--import` / `--export-on-exit`** — dados criados nos emulators (usuários de teste, documentos) são **persistidos** entre restarts. Sem isso, você perderia todos os dados de teste toda vez que rodasse `docker compose down`.

---

### 2. `api/Dockerfile.dev` — Go com Air (hot-reload)

```dockerfile
FROM golang:1.22-alpine

# Air: hot-reload para Go
RUN go install github.com/air-verse/air@latest

WORKDIR /app

# Copia go.mod/go.sum primeiro (cache de dependências)
COPY go.mod go.sum ./
RUN go mod download

# Não copia o código — ele vem via volume mount
# Isso permite que mudanças locais sejam refletidas instantaneamente

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
```

#### Por que Air?

Sem Air, o ciclo de desenvolvimento seria:

1. Editar código
2. `Ctrl+C` no terminal
3. `go run cmd/server/main.go`
4. Esperar compilar (~2-3s)
5. Testar

Com Air:

1. Editar código
2. ✅ API já reiniciou automaticamente (~1s)

Air monitora arquivos `.go` e recompila quando detecta mudanças. É o equivalente do `nodemon` para Node.js.

---

### 3. `api/.air.toml` — Configuração do Air

```toml
root = "."
tmp_dir = "tmp"

[build]
  bin = "./tmp/server"
  cmd = "go build -o ./tmp/server ./cmd/server/main.go"
  delay = 1000                        # ms antes de rebuildar (debounce)
  exclude_dir = ["tmp", "vendor", "docs"]
  exclude_regex = ["_test\\.go$"]
  include_ext = ["go", "toml"]
  kill_delay = 500                    # ms para matar processo anterior

[log]
  time = false

[color]
  build = "yellow"
  runner = "green"
  main = "magenta"
```

#### O que cada campo faz

- **`cmd`** — comando de build. Compila o `main.go` e coloca o binário em `tmp/`.
- **`delay`** — espera 1 segundo após detectar mudança antes de rebuildar. Evita rebuilds múltiplos quando você salva vários arquivos em sequência.
- **`exclude_dir`** — não monitora esses diretórios (evita loops e rebuilds desnecessários).
- **`exclude_regex`** — ignora arquivos de teste (mudar um test não precisa reiniciar o servidor).
- **`include_ext`** — só monitora arquivos `.go` e `.toml`.

---

### 4. `web/Dockerfile.dev` — React com Vite dev server

```dockerfile
FROM node:20-alpine

WORKDIR /app

# Copia package.json/lock primeiro (cache de dependências)
COPY package.json package-lock.json ./
RUN npm ci

# Não copia o código — ele vem via volume mount

EXPOSE 5173

# --host 0.0.0.0: aceita conexões de fora do container
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]
```

#### Por que `--host 0.0.0.0`?

Por padrão, o Vite escuta apenas em `localhost` (127.0.0.1). Dentro de um container Docker, `localhost` significa "dentro do container" — o browser no host não consegue acessar.

`--host 0.0.0.0` faz o Vite aceitar conexões de qualquer interface, incluindo a ponte Docker.

---

### 5. `firebase/firebase.json` — Configuração dos Emulators

```json
{
  "emulators": {
    "auth": {
      "port": 9099,
      "host": "0.0.0.0"
    },
    "firestore": {
      "port": 8081,
      "host": "0.0.0.0"
    },
    "storage": {
      "port": 9199,
      "host": "0.0.0.0"
    },
    "ui": {
      "enabled": true,
      "port": 4000,
      "host": "0.0.0.0"
    }
  },
  "firestore": {
    "rules": "firestore.rules"
  },
  "storage": {
    "rules": "storage.rules"
  }
}
```

### 6. `firebase/firestore.rules` — Regras iniciais

```
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    // Dev: permite tudo. Produção terá regras reais.
    match /{document=**} {
      allow read, write: if true;
    }
  }
}
```

### 7. `firebase/storage.rules` — Regras iniciais

```
rules_version = '2';
service firebase.storage {
  match /b/{bucket}/o {
    match /{allPaths=**} {
      allow read, write: if true;
    }
  }
}
```

> ⚠️ Essas regras são **só para desenvolvimento local**. Em produção (Fase 1+), teremos regras restritivas com validação de autenticação.

---

### 8. `.env.example` — Template de variáveis

```bash
# ─── Ambiente ────────────────────────────────────────
ENVIRONMENT=development
PORT=8080

# ─── Firebase ────────────────────────────────────────
FIREBASE_PROJECT_ID=traceo-dev
# Em dev, os emulators são usados automaticamente.
# Em produção, defina o path para o service account:
# GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json

# ─── Resend (email) ─────────────────────────────────
# Não obrigatório em dev — emails são logados no console.
# RESEND_API_KEY=re_xxxxxxxxxxxx
# RESEND_FROM_EMAIL=noreply@traceo.me

# ─── WhatsApp Business ──────────────────────────────
# Não obrigatório em dev — notificações são logadas no console.
# WHATSAPP_TOKEN=xxxxxxxx
# WHATSAPP_PHONE_ID=xxxxxxxx

# ─── Telegram ───────────────────────────────────────
# Não obrigatório em dev — mensagens são logadas no console.
# TELEGRAM_BOT_TOKEN=xxxxxxxx
# TELEGRAM_CHAT_ID=xxxxxxxx

# ─── Google Maps ────────────────────────────────────
# Necessário para geocoding e mapas no frontend.
# VITE_GOOGLE_MAPS_API_KEY=AIzaxxxxxxxx

# ─── Gemini AI ──────────────────────────────────────
# Não obrigatório em dev — AI features retornam mocks.
# GEMINI_API_KEY=xxxxxxxx
```

#### Filosofia: zero config para dev

Note que **nenhuma variável é obrigatória para rodar localmente**. O sistema detecta `ENVIRONMENT=development` e:

- Usa Firebase Emulators em vez de Firebase real
- Loga emails/notificações no console em vez de enviar
- Retorna mocks para features de AI
- Aceita qualquer origem no CORS

Isso garante que `make dev` funciona **sem configurar absolutamente nada**.

---

### 9. `Makefile` — Comandos padronizados

```makefile
.PHONY: dev dev-build down logs logs-api logs-web logs-firebase \
        test test-api test-web lint lint-api lint-web \
        clean seed prod-build help

# ─── Desenvolvimento ──────────────────────────────────

## Sobe todo o ambiente local (API + Web + Firebase Emulators)
dev:
	docker compose up

## Rebuilda imagens e sobe o ambiente (usar após mudar Dockerfile ou dependências)
dev-build:
	docker compose up --build

## Para todos os containers
down:
	docker compose down

## Para e remove volumes (reset completo — perde dados dos emulators)
clean:
	docker compose down -v

# ─── Logs ─────────────────────────────────────────────

## Mostra logs de todos os serviços
logs:
	docker compose logs -f

## Mostra logs só da API Go
logs-api:
	docker compose logs -f api

## Mostra logs só do frontend React
logs-web:
	docker compose logs -f web

## Mostra logs só do Firebase Emulators
logs-firebase:
	docker compose logs -f firebase

# ─── Testes ───────────────────────────────────────────

## Roda testes do Go
test-api:
	docker compose exec api go test ./... -v -count=1

## Roda testes do React
test-web:
	docker compose exec web npm test

## Roda todos os testes
test: test-api test-web

# ─── Linting ──────────────────────────────────────────

## Lint do Go (golangci-lint)
lint-api:
	docker compose exec api golangci-lint run ./...

## Lint do React (eslint)
lint-web:
	docker compose exec web npm run lint

## Lint de tudo
lint: lint-api lint-web

# ─── Utilitários ──────────────────────────────────────

## Abre um shell dentro do container da API
shell-api:
	docker compose exec api sh

## Abre um shell dentro do container do frontend
shell-web:
	docker compose exec web sh

## Gera documentação Swagger
swagger:
	docker compose exec api swag init -g cmd/server/main.go -o docs/swagger

## Seed: popula emulators com dados de teste
seed:
	docker compose exec api go run cmd/seed/main.go

# ─── Produção (teste local) ──────────────────────────

## Builda imagem de produção do Go (para testar localmente)
prod-build:
	docker build -t desaparecidos-api:local ./api

## Roda imagem de produção localmente
prod-run:
	docker run --rm -p 8080:8080 --env-file .env desaparecidos-api:local

# ─── Help ─────────────────────────────────────────────

## Mostra esta ajuda
help:
	@echo ""
	@echo "Comandos disponíveis:"
	@echo ""
	@echo "  make dev          Sobe o ambiente completo"
	@echo "  make dev-build    Rebuilda e sobe (após mudar deps)"
	@echo "  make down         Para tudo"
	@echo "  make clean        Para e apaga dados (reset total)"
	@echo ""
	@echo "  make logs         Logs de todos os serviços"
	@echo "  make logs-api     Logs só da API"
	@echo "  make logs-web     Logs só do frontend"
	@echo ""
	@echo "  make test         Roda todos os testes"
	@echo "  make test-api     Testes do Go"
	@echo "  make test-web     Testes do React"
	@echo ""
	@echo "  make lint         Lint de tudo"
	@echo "  make shell-api    Shell no container Go"
	@echo "  make shell-web    Shell no container React"
	@echo "  make swagger      Gera doc Swagger"
	@echo ""
```

---

## Fluxo de Trabalho Diário

### Primeiro setup (uma vez)

```bash
# 1. Clonar o repositório
git clone https://github.com/seu-usuario/desaparecidos.git
cd desaparecidos

# 2. Criar .env a partir do template
cp .env.example .env

# 3. Subir tudo
make dev
```

Na primeira execução, Docker vai:
1. Baixar imagens base (golang, node, etc.) — ~2-3 minutos
2. Instalar dependências Go (`go mod download`) — ~1 minuto
3. Instalar dependências Node (`npm ci`) — ~1 minuto
4. Iniciar Firebase Emulators — ~10 segundos
5. Compilar e iniciar a API Go — ~3 segundos
6. Iniciar o Vite dev server — ~2 segundos

**Total primeira vez: ~5 minutos. Execuções seguintes: ~10 segundos.**

### Desenvolvimento normal

```bash
# Terminal 1: subir ambiente (roda em foreground com logs)
make dev

# Terminal 2: (opcional) rodar testes enquanto desenvolve
make test-api

# Quando terminar:
# Ctrl+C no terminal 1, ou:
make down
```

### Quando rebuildar?

| Situação | Comando |
|---|---|
| Editou código Go ou React | Nada — hot-reload automático |
| Adicionou dependência Go (`go get`) | `make dev-build` |
| Adicionou dependência Node (`npm install`) | `make dev-build` |
| Mudou Dockerfile.dev | `make dev-build` |
| Mudou docker-compose.yml | `make down && make dev` |
| Quer resetar dados dos emulators | `make clean && make dev` |

---

## Tarefas Detalhadas

### Tarefa 0B.1 — Instalar Docker Desktop

```bash
# macOS
brew install --cask docker

# Verificar
docker --version
docker compose version
```

> Docker Desktop inclui Docker Engine + Docker Compose v2. Não é necessário instalar `docker-compose` separadamente (v1 legado). O comando é `docker compose` (sem hífen).

### Tarefa 0B.2 — Criar Dockerfile.dev para Go

Criar `api/Dockerfile.dev` com Go + Air conforme seção acima.

### Tarefa 0B.3 — Criar `.air.toml`

Criar `api/.air.toml` conforme seção acima.

### Tarefa 0B.4 — Criar Dockerfile.dev para React

Criar `web/Dockerfile.dev` conforme seção acima.

### Tarefa 0B.5 — Configurar Firebase Emulators

Criar `firebase/firebase.json`, `firebase/firestore.rules` e `firebase/storage.rules` conforme seções acima.

### Tarefa 0B.6 — Criar docker-compose.yml

Criar `docker-compose.yml` na raiz conforme seção acima.

### Tarefa 0B.7 — Criar `.env.example` e `.env`

```bash
cp .env.example .env
# .env já está no .gitignore
```

### Tarefa 0B.8 — Criar Makefile

Criar `Makefile` na raiz conforme seção acima.

### Tarefa 0B.9 — Validar o ambiente

Checklist de validação:

```bash
# 1. Subir tudo
make dev

# 2. Verificar que a API responde
curl http://localhost:8080/api/v1/health
# → {"status": "ok"}

# 3. Verificar que o frontend carrega
# Abrir http://localhost:5173 no browser

# 4. Verificar que os emulators estão rodando
# Abrir http://localhost:4000 no browser (Firebase Emulator UI)

# 5. Verificar hot-reload do Go
# Editar qualquer .go → logs mostram "rebuilding..."

# 6. Verificar HMR do React
# Editar qualquer .tsx → browser atualiza instantaneamente

# 7. Verificar que testes rodam
make test-api
```

### Tarefa 0B.10 — Atualizar `.gitignore`

Garantir que o `.gitignore` na raiz contenha:

```
# Docker
.env

# Firebase Emulator data
firebase/data/

# Go Air temp
api/tmp/
```

---

## Troubleshooting

### "Port already in use"

```bash
# Verificar o que está usando a porta
lsof -i :8080

# Parar containers antigos
make down
```

### "Permission denied" em volumes

```bash
# macOS: garantir que Docker Desktop tem acesso ao diretório
# Docker Desktop → Settings → Resources → File Sharing
```

### "go mod download" muito lento

Verificar se o volume `go-modules` existe:

```bash
docker volume ls | grep go-modules
```

Se não existir, o Docker recria a cada build. O `docker-compose.yml` já define o named volume para cache.

### Firebase Emulators não iniciam

```bash
# Ver logs detalhados
make logs-firebase

# Reset completo dos emulators
make clean && make dev
```

### Hot-reload não funciona (Go)

Verificar se o Air está monitorando os arquivos:

```bash
make logs-api
# Deve mostrar: "watching ." e listar diretórios
```

Se não funcionar, verificar se o volume está montado:

```bash
docker compose exec api ls -la /app
# Deve mostrar os arquivos do projeto
```

---

## Diferença entre Dev e Produção

| Aspecto | Dev (docker-compose) | Prod (Cloud Run) |
|---|---|---|
| **Go runtime** | `golang:1.22-alpine` (SDK completo) | `scratch` (só binário) |
| **Hot-reload** | Air monitora arquivos | Não existe — binário estático |
| **Firebase** | Emulators locais | Firebase real (GCP) |
| **Variáveis** | `.env` file | Secret Manager |
| **CORS** | Permite tudo | Só origens configuradas |
| **Logs** | stdout no terminal | Cloud Logging (JSON) |
| **Imagem** | ~800MB (Go SDK) | ~15MB (binário) |
| **Notificações** | Logadas no console | Enviadas de verdade |

O **Dockerfile de produção** (`api/Dockerfile`) continua sendo o multi-stage build da Fase 0 — ele compila o binário e gera uma imagem mínima. O `Dockerfile.dev` é **só para desenvolvimento**.

---

## Entregáveis da Fase 0B

- [ ] Docker Desktop instalado e funcionando
- [ ] `api/Dockerfile.dev` com Go + Air
- [ ] `api/.air.toml` configurado
- [ ] `web/Dockerfile.dev` com Vite dev server
- [ ] `firebase/firebase.json` + regras dos emulators
- [ ] `docker-compose.yml` orquestrando os 3 serviços
- [ ] `.env.example` documentando todas as variáveis
- [ ] `Makefile` com comandos padronizados
- [ ] `make dev` sobe tudo e responde em `localhost:8080` e `localhost:5173`
- [ ] Hot-reload funcionando no Go (Air) e React (Vite HMR)
- [ ] Firebase Emulator UI acessível em `localhost:4000`
- [ ] `.gitignore` atualizado

---

## Próxima Fase

→ [FASE_01_AUTH_USUARIO.md](./FASE_01_AUTH_USUARIO.md) — Autenticação completa + CRUD de Usuário

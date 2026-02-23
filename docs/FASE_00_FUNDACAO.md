# Fase 0 — Fundação & Setup

> **Duração estimada**: 2 semanas
> **Pré-requisitos**: Go instalado, Node.js instalado, conta Google Cloud

---

## Objetivo

Ter o ambiente de desenvolvimento funcional e os **primeiros bytes de Go rodando**. Ao final desta fase, você terá:

- Um monorepo organizado
- Um servidor Go respondendo `{"status": "ok"}` no endpoint `/api/v1/health`
- Um app React renderizando uma página inicial
- Firebase configurado (Firestore + Auth + Storage)
- Docker para desenvolvimento local

---

## Conceitos Go que você vai aprender nesta fase

### 1. O que é Go e por que ele existe?

Go foi criado em 2009 no Google por **Rob Pike**, **Ken Thompson** (criador do Unix e UTF-8) e **Robert Griesemer**. Eles estavam frustrados com a complexidade do C++ e a lentidão de compilação.

Go foi projetado para ser:
- **Simples** — a linguagem inteira cabe em uma especificação curta. Não tem herança, não tem generics complexos, não tem exceções.
- **Rápido para compilar** — projetos enormes compilam em segundos
- **Fácil de ler** — Go tem **uma forma correta** de formatar código (`gofmt`). Não existe discussão sobre tabs vs espaços, onde colocar a chave, etc.
- **Ótimo para servidores** — goroutines tornam concorrência trivial

#### A filosofia do Go em uma frase:

> *"Less is exponentially more."* — Rob Pike

Se você vem de Java/Python e sente que Go "falta coisa" (herança, exceções, decorators), é **intencional**. Go remove features para forçar simplicidade.

---

### 2. Packages e Modules — como Go organiza código

Em Python, você tem `import flask`. Em Go, o sistema é diferente.

#### Module

Um **module** é o seu projeto inteiro. Ele é definido pelo arquivo `go.mod`:

```go
module github.com/seu-usuario/traceo-api

go 1.22
```

O nome do module é geralmente a URL do repositório. Isso permite que outros projetos importem seu código.

#### Package

Um **package** é uma pasta com arquivos `.go`. Todos os arquivos numa mesma pasta **pertencem ao mesmo package**.

```
internal/
  domain/
    user/           ← package "user"
      entity.go     ← pertence ao package "user"
      service.go    ← pertence ao package "user"
```

No topo de cada arquivo:

```go
package user  // declara que este arquivo pertence ao package "user"
```

**Regra importante**: em Go, **visibilidade é controlada pela primeira letra**:
- `User` (maiúscula) → **exportado** (público) — qualquer package pode acessar
- `user` (minúscula) → **não exportado** (privado ao package)

Não existe `public`, `private`, `protected`. Só maiúscula vs minúscula. Simples.

```go
type User struct {     // Exportado — outros packages veem isso
    Name  string       // Exportado
    email string       // NÃO exportado — só o package "user" acessa
}
```

#### Por que isso importa para nós?

Na nossa estrutura, o package `handler` vai importar o package `domain/user`:

```go
package handler

import "github.com/seu-usuario/traceo-api/internal/domain/user"

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
    u := user.User{Name: "João"}  // OK — User é exportado
    // u.email = "x"              // ERRO — email não é exportado
}
```

---

### 3. O diretório `internal/` — uma convenção especial do Go

Em Go, se você coloca código dentro de `internal/`, **nenhum package externo pode importar esse código**. É uma regra enforçada pelo compilador.

```
traceo-api/
├── cmd/              ← executáveis (main.go)
├── internal/         ← código privado do projeto (ninguém de fora importa)
│   ├── domain/
│   ├── handler/
│   └── infrastructure/
├── pkg/              ← código que PODE ser importado por outros projetos
├── go.mod
└── go.sum
```

**Por que usar `internal/`?**

Porque queremos que nossas regras de domínio (`domain/user`) **nunca** sejam acessadas diretamente por código externo. É uma barreira arquitetural enforçada pelo compilador.

`pkg/` é para utilitários genéricos que poderiam ser usados em outros projetos (ex: um helper de slug, um validador de data).

---

### 4. `main.go` — o ponto de entrada

Todo programa Go começa com `package main` e uma função `main()`:

```go
package main

import "fmt"

func main() {
    fmt.Println("Olá, mundo!")
}
```

Para executar:
```bash
go run cmd/server/main.go
```

Para compilar em binário:
```bash
go build -o server cmd/server/main.go
./server  # executa o binário compilado
```

O binário é **estático** — não precisa de runtime, não precisa de dependências instaladas. Copia o arquivo para qualquer máquina Linux e roda. Isso é o que torna Go perfeito para Docker.

---

### 5. Structs — o "modelo" do Go

Go não tem classes. Tem **structs** (estruturas de dados) e **métodos associados**.

```go
// Definindo uma struct
type User struct {
    ID    string
    Name  string
    Email string
}

// Definindo um método na struct
func (u User) FullInfo() string {
    return u.Name + " (" + u.Email + ")"
}

// Usando
user := User{
    ID:    "123",
    Name:  "João",
    Email: "joao@email.com",
}
fmt.Println(user.FullInfo()) // "João (joao@email.com)"
```

Diferença fundamental vs Python/Java:
- **Não tem herança** — Go usa **composição** (embedding)
- **Não tem construtor** — você cria a struct diretamente ou faz uma função `NewUser()`
- **Não tem `self`/`this`** — o receiver `(u User)` é explícito

Vamos nos aprofundar em structs e métodos na Fase 1. Por agora, basta saber que é assim que Go modela dados.

---

### 6. Error handling — a filosofia Go que mais surpreende

Em Python, você usa `try/except`. Em Java, `try/catch`. Em Go, **não existem exceções**.

Go retorna erros como **valores normais**:

```go
// Função que pode falhar
func findUser(id string) (User, error) {
    if id == "" {
        return User{}, fmt.Errorf("id cannot be empty")
    }
    // ... busca no banco
    return user, nil  // nil = sem erro
}

// Quem chama DEVE verificar o erro
user, err := findUser("123")
if err != nil {
    // trata o erro
    log.Printf("failed to find user: %v", err)
    return
}
// usa user normalmente
```

**Por que Go faz isso?**

1. **Erros são explícitos** — você vê exatamente onde cada erro pode acontecer
2. **Sem surpresas** — exceções em Python/Java podem "voar" por 10 frames de stack sem tratamento. Em Go, o erro está ali na sua frente.
3. **Forçar tratamento** — se você ignora o `err`, o linter reclama

Isso parece verboso no começo (e é). Mas depois de usar por um tempo, você percebe que o código fica **extremamente previsível**. Sem exceções surpresa quebrando em produção.

---

## Estrutura do Monorepo

```
desaparecidos/
├── api/                          ← Projeto Go (backend)
│   ├── cmd/
│   │   └── server/
│   │       └── main.go           ← entry point
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go         ← leitura de variáveis de ambiente
│   │   ├── domain/
│   │   │   ├── user/
│   │   │   ├── missing/
│   │   │   ├── sighted/
│   │   │   └── homeless/
│   │   ├── handler/
│   │   │   ├── health_handler.go
│   │   │   └── middleware/
│   │   └── infrastructure/
│   │       ├── firestore/
│   │       ├── storage/
│   │       ├── notification/
│   │       └── auth/
│   ├── pkg/
│   │   ├── slug/
│   │   ├── validator/
│   │   └── dateutil/
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
│
├── web/                          ← Projeto React (frontend)
│   ├── src/
│   │   ├── app/
│   │   ├── features/
│   │   ├── shared/
│   │   ├── main.tsx
│   │   └── index.css
│   ├── public/
│   ├── package.json
│   ├── tsconfig.json
│   ├── tailwind.config.ts
│   └── vite.config.ts
│
├── docs/                         ← Documentação (estes arquivos)
│   ├── ROADMAP.md
│   ├── FASE_00_FUNDACAO.md
│   └── ...
│
├── .gitignore
├── docker-compose.yml            ← Firebase Emulators + dev tools
└── README.md
```

### Por que essa estrutura e não outra?

**`cmd/server/main.go`** — Convenção Go para executáveis. Se no futuro tivéssemos um CLI ou um worker, ficaria em `cmd/worker/main.go`. Cada subdiretório de `cmd/` é um binário separado.

**`internal/`** — Tudo que é específico deste projeto. O compilador Go impede importação externa.

**`pkg/`** — Utilitários genéricos. Poderiam ser extraídos para uma lib separada no futuro.

**`web/` ao lado de `api/`** — Não dentro. São projetos independentes com toolchains diferentes (Go vs Node). Ficam lado a lado no monorepo.

---

## Tarefas Detalhadas

### Backend (Go)

#### Tarefa 0.1 — Instalar Go

```bash
# macOS com Homebrew
brew install go

# Verificar
go version
# go version go1.22.x darwin/arm64
```

Configurar no shell (~/.zshrc):
```bash
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

#### Tarefa 0.2 — Inicializar o module Go

```bash
mkdir -p desaparecidos/api
cd desaparecidos/api
go mod init github.com/seu-usuario/traceo-api
```

Isso cria o `go.mod`. É o equivalente do `package.json` (Node) ou `requirements.txt` (Python).

#### Tarefa 0.3 — Criar o `main.go` com health check

Este será o primeiro código Go que escrevemos. Vou explicar cada linha quando implementarmos juntos.

O objetivo é:
- Criar um servidor HTTP na porta 8080
- Ter um endpoint `GET /api/v1/health` que retorna `{"status": "ok"}`
- Usar o router **Chi** (vou explicar por que na implementação)

#### Tarefa 0.4 — Criar o `config.go`

Leitura de variáveis de ambiente de forma segura. Conceitos:
- `os.Getenv()` — como Go lê variáveis de ambiente
- Validação na inicialização (fail fast)

#### Tarefa 0.5 — Criar o `Dockerfile`

Multi-stage build:
```dockerfile
# Stage 1: Build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server cmd/server/main.go

# Stage 2: Run
FROM alpine:3.19
COPY --from=builder /app/server /server
EXPOSE 8080
CMD ["/server"]
```

**Por que multi-stage?**
- Stage 1 tem o SDK Go (~800MB) para compilar
- Stage 2 tem só o Alpine (~5MB) + binário (~15MB)
- Imagem final: **~20MB** (vs ~200MB+ para Python/Node)

Isso significa cold start **sub-segundo** no Cloud Run.

### Frontend (React)

#### Tarefa 0.6 — Inicializar o projeto React

```bash
cd desaparecidos
npm create vite@latest web -- --template react-ts
cd web
npm install
npm install -D tailwindcss @tailwindcss/vite
npm install lucide-react
```

#### Tarefa 0.7 — Configurar Tailwind + shadcn/ui

Instalar e configurar os componentes base do design system.

#### Tarefa 0.8 — Criar página inicial (loading/splash)

Um componente React simples que chama `GET /api/v1/health` e mostra o status.

### Infraestrutura

#### Tarefa 0.9 — Criar projeto Firebase

1. Ir ao [Firebase Console](https://console.firebase.google.com)
2. Criar projeto "desaparecidos"
3. Ativar Firestore, Authentication e Storage
4. Gerar credenciais de serviço (service account JSON)

#### Tarefa 0.10 — Configurar Firebase Emulators

Para desenvolver **localmente** sem depender do Firebase real:

```bash
npm install -g firebase-tools
firebase init emulators
```

A configuração completa do ambiente dockerizado (docker-compose com API + frontend + emulators, hot-reload, Makefile, `.env.example`) está na **Fase 0B**:

→ [FASE_00B_DOCKER_LOCAL.md](./FASE_00B_DOCKER_LOCAL.md) — Ambiente Local Dockerizado

#### Tarefa 0.11 — Configurar linting e formatação

**Go:**
```bash
# gofmt já vem com Go — formata automaticamente
gofmt -w .

# golangci-lint — linter completo
brew install golangci-lint
```

**React:**
```bash
npm install -D eslint prettier
```

---

## O que instalar no Chi (e por que Chi?)

### Por que Chi e não outro router?

Go tem vários routers HTTP. Os mais populares:

| Router | Estilo | Prós | Contras |
|---|---|---|---|
| **net/http** (stdlib) | Minimalista | Zero dependências, é o padrão | Sem path params (`/user/:id`), sem middleware chain |
| **Chi** | Composável | Compatível com `net/http`, leve, middleware chain | Menos features que um "framework" |
| **Echo** | Framework | Muitas features prontas, binding, validation | Menos idiomático, API própria |
| **Gin** | Framework | Mais popular, rápido | API própria, menos compatível com stdlib |
| **Fiber** | Framework | Muito rápido, API similar ao Express | Usa fasthttp (incompatível com net/http) |

**Escolhemos Chi porque:**

1. **É compatível com `net/http`** — os handlers Chi são handlers padrão do Go. Se amanhã você quiser trocar Chi por outra coisa, seus handlers continuam funcionando. Echo, Gin e Fiber têm APIs próprias que criam lock-in.

2. **Middleware chain elegante** — Chi permite compor middlewares de forma declarativa:
```go
r := chi.NewRouter()
r.Use(middleware.Logger)      // log de todas as requisições
r.Use(middleware.Recoverer)   // captura panics
r.Use(cors.Handler(corsOpts)) // CORS

r.Route("/api/v1", func(r chi.Router) {
    r.Get("/health", healthHandler.Check)

    r.Route("/users", func(r chi.Router) {
        r.Use(authMiddleware) // só rotas de user precisam de auth
        r.Post("/", userHandler.Create)
        r.Get("/{id}", userHandler.FindByID)
    })
})
```

3. **É leve** — ~1000 linhas de código. Não é um framework, é um router. Você entende todo o código-fonte em uma tarde.

4. **É o mais recomendado pela comunidade Go** para projetos que querem se manter idiomáticos.

#### Comparação com o que você conhece

Se você conhece Flask:
- Flask = framework com template engine, sessions, etc
- Chi = só o roteamento + middleware, como o `Flask.route()` sem o resto

Se você conhece Express.js:
- Chi é o equivalente Go do Express, mas sem o baggage do Node

---

## Documentação de API — Swagger/OpenAPI

Uma API sem documentação é uma API que ninguém usa. Vamos configurar **Swagger** desde o primeiro endpoint.

### swaggo/swag — Swagger automático em Go

Em Go, a ferramenta padrão é o `swaggo/swag`. Ela gera a spec OpenAPI automaticamente a partir de **comentários nos handlers**:

```go
// @Summary      Health check
// @Description  Verifica se a API está no ar
// @Tags         system
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /api/v1/health [get]
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
```

Depois de anotar os handlers, basta rodar:

```bash
# Instalar (uma vez)
go install github.com/swaggo/swag/cmd/swag@latest

# Gerar spec (roda sempre que mudar algo)
swag init -g cmd/server/main.go -o api/docs/swagger
```

Isso gera `swagger.json` + `swagger.yaml` + `docs.go`. Para servir a UI interativa:

```go
import (
    httpSwagger "github.com/swaggo/http-swagger"
    _ "github.com/seuuser/desaparecidos/api/docs/swagger"
)

// No router Chi
r.Get("/swagger/*", httpSwagger.WrapHandler)
```

Resultado: acessar `http://localhost:8080/swagger/` abre uma UI interativa onde você pode ver todos os endpoints, schemas e até testar chamadas direto do browser.

### Por que não Postman collections?

Postman é ótimo, mas a documentação fica **separada** do código. Se muda o endpoint no Go e esquece de atualizar o Postman, a doc fica desatualizada. Com swaggo, a doc **é** o código — se os comentários estão corretos, a spec está correta.

### Quando anotar

A partir da Fase 1, **todo handler novo** deve ter anotações Swagger. Na Fase 0, configuramos o setup e anotamos o health check como exemplo.

---

## Entregáveis da Fase 0

Ao final desta fase, você terá:

- [x] Monorepo criado com `api/`, `web/`, `shared/`, `firebase/`
- [x] `go mod init` executado (Go 1.26, Chi, go-i18n, cors, swaggo)
- [x] `main.go` com servidor HTTP + graceful shutdown + structured logging (slog)
- [x] `GET /api/v1/health` retornando status i18n (PT-BR e EN via `Accept-Language`)
- [x] i18n backend: `go-i18n/v2` com arquivos TOML embeddados, middleware `Accept-Language`
- [x] i18n frontend: `react-i18next` com detecção de idioma do browser + language switcher
- [x] Projeto React inicializado com Vite + TypeScript + Tailwind v4
- [x] shadcn/ui configurado (Button, Card, Badge)
- [x] Página React que chama o health check e alterna idioma
- [x] Firebase emulators configurados (firebase.json + regras dev)
- [x] Dockerfile produção (multi-stage, scratch, ~15MB)
- [x] Dockerfile.dev com Air (hot-reload Go)
- [x] docker-compose.yml (api + web + firebase emulators)
- [x] Makefile com comandos padronizados
- [x] `.env.example` + `.gitignore`
- [x] Swagger (swaggo/swag) configurado — `GET /swagger/*` servindo a UI
- [ ] Linting configurado (Go + React)

---

## Próxima Fase

→ [FASE_01_AUTH_USUARIO.md](./FASE_01_AUTH_USUARIO.md) — Autenticação completa + CRUD de Usuário

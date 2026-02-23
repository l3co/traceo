# Fase 8 — Deploy no Cloud Run & Observabilidade

> **Duração estimada**: 2 semanas
> **Pré-requisito**: Fase 7 concluída (plataforma polida e funcional)

---

## Objetivo

Colocar a plataforma em produção com infraestrutura robusta, observabilidade e CI/CD. Ao final desta fase:

- API Go rodando no Cloud Run com auto-scaling
- Frontend React no Firebase Hosting com CDN global
- CI/CD automatizado (push → build → deploy)
- Logs estruturados com correlação de request
- Métricas e alertas
- Domínio customizado com HTTPS

Esta é a fase onde o projeto sai do "funciona na minha máquina" para "funciona para o mundo".

---

## Conceitos Go que você vai aprender nesta fase

### 1. Structured Logging — logs legíveis por máquinas

No desenvolvimento, `log.Printf` é suficiente. Em produção, precisamos de **logs estruturados** — JSON que pode ser filtrado e buscado no Cloud Logging.

#### O problema com logs textuais

```
2024-03-15 10:30:45 ERROR: failed to find user abc123: firestore: document not found
```

Isso é legível por humanos, mas:
- Como filtrar todos os erros de "find user"?
- Qual request causou esse erro?
- Qual endpoint?
- Quanto tempo levou?

#### A solução: `slog` (standard library, Go 1.21+)

```go
import "log/slog"

// Setup no main.go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
slog.SetDefault(logger)

// Uso em qualquer lugar
slog.Info("user found",
    slog.String("user_id", user.ID),
    slog.String("email", user.Email),
    slog.Duration("duration", elapsed),
)

slog.Error("failed to find user",
    slog.String("user_id", id),
    slog.String("error", err.Error()),
    slog.String("request_id", requestID),
)
```

Output (JSON):
```json
{
    "time": "2024-03-15T10:30:45Z",
    "level": "INFO",
    "msg": "user found",
    "user_id": "abc123",
    "email": "joao@email.com",
    "duration": "2.5ms"
}
```

#### Por que `slog` e não zerolog/zap?

| Biblioteca | Prós | Contras |
|---|---|---|
| `slog` (stdlib) | Nativo Go 1.21+, zero dependência, API oficial | Mais recente, menos features |
| zerolog | Muito rápido, zero allocation | Dependência externa |
| zap (Uber) | Muito popular, battle-tested | Dependência externa, API mais complexa |

**Decisão**: `slog` — é a solução oficial do Go e suficiente para nosso caso. Usar a stdlib é sempre preferível quando possível. No Cloud Logging do GCP, logs JSON são parseados automaticamente.

#### Middleware de logging para HTTP

```go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        requestID := uuid.New().String()

        // Adiciona request_id no context
        ctx := context.WithValue(r.Context(), "request_id", requestID)

        // Wrapper para capturar o status code
        ww := &responseWriter{ResponseWriter: w, statusCode: 200}

        next.ServeHTTP(ww, r.WithContext(ctx))

        slog.Info("http request",
            slog.String("method", r.Method),
            slog.String("path", r.URL.Path),
            slog.Int("status", ww.statusCode),
            slog.Duration("duration", time.Since(start)),
            slog.String("request_id", requestID),
            slog.String("remote_addr", r.RemoteAddr),
        )
    })
}
```

Cada request gera uma linha de log com método, path, status, duração e request_id. No Cloud Logging, você pode filtrar: "mostre todos os requests que retornaram 500 nos últimos 30 minutos".

---

### 2. Graceful Shutdown — desligar sem perder requests

Quando o Cloud Run faz deploy de uma nova versão, ele envia um sinal `SIGTERM` para a instância antiga. Ela tem **10 segundos** para terminar requests em andamento antes de ser finalizada.

Se não tratamos o `SIGTERM`, requests em andamento são cortados no meio → erro 502 para o cliente.

```go
func main() {
    // ... setup do router, services, etc ...

    server := &http.Server{
        Addr:    ":" + port,
        Handler: router,
    }

    // Channel que recebe sinais do OS
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    // Inicia o servidor em uma goroutine
    go func() {
        slog.Info("server starting", slog.String("port", port))
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            slog.Error("server failed", slog.String("error", err.Error()))
            os.Exit(1)
        }
    }()

    // Bloqueia até receber SIGTERM ou SIGINT
    sig := <-quit
    slog.Info("shutdown signal received", slog.String("signal", sig.String()))

    // Dá um tempo para requests em andamento terminarem
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        slog.Error("forced shutdown", slog.String("error", err.Error()))
    }

    slog.Info("server stopped gracefully")
}
```

#### Como isso funciona:

1. O servidor roda em uma goroutine
2. A main goroutine fica bloqueada esperando um sinal (`<-quit`)
3. Quando Cloud Run envia `SIGTERM`, `quit` recebe o sinal
4. `server.Shutdown(ctx)` para de aceitar novas conexões e espera as em andamento terminarem
5. Se levar mais de 10 segundos, o context expira e força a parada

**Esse pattern é obrigatório para Cloud Run.** Sem ele, deploys causam erros intermitentes.

---

### 3. Rate Limiting — proteção contra abuso

Rate limiting impede que um IP faça muitas requisições em pouco tempo. Importante para:
- Proteger contra brute force no login
- Proteger contra scraping dos dados
- Proteger endpoints públicos (cadastro de homeless) contra spam

```go
import "golang.org/x/time/rate"

type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.Mutex
    rate     rate.Limit
    burst    int
}

func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
    return &RateLimiter{
        limiters: make(map[string]*rate.Limiter),
        rate:     r,
        burst:    burst,
    }
}

func (rl *RateLimiter) GetLimiter(ip string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    limiter, exists := rl.limiters[ip]
    if !exists {
        limiter = rate.NewLimiter(rl.rate, rl.burst)
        rl.limiters[ip] = limiter
    }
    return limiter
}

func RateLimitMiddleware(rl *RateLimiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            limiter := rl.GetLimiter(r.RemoteAddr)
            if !limiter.Allow() {
                http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

#### `sync.Mutex` — a trava de exclusão mútua

O `RateLimiter` usa um `map` compartilhado entre goroutines (cada request HTTP roda em sua goroutine). Maps em Go **não são thread-safe** — se duas goroutines lerem/escreverem ao mesmo tempo, o programa pode crashar.

`sync.Mutex` é a solução:

```go
rl.mu.Lock()    // "tranquei — só eu estou acessando o map"
defer rl.mu.Unlock()  // "quando sair desta função, destranco"
// ... acessa o map com segurança ...
```

Mutex é mais simples que channels para proteger acesso a dados compartilhados. A regra prática:
- **Channels** → para comunicar dados entre goroutines
- **Mutex** → para proteger acesso concorrente a dados compartilhados

---

### 4. Configuration Management — variáveis de ambiente

Em produção, **toda configuração vem de variáveis de ambiente**. Nunca hardcoded, nunca em arquivo commitado.

```go
// internal/config/config.go

type Config struct {
    Port              string
    FirebaseProjectID string
    SendGridAPIKey    string
    SendGridFromEmail string
    TelegramBotToken  string
    TelegramChatID    string
    GCSBucket         string
    AllowedOrigins    []string
    Environment       string // "development" | "production"
}

func Load() (*Config, error) {
    cfg := &Config{
        Port:              getEnv("PORT", "8080"),
        FirebaseProjectID: requireEnv("FIREBASE_PROJECT_ID"),
        SendGridAPIKey:    requireEnv("SENDGRID_API_KEY"),
        SendGridFromEmail: requireEnv("SENDGRID_FROM_EMAIL"),
        TelegramBotToken:  requireEnv("TELEGRAM_BOT_TOKEN"),
        TelegramChatID:    requireEnv("TELEGRAM_CHAT_ID"),
        GCSBucket:         requireEnv("GCS_BUCKET"),
        AllowedOrigins:    strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:5173"), ","),
        Environment:       getEnv("ENVIRONMENT", "development"),
    }
    return cfg, nil
}

func getEnv(key, fallback string) string {
    if val := os.Getenv(key); val != "" {
        return val
    }
    return fallback
}

func requireEnv(key string) string {
    val := os.Getenv(key)
    if val == "" {
        slog.Error("required environment variable not set", slog.String("key", key))
        os.Exit(1) // fail fast — não sobe sem config
    }
    return val
}
```

#### Fail Fast

Se uma variável obrigatória não está configurada, o programa **não sobe**. Isso é intencional:
- Melhor descobrir que falta configuração no deploy do que em runtime 3 horas depois
- Cloud Run mostra o erro nos logs de startup
- Previne estados inconsistentes

No Cloud Run, variáveis de ambiente são configuradas via Secret Manager ou diretamente no serviço.

---

## Infraestrutura: Cloud Run + Firebase Hosting

### API Go no Cloud Run

```
GitHub Push → Cloud Build → Container Image → Cloud Run
```

#### Cloud Build (`cloudbuild.yaml`)

```yaml
steps:
  # Build da imagem Docker
  - name: 'gcr.io/cloud-builders/docker'
    args: ['build', '-t', 'gcr.io/$PROJECT_ID/desaparecidos-api', './api']

  # Push para o Container Registry
  - name: 'gcr.io/cloud-builders/docker'
    args: ['push', 'gcr.io/$PROJECT_ID/desaparecidos-api']

  # Deploy no Cloud Run
  - name: 'gcr.io/google.com/cloudsdktool/cloud-sdk'
    entrypoint: gcloud
    args:
      - 'run'
      - 'deploy'
      - 'desaparecidos-api'
      - '--image=gcr.io/$PROJECT_ID/desaparecidos-api'
      - '--region=southamerica-east1'
      - '--platform=managed'
      - '--allow-unauthenticated'
      - '--min-instances=0'
      - '--max-instances=10'
      - '--memory=256Mi'
      - '--cpu=1'
      - '--set-secrets=SENDGRID_API_KEY=sendgrid-key:latest,TELEGRAM_BOT_TOKEN=telegram-token:latest'
```

#### Por que `southamerica-east1`?

É a região mais próxima do Brasil (São Paulo). Latência ~10-20ms para usuários brasileiros, vs ~150ms para `us-central1`.

#### Por que `min-instances=0`?

Scale-to-zero: se ninguém acessar o site por um tempo, o Cloud Run desliga todas as instâncias. Custo = zero quando não há tráfego.

**Trade-off**: o primeiro request após período inativo tem **cold start** (~100-300ms para Go). Aceitável para nosso caso. Se quisermos eliminar cold start: `min-instances=1` (~R$20-30/mês).

### Frontend React no Firebase Hosting

```bash
# No diretório web/
npm run build
firebase deploy --only hosting
```

Firebase Hosting serve arquivos estáticos via CDN global. Performance excelente, custo quase zero.

---

## Tarefas Detalhadas

### Backend

#### Tarefa 7.1 — Structured Logging com slog

- Configurar `slog` com JSON handler no `main.go`
- Substituir todos os `log.Printf` por `slog.Info/Error/Warn`
- Adicionar request_id em todos os logs
- Middleware de logging HTTP

#### Tarefa 7.2 — Graceful Shutdown

- Implementar signal handling (SIGTERM, SIGINT)
- `server.Shutdown()` com timeout de 10 segundos
- Testar: iniciar servidor, fazer request longo, enviar SIGTERM → request deve completar

#### Tarefa 7.3 — Rate Limiting (revisão de produção)

> Rate limiting base já foi implementado na Fase 1 (ver [SECURITY.md](./SECURITY.md) seção 1.1).
> Nesta tarefa, revisar e ajustar limites para produção.

- Revisar limites de rate limiting baseado em métricas reais de uso
- Adicionar monitoramento: alertar quando IPs atingem rate limit com frequência (possível ataque)
- Considerar rate limiting distribuído se escalar para múltiplas instâncias Cloud Run (Redis ou Memorystore)

#### Tarefa 7.4 — Health Check avançado

Expandir `GET /api/v1/health` para verificar dependências:

```json
{
    "status": "ok",
    "version": "1.0.0",
    "uptime": "2h30m",
    "dependencies": {
        "firestore": "ok",
        "firebase_auth": "ok"
    }
}
```

#### Tarefa 7.5 — CORS e Security Headers (revisão de produção)

> CORS e Security Headers já foram implementados na Fase 1 (ver [SECURITY.md](./SECURITY.md) seções 1.3 e 1.5).
> Nesta tarefa, revisar configurações para produção.

- Confirmar `AllowedOrigins` para domínio final (`https://desaparecidos.me`)
- Revisar CSP no Firebase Hosting headers para o frontend
- Testar headers com [securityheaders.com](https://securityheaders.com)

#### Tarefa 7.7 — Dockerfile otimizado

Multi-stage build com imagem mínima:

```dockerfile
# Stage 1: Build
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server cmd/server/main.go

# Stage 2: Run
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
```

**`FROM scratch`** — imagem vazia. Só o binário Go + certificados SSL. Imagem final: **~10-15MB**.

**`-ldflags="-s -w"`** — remove debug symbols, reduz o tamanho do binário em ~30%.

#### Tarefa 7.8 — CI/CD com Cloud Build

- `cloudbuild.yaml` para build + deploy automático
- Trigger: push na branch `main`
- Secrets via Secret Manager

#### Tarefa 7.9 — Configurar Secret Manager

Armazenar todas as chaves sensíveis:
- `SENDGRID_API_KEY`
- `TELEGRAM_BOT_TOKEN`
- `FIREBASE_SERVICE_ACCOUNT` (se necessário)

#### Tarefa 7.10 — Domínio customizado

- Configurar `desaparecidos.me` no Firebase Hosting (frontend)
- Configurar `api.desaparecidos.me` no Cloud Run (backend)
- SSL automático (Let's Encrypt via Cloud Run / Firebase)

### Frontend

#### Tarefa 7.11 — Build de produção

```bash
npm run build
# Verificar tamanho do bundle
npx vite-bundle-visualizer
```

#### Tarefa 7.12 — Firebase Hosting setup

```bash
firebase init hosting
# public directory: dist
# SPA: Yes (rewrite all URLs to index.html)
# Automatic builds: Yes (connect to GitHub)
```

#### Tarefa 7.13 — Environment variables no frontend

```
# .env.production
VITE_API_URL=https://api.desaparecidos.me
VITE_FIREBASE_PROJECT_ID=desaparecidos
VITE_MAPBOX_TOKEN=pk.xxx
```

---

## Monitoramento e Alertas

### Cloud Monitoring

Configurar alertas para:
- **Error rate > 5%** nos últimos 5 minutos → alerta email
- **Latência p99 > 2 segundos** → alerta email
- **Cloud Run instâncias > 5** → aviso (possível pico ou ataque)
- **Firestore reads > 50.000/dia** → aviso de custo

### Dashboard de métricas

No Cloud Console, criar dashboard com:
- Requests por minuto
- Latência p50/p95/p99
- Taxa de erros
- Instâncias ativas
- Memória e CPU por instância

---

## Entregáveis da Fase 7

- [ ] Structured logging com slog (JSON)
- [ ] Graceful shutdown
- [ ] Rate limiting revisado para produção (métricas reais + alertas)
- [ ] Health check avançado com dependências
- [ ] CORS e Security Headers revisados para domínio final
- [ ] Dockerfile otimizado (scratch, ~15MB)
- [ ] CI/CD com Cloud Build
- [ ] Secrets no Secret Manager (todas as API keys)
- [ ] Firestore Security Rules de produção (ver [SECURITY.md](./SECURITY.md))
- [ ] Cloud Storage Rules de produção
- [ ] Budget alerts no GCP ($50/mês)
- [ ] Domínio customizado com HTTPS
- [ ] Firebase Hosting para frontend
- [ ] Alertas de monitoramento configurados
- [ ] Dashboard de métricas no Cloud Console

---

## Próxima Fase

→ [FASE_09_MOBILE.md](./FASE_09_MOBILE.md) — React Native (App Mobile)

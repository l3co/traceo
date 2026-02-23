# Segurança — Estratégia Completa

> **Documento complementar obrigatório** do sistema de engenharia.
> Define proteção, risco e mitigação em todas as camadas.

---

## Visão Geral

A segurança deste projeto opera em **5 camadas**:

```
┌─────────────────────────────────────────────────┐
│  1. Perímetro        Rate limiting, CORS, bot   │
│                      protection, request limits  │
├─────────────────────────────────────────────────┤
│  2. Autenticação     Firebase Auth, JWT, roles   │
├─────────────────────────────────────────────────┤
│  3. Autorização      Ownership, RBAC, policies   │
├─────────────────────────────────────────────────┤
│  4. Dados            Validation, sanitization,   │
│                      Firestore rules, encryption │
├─────────────────────────────────────────────────┤
│  5. Infraestrutura   API key restrictions, cost  │
│                      protection, secrets, logs   │
└─────────────────────────────────────────────────┘
```

Cada camada é independente. Se uma falhar, as outras ainda protegem.

---

## 1. Perímetro — Primeira Linha de Defesa

### 1.1 Rate Limiting (Fase 1 — não Fase 8)

Rate limiting é **fundação**, não polimento. Deve existir desde o primeiro endpoint público.

#### Estratégia em 3 camadas

| Camada | Escopo | Limite | Implementação |
|---|---|---|---|
| **Global** | Por IP, todos os endpoints | 200 req/min | Middleware Chi (primeiro da chain) |
| **Por endpoint** | Endpoints sensíveis | Variável (tabela abaixo) | Middleware por rota |
| **Por usuário** | Usuário autenticado | 100 req/min | Middleware pós-auth |

#### Limites por categoria de endpoint

| Categoria | Endpoints | Limite por IP | Por quê |
|---|---|---|---|
| **Auth** | `POST /auth/login`, `POST /auth/forgot-password` | 5 req/min | Brute force |
| **Cadastro** | `POST /users` | 3 req/min | Criação massiva de contas |
| **Write público** | `POST /tips` (denúncia anônima) | 5 req/min | Spam sem auth |
| **Write autenticado** | `POST /missing`, `POST /sightings` | 20 req/min | Abuso de contas comprometidas |
| **Upload** | `POST /upload` | 10 req/min | Consumo de storage |
| **AI** | `POST /missing/:id/age-progression` | 2 req/min por user | Custo de API Gemini/Imagen |
| **Leitura** | `GET /missing`, `GET /search` | 60 req/min | Scraping |
| **Health** | `GET /health` | Sem limite | Monitoramento |

#### Implementação em Go

```go
// internal/handler/middleware/rate_limit.go

type RateLimitConfig struct {
    GlobalRate     rate.Limit // requests por segundo (global por IP)
    GlobalBurst    int
    EndpointRates  map[string]EndpointLimit
}

type EndpointLimit struct {
    Rate  rate.Limit
    Burst int
}
```

Usamos `golang.org/x/time/rate` (stdlib estendida) com um map de limiters por IP. Limpeza periódica de IPs inativos para evitar memory leak.

#### Headers de resposta

Quando o rate limit é atingido, retornar:

```
HTTP/1.1 429 Too Many Requests
Retry-After: 30
X-RateLimit-Limit: 200
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1709312400
```

Esses headers permitem que o frontend (e bots bem-comportados) respeitem o limite sem bater repetidamente.

#### Por que na Fase 1 e não na Fase 8?

O rate limiting estava planejado para a Fase 8 (Deploy). Isso é um erro de priorização. Motivos para antecipar:

1. **Endpoints públicos existem desde a Fase 1** — login, cadastro, health check
2. **Denúncias anônimas (Fase 4) não têm auth** — sem rate limit, qualquer bot spamma
3. **Google Maps e Gemini têm custo por request** — sem limite, uma conta comprometida gera fatura
4. **É um middleware** — implementar cedo não atrasa nada; é uma função que envolve o handler

---

### 1.2 Proteção contra Bots (Fase 4 — endpoints públicos)

Endpoints **sem autenticação** são alvos naturais de bots:

| Endpoint | Risco | Proteção |
|---|---|---|
| `POST /users` (cadastro) | Criação massiva de contas fake | reCAPTCHA v3 |
| `POST /tips` (denúncia anônima) | Spam de denúncias | reCAPTCHA v3 |
| `POST /sightings` (se permitir anônimo) | Avistamentos falsos | reCAPTCHA v3 + rate limit |
| `GET /missing/search` | Scraping da base | Rate limit agressivo |

#### reCAPTCHA v3 (invisível)

Diferente do reCAPTCHA v2 (checkbox "não sou um robô"), o v3 é **invisível** — roda em background e retorna um **score de 0.0 a 1.0**:

- **0.9–1.0** → provavelmente humano → liberar
- **0.5–0.8** → suspeito → liberar mas monitorar
- **0.0–0.4** → provavelmente bot → bloquear

```go
// Middleware para endpoints públicos sensíveis
func RecaptchaMiddleware(secretKey string, threshold float64) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("X-Recaptcha-Token")
            if token == "" {
                http.Error(w, "recaptcha token required", http.StatusForbidden)
                return
            }

            score, err := verifyRecaptcha(secretKey, token)
            if err != nil || score < threshold {
                http.Error(w, "bot detected", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

No frontend React:

```tsx
import { useGoogleReCaptcha } from 'react-google-recaptcha-v3'

const { executeRecaptcha } = useGoogleReCaptcha()
const token = await executeRecaptcha('register')
// Enviar token no header X-Recaptcha-Token
```

**Custo**: gratuito até 1M de verificações/mês. Suficiente para o nosso caso.

**Quando implementar**: junto com o primeiro endpoint público sem auth (Fase 4 — denúncias anônimas). Opcionalmente, já no cadastro de usuário (Fase 1) se quisermos proteção desde o início.

---

### 1.3 CORS (Fase 1)

CORS deve ser configurado desde o primeiro handler. Não deixar para a Fase 8.

```go
// Desenvolvimento
AllowedOrigins: []string{"http://localhost:5173"}

// Produção
AllowedOrigins: []string{"https://desaparecidos.me", "https://www.desaparecidos.me"}
```

**Regra**: nunca `AllowedOrigins: ["*"]` em produção. Isso permite que qualquer site faça requests à nossa API usando cookies/tokens do usuário.

---

### 1.4 Limites de Request (Fase 1)

| Limite | Valor | Por quê |
|---|---|---|
| **JSON body** | 1 MB | Previne payloads gigantes que consomem memória |
| **Upload de foto** | 10 MB por arquivo | Fotos de boa qualidade raramente passam de 5 MB |
| **Máximo de fotos por request** | 5 | Array `photoURLs` limitado |
| **Timeout de request** | 30 segundos | Previne conexões penduradas |
| **Header size** | 8 KB | Padrão HTTP |

```go
// Middleware de body limit
func MaxBodySize(maxBytes int64) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
            next.ServeHTTP(w, r)
        })
    }
}

// No router
r.Use(MaxBodySize(1 << 20)) // 1 MB para JSON
r.With(MaxBodySize(10 << 20)).Post("/upload", uploadHandler) // 10 MB para uploads
```

---

### 1.5 Security Headers (Fase 1)

Headers de segurança HTTP devem estar presentes desde o primeiro deploy, não na Fase 8.

```go
func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(self)")
        next.ServeHTTP(w, r)
    })
}
```

**Nota sobre CSP (Content-Security-Policy)**: não definir no backend para SPA. O CSP do React é melhor controlado via meta tag no `index.html` ou via Firebase Hosting headers, porque precisa permitir scripts inline do Vite em dev.

---

## 2. Autenticação — Firebase Auth

### 2.1 Por que Firebase Auth?

| Aspecto | Auth custom (Go + bcrypt + JWT) | Firebase Auth |
|---|---|---|
| **Implementação** | ~2 semanas de código | ~2 horas de integração |
| **Segurança de senha** | Você implementa bcrypt | Google implementa e audita |
| **Brute force** | Você implementa | Built-in |
| **Multi-provider** | Você integra cada OAuth | Toggle no console |
| **Refresh tokens** | Você implementa rotação | Automático |
| **Password reset email** | Você configura SMTP | Firebase envia |
| **MFA** | Você implementa TOTP | SDK pronto |
| **Auditoria** | Você implementa logs | Cloud Logging automático |

**Decisão**: Firebase Auth. A plataforma lida com dados sensíveis de pessoas desaparecidas — não é o lugar para experimentar com crypto custom.

### 2.2 Providers planejados

| Provider | Fase | Por quê |
|---|---|---|
| **Email/Senha** | Fase 1 | Base — todo usuário tem email |
| **Google** | Fase 7 | Conveniência — login com um clique |
| **Apple** | Fase 9 | Obrigatório para apps iOS na App Store |
| **Phone (SMS)** | Avaliar | Útil para familiares sem email |

O Firebase Auth suporta todos com **zero mudança no backend**. O middleware JWT verifica o token independente do provider que o gerou.

### 2.3 Fluxo de autenticação

```
Cliente                 Go API                  Firebase Auth
  │                       │                         │
  │ POST /auth/login      │                         │
  │ {email, password}     │                         │
  │──────────────────────►│                         │
  │                       │  verifyPassword()       │
  │                       │────────────────────────►│
  │                       │  ◄─── JWT (id_token)    │
  │                       │                         │
  │ ◄─── {token, user}   │                         │
  │                       │                         │
  │ GET /missing          │                         │
  │ Authorization: Bearer │                         │
  │──────────────────────►│                         │
  │                       │  VerifyIDToken(jwt)     │
  │                       │────────────────────────►│
  │                       │  ◄─── {uid, claims}     │
  │                       │                         │
  │ ◄─── [missing data]  │                         │
```

### 2.4 JWT — Detalhes

- **Expiração**: 1 hora (padrão Firebase)
- **Refresh**: SDK do Firebase no frontend faz refresh automático antes de expirar
- **Verificação no Go**: `auth.VerifyIDToken(ctx, idToken)` — verifica assinatura, expiração e revogação
- **Claims customizados**: podemos adicionar `role` (user/volunteer/ong/admin) como custom claim para RBAC

```go
// Adicionar role como custom claim
err := authClient.SetCustomUserClaims(ctx, uid, map[string]interface{}{
    "role": "ong",
})
```

O middleware de auth lê o claim do token — sem query extra no Firestore para cada request.

---

## 3. Autorização — Quem pode o quê

### 3.1 Roles

| Role | Quem | Pode |
|---|---|---|
| **user** | Familiar que cadastra desaparecido | CRUD dos próprios missing, ver sightings |
| **volunteer** | Voluntário verificado | Cadastrar homeless, registrar avistamentos |
| **ong** | ONG parceira | Tudo do volunteer + revisar denúncias + revisar matches |
| **admin** | Administrador da plataforma | Tudo |

### 3.2 Ownership — "só o dono pode editar"

Além do role, a maioria das operações de escrita exige **ownership**:

```go
func (s *MissingService) Update(ctx context.Context, id string, input UpdateInput) error {
    missing, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return err
    }

    // Ownership check — só quem cadastrou pode editar
    callerID := auth.UserIDFromContext(ctx)
    callerRole := auth.RoleFromContext(ctx)

    if missing.UserID != callerID && callerRole != "admin" {
        return ErrForbidden
    }

    // ... update
}
```

**Regra**: toda operação de escrita verifica ownership OU role admin. Sem exceções.

### 3.3 Políticas por endpoint

| Endpoint | Auth? | Role mínimo | Ownership? |
|---|---|---|---|
| `GET /missing` | Não | — | — |
| `POST /missing` | Sim | user | — |
| `PUT /missing/:id` | Sim | user | Sim (ou admin) |
| `DELETE /missing/:id` | Sim | user | Sim (ou admin) |
| `POST /sightings` | Sim | user | — |
| `POST /homeless` | Sim | volunteer | — |
| `POST /tips` | **Não** | — | — |
| `PATCH /tips/:id` | Sim | ong | — |
| `GET /missing/:id/timeline` | Sim | user | Sim (ou ong/admin) |
| `POST /missing/:id/age-progression` | Sim | user | Sim (ou admin) |
| `GET /missing/:id/poster` | Não | — | — |

---

## 4. Dados — Validação, Sanitização e Regras

### 4.1 Input Validation (toda fase)

**Toda entrada é hostil até prova em contrário.**

```go
type CreateMissingRequest struct {
    Name                string `json:"name" validate:"required,min=2,max=150"`
    BirthDate           string `json:"birth_date" validate:"required,datetime=2006-01-02"`
    Gender              string `json:"gender" validate:"required,oneof=male female other"`
    Height              string `json:"height,omitempty" validate:"omitempty,max=10"`
    Circumstance        string `json:"circumstance" validate:"required,oneof=left_home ran_away abduction hospital disaster unknown"`
    PoliceReportNumber  string `json:"police_report_number,omitempty" validate:"omitempty,max=50"`
}
```

Usamos `go-playground/validator` — a lib de validação mais usada no ecossistema Go. Validação acontece **no handler**, antes de chegar ao service.

### 4.2 Sanitização

Além de validar formato, sanitizar conteúdo:

```go
import "github.com/microcosm-cc/bluemonday"

var policy = bluemonday.StrictPolicy() // remove TODO HTML

func sanitize(input string) string {
    return policy.Sanitize(strings.TrimSpace(input))
}
```

Aplicar em: `name`, `observation`, `description`, `circumstanceDetails`, `selfReportedInfo` — todo campo de texto livre.

**Por quê?** Mesmo com React escapando output por padrão (proteção XSS no render), dados sujos no banco são um risco se consumidos por outros clientes (mobile, relatórios PDF, emails).

### 4.3 Firestore Security Rules (produção)

As regras de desenvolvimento (allow all) devem ser substituídas antes de qualquer deploy:

```
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {

    // Users — só o próprio usuário lê/edita
    match /users/{userId} {
      allow read: if request.auth != null && request.auth.uid == userId;
      allow write: if request.auth != null && request.auth.uid == userId;
    }

    // Missing — leitura pública, escrita autenticada
    match /missing/{missingId} {
      allow read: if true;
      allow create: if request.auth != null;
      allow update, delete: if request.auth != null &&
        (resource.data.userId == request.auth.uid ||
         request.auth.token.role == 'admin');
    }

    // Sightings — leitura autenticada, escrita autenticada
    match /sightings/{sightingId} {
      allow read: if request.auth != null;
      allow create: if request.auth != null;
    }

    // Tips — criação pública, leitura por role
    match /tips/{tipId} {
      allow create: if true;
      allow read: if request.auth != null &&
        (request.auth.token.role in ['ong', 'admin']);
      allow update: if request.auth != null &&
        (request.auth.token.role in ['ong', 'admin']);
    }

    // Timeline — leitura autenticada
    match /timeline/{eventId} {
      allow read: if request.auth != null;
      allow write: if false; // só o backend escreve
    }

    // Default: negar tudo
    match /{document=**} {
      allow read, write: if false;
    }
  }
}
```

**Nota importante**: as Firestore Rules são uma **segunda linha de defesa**. A validação principal acontece no Go (service layer). As rules protegem contra acesso direto ao Firestore (ex: se alguém obtiver as credenciais do projeto Firebase).

### 4.4 Cloud Storage Rules

```
rules_version = '2';
service firebase.storage {
  match /b/{bucket}/o {
    // Fotos de desaparecidos — leitura pública, upload autenticado
    match /missing/{missingId}/{fileName} {
      allow read: if true;
      allow write: if request.auth != null
        && request.resource.size < 10 * 1024 * 1024  // max 10 MB
        && request.resource.contentType.matches('image/.*');  // só imagens
    }

    // Fotos de homeless — mesma regra
    match /homeless/{homelessId}/{fileName} {
      allow read: if true;
      allow write: if request.auth != null
        && request.resource.size < 10 * 1024 * 1024
        && request.resource.contentType.matches('image/.*');
    }

    // Avatares — só o dono
    match /avatars/{userId}/{fileName} {
      allow read: if true;
      allow write: if request.auth != null
        && request.auth.uid == userId
        && request.resource.size < 5 * 1024 * 1024
        && request.resource.contentType.matches('image/.*');
    }

    // Default: negar
    match /{allPaths=**} {
      allow read, write: if false;
    }
  }
}
```

### 4.5 Dados sensíveis nos logs

**Nunca logar**:
- Tokens JWT
- Senhas (mesmo hash)
- Emails completos (mascarar: `j***@email.com`)
- Telefones completos
- IPs em logs de aplicação (só no access log HTTP)

```go
// ❌ Errado
slog.Info("user login", slog.String("email", user.Email))

// ✅ Correto
slog.Info("user login", slog.String("user_id", user.ID))
```

---

## 5. Infraestrutura — Proteção de APIs Externas e Custos

### 5.1 Proteção de API Keys — Google Maps

A API key do Google Maps é exposta no frontend (inevitável — o JavaScript precisa da key para renderizar o mapa). Sem restrição, qualquer pessoa copia a key e usa na própria aplicação, gerando custo na **sua** conta Google Cloud.

#### Restrições obrigatórias no Console do GCP

| Restrição | Configuração | Por quê |
|---|---|---|
| **HTTP referrer** | `desaparecidos.me/*`, `localhost:5173/*` | Só aceita requests do nosso domínio |
| **API restrictions** | Só Maps JS API, Geocoding, Places | Impede uso da key em outras APIs |
| **Quotas** | 10.000 loads/dia (Maps JS), 1.000 req/dia (Geocoding) | Cap de custo |

**Como configurar**:
1. Google Cloud Console → APIs & Services → Credentials
2. Selecionar a API key
3. Application restrictions → HTTP referrers
4. API restrictions → Restrict key → selecionar só as APIs necessárias

**Estimativa de custo com proteção**:
- Maps JS API: $7 por 1.000 loads → 10.000 loads/dia ≈ $70/dia max
- Geocoding: $5 por 1.000 requests → 1.000 req/dia ≈ $5/dia max
- Com as quotas, o custo máximo mensal é previsível e limitado

#### Keys separadas por ambiente

| Ambiente | Key | Restrições |
|---|---|---|
| **Development** | `VITE_GOOGLE_MAPS_API_KEY_DEV` | Referrer: `localhost:*` |
| **Production** | `VITE_GOOGLE_MAPS_API_KEY` | Referrer: `desaparecidos.me/*` |

Nunca usar a mesma key em dev e prod. Se a key de dev vazar (commit acidental), a de prod não é afetada.

---

### 5.2 Proteção de API Keys — Gemini / Imagen

A key do Gemini é usada **apenas no backend** (Go) — nunca exposta no frontend.

| Proteção | Configuração |
|---|---|
| **Secret Manager** | Key armazenada no GCP Secret Manager, injetada como env var no Cloud Run |
| **IP restriction** | Restringir a key para o IP do Cloud Run (se possível) |
| **Quotas no GCP** | Requests por minuto e por dia |
| **Worker pool limitado** | Máximo de 3 processamentos simultâneos de AI (já planejado na Fase 6) |
| **Rate limit por user** | 2 req/min para age progression (tabela da seção 1.1) |

#### Custo estimado e caps

| Operação | Custo unitário | Cap diário | Custo máximo/dia |
|---|---|---|---|
| Age progression (Imagen) | ~$0.04/imagem | 100 imagens | $4.00 |
| Face matching (Gemini Vision) | ~$0.01/comparação | 500 comparações | $5.00 |
| Descrição facial (Gemini) | ~$0.005/request | 200 requests | $1.00 |

**Total máximo diário de AI: ~$10.** Definir alerta no GCP se passar de $5/dia.

---

### 5.3 Proteção de API Keys — Resend (Email)

| Proteção | Configuração |
|---|---|
| **Secret Manager** | Key no backend, nunca no frontend |
| **Domain verification** | Resend verifica que só enviamos de `@desaparecidos.me` |
| **Rate limit de envio** | Máximo 100 emails/hora no código Go |
| **Template fixo** | Não permitir conteúdo dinâmico arbitrário em emails |

---

### 5.4 Proteção de API Keys — WhatsApp Business

| Proteção | Configuração |
|---|---|
| **Secret Manager** | Token no backend |
| **Template approved** | Meta revisa e aprova cada template de mensagem |
| **Rate limit** | Máximo 50 mensagens/hora |
| **Webhook verification** | Validar assinatura de webhooks do Meta |

---

### 5.5 Budget Alerts no GCP

Configurar alertas de custo no Google Cloud Billing:

| Threshold | Ação |
|---|---|
| **50% do budget mensal** | Email de aviso |
| **80% do budget mensal** | Email + Slack/Telegram de alerta |
| **100% do budget mensal** | Email + alerta urgente |
| **120% do budget mensal** | Desabilitar APIs não-essenciais (programmatic budget action) |

**Budget sugerido para início**: $50/mês (Cloud Run + Firestore + Maps + AI).

Para configurar:
1. Cloud Console → Billing → Budgets & Alerts
2. Definir budget → $50/mês
3. Adicionar thresholds com notificação por email
4. Opcional: Cloud Function que desabilita APIs ao atingir 120%

---

### 5.6 Proxy de API para o Frontend

Para APIs externas que o frontend consome diretamente (Google Maps é inevitável), considerar **proxy endpoints** no backend para serviços de geocoding:

```
Frontend → GET /api/v1/geocode?address=...  → Go API → Google Geocoding API
```

**Vantagens do proxy**:
- A API key do Geocoding fica no backend (nunca exposta)
- Rate limit controlado no Go
- Cache de resultados (mesmo endereço = mesmo resultado)
- Auditoria de uso

**Quando usar proxy**: Geocoding, Places Autocomplete.
**Quando NÃO usar proxy**: Maps JavaScript API (precisa rodar no browser).

---

## 6. Resumo — Quando implementar cada medida

### Fase 0 (Fundação)

- [ ] `.env` com secrets (nunca commitar)
- [ ] `.gitignore` para `.env`, service accounts, keys

### Fase 1 (Auth + User)

- [ ] Firebase Auth integrado (email/senha)
- [ ] Middleware JWT com verificação de token
- [ ] Middleware de rate limiting global (200 req/min por IP)
- [ ] Rate limiting específico para auth (5 req/min no login)
- [ ] CORS configurado (origem específica, não wildcard)
- [ ] Security Headers middleware
- [ ] Body size limit (1 MB JSON, 10 MB upload)
- [ ] Request timeout (30s)
- [ ] Input validation com `go-playground/validator`
- [ ] Sanitização de texto livre com `bluemonday`
- [ ] Ownership check no service (user só edita o que é dele)
- [ ] Custom claims para role no Firebase Auth
- [ ] Logs sem dados sensíveis

### Fase 2 (Desaparecidos + Maps)

- [ ] Google Maps API key com HTTP referrer restriction
- [ ] API keys separadas dev/prod
- [ ] Quotas configuradas no GCP Console para Maps
- [ ] Proxy endpoint para Geocoding (key no backend)
- [ ] Upload com validação de content-type (só imagens)
- [ ] Cloud Storage rules (autenticado + size limit + content-type)

### Fase 4 (Avistamentos + Tips)

- [ ] reCAPTCHA v3 nos endpoints públicos sem auth
- [ ] Rate limiting específico para `POST /tips` (5 req/min)
- [ ] Validação de denúncia anônima (texto mínimo, anti-spam)

### Fase 6 (AI)

- [ ] Gemini API key no Secret Manager
- [ ] Worker pool com concurrency limitada (3)
- [ ] Rate limiting específico para age progression (2 req/min por user)
- [ ] Quotas de AI no GCP Console

### Fase 8 (Deploy)

- [ ] Firestore Security Rules (produção)
- [ ] Cloud Storage Rules (produção)
- [ ] Secret Manager para todas as keys
- [ ] Budget alerts no GCP ($50/mês)
- [ ] Cloud Function de circuit breaker ao atingir 120% do budget
- [ ] Monitoramento de tentativas de login falhadas
- [ ] Alerta de rate limit excedido (possível ataque)
- [ ] Revisão final de toda a checklist

### Fase 9 (Mobile)

- [ ] Firebase Auth com Google e Apple providers
- [ ] App Check (atesta que o request vem do app real, não de emulador)
- [ ] Certificate pinning no React Native

---

## 7. Threat Model — Ameaças e Mitigações

| Ameaça | Probabilidade | Impacto | Mitigação |
|---|---|---|---|
| Brute force no login | Alta | Médio | Rate limit 5/min + Firebase built-in |
| Scraping de dados de desaparecidos | Alta | Médio | Rate limit leitura + reCAPTCHA |
| API key do Google Maps roubada | Alta | Alto (custo) | HTTP referrer restriction + quotas |
| Spam em denúncias anônimas | Alta | Baixo | reCAPTCHA + rate limit + revisão ONG |
| Abuso de AI (age progression) | Média | Alto (custo) | Rate limit 2/min + worker pool + budget alert |
| XSS em campos de texto | Média | Alto | Sanitização backend + React escape + CSP |
| CSRF | Baixa (SPA) | Médio | CORS strict + token Bearer (não cookie) |
| Vazamento de dados do Firestore | Baixa | Crítico | Security Rules + acesso só via API Go |
| Conta admin comprometida | Baixa | Crítico | MFA obrigatório para admin (Firebase Auth) |
| DDoS | Baixa | Alto | Cloud Run auto-scaling + rate limit + Cloudflare futuro |

---

## 8. Decisões Conscientes

### Por que NÃO implementar agora

| Medida | Por quê deixar para depois |
|---|---|
| **WAF (Web Application Firewall)** | Custo de Cloudflare Pro. Avaliar na Fase 8 se o tráfego justificar |
| **mTLS entre serviços** | Só temos um serviço backend. Relevante se virar microsserviços |
| **Encrypted at rest custom** | Firestore já encripta at rest por padrão (Google-managed keys) |
| **Pen test formal** | Fazer antes do launch público, mas não precisa durante o desenvolvimento |
| **SOC 2 / LGPD formal** | Avaliar quando tiver tráfego real. LGPD sim, compliance formal depois |

### LGPD — Considerações básicas

O projeto lida com dados pessoais sensíveis (fotos, localização, dados de saúde). Considerações mínimas:

- **Consentimento** — checkbox de termos de uso no cadastro (já planejado)
- **Direito ao esquecimento** — endpoint `DELETE /users/:id` que remove todos os dados do usuário
- **Minimização** — só coletar dados necessários (campos opcionais são opcionais de verdade)
- **Portabilidade** — futuro: endpoint que exporta dados do usuário em JSON
- **Logs** — não logar dados pessoais identificáveis

> Nota: compliance LGPD formal requer avaliação jurídica. Estas são práticas técnicas de base.

---

## Referências

- [OWASP Top 10 (2021)](https://owasp.org/www-project-top-ten/)
- [Firebase Auth docs](https://firebase.google.com/docs/auth)
- [Google Maps API key best practices](https://developers.google.com/maps/api-security-best-practices)
- [GCP Budget alerts](https://cloud.google.com/billing/docs/how-to/budgets)
- [reCAPTCHA v3 docs](https://developers.google.com/recaptcha/docs/v3)
- [Go security best practices](https://go.dev/doc/security/best-practices)

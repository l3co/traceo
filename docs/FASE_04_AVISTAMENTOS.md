# Fase 4 — Avistamentos & Notificações Assíncronas

> **Duração estimada**: 3 semanas
> **Pré-requisito**: Fase 3 concluída (busca + dashboard funcionando)

---

## Objetivo

Implementar o sistema de avistamentos (sightings) — quando alguém vê uma pessoa desaparecida e informa a localização — e o fluxo de notificações assíncronas que alerta o familiar. Ao final desta fase:

- Qualquer pessoa pode registrar um avistamento em um desaparecido
- O familiar recebe email automático com observação e localização
- O familiar pode ver um mapa com todos os pontos onde o desaparecido foi visto

Esta fase é onde a assincronicidade do Go brilha de verdade — processamento em background sem infraestrutura extra.

---

## Conceitos Go que você vai aprender nesta fase

### 1. Channels em profundidade — padrões de uso real

Na Fase 3, vimos channels básicos. Agora vamos aplicar padrões que a comunidade Go usa em produção.

#### Fan-out: uma goroutine envia para múltiplos consumers

Quando um avistamento é registrado, precisamos fazer duas coisas em paralelo:
- Enviar email para o familiar
- Atualizar contadores de estatísticas

```go
type NotificationEvent struct {
    Type    string
    Payload interface{}
}

// Um "dispatcher" que distribui eventos para múltiplos handlers
type EventDispatcher struct {
    handlers []func(ctx context.Context, event NotificationEvent) error
}

func (d *EventDispatcher) Dispatch(ctx context.Context, event NotificationEvent) {
    for _, handler := range d.handlers {
        h := handler // captura para a closure
        go h(ctx, event)
    }
}
```

#### A armadilha da closure em loops

Um erro **muito comum** em Go (e JavaScript):

```go
// ❌ ERRADO — todas as goroutines usam o mesmo `handler`
for _, handler := range d.handlers {
    go handler(ctx, event) // `handler` muda a cada iteração!
}

// ✅ CORRETO — captura o valor da iteração
for _, handler := range d.handlers {
    h := handler // cria uma cópia local
    go h(ctx, event)
}
```

Isso acontece porque a goroutine é uma closure que captura a **variável** (não o valor). Se a variável muda no loop, todas as goroutines veem o último valor.

> **Nota**: A partir do Go 1.22, esse comportamento mudou — o loop variable é scoped por iteração. Mas como é um erro tão comum e recente, vale entender o motivo.

---

### 2. Integração com APIs externas — SendGrid e Telegram

Em Go, fazer uma requisição HTTP é nativo (package `net/http`). Não precisa de requests, axios ou fetch — está na standard library.

#### Enviando email via Resend

Usamos **Resend** em vez de SendGrid — API mais moderna, melhor DX e free tier de 3.000 emails/mês.

```go
package notification

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type ResendSender struct {
    apiKey    string
    fromEmail string
    client    *http.Client
}

func NewResendSender(apiKey, fromEmail string) *ResendSender {
    return &ResendSender{
        apiKey:    apiKey,
        fromEmail: fromEmail,
        client:    &http.Client{Timeout: 10 * time.Second},
    }
}

func (s *ResendSender) Send(ctx context.Context, to, subject, htmlBody string) error {
    payload := map[string]interface{}{
        "from":    s.fromEmail,
        "to":     []string{to},
        "subject": subject,
        "html":    htmlBody,
    }

    body, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("marshaling email payload: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewReader(body))
    if err != nil {
        return fmt.Errorf("creating request: %w", err)
    }

    req.Header.Set("Authorization", "Bearer "+s.apiKey)
    req.Header.Set("Content-Type", "application/json")

    resp, err := s.client.Do(req)
    if err != nil {
        return fmt.Errorf("sending email: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return fmt.Errorf("resend returned status %d", resp.StatusCode)
    }

    return nil
}
```

Perceba como a API do Resend é mais limpa que a do SendGrid — sem `personalizations`, sem arrays aninhados. Direto ao ponto: `from`, `to`, `subject`, `html`.

#### Pontos importantes do código acima:

1. **`http.NewRequestWithContext(ctx, ...)`** — passa o context. Se o context for cancelado (ex: timeout), a requisição HTTP é abortada automaticamente.

2. **`defer resp.Body.Close()`** — sempre fechar o body da resposta. Se não fechar, vaza conexão. O `defer` garante que fecha mesmo se der erro depois.

3. **`&http.Client{Timeout: 10 * time.Second}`** — timeout global. O client padrão (`http.DefaultClient`) **não tem timeout** — pode ficar pendurado para sempre. Sempre criar um client com timeout.

4. **`resp.StatusCode >= 400`** — em Go, status HTTP de erro **não** é um erro automático. Diferente de Python requests que lança exceção com `raise_for_status()`, em Go você precisa verificar manualmente.

#### Enviando mensagem via WhatsApp Business API

O WhatsApp Business API funciona via templates pré-aprovados pela Meta. Você não manda texto livre — manda um template com variáveis.

```go
type WhatsAppSender struct {
    phoneNumberID string
    accessToken   string
    client        *http.Client
}

func NewWhatsAppSender(phoneNumberID, accessToken string) *WhatsAppSender {
    return &WhatsAppSender{
        phoneNumberID: phoneNumberID,
        accessToken:   accessToken,
        client:        &http.Client{Timeout: 10 * time.Second},
    }
}

func (w *WhatsAppSender) Send(ctx context.Context, toPhone, templateName string, params []string) error {
    url := fmt.Sprintf("https://graph.facebook.com/v18.0/%s/messages", w.phoneNumberID)

    // Monta os parâmetros do template
    components := []map[string]interface{}{}
    if len(params) > 0 {
        parameters := make([]map[string]string, len(params))
        for i, p := range params {
            parameters[i] = map[string]string{"type": "text", "text": p}
        }
        components = append(components, map[string]interface{}{
            "type":       "body",
            "parameters": parameters,
        })
    }

    payload := map[string]interface{}{
        "messaging_product": "whatsapp",
        "to":               toPhone,
        "type":             "template",
        "template": map[string]interface{}{
            "name":       templateName,
            "language":   map[string]string{"code": "pt_BR"},
            "components": components,
        },
    }

    body, _ := json.Marshal(payload)
    req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
    req.Header.Set("Authorization", "Bearer "+w.accessToken)
    req.Header.Set("Content-Type", "application/json")

    resp, err := w.client.Do(req)
    if err != nil {
        return fmt.Errorf("sending whatsapp message: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return fmt.Errorf("whatsapp api returned status %d", resp.StatusCode)
    }

    return nil
}
```

**Conceito importante**: a API do WhatsApp usa **templates pré-aprovados** com variáveis. Por exemplo, um template `sighting_alert` poderia ser:

> "Olá! Alguém informou que *{{1}}* foi visto(a) recentemente. Observação: {{2}}. Acesse a plataforma para ver a localização."

Onde `{{1}}` = nome do desaparecido e `{{2}}` = observação do avistamento. O template precisa ser submetido e aprovado pela Meta antes de poder ser usado.

Esse processo de aprovação é burocrático, mas garante que o WhatsApp não é usado para spam.

#### Enviando mensagem via Telegram

```go
type TelegramSender struct {
    botToken string
    chatID   string
    client   *http.Client
}

func (t *TelegramSender) SendMessage(ctx context.Context, message string) error {
    url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)

    payload := map[string]string{
        "chat_id":    t.chatID,
        "text":       message,
        "parse_mode": "Markdown",
    }

    body, _ := json.Marshal(payload)
    req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")

    resp, err := t.client.Do(req)
    if err != nil {
        return fmt.Errorf("sending telegram message: %w", err)
    }
    defer resp.Body.Close()

    return nil
}
```

---

### 3. Templates HTML em Go — emails bonitos

Go tem um package `html/template` na standard library para gerar HTML seguro (com escaping automático contra XSS):

```go
package notification

import (
    "bytes"
    "html/template"
)

const sightingEmailTpl = `
<div style="margin: 5px auto; text-align:center; padding: 10px; font-family: 'Raleway', sans-serif;">
    <h1>Desaparecidos.me</h1>
    <p>Alguém informou que a pessoa desaparecida foi avistada.</p>
    <h3>Observação</h3>
    <p style="background-color: #0097D6; padding: 5px; border-radius: 5px; color: black; font-weight: bold">
        {{.Observation}}
    </p>
    <p>Acesse a plataforma para ver a localização no mapa.</p>
</div>
`

const passwordResetTpl = `
<div style="margin: 5px auto; text-align:center; padding: 10px; font-family: 'Raleway', sans-serif;">
    <h1>Desaparecidos.me</h1>
    <p>Você solicitou uma nova senha.</p>
    <p>Clique no link abaixo para redefinir:</p>
    <a href="{{.ResetLink}}" style="background-color: #FF9800; padding: 10px 20px; border-radius: 5px; color: white; font-weight: bold; text-decoration: none;">
        Redefinir Senha
    </a>
</div>
`

func renderTemplate(tplStr string, data interface{}) (string, error) {
    tpl, err := template.New("email").Parse(tplStr)
    if err != nil {
        return "", err
    }

    var buf bytes.Buffer
    if err := tpl.Execute(&buf, data); err != nil {
        return "", err
    }

    return buf.String(), nil
}
```

#### `{{.Observation}}` — template syntax do Go

- `{{.Field}}` — acessa um campo da struct passada como dado
- `{{if .Condition}}...{{end}}` — condicional
- `{{range .Items}}...{{end}}` — iteração
- `{{.Field | html}}` — pipeline com funções (escaping automático em html/template)

É similar ao Jinja2 do Flask, mas mais simples. O `html/template` faz escaping automático — previne XSS sem esforço.

---

### 4. Interfaces como contratos de notificação

Definimos uma interface para o serviço de notificação no domínio:

```go
// domain/notification/notifier.go
package notification

import "context"

type Notifier interface {
    NotifySighting(ctx context.Context, userEmail string, observation string) error
    NotifyPasswordReset(ctx context.Context, userEmail string, resetLink string) error
    NotifyNewHomeless(ctx context.Context, name string, birthDate string, photoURL string, id string) error
}
```

A implementação concreta usa SendGrid + Telegram, mas o domínio não sabe disso:

```go
// infrastructure/notification/service.go
type Service struct {
    email    *EmailSender
    telegram *TelegramSender
}

func (s *Service) NotifySighting(ctx context.Context, userEmail, observation string) error {
    html, err := renderTemplate(sightingEmailTpl, map[string]string{
        "Observation": observation,
    })
    if err != nil {
        return err
    }
    return s.email.Send(ctx, userEmail, "Desaparecido foi avistado!", html)
}

func (s *Service) NotifyNewHomeless(ctx context.Context, name, birthDate, photoURL, id string) error {
    msg := fmt.Sprintf("🆕 *Novo cadastro*\n*Nome*: _%s_\n*Nascimento*: _%s_\n[Saiba mais](https://desaparecidos.me/homeless/%s)", name, birthDate, id)
    return s.telegram.SendMessage(ctx, msg)
}
```

#### Por que interface para notificação?

1. **Testes** — nos testes, criamos um mock que implementa `Notifier` e registra as chamadas sem enviar email/telegram de verdade
2. **Evolução** — amanhã podemos adicionar push notification (mobile) sem mudar o domínio
3. **Desligamento** — em desenvolvimento local, podemos usar um notifier que só loga no console

---

### 5. Goroutines para fire-and-forget vs Cloud Tasks para garantia

Quando um avistamento é registrado, o fluxo é:

```
1. Salvar avistamento no Firestore         ← SÍNCRONO (precisa dar certo)
2. Enviar email para o familiar             ← ASSÍNCRONO (não pode bloquear a resposta)
```

Temos duas opções para o passo 2:

#### Opção A: Goroutine simples (fire-and-forget)

```go
func (s *SightingService) Create(ctx context.Context, input CreateInput) (*Sighting, error) {
    sighting := // ... cria e salva ...

    // Dispara email em background
    go func() {
        bgCtx := context.Background() // novo context, independente do request
        if err := s.notifier.NotifySighting(bgCtx, user.Email, input.Observation); err != nil {
            log.Printf("ERROR: failed to send sighting email: %v", err)
        }
    }()

    return sighting, nil
}
```

**Prós**: zero infraestrutura, instantâneo
**Contras**: se o processo morrer durante o envio, o email se perde; se o SendGrid estiver fora, não retenta

#### Opção B: Cloud Tasks (garantia de entrega)

```go
func (s *SightingService) Create(ctx context.Context, input CreateInput) (*Sighting, error) {
    sighting := // ... cria e salva ...

    // Cria uma task que será processada por outro endpoint
    task := &cloudtasks.CreateTaskRequest{
        Parent: "projects/desaparecidos/locations/us-central1/queues/notifications",
        Task: &taskspb.Task{
            MessageType: &taskspb.Task_HttpRequest{
                HttpRequest: &taskspb.HttpRequest{
                    HttpMethod: taskspb.HttpMethod_POST,
                    Url:        "https://api.desaparecidos.me/internal/notify-sighting",
                    Body:       jsonPayload,
                },
            },
        },
    }
    s.taskClient.CreateTask(ctx, task)

    return sighting, nil
}
```

**Prós**: garantia de at-least-once delivery, retentativas automáticas
**Contras**: mais infraestrutura, latência para criar a task

#### Nossa decisão

**Começar com goroutine simples + logging robusto.** Para o volume do projeto (dezenas de notificações por dia no máximo), a chance de perder um email é mínima. Se o SendGrid retornar erro, logamos e podemos reenviar manualmente.

**Migrar para Cloud Tasks** se:
- O volume crescer significativamente
- Algum email se perder em produção
- Precisarmos de retentativas automáticas

A interface `Notifier` nos permite fazer essa troca sem mudar o domínio.

---

### 6. `context.Background()` vs `r.Context()` — cuidado importante

Quando disparamos uma goroutine a partir de um handler HTTP, precisamos usar um **novo context**:

```go
// ❌ ERRADO — usa o context do request
go func() {
    s.notifier.NotifySighting(ctx, email, obs) // ctx do request!
}()
// Problema: quando o handler retornar, o context do request é CANCELADO.
// A goroutine tenta enviar email com context cancelado → falha.

// ✅ CORRETO — cria um context independente
go func() {
    bgCtx := context.Background()
    // Opcionalmente, com timeout
    bgCtx, cancel := context.WithTimeout(bgCtx, 30*time.Second)
    defer cancel()
    s.notifier.NotifySighting(bgCtx, email, obs)
}()
```

Isso é um erro tão comum que merece destaque. O context do HTTP request vive **enquanto o request está sendo processado**. Quando o handler retorna a resposta, o context é cancelado. Goroutines que continuam rodando depois precisam do seu próprio context.

---

## Tarefas Detalhadas

### Backend

#### Tarefa 4.1 — Entity Sighting

Criar `internal/domain/sighted/entity.go`:
- Struct: ID, MissingID, Location (GeoPoint), Observation, CreatedAt
- Validação: MissingID obrigatório, Location válido
- Erros sentinela: `ErrSightingNotFound`

#### Tarefa 4.2 — Interface SightingRepository

```go
type Repository interface {
    Create(ctx context.Context, s *Sighting) error
    FindByID(ctx context.Context, id string) (*Sighting, error)
    FindByMissingID(ctx context.Context, missingID string) ([]*Sighting, error)
}
```

#### Tarefa 4.3 — SightingService

- Create: valida que o desaparecido existe, salva o avistamento, dispara notificação em goroutine
- FindByMissingID: retorna todos os avistamentos
- FindByID: retorna um avistamento específico

#### Tarefa 4.4 — Repositório Firestore para Sighting

Implementar a interface com Firestore.

#### Tarefa 4.5 — EmailSender (SendGrid)

Criar `internal/infrastructure/notification/email_sender.go`:
- Método `Send(ctx, to, subject, htmlBody) error`
- Timeout de 10 segundos
- Logging de sucesso/falha

#### Tarefa 4.6 — TelegramSender

Criar `internal/infrastructure/notification/telegram_sender.go`:
- Método `SendMessage(ctx, message) error`
- Formatação Markdown para Telegram

#### Tarefa 4.7 — Email templates

Criar templates HTML para:
- Avistamento de desaparecido
- (Recuperação de senha já feita na Fase 1 via Firebase Auth)

#### Tarefa 4.8 — Notification Service (implementação do Notifier)

Criar `internal/infrastructure/notification/service.go`:
- Implementa a interface `Notifier`
- Compõe EmailSender + TelegramSender
- Renderiza templates antes de enviar

#### Tarefa 4.9 — Handlers REST de Sighting

- `POST /api/v1/missing/{id}/sightings` — registrar avistamento
- `GET /api/v1/missing/{id}/sightings` — listar avistamentos de um desaparecido
- `GET /api/v1/sightings/{id}` — buscar avistamento por ID

#### Tarefa 4.10 — Testes

- Testar criação de avistamento (com mock do Notifier)
- Testar que notificação é chamada corretamente
- Testar listagem por desaparecido
- Testar com race detector (`go test -race`)

### Frontend (React)

#### Tarefa 4.11 — Formulário de avistamento

- Acessível a partir do modal de detalhes do desaparecido
- Campo de observação (textarea)
- Mapa para selecionar localização (reutilizar MapPicker)
- Botão "Informar Avistamento"
- Feedback de sucesso (toast notification)

#### Tarefa 4.12 — Página de notificações do usuário

- Rota: `/notifications`
- Mapa fullscreen com markers de avistamentos
- Cada marker mostra popup com: data, observação
- Lista lateral com avistamentos ordenados por data

#### Tarefa 4.13 — Botão "Informar avistamento" na listagem

- Na listagem de desaparecidos, cada card tem um link/botão
- Abre formulário de avistamento em modal ou nova página

---

## Decisões Específicas desta Fase

### Por que não usar Cloud Pub/Sub desde o início?

Cloud Pub/Sub é poderoso mas adiciona complexidade:
- Precisa criar topics e subscriptions
- Precisa de um endpoint HTTP para receber os messages
- Precisa lidar com acknowledgment e dead-letter queues
- Precisa pagar (mínimo, mas ainda assim)

Para **2-3 notificações por evento** (email + telegram), uma goroutine é a solução proporcional ao problema. Over-engineering seria usar Pub/Sub para isso.

### Logging robusto para goroutines de notificação

Como goroutines fire-and-forget podem falhar silenciosamente, o logging é **essencial**:

```go
go func() {
    bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    log.Printf("INFO: sending sighting notification to %s for missing %s", user.Email, missingID)

    if err := s.notifier.NotifySighting(bgCtx, user.Email, observation); err != nil {
        log.Printf("ERROR: failed to send sighting notification to %s: %v", user.Email, err)
        // TODO: Se isso acontecer frequentemente, considerar migrar para Cloud Tasks
    } else {
        log.Printf("INFO: sighting notification sent successfully to %s", user.Email)
    }
}()
```

Na Fase 7 (Deploy), vamos substituir esses logs por structured logging com correlação de request IDs.

---

## Entregáveis da Fase 4

- [ ] Entity Sighting com validações
- [ ] Interface e implementação do SightingRepository
- [ ] SightingService com notificação assíncrona
- [ ] EmailSender (SendGrid) com templates HTML
- [ ] TelegramSender com mensagem formatada
- [ ] Interface Notifier implementada
- [ ] Handlers REST para Sighting
- [ ] Testes com mock do Notifier + race detector
- [ ] React: Formulário de avistamento com mapa
- [ ] React: Página de notificações com mapa
- [ ] React: Botão de avistamento na listagem

---

## Próxima Fase

→ [FASE_05_HOMELESS.md](./FASE_05_HOMELESS.md) — Módulo Homeless ("Quero Ser Encontrado")

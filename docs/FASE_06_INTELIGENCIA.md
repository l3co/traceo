# Fase 6 — Inteligência Artificial: Age Progression & Face Matching

> **Duração estimada**: 4 semanas
> **Pré-requisito**: Fase 5 concluída (Missing + Homeless funcionando com dados reais)

---

## Objetivo

Adicionar a camada inteligente da plataforma usando **Gemini** e **Imagen** do Google. Ao final desta fase:

- Quando um desaparecido é cadastrado, a plataforma gera automaticamente **projeções visuais** de como a pessoa estaria após 1, 3, 5 e 10 anos
- Quando um homeless é cadastrado, a plataforma **compara automaticamente** com a base de desaparecidos usando visão computacional
- Se um match com score alto é encontrado, o familiar é notificado por WhatsApp/email
- O familiar pode confirmar ou rejeitar o match

Esta é a fase que transforma a plataforma de um "CRUD de cadastros" em uma **ferramenta ativa de busca**. É aqui que a tecnologia moderna faz diferença real na vida das pessoas.

---

## Conceitos Go que você vai aprender nesta fase

### 1. Worker Pattern — processamento em background estruturado

Na Fase 4, usamos goroutines fire-and-forget para notificações. Para AI, precisamos de algo mais robusto: as operações são **demoradas** (5-30 segundos por chamada ao Gemini) e **não podem bloquear** o handler HTTP.

#### O problema

```go
// ❌ RUIM — bloqueia o handler por 30 segundos
func (h *MissingHandler) Create(w http.ResponseWriter, r *http.Request) {
    missing := // ... cria o desaparecido ...

    // Isso leva 30 segundos!
    ageImages, err := h.aiService.GenerateAgeProgression(ctx, missing.PhotoURL, missing.BirthDate)

    // O usuário ficou esperando 30 segundos...
    json.NewEncoder(w).Encode(response)
}
```

```go
// ⚠️ MELHOR, mas sem controle — goroutine fire-and-forget
func (h *MissingHandler) Create(w http.ResponseWriter, r *http.Request) {
    missing := // ... cria o desaparecido ...

    go h.aiService.GenerateAgeProgression(context.Background(), missing.ID, missing.PhotoURL)
    // E se der erro? E se quisermos retry? E se quisermos limitar a 3 processamentos simultâneos?

    json.NewEncoder(w).Encode(response)
}
```

#### A solução: Worker com channel

```go
// internal/worker/ai_worker.go

type AIJob struct {
    Type      string // "age_progression" | "face_matching"
    MissingID string
    PhotoURL  string
    BirthDate time.Time
    // ... outros dados necessários
}

type AIWorker struct {
    jobs      chan AIJob
    aiService *ai.Service
    logger    *slog.Logger
    wg        sync.WaitGroup
}

func NewAIWorker(aiService *ai.Service, logger *slog.Logger, concurrency int) *AIWorker {
    w := &AIWorker{
        jobs:      make(chan AIJob, 100), // buffer de 100 jobs
        aiService: aiService,
        logger:    logger,
    }

    // Inicia N goroutines workers
    for i := 0; i < concurrency; i++ {
        w.wg.Add(1)
        go w.run(i)
    }

    return w
}

func (w *AIWorker) run(id int) {
    defer w.wg.Done()
    w.logger.Info("ai worker started", slog.Int("worker_id", id))

    for job := range w.jobs {
        ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

        w.logger.Info("processing ai job",
            slog.String("type", job.Type),
            slog.String("missing_id", job.MissingID),
            slog.Int("worker_id", id),
        )

        var err error
        switch job.Type {
        case "age_progression":
            err = w.aiService.ProcessAgeProgression(ctx, job.MissingID, job.PhotoURL, job.BirthDate)
        case "face_matching":
            err = w.aiService.ProcessFaceMatching(ctx, job.MissingID, job.PhotoURL)
        }

        if err != nil {
            w.logger.Error("ai job failed",
                slog.String("type", job.Type),
                slog.String("missing_id", job.MissingID),
                slog.String("error", err.Error()),
            )
            // TODO: retry logic ou dead-letter queue
        }

        cancel()
    }
}

// Enqueue adiciona um job na fila
func (w *AIWorker) Enqueue(job AIJob) {
    w.jobs <- job
}

// Shutdown espera todos os jobs em andamento terminarem
func (w *AIWorker) Shutdown() {
    close(w.jobs) // sinaliza para os workers pararem
    w.wg.Wait()   // espera todos terminarem
}
```

#### Como isso funciona:

1. **`make(chan AIJob, 100)`** — cria um channel buffered que funciona como fila. Até 100 jobs podem ser enfileirados sem bloquear.

2. **`for i := 0; i < concurrency; i++`** — cria N workers. Se `concurrency=3`, temos 3 goroutines processando jobs em paralelo. Isso limita a pressão na API do Gemini.

3. **`for job := range w.jobs`** — cada worker fica em loop esperando jobs do channel. Quando o channel é fechado (`close(w.jobs)`), o loop termina.

4. **`Shutdown()`** — chamado no graceful shutdown (Fase 8). Fecha o channel e espera os jobs em andamento terminarem.

#### No handler, o uso é simples:

```go
func (h *MissingHandler) Create(w http.ResponseWriter, r *http.Request) {
    missing := // ... cria e salva o desaparecido ...

    // Enfileira o job de AI — retorna instantaneamente
    h.aiWorker.Enqueue(AIJob{
        Type:      "age_progression",
        MissingID: missing.ID,
        PhotoURL:  missing.PhotoURL,
        BirthDate: missing.BirthDate,
    })

    // Responde imediatamente ao usuário
    httputil.JSON(w, http.StatusCreated, toMissingResponse(missing))
}
```

O usuário vê o cadastro criado imediatamente. As imagens de age progression aparecem depois (segundos a minutos), e o frontend pode fazer polling ou usar um banner "Gerando projeções de idade...".

#### Comparação com Celery

| Aspecto | Celery (legado) | Go Worker Pattern |
|---|---|---|
| Infraestrutura | Redis + worker process separado | Zero — goroutines no mesmo processo |
| Configuração | celeryconfig.py, BROKER_URL, etc. | ~50 linhas de código Go |
| Monitoramento | Flower (dashboard separado) | Logs estruturados (slog) |
| Retry | Decorators complexos | Lógica simples no worker |
| Escala | Mais workers = mais processos | Mais goroutines = um parâmetro |

---

### 2. Chamando o Gemini API — multimodal em Go

O Gemini é um modelo **multimodal** — aceita texto E imagens como input. Isso é fundamental para nossas duas features.

```go
// internal/infrastructure/ai/gemini_client.go

import (
    "context"
    "fmt"

    "github.com/google/generative-ai-go/genai"
    "google.golang.org/api/option"
)

type GeminiClient struct {
    client *genai.Client
    model  *genai.GenerativeModel
}

func NewGeminiClient(ctx context.Context, apiKey string) (*GeminiClient, error) {
    client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
    if err != nil {
        return nil, fmt.Errorf("creating gemini client: %w", err)
    }

    model := client.GenerativeModel("gemini-2.0-flash")
    model.SetTemperature(0.4) // mais determinístico para análise

    return &GeminiClient{client: client, model: model}, nil
}

func (g *GeminiClient) Close() error {
    return g.client.Close()
}
```

#### Análise facial para face matching

```go
func (g *GeminiClient) CompareFaces(ctx context.Context, photo1URL, photo2URL string) (*FaceComparison, error) {
    // Baixa as imagens
    img1, err := downloadImage(ctx, photo1URL)
    if err != nil {
        return nil, fmt.Errorf("downloading photo1: %w", err)
    }
    img2, err := downloadImage(ctx, photo2URL)
    if err != nil {
        return nil, fmt.Errorf("downloading photo2: %w", err)
    }

    prompt := `Analyze these two facial photos and compare them.
Consider: facial structure, eye shape and color, nose shape, lip shape, 
skin tone, face shape, distinguishing features.

Respond in JSON format:
{
    "similarity_score": 0.0-1.0,
    "analysis": "detailed explanation in Portuguese",
    "matching_features": ["feature1", "feature2"],
    "different_features": ["feature1", "feature2"],
    "confidence": "high" | "medium" | "low"
}

Be conservative with scores. Only score above 0.7 if there is strong facial resemblance.
Consider that one photo may be older (age difference is expected).`

    resp, err := g.model.GenerateContent(ctx,
        genai.ImageData("jpeg", img1),
        genai.ImageData("jpeg", img2),
        genai.Text(prompt),
    )
    if err != nil {
        return nil, fmt.Errorf("gemini compare faces: %w", err)
    }

    // Parse a resposta JSON do Gemini
    result, err := parseFaceComparison(resp)
    if err != nil {
        return nil, fmt.Errorf("parsing gemini response: %w", err)
    }

    return result, nil
}

type FaceComparison struct {
    SimilarityScore   float64  `json:"similarity_score"`
    Analysis          string   `json:"analysis"`
    MatchingFeatures  []string `json:"matching_features"`
    DifferentFeatures []string `json:"different_features"`
    Confidence        string   `json:"confidence"`
}
```

#### Pontos importantes:

1. **`genai.ImageData("jpeg", img1)`** — o Gemini aceita imagens diretamente como parte do prompt. Não precisa converter para base64 manualmente — o SDK Go faz isso.

2. **Prompt engineering** — o prompt pede resposta em JSON estruturado. Isso facilita o parsing no Go. Pedimos que seja conservador nos scores para evitar falsos positivos (seria cruel notificar um familiar com um match errado).

3. **`SetTemperature(0.4)`** — temperatura baixa = respostas mais determinísticas e consistentes. Para análise facial, queremos consistência, não criatividade.

---

### 3. Imagen 3 — geração de imagens de age progression

Imagen é o modelo de geração de imagem do Google. Diferente do Gemini (que analisa), Imagen **cria** imagens.

```go
// internal/infrastructure/ai/imagen_client.go

func (g *GeminiClient) GenerateAgeProgression(
    ctx context.Context,
    photoURL string,
    currentAge int,
    targetAge int,
    gender string,
) ([]byte, error) {
    photo, err := downloadImage(ctx, photoURL)
    if err != nil {
        return nil, err
    }

    // Primeiro: Gemini analisa a foto e descreve características
    descPrompt := fmt.Sprintf(`Describe this person's facial features in detail for age progression.
Current age: %d years old. Gender: %s.
Focus on: bone structure, eye shape, nose shape, lip shape, skin characteristics, 
hair pattern, distinguishing marks.
Be specific and detailed. Respond in English.`, currentAge, gender)

    descResp, err := g.model.GenerateContent(ctx,
        genai.ImageData("jpeg", photo),
        genai.Text(descPrompt),
    )
    if err != nil {
        return nil, err
    }

    description := extractText(descResp)

    // Segundo: Imagen gera a imagem envelhecida
    imagenModel := g.client.GenerativeModel("imagen-3.0-generate-002")

    genPrompt := fmt.Sprintf(`Photorealistic portrait of the same person described below, 
but aged to %d years old. Maintain the same facial features, just naturally aged.
Keep the same ethnicity, bone structure, and distinguishing features.

Description: %s

Style: photorealistic ID photo, front-facing, neutral background, natural lighting.`, 
        targetAge, description)

    imgResp, err := imagenModel.GenerateContent(ctx, genai.Text(genPrompt))
    if err != nil {
        return nil, fmt.Errorf("imagen generation: %w", err)
    }

    // Extrai a imagem gerada
    imageBytes := extractImage(imgResp)
    return imageBytes, nil
}
```

#### O fluxo completo de age progression:

```
1. Foto original entra
2. Gemini Vision analisa e descreve características faciais detalhadamente
3. A descrição é usada como prompt para o Imagen
4. Imagen gera a imagem envelhecida para cada faixa (+1, +3, +5, +10 anos)
5. Imagens são salvas no Cloud Storage
6. URLs são armazenadas no campo ageProgressionURLs do documento Missing
```

#### Cuidados éticos e técnicos:

1. **Não substituir a foto original** — as projeções são **complementares**, nunca substituem a foto real
2. **Disclaimer visível** — no frontend, mostrar: "Imagem gerada por IA — projeção aproximada de aparência"
3. **Qualidade varia** — para fotos de baixa qualidade, as projeções podem ser imprecisas. Mostrar um indicador de confiança
4. **Custo de API** — cada geração de imagem tem custo. Limitar a 4 projeções por desaparecido (+1, +3, +5, +10 anos)
5. **Cache** — uma vez geradas, as imagens ficam no Cloud Storage. Não regenerar a cada acesso

---

### 4. Face Matching Pipeline — o fluxo completo

Quando um homeless é cadastrado, o seguinte pipeline é executado em background:

```
┌─────────────────┐
│ Novo Homeless    │
│ cadastrado       │
└────────┬────────┘
         │
         ▼
┌─────────────────────┐
│ Filtrar candidatos   │  ← Firestore: mesmo gênero, cor de pele,
│ na base de Missing   │    faixa etária compatível (±15 anos)
└────────┬────────────┘
         │
         ▼ (máx 20 candidatos)
┌─────────────────────┐
│ Para cada candidato: │
│ Gemini CompareFaces  │  ← Compara foto homeless vs foto missing
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│ Filtrar por score    │  ← score > 0.6 → salva como match
│ > threshold          │    score > 0.8 → notifica familiar
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│ Salvar matches no    │
│ Firestore            │
│ Notificar familiares │
└─────────────────────┘
```

Em Go:

```go
// internal/domain/matching/service.go

type Service struct {
    missingRepo missing.Repository
    homelessRepo homeless.Repository
    matchRepo   MatchRepository
    gemini      *ai.GeminiClient
    notifier    notification.Notifier
    logger      *slog.Logger
}

func (s *Service) ProcessFaceMatching(ctx context.Context, homelessID string) error {
    // 1. Buscar o homeless recém-cadastrado
    h, err := s.homelessRepo.FindByID(ctx, homelessID)
    if err != nil {
        return fmt.Errorf("finding homeless %s: %w", homelessID, err)
    }

    // 2. Buscar candidatos com características similares
    candidates, err := s.missingRepo.FindCandidates(ctx, CandidateFilter{
        Gender:   h.Gender,
        Skin:     h.Skin,
        MinAge:   h.Age() - 15,
        MaxAge:   h.Age() + 15,
        Status:   missing.StatusDisappeared, // só os que ainda estão desaparecidos
        Limit:    20,
    })
    if err != nil {
        return fmt.Errorf("finding candidates: %w", err)
    }

    s.logger.Info("face matching started",
        slog.String("homeless_id", homelessID),
        slog.Int("candidates", len(candidates)),
    )

    // 3. Comparar cada candidato com Gemini
    for _, candidate := range candidates {
        comparison, err := s.gemini.CompareFaces(ctx, h.PhotoURL, candidate.PhotoURL)
        if err != nil {
            s.logger.Error("face comparison failed",
                slog.String("homeless_id", homelessID),
                slog.String("missing_id", candidate.ID),
                slog.String("error", err.Error()),
            )
            continue // não para o pipeline por um erro individual
        }

        s.logger.Info("face comparison result",
            slog.String("homeless_id", homelessID),
            slog.String("missing_id", candidate.ID),
            slog.Float64("score", comparison.SimilarityScore),
        )

        // 4. Se score relevante, salvar como match
        if comparison.SimilarityScore >= 0.6 {
            match := &Match{
                ID:             generateID(),
                HomelessID:     homelessID,
                MissingID:      candidate.ID,
                Score:          comparison.SimilarityScore,
                Status:         MatchStatusPending,
                GeminiAnalysis: comparison.Analysis,
                CreatedAt:      time.Now(),
            }

            if err := s.matchRepo.Create(ctx, match); err != nil {
                s.logger.Error("saving match failed", slog.String("error", err.Error()))
                continue
            }

            // 5. Se score alto, notificar o familiar
            if comparison.SimilarityScore >= 0.8 {
                user, _ := s.missingRepo.FindOwner(ctx, candidate.ID)
                if user != nil {
                    go func(userEmail, userPhone, missingName string, score float64) {
                        bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
                        defer cancel()
                        s.notifier.NotifyPotentialMatch(bgCtx, notification.MatchNotification{
                            UserEmail:   userEmail,
                            UserPhone:   userPhone,
                            MissingName: missingName,
                            Score:       score,
                            Analysis:    comparison.Analysis,
                        })
                    }(user.Email, user.CellPhone, candidate.Name, comparison.SimilarityScore)
                }
            }
        }
    }

    return nil
}
```

#### Pontos críticos de design:

1. **`continue` em vez de `return`** — se a comparação de UM candidato falhar, o pipeline continua com os outros. Não sacrificamos 19 comparações por causa de 1 erro.

2. **Threshold duplo (0.6 e 0.8)** — score ≥ 0.6 salva no banco para revisão manual. Score ≥ 0.8 notifica o familiar automaticamente. Isso balanceia entre não perder matches reais e não alarmar familiares com falsos positivos.

3. **Limit de 20 candidatos** — não comparar com a base inteira. Filtrar primeiro por características físicas básicas reduz dramaticamente o número de chamadas à API (custo + tempo).

4. **Rate limiting implícito** — o worker tem concurrency limitada (ex: 3). Isso evita estourar limites da API do Gemini.

---

### 5. JSON parsing de respostas do Gemini

O Gemini retorna texto. Quando pedimos JSON, ele geralmente retorna JSON válido, mas pode incluir markdown fences (\`\`\`json ... \`\`\`). Precisamos de um parser robusto:

```go
// internal/infrastructure/ai/parser.go

import (
    "encoding/json"
    "regexp"
    "strings"

    "github.com/google/generative-ai-go/genai"
)

func extractText(resp *genai.GenerateContentResponse) string {
    if resp == nil || len(resp.Candidates) == 0 {
        return ""
    }
    var parts []string
    for _, part := range resp.Candidates[0].Content.Parts {
        if text, ok := part.(genai.Text); ok {
            parts = append(parts, string(text))
        }
    }
    return strings.Join(parts, "")
}

var jsonFenceRegex = regexp.MustCompile("(?s)```(?:json)?\\s*(.+?)```")

func parseJSONFromGemini[T any](resp *genai.GenerateContentResponse) (*T, error) {
    text := extractText(resp)

    // Remove markdown fences se presentes
    if matches := jsonFenceRegex.FindStringSubmatch(text); len(matches) > 1 {
        text = matches[1]
    }

    text = strings.TrimSpace(text)

    var result T
    if err := json.Unmarshal([]byte(text), &result); err != nil {
        return nil, fmt.Errorf("parsing gemini json response: %w\nraw: %s", err, text)
    }

    return &result, nil
}
```

Uso:

```go
comparison, err := parseJSONFromGemini[FaceComparison](resp)
```

Este é um uso prático de **generics** (conceito da Fase 5). A função `parseJSONFromGemini[T]` funciona para qualquer tipo de resposta JSON que esperamos do Gemini.

---

### 6. Google Maps no Go — não tem (e não precisa)

Google Maps é **100% frontend**. O backend Go não precisa de SDK de mapas. O que o backend faz é:

- Armazenar coordenadas como `GeoPoint` no Firestore
- Retornar coordenadas na API REST como `{ "lat": -15.85, "lng": -47.91 }`
- Calcular distâncias se necessário (`math` package do Go)

O frontend React usa `@vis.gl/react-google-maps` para renderizar os mapas com os dados que recebe da API.

Para features avançadas como **heatmaps de áreas de risco** e **raio de busca**, o backend precisa retornar os dados agregados:

```go
// GET /api/v1/missing/heatmap
type HeatmapPoint struct {
    Lat    float64 `json:"lat"`
    Lng    float64 `json:"lng"`
    Weight float64 `json:"weight"` // intensidade (ex: número de desaparecimentos na região)
}

func (h *MissingHandler) Heatmap(w http.ResponseWriter, r *http.Request) {
    locations, err := h.service.GetLocationsForHeatmap(r.Context())
    if err != nil {
        httputil.Error(w, http.StatusInternalServerError, "failed to get heatmap data")
        return
    }
    httputil.JSON(w, http.StatusOK, locations)
}
```

No React:

```tsx
import { Map, useMap } from '@vis.gl/react-google-maps'

function DisappearanceHeatmap({ points }: { points: HeatmapPoint[] }) {
    const map = useMap()

    useEffect(() => {
        if (!map || !points.length) return

        const heatmap = new google.maps.visualization.HeatmapLayer({
            data: points.map(p => ({
                location: new google.maps.LatLng(p.lat, p.lng),
                weight: p.weight,
            })),
            radius: 30,
            opacity: 0.7,
        })
        heatmap.setMap(map)

        return () => heatmap.setMap(null)
    }, [map, points])

    return (
        <Map
            defaultCenter={{ lat: -15.77, lng: -47.92 }}
            defaultZoom={4}
            mapId="heatmap"
        />
    )
}
```

---

## Tarefas Detalhadas

### Backend

#### Tarefa 6.1 — GeminiClient

Criar `internal/infrastructure/ai/gemini_client.go`:
- Inicialização com API key
- Método `Close()`
- Método `DescribeFace(ctx, photoURL)` → descrição textual
- Método `CompareFaces(ctx, photo1URL, photo2URL)` → FaceComparison
- Parser robusto para JSON do Gemini (com tratamento de markdown fences)

#### Tarefa 6.2 — Age Progression Service

Criar `internal/domain/ai/age_progression.go`:
- `ProcessAgeProgression(ctx, missingID, photoURL, birthDate)`:
  1. Calcular idade atual
  2. Definir faixas (+1, +3, +5, +10 anos a partir da data de desaparecimento)
  3. Para cada faixa: chamar Gemini para descrever → Imagen para gerar
  4. Upload das imagens geradas para Cloud Storage
  5. Atualizar campo `ageProgressionURLs` no documento Missing

#### Tarefa 6.3 — Face Matching Service

Criar `internal/domain/matching/service.go`:
- Entity `Match` com campos: ID, HomelessID, MissingID, Score, Status, GeminiAnalysis, CreatedAt, ReviewedAt
- Interface `MatchRepository`
- `ProcessFaceMatching(ctx, homelessID)`:
  1. Buscar homeless
  2. Filtrar candidatos missing por características
  3. Comparar faces com Gemini
  4. Salvar matches com score ≥ 0.6
  5. Notificar familiares para score ≥ 0.8

#### Tarefa 6.4 — Match Repository (Firestore)

Criar `internal/infrastructure/firestore/match_repository.go`:
- Create, FindByID, FindByHomelessID, FindByMissingID, UpdateStatus

#### Tarefa 6.5 — AI Worker

Criar `internal/worker/ai_worker.go`:
- Worker com channel e concurrency configurável
- Processa jobs de age_progression e face_matching
- Shutdown graceful
- Integrar no main.go

#### Tarefa 6.6 — Endpoint para status do desaparecido

Criar `PATCH /api/v1/missing/:id/status`:
- Body: `{ "status": "found" }` ou `{ "status": "disappeared" }`
- Só o dono do cadastro (userId) pode alterar
- Quando marcado como "found", cancelar processamentos AI pendentes

#### Tarefa 6.7 — Endpoints de age progression

- `GET /api/v1/missing/:id/age-progression` — retorna URLs das imagens geradas
- `POST /api/v1/missing/:id/age-progression` — regenerar (trigger manual)

#### Tarefa 6.8 — Endpoints de matching

- `GET /api/v1/homeless/:id/matches` — lista candidatos de match para um homeless
- `GET /api/v1/missing/:id/matches` — lista matches encontrados para um missing
- `PATCH /api/v1/matches/:id` — confirmar ou rejeitar match (body: `{ "status": "confirmed" | "rejected" }`)

#### Tarefa 6.9 — Endpoint de heatmap

- `GET /api/v1/missing/heatmap` — retorna array de pontos com peso para heatmap

#### Tarefa 6.10 — Notificação de match

Adicionar ao `MultiChannelNotifier`:
- `NotifyPotentialMatch(ctx, params MatchNotification)` — envia WhatsApp + email quando score ≥ 0.8
- Template WhatsApp: "Encontramos uma possível correspondência para {nome}. Acesse a plataforma para verificar."
- Template email: inclui as duas fotos lado a lado + análise do Gemini

#### Tarefa 6.11 — Testes

- Mock do GeminiClient (interface) para testes unitários
- Testar pipeline de matching: filtro de candidatos → comparação → salvamento
- Testar thresholds: score < 0.6 não salva, 0.6-0.8 salva sem notificar, > 0.8 notifica
- Testar age progression: verificar que imagens são salvas e URLs atualizadas
- Testar worker: enqueue, processamento, shutdown graceful

### Frontend (React)

#### Tarefa 6.12 — Google Maps Setup

Substituir Mapbox por Google Maps em todo o projeto:

```bash
npm install @vis.gl/react-google-maps
npm uninstall react-map-gl mapbox-gl  # se já instalados
```

Componentes a criar/migrar:
- `MapPicker` — seleção de localização (click to place marker)
- `MapView` — visualização estática (marker em posição)
- `MapHeatmap` — mapa de calor com áreas de risco
- `MapClusters` — markers agrupados por proximidade

#### Tarefa 6.13 — Galeria de age progression no card de detalhes

No modal de detalhes do desaparecido:
- Seção "Como pode estar hoje" com slider de tempo
- Foto original à esquerda, projeção à direita
- Slider: +1 ano, +3 anos, +5 anos, +10 anos
- Disclaimer: "Imagem gerada por IA — projeção aproximada"
- Loading state: "Gerando projeções..." com skeleton
- Botão "Regenerar" para o dono do cadastro

#### Tarefa 6.14 — Página de matches para o familiar

Rota: `/missing/:id/matches`
- Lista de candidatos de match ordenados por score
- Cada card: foto do homeless + foto do missing lado a lado
- Score de similaridade visual (barra de progresso)
- Análise do Gemini em texto
- Botões: "Confirmar — é ele/ela!" e "Não é"
- Ao confirmar: atualiza status do match + opção de marcar missing como "found"

#### Tarefa 6.15 — Heatmap de áreas de risco

Nova página ou seção no mapa geral:
- Mapa fullscreen com heatmap layer
- Toggle: markers individuais ↔ heatmap
- Legenda de intensidade (verde → amarelo → vermelho)
- Filtro por período (último ano, últimos 5 anos, todos)

#### Tarefa 6.16 — Botão de alterar status

No card de detalhes do desaparecido (só visível para o dono):
- Botão "Marcar como encontrado" (quando status = disappeared)
- Botão "Reativar busca" (quando status = found)
- Confirmação antes de alterar ("Tem certeza?")
- Ao marcar como encontrado: celebração visual (confetti? mensagem positiva)

---

## Decisões Específicas desta Fase

### Por que Gemini e não OpenAI Vision?

| Aspecto | Gemini | OpenAI Vision |
|---|---|---|
| **Integração GCP** | ✅ Nativa (mesmo projeto, mesmo billing) | ❌ API separada |
| **SDK Go** | Oficial (`google/generative-ai-go`) | Não-oficial ou HTTP direto |
| **Custo** | Free tier generoso, depois competitivo | Sem free tier, mais caro |
| **Imagen** | ✅ Integrado (geração de imagem) | DALL-E (separado) |
| **Multimodal** | ✅ Nativo | ✅ Nativo |

Estamos no ecossistema Google. Gemini + Imagen é a escolha natural — um SDK, uma conta, um billing.

### Threshold de matching: por que 0.6 e 0.8?

- **< 0.6** — descarta. Sem semelhança relevante.
- **0.6 - 0.8** — salva como match pendente. O familiar pode ver na seção "Possíveis correspondências", mas não recebe notificação push. Evita alarme falso.
- **≥ 0.8** — notifica automaticamente. Alta confiança de que pode ser a mesma pessoa.

Esses valores são **iniciais**. Em produção, ajustaremos baseado em feedback real. Se familiares reportarem muitos falsos positivos em 0.8, subimos para 0.85. Se reportarem que perdemos matches reais, descemos para 0.75.

### Custo estimado da API

| Operação | Custo estimado | Frequência |
|---|---|---|
| Age progression (4 imagens) | ~$0.05-0.10 | 1x por cadastro de missing |
| Face matching (20 comparações) | ~$0.02-0.05 | 1x por cadastro de homeless |
| Gemini vision (análise) | ~$0.001-0.005 por imagem | Por comparação |

Para o volume esperado (dezenas de cadastros por mês), o custo é negligível — menos de $5/mês.

### Privacidade e LGPD

Usar AI para comparação facial levanta questões de privacidade:

1. **Consentimento** — o familiar que cadastra o desaparecido consente ao aceitar os termos
2. **Transparência** — explicar na política de privacidade que fotos são analisadas por AI
3. **Propósito legítimo** — encontrar pessoas desaparecidas é um propósito legítimo sob LGPD
4. **Fotos não saem do GCP** — Gemini roda no mesmo GCP, as fotos não vão para terceiros
5. **Direito de exclusão** — ao deletar um cadastro, deletar também os dados de AI

---

## Entregáveis da Fase 6

- [ ] GeminiClient com análise facial e comparação
- [ ] Age Progression: descrição + geração de imagens por faixa etária
- [ ] Face Matching: pipeline completo (filtro → comparação → salvamento → notificação)
- [ ] Match entity + repository + endpoints
- [ ] AI Worker com channel e concurrency
- [ ] Endpoint PATCH de status (disappeared ↔ found)
- [ ] Endpoint de heatmap
- [ ] Testes com mock do GeminiClient
- [ ] React: Google Maps (MapPicker, MapView, MapHeatmap, MapClusters)
- [ ] React: Galeria de age progression com slider
- [ ] React: Página de matches com confirmação/rejeição
- [ ] React: Heatmap de áreas de risco
- [ ] React: Botão de alterar status do desaparecido

---

## Próxima Fase

→ [FASE_07_POLISH.md](./FASE_07_POLISH.md) — Páginas Institucionais, SEO & Acessibilidade

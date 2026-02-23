# Fase 3 — Busca Textual & Dashboard

> **Duração estimada**: 3 semanas
> **Pré-requisito**: Fase 2 concluída (CRUD de desaparecidos funcionando)

---

## Objetivo

Implementar busca textual de desaparecidos e um dashboard com estatísticas visuais. Ao final desta fase:

- Barra de busca global que encontra desaparecidos por nome/apelido
- Dashboard com totais, gráficos por gênero, crianças, e desaparecidos por ano
- Mapa geral com clusters de regiões de maior incidência de desaparecimento

Esta é a fase onde a concorrência do Go começa a brilhar — buscar dados de múltiplas fontes em paralelo.

---

## Conceitos Go que você vai aprender nesta fase

### 1. Goroutines — concorrência leve e nativa

Uma goroutine é a forma Go de executar algo em paralelo. Diferente de threads do sistema operacional (que custam ~1MB cada), goroutines custam **~2KB** cada. Você pode criar milhares sem problema.

```go
// Executar uma função em paralelo
go minhaFuncao()

// Executar uma função anônima em paralelo
go func() {
    fmt.Println("rodando em paralelo")
}()
```

A palavra `go` antes de qualquer chamada de função cria uma goroutine. É isso. Não precisa de library, não precisa de framework, não precisa de `async/await`.

#### Comparação com outras linguagens

| Linguagem | Mecanismo | Custo por "thread" | Complexidade |
|---|---|---|---|
| Python | asyncio / threading | ~8MB (thread) ou event loop | `async/await` em toda a cadeia |
| Node.js | Event loop + callbacks | 1 thread | Promises, async/await |
| Java | Threads / Virtual Threads | ~1MB (thread) / ~2KB (virtual) | Thread pools, executors |
| **Go** | **Goroutines** | **~2KB** | **`go func()`** |

Em Python/Flask (projeto legado), para enviar um email sem bloquear a resposta HTTP, você precisou de **Redis + Celery + worker process separado**. Em Go:

```go
func (h *SightingHandler) Create(w http.ResponseWriter, r *http.Request) {
    // ... cria o avistamento ...

    // Envia email em paralelo — NÃO bloqueia a resposta
    go h.notifier.SendSightingEmail(ctx, userEmail, observation)

    // Responde imediatamente
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(response)
}
```

Zero infraestrutura extra. A goroutine roda em background enquanto o handler já respondeu.

#### O perigo: goroutines sem controle

Se você cria goroutines sem esperar que terminem, não sabe se deram erro, se travaram, se estão vazando memória. Para isso existem mecanismos de sincronização.

---

### 2. `sync.WaitGroup` — esperando múltiplas goroutines

No dashboard, precisamos buscar estatísticas de **múltiplas fontes em paralelo**:
- Total de desaparecidos
- Total de homeless
- Contagem por gênero
- Contagem por ano
- Contagem de crianças

Fazer essas 5 queries **sequencialmente** é lento. Fazer em paralelo é o ideal.

`WaitGroup` é um contador de goroutines pendentes:

```go
func (s *StatsService) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
    var (
        wg           sync.WaitGroup
        stats        DashboardStats
        missingCount int64
        homelessCount int64
        genderStats  []GenderStat
        yearStats    []YearStat
        errCh        = make(chan error, 5) // channel para coletar erros
    )

    // Query 1: total de desaparecidos
    wg.Add(1)
    go func() {
        defer wg.Done()
        count, err := s.missingRepo.Count(ctx)
        if err != nil {
            errCh <- fmt.Errorf("counting missing: %w", err)
            return
        }
        missingCount = count
    }()

    // Query 2: total de homeless
    wg.Add(1)
    go func() {
        defer wg.Done()
        count, err := s.homelessRepo.Count(ctx)
        if err != nil {
            errCh <- fmt.Errorf("counting homeless: %w", err)
            return
        }
        homelessCount = count
    }()

    // Query 3: por gênero
    wg.Add(1)
    go func() {
        defer wg.Done()
        gs, err := s.missingRepo.CountByGender(ctx)
        if err != nil {
            errCh <- fmt.Errorf("counting by gender: %w", err)
            return
        }
        genderStats = gs
    }()

    // ... mais queries ...

    // Espera TODAS as goroutines terminarem
    wg.Wait()
    close(errCh)

    // Verifica se alguma deu erro
    for err := range errCh {
        if err != nil {
            return nil, err
        }
    }

    stats.MissingTotal = missingCount
    stats.HomelessTotal = homelessCount
    stats.ByGender = genderStats
    // ...

    return &stats, nil
}
```

#### Como isso funciona passo a passo:

1. `wg.Add(1)` → incrementa o contador ("tenho mais 1 goroutine pendente")
2. `go func() { defer wg.Done(); ... }()` → cria goroutine. `defer wg.Done()` decrementa o contador quando terminar
3. `wg.Wait()` → bloqueia até o contador chegar a zero (todas terminaram)

As 5 queries rodam **ao mesmo tempo**. Se cada uma leva 100ms, o total é ~100ms (não 500ms).

#### O `defer` — executar algo ao sair da função

`defer` agenda uma chamada para ser executada quando a função encerrar (por qualquer motivo):

```go
func doSomething() {
    file, _ := os.Open("arquivo.txt")
    defer file.Close() // será executado ao sair da função, não importa como

    // ... usa o arquivo ...
    // se der panic, retornar, ou chegar ao fim, file.Close() é chamado
}
```

É o equivalente do `try/finally` em Python/Java, mas mais elegante. Você coloca o "cleanup" logo após o "open", não lá embaixo no `finally`.

---

### 3. Channels — comunicação entre goroutines

Channels são como "tubos" por onde goroutines trocam dados de forma segura:

```go
// Criar um channel que transporta strings
ch := make(chan string)

// Goroutine envia dados pelo channel
go func() {
    ch <- "resultado da busca"  // envia
}()

// Main thread recebe
resultado := <-ch  // bloqueia até receber
fmt.Println(resultado)
```

#### Channels buffered vs unbuffered

```go
// Unbuffered — envia BLOQUEIA até alguém receber
ch := make(chan int)

// Buffered — envia NÃO bloqueia se o buffer não está cheio
ch := make(chan int, 5) // buffer de 5 itens
```

No exemplo do dashboard, usamos um channel buffered para coletar erros:

```go
errCh := make(chan error, 5) // pode armazenar até 5 erros sem bloquear
```

#### A frase mais importante do Go sobre concorrência

> *"Don't communicate by sharing memory; share memory by communicating."*

Em vez de usar mutex/locks para proteger dados compartilhados (como em Java/C++), em Go preferimos enviar dados entre goroutines via channels. Isso elimina race conditions by design.

---

### 4. Busca textual — por que Firestore não resolve sozinho

Firestore não tem full-text search. Ele faz queries exatas, com prefixo (`>=` e `<`), mas não busca "João" dentro de "João Carlos Silva".

#### Opções consideradas

| Opção | Prós | Contras | Veredicto |
|---|---|---|---|
| **Firestore prefix query** | Nativo, sem custo extra | Só busca do início ("João" acha, "Carlos" não) | ❌ Para busca real |
| **Algolia** | Full-text search excelente, typo-tolerant | Pago após free tier (10K queries/mês) | ✅ Se volume justificar |
| **Typesense** | Open-source, self-hosted ou cloud | Precisa hospedar ou pagar | ✅ Alternativa ao Algolia |
| **Meilisearch** | Open-source, rápido | Precisa hospedar | ⚠️ Mais novo, menos maduro |
| **Array de keywords no Firestore** | Nativo | Gambiarra, não escala | ❌ |

#### Decisão: começar com Firestore prefix query + evoluir para Algolia

Para o volume inicial (centenas de registros), uma **busca por prefixo** é suficiente:

```go
func (r *firestoreMissingRepo) Search(ctx context.Context, query string) ([]*Missing, error) {
    query = strings.ToLower(query)
    docs, err := r.client.Collection("missing").
        Where("name_lowercase", ">=", query).
        Where("name_lowercase", "<=", query+"\uf8ff").
        Limit(20).
        Documents(ctx).GetAll()
    // ...
}
```

O truque do `\uf8ff` é que esse é o último caractere Unicode. Então `"joão"` a `"joão\uf8ff"` captura tudo que começa com "joão".

**Quando migrar para Algolia**: quando a busca por prefixo não for suficiente (ex: buscar por "Silva" no meio do nome, busca com typo tolerance). Isso pode ser na Fase 6 ou posterior.

#### A interface protege a troca

```go
// domain/missing/repository.go
type Repository interface {
    // ...
    Search(ctx context.Context, query string, limit int) ([]*Missing, error)
}
```

Hoje, `Search` usa Firestore prefix query. Amanhã, pode usar Algolia. O service e o handler não mudam.

---

### 5. Aggregations no Firestore — distributed counters

O projeto legado usa MongoDB aggregation pipelines para calcular estatísticas:

```python
# Legado: MongoDB
pipeline = [{"$group": {"_id": "$gender", "count": {"$sum": 1}}}]
db.disappeared.aggregate(pipeline)
```

Firestore **não tem aggregation pipeline**. Temos duas abordagens:

#### Opção A: Contar no cliente

```go
docs, _ := client.Collection("missing").Documents(ctx).GetAll()
maleCount := 0
for _, doc := range docs {
    if doc.Data()["gender"] == "male" {
        maleCount++
    }
}
```

**Problema**: lê TODOS os documentos. Se tiver 10.000, lê 10.000. Caro e lento.

#### Opção B: Distributed Counters (nossa escolha)

Manter um documento de "contadores" que é atualizado sempre que um registro é criado/deletado:

```
stats/missing
  ├── total: 1234
  ├── byGender:
  │   ├── male: 678
  │   └── female: 556
  ├── byYear:
  │   ├── "2018": 150
  │   ├── "2019": 200
  │   └── ...
  └── childCount: 342
```

Quando criamos um desaparecido, **atualizamos os contadores atomicamente**:

```go
func (s *Service) Create(ctx context.Context, input CreateInput) (*Missing, error) {
    m := // ... cria a entidade ...

    // Transação: cria o documento E atualiza contadores
    err := s.firestore.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
        // 1. Cria o desaparecido
        if err := tx.Set(missingRef, m); err != nil {
            return err
        }

        // 2. Atualiza contadores
        statsRef := s.firestore.Collection("stats").Doc("missing")
        tx.Update(statsRef, []firestore.Update{
            {Path: "total", Value: firestore.Increment(1)},
            {Path: "byGender." + string(m.Gender), Value: firestore.Increment(1)},
            {Path: "byYear." + strconv.Itoa(m.DateOfDisappearance.Year()), Value: firestore.Increment(1)},
        })
        if m.WasChild {
            tx.Update(statsRef, []firestore.Update{
                {Path: "childCount", Value: firestore.Increment(1)},
            })
        }
        return nil
    })

    return m, err
}
```

**Vantagem**: consultar estatísticas é **uma leitura de um documento**. Instantâneo.

**Trade-off**: toda escrita precisa atualizar os contadores. Mais complexo, mas o volume do projeto não justifica preocupação com performance de escrita.

#### Firestore Count Queries (novidade recente)

Firestore adicionou `Count()` queries em 2023:

```go
countQuery := client.Collection("missing").Where("gender", "==", "male").Count(ctx)
```

Isso é mais simples que distributed counters e pode ser suficiente para nosso volume. Vamos avaliar durante a implementação.

---

## Tarefas Detalhadas

### Backend

#### Tarefa 3.1 — Endpoint de busca

Criar `GET /api/v1/missing/search?q=texto`:
- Firestore prefix query por nome (lowercase)
- Limite de 20 resultados
- Campo `name_lowercase` no documento (gerado no Create/Update)

#### Tarefa 3.2 — Endpoint de estatísticas

Criar `GET /api/v1/missing/stats`:
- Retorna: total, por gênero, crianças, por ano
- Usa distributed counters ou Count queries
- Goroutines paralelas para buscar múltiplos dados

Resposta esperada:
```json
{
  "total": 1234,
  "by_gender": {
    "male": 678,
    "female": 556
  },
  "child_count": 342,
  "by_year": [
    {"year": 2018, "count": 150},
    {"year": 2019, "count": 200}
  ]
}
```

#### Tarefa 3.3 — Endpoint de geolocalização

Criar `GET /api/v1/missing/locations`:
- Retorna array de coordenadas com dados mínimos (nome, data, status)
- Usado pelo mapa de desaparecimento
- Paginação por lote (batch de 100)

Resposta esperada:
```json
{
  "locations": [
    {
      "id": "abc123",
      "name": "João Silva",
      "lat": -15.85,
      "lng": -47.91,
      "status": "disappeared"
    }
  ]
}
```

#### Tarefa 3.4 — Distributed counters ou Count queries

Avaliar e implementar a melhor abordagem para o volume esperado:
- Se < 1000 documentos: Count queries são suficientes
- Se > 1000 documentos: distributed counters para performance

#### Tarefa 3.5 — Testes

- Testar busca com diferentes queries
- Testar estatísticas com dados mock
- Testar goroutines paralelas (race detector: `go test -race ./...`)

### Frontend (React)

#### Tarefa 3.6 — Barra de busca global

- Input com debounce (300ms) para não fazer request a cada tecla
- Dropdown com resultados (nome + foto miniatura)
- Click abre o modal de detalhes
- Keyboard navigation (seta cima/baixo, Enter)
- Usando TanStack Query para cache

#### Tarefa 3.7 — Página de Dashboard

Layout baseado no legado (`info.html`), mas modernizado:

- **Card "Total de Desaparecidos"** — número grande, ícone
- **Card "Total de Moradores de Rua"** — número grande, ícone
- **Card "Crianças Desaparecidas"** — número, ícone
- **Gráfico de pizza/donut** — por gênero
- **Gráfico de barras** — desaparecidos por ano

Usar **Recharts** (biblioteca de gráficos para React):
```bash
npm install recharts
```

#### Tarefa 3.8 — Mapa geral de desaparecimento

- Mapa fullscreen com clusters
- Cada ponto é um desaparecido
- Zoom agrupa em clusters (Mapbox clustering ou deck.gl)
- Hover mostra popup com nome + data
- Click abre modal de detalhes

#### Tarefa 3.9 — Integração busca + listagem

- Resultado da busca redireciona para listagem filtrada
- URL reflete o estado da busca: `/missing?q=joão`
- Limpar busca retorna para listagem completa

---

## Decisões Específicas desta Fase

### Por que Recharts e não Chart.js ou D3?

| Biblioteca | Estilo | Prós | Contras |
|---|---|---|---|
| Recharts | Componentes React | Declarativo, fácil, responsivo | Menos customizável |
| Chart.js | Canvas-based | Popular, muitos exemplos | Imperativo, não React-native |
| D3 | SVG manipulation | Poder total | Curva de aprendizado enorme |
| Nivo | Componentes React (D3) | Bonito, temas | Pesado |

**Recharts** é a escolha porque:
- API declarativa (componentes React)
- Responsivo por padrão
- Leve
- Suficiente para gráficos simples (barras, pizza)

O legado usava **Chartist.js** (lib antiga, não mantida). Recharts é a evolução natural.

### Race detector — ferramenta essencial para concorrência

Go tem uma ferramenta built-in que detecta race conditions:

```bash
go test -race ./...
```

Isso roda os testes com instrumentação especial que detecta acessos simultâneos à mesma memória. **Sempre rodar com `-race` durante desenvolvimento.**

Se o race detector encontrar algo, o teste falha com um relatório detalhado mostrando exatamente quais goroutines acessaram o mesmo dado.

---

## Entregáveis da Fase 3

- [ ] Endpoint de busca textual (prefix query)
- [ ] Endpoint de estatísticas (com goroutines paralelas)
- [ ] Endpoint de geolocalização
- [ ] Distributed counters ou Count queries implementados
- [ ] Testes com race detector
- [ ] React: Barra de busca global com debounce
- [ ] React: Dashboard com cards e gráficos
- [ ] React: Mapa geral com clusters

---

## Próxima Fase

→ [FASE_04_AVISTAMENTOS.md](./FASE_04_AVISTAMENTOS.md) — Avistamentos & Notificações Assíncronas

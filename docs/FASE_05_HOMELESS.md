# Fase 5 — Homeless ("Quero Ser Encontrado")

> **Duração estimada**: 2 semanas
> **Pré-requisito**: Fase 4 concluída (avistamentos + notificações funcionando)

---

## Objetivo

Implementar o módulo de moradores de rua/desabrigados — pessoas que querem ser encontradas pela família. ONGs e voluntários cadastram essas pessoas na plataforma. Ao final desta fase:

- Cadastro de homeless com foto, dados e localização no mapa
- Listagem com cards e mapa individual por registro
- Notificação via Telegram quando novo homeless é cadastrado
- Seção de estatísticas no dashboard

Esta é uma fase de **consolidação**. Os patterns já foram estabelecidos nas fases anteriores. Agora vamos aplicá-los com fluência.

---

## Conceitos Go que você vai aprender nesta fase

### 1. Reutilização de código — quando extrair e quando copiar

Um dos princípios mais citados em engenharia de software é o DRY (Don't Repeat Yourself). Mas em Go, a comunidade tem uma visão mais pragmática:

> *"A little copying is better than a little dependency."* — Go Proverbs

#### O que isso significa na prática?

Olhando nosso código até agora, `Missing` e `Homeless` compartilham muitos campos:

```go
// Missing
type Missing struct {
    Name     string
    Nickname string
    BirthDate time.Time
    Gender   Gender
    Eyes     EyeColor
    Hair     HairColor
    Skin     SkinColor
    PhotoURL string
    Location GeoPoint
    Slug     string
    // ... campos exclusivos de Missing (status, clothes, height, etc.)
}

// Homeless
type Homeless struct {
    Name     string
    Nickname string
    BirthDate time.Time
    Gender   Gender
    Eyes     EyeColor
    Hair     HairColor
    Skin     SkinColor
    PhotoURL string
    Location GeoPoint
    Slug     string
}
```

A tentação é criar uma struct base e embeddar:

```go
// ⚠️ Tentação: abstrair prematuramente
type PersonProfile struct {
    Name     string
    Nickname string
    BirthDate time.Time
    Gender   Gender
    Eyes     EyeColor
    Hair     HairColor
    Skin     SkinColor
    PhotoURL string
    Location GeoPoint
    Slug     string
}

type Missing struct {
    PersonProfile          // embed
    UserID string
    Status Status
    // ...
}

type Homeless struct {
    PersonProfile          // embed
}
```

#### Por que NÃO vamos fazer isso (agora)?

1. **Acoplamento oculto** — se mudarmos algo no `PersonProfile` por causa do `Missing`, o `Homeless` também muda. Os dois domínios evoluem de forma independente.

2. **Complexidade prematura** — temos **dois** tipos que compartilham campos. Se tivéssemos 5 ou 10, aí sim faria sentido extrair. Com 2, a duplicação é aceitável.

3. **Clareza** — abrindo `homeless/entity.go`, eu vejo TODOS os campos do Homeless sem precisar navegar para outra struct. Auto-contido.

4. **Go idiomático** — a comunidade Go prefere duplicar 10 linhas de struct a criar uma abstração que adiciona indireção.

#### Quando abstrair?

- Se um **terceiro** tipo aparecer com os mesmos campos → hora de extrair
- Se a lógica de **validação** for idêntica e mudar junto → extrair para um shared validator
- Se os **value objects** (Gender, EyeColor, etc.) já estão compartilhados → isso é bom, eles são do domínio compartilhado

Os value objects (`Gender`, `EyeColor`, etc.) SIM são compartilhados. Eles ficam em um package acessível por ambos os domínios:

```
internal/
  domain/
    shared/
      value_objects.go    ← Gender, EyeColor, HairColor, SkinColor
    missing/
      entity.go           ← usa shared.Gender
    homeless/
      entity.go           ← usa shared.Gender
```

Ou ficam no package de quem definiu primeiro (missing) e o homeless importa. Vamos decidir durante a implementação.

---

### 2. Generics em Go — quando usar (e quando não)

Go 1.18 (2022) introduziu **generics** (type parameters). Vamos ver onde faz sentido no nosso projeto.

#### Exemplo sem generics (repetição)

Nossos repositórios fazem coisas similares:

```go
// user_repository.go
func docsToUsers(docs []*firestore.DocumentSnapshot) ([]*User, error) {
    users := make([]*User, 0, len(docs))
    for _, doc := range docs {
        var u User
        if err := doc.DataTo(&u); err != nil {
            return nil, err
        }
        u.ID = doc.Ref.ID
        users = append(users, &u)
    }
    return users, nil
}

// missing_repository.go — MESMA lógica, tipo diferente
func docsToMissing(docs []*firestore.DocumentSnapshot) ([]*Missing, error) {
    items := make([]*Missing, 0, len(docs))
    for _, doc := range docs {
        var m Missing
        if err := doc.DataTo(&m); err != nil {
            return nil, err
        }
        m.ID = doc.Ref.ID
        items = append(items, &m)
    }
    return items, nil
}
```

#### Com generics

```go
// pkg/firestoreutil/convert.go

type Identifiable interface {
    SetID(id string)
}

func DocsTo[T Identifiable](docs []*firestore.DocumentSnapshot) ([]*T, error) {
    items := make([]*T, 0, len(docs))
    for _, doc := range docs {
        var item T
        if err := doc.DataTo(&item); err != nil {
            return nil, err
        }
        item.SetID(doc.Ref.ID)
        items = append(items, &item)
    }
    return items, nil
}
```

Uso:

```go
users, err := firestoreutil.DocsTo[User](docs)
missing, err := firestoreutil.DocsTo[Missing](docs)
homeless, err := firestoreutil.DocsTo[Homeless](docs)
```

#### Quando usar generics em Go?

| Situação | Generics? | Por quê |
|---|---|---|
| Funções utilitárias (map, filter, contains) | ✅ Sim | Evita repetir a mesma lógica para cada tipo |
| Conversão de documentos Firestore | ✅ Sim | Lógica idêntica, tipo diferente |
| Entidades de domínio | ❌ Não | Cada entidade tem comportamento específico |
| Services | ❌ Não | Cada service tem regras de negócio diferentes |
| Handlers | ❌ Não | Cada handler lida com DTOs diferentes |

A regra: **generics para infraestrutura/utilitário, não para domínio**.

---

### 3. Refatoração guiada — identificando smells

Neste ponto do projeto, com 4 domínios implementados, é hora de olhar para trás e refatorar. Smells comuns que podem ter aparecido:

#### Código duplicado entre handlers

Todos os handlers fazem: decodificar JSON → validar → chamar service → codificar response. Se o boilerplate for grande, podemos extrair helpers:

```go
// pkg/httputil/respond.go

func JSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func Error(w http.ResponseWriter, status int, message string) {
    JSON(w, status, map[string]string{"error": message})
}

func Decode(r *http.Request, dst interface{}) error {
    if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
        return fmt.Errorf("decoding request body: %w", err)
    }
    return nil
}
```

Uso no handler:

```go
func (h *HomelessHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateHomelessRequest
    if err := httputil.Decode(r, &req); err != nil {
        httputil.Error(w, http.StatusBadRequest, "invalid request body")
        return
    }

    homeless, err := h.service.Create(r.Context(), req.toInput())
    if err != nil {
        // mapear erro de domínio para HTTP status
        httputil.Error(w, mapErrorToStatus(err), err.Error())
        return
    }

    httputil.JSON(w, http.StatusCreated, toHomelessResponse(homeless))
}
```

#### Mapeamento de erros de domínio para HTTP

Em vez de cada handler ter `if/else` para mapear erros, centralizamos:

```go
func mapErrorToStatus(err error) int {
    switch {
    case errors.Is(err, user.ErrUserNotFound),
         errors.Is(err, missing.ErrMissingNotFound),
         errors.Is(err, homeless.ErrHomelessNotFound):
        return http.StatusNotFound

    case errors.Is(err, user.ErrEmailAlreadyExists):
        return http.StatusConflict

    case errors.Is(err, user.ErrInvalidPassword),
         errors.Is(err, missing.ErrInvalidMissing):
        return http.StatusBadRequest

    default:
        return http.StatusInternalServerError
    }
}
```

---

## Tarefas Detalhadas

### Backend

#### Tarefa 5.1 — Entity Homeless

Criar `internal/domain/homeless/entity.go`:
- Struct com campos: ID, Name, Nickname, BirthDate, Gender, Eyes, Hair, Skin, PhotoURL, Location, Slug, CreatedAt, UpdatedAt
- Métodos: `Age()`, `GenerateSlug()`, `Validate()`
- Erros sentinela: `ErrHomelessNotFound`
- Reutilizar value objects (Gender, EyeColor, etc.) da Fase 2

#### Tarefa 5.2 — Interface HomelessRepository

```go
type Repository interface {
    Create(ctx context.Context, h *Homeless) error
    FindByID(ctx context.Context, id string) (*Homeless, error)
    FindAll(ctx context.Context) ([]*Homeless, error)
    Count(ctx context.Context) (int64, error)
    CountByGender(ctx context.Context) (map[Gender]int64, error)
}
```

**Nota**: diferente do Missing, Homeless **não tem paginação** (o legado não tinha). Se o volume crescer, adicionamos depois.

#### Tarefa 5.3 — HomelessService

- Create: valida, gera slug, salva, dispara notificação Telegram
- FindByID, FindAll, Count, CountByGender

A notificação Telegram usa a mesma interface `Notifier` da Fase 4:

```go
func (s *Service) Create(ctx context.Context, input CreateInput) (*Homeless, error) {
    h := // ... cria e valida a entidade ...

    if err := s.repo.Create(ctx, h); err != nil {
        return nil, err
    }

    // Notifica via Telegram em background
    go func() {
        bgCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
        defer cancel()
        if err := s.notifier.NotifyNewHomeless(bgCtx, h.Name, h.BirthDate.Format("02/01/2006"), h.PhotoURL, h.ID); err != nil {
            log.Printf("ERROR: failed to notify new homeless: %v", err)
        }
    }()

    // Enfileira job de face matching com a base de desaparecidos (Fase 6)
    s.aiWorker.Enqueue(worker.AIJob{
        Type:      "face_matching",
        MissingID: h.ID, // neste contexto, é o homelessID
        PhotoURL:  h.PhotoURL,
    })

    return h, nil
}
```

#### Tarefa 5.4 — Repositório Firestore para Homeless

Implementar a interface. Mesmos patterns do MissingRepository.

#### Tarefa 5.5 — Handlers REST de Homeless

- `POST /api/v1/homeless` — criar
- `GET /api/v1/homeless` — listar todos
- `GET /api/v1/homeless/{id}` — buscar por ID
- `GET /api/v1/homeless/stats` — estatísticas (total, por gênero)

#### Tarefa 5.6 — Atualizar Dashboard stats

Adicionar dados de homeless no endpoint `GET /api/v1/missing/stats` (ou criar endpoint unificado `GET /api/v1/stats`).

#### Tarefa 5.7 — Refatoração

- Extrair `httputil` (JSON, Error, Decode) se ainda não feito
- Extrair mapeamento de erros centralizado
- Avaliar extração de generics para conversão Firestore
- Revisar e limpar código das fases anteriores

#### Tarefa 5.8 — Testes

- Entity: Age, Validate, GenerateSlug
- Service: Create com mock de Notifier e Repository
- Handler: endpoints com httptest

### Frontend (React)

#### Tarefa 5.9 — Página de cadastro de homeless

- Formulário: foto (obrigatória), nome, apelido, nascimento
- Selects: pele, cabelo, olhos, gênero
- Mapa para selecionar localização (reutilizar MapPicker)
- Validação de data (não pode ser futura)
- Máscara de data (DD/MM/YYYY)

#### Tarefa 5.10 — Página de listagem de homeless

- Grid de cards
- Cada card: foto, nome, idade, gênero, olhos, cabelo
- Mapa individual por card (reutilizar MapView)
- Sem paginação (lista tudo)

#### Tarefa 5.11 — Atualizar Dashboard

- Adicionar card "Total de Moradores de Rua"
- Adicionar gráfico de gênero para homeless (se relevante)

#### Tarefa 5.12 — Link "Quero ser encontrado" no menu lateral

- Adicionar link no sidebar para `/homeless/create`
- Página de escolha "Registrar familiar" vs "Quero ser encontrado" (similar ao `start_user.html` do legado)

---

## Decisões Específicas desta Fase

### Homeless SEM autenticação obrigatória no cadastro?

No projeto legado, o cadastro de homeless **não exige login**. Faz sentido porque quem cadastra é um voluntário/ONG que encontrou a pessoa na rua — não necessariamente tem conta na plataforma.

**Decisão**: manter o cadastro de homeless **sem autenticação obrigatória**. Mas adicionar proteção contra spam:
- Rate limiting por IP (máximo 5 cadastros por hora)
- Campo honeypot ou reCAPTCHA
- Validação de foto obrigatória (dificulta bots)

### Por que Homeless não tem status (disappeared/found)?

No domínio Missing, a pessoa pode ser "desaparecida" ou "encontrada". No Homeless, não existe esse conceito — a pessoa está registrada como "querendo ser encontrada". Se a família a encontrar, o registro pode ser removido ou marcado como "reunited" no futuro.

Por enquanto, mantemos simples: só existe ou não existe. Se surgir a necessidade de um status, adicionamos.

### Organização dos value objects compartilhados

Os value objects (Gender, EyeColor, etc.) são usados por Missing e Homeless. Duas opções:

**Opção A**: package `domain/shared/`
```
domain/
  shared/
    gender.go
    eye_color.go
    hair_color.go
    skin_color.go
```

**Opção B**: manter no `missing/` e importar no `homeless/`
```
domain/
  missing/
    value_objects.go  ← Gender, EyeColor, etc.
  homeless/
    entity.go         ← importa missing.Gender
```

**Decisão**: Opção A (`domain/shared/`). Motivo: value objects de características físicas não "pertencem" ao domínio Missing — são conceitos do domínio da plataforma como um todo. Se amanhã criarmos um terceiro tipo de pessoa, os mesmos value objects são usados.

---

## Entregáveis da Fase 5

- [ ] Entity Homeless com comportamento
- [ ] Interface e implementação HomelessRepository
- [ ] HomelessService com notificação Telegram
- [ ] Handlers REST para Homeless
- [ ] Refatoração: httputil, error mapping, generics (se aplicável)
- [ ] Testes unitários e de integração
- [ ] React: Cadastro de homeless com mapa e upload
- [ ] React: Listagem de homeless com cards e mapas
- [ ] React: Dashboard atualizado com dados de homeless
- [ ] React: Página de escolha (registrar familiar vs quero ser encontrado)

---

## Próxima Fase

→ [FASE_06_INTELIGENCIA.md](./FASE_06_INTELIGENCIA.md) — AI: Age Progression & Face Matching

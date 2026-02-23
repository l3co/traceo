# Fase 2 — Desaparecidos: CRUD Completo

> **Duração estimada**: 4 semanas
> **Pré-requisito**: Fase 1 concluída (auth + user funcionando)

---

## Objetivo

Funcionalidade **core** do sistema — cadastro, listagem, edição e remoção de pessoas desaparecidas, com geolocalização no mapa. Ao final desta fase:

- Usuário cadastra um desaparecido com todos os detalhes (características, foto, local no mapa)
- Listagem paginada com cards
- Modal de detalhes com mapa do local de desaparecimento
- Edição e remoção de registros

Esta é a fase onde o domínio mais rico do sistema toma forma. É aqui que você vai sentir o valor de ter entidades com **comportamento**, e não apenas sacos de dados.

---

## Conceitos Go que você vai aprender nesta fase

### 1. Composição vs Herança — a decisão de design mais importante do Go

Go **não tem herança**. Ponto final. Não existe `extends`, não existe `class`, não existe hierarquia de tipos.

#### Por que Go rejeitou herança?

Os criadores do Go observaram que herança em projetos grandes gera:
- **Hierarquias profundas** que ninguém entende (`Animal → Mammal → Dog → GoldenRetriever`)
- **Fragilidade** — mudar a classe base quebra todas as filhas
- **Acoplamento invisível** — a filha depende de detalhes internos da mãe

#### O que Go usa no lugar: Composição (Embedding)

Em vez de "Missing **é um** Entity", pensamos "Missing **tem** campos de timestamp":

```go
// Campos comuns que vários domínios usam
type Timestamps struct {
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Missing CONTÉM Timestamps (embedding)
type Missing struct {
    ID                   string
    UserID               string
    Name                 string
    Slug                 string
    Status               Status
    // ... outros campos
    Timestamps           // ← embedded (sem nome de campo)
}
```

Com embedding, os campos de `Timestamps` ficam acessíveis diretamente:

```go
m := Missing{Name: "João"}
m.CreatedAt = time.Now() // acessa direto, como se fosse campo do Missing
m.UpdatedAt = time.Now()
```

#### Quando usar embedding vs campo nomeado?

```go
// Embedding (promoção de campos) — "Missing TEM timestamps"
type Missing struct {
    Timestamps  // campos promovidos: m.CreatedAt funciona
}

// Campo nomeado — "Missing TEM uma localização"
type Missing struct {
    Location GeoPoint  // acesso: m.Location.Lat
}
```

Regra prática:
- **Embedding** quando os campos fazem parte da identidade (timestamps, audit)
- **Campo nomeado** quando é uma composição explícita (localização é um conceito separado)

#### Comparação com Python

No projeto legado, `disappeared` é um dicionário Python sem estrutura. Qualquer campo pode ser adicionado ou removido a qualquer momento. Erros aparecem em runtime.

No Go, a struct define **exatamente** quais campos existem. Se você tentar acessar um campo que não existe, o compilador te avisa na hora. Segurança em tempo de compilação.

---

### 2. Value Objects — tipos que representam conceitos

No projeto legado, gênero é uma string: `"male"` ou `"female"`. Nada impede alguém de passar `"banana"`. O erro só aparece (se aparecer) lá na frente.

Em Go, podemos criar **tipos customizados** que restringem os valores válidos:

```go
// Tipo customizado baseado em string
type Gender string

const (
    GenderMale   Gender = "male"
    GenderFemale Gender = "female"
)

// Método de validação no próprio tipo
func (g Gender) IsValid() bool {
    switch g {
    case GenderMale, GenderFemale:
        return true
    }
    return false
}

// Método de exibição em PT-BR
func (g Gender) Label() string {
    switch g {
    case GenderMale:
        return "Masculino"
    case GenderFemale:
        return "Feminino"
    }
    return "Não informado"
}
```

Fazemos o mesmo para Eyes, Hair, Skin, Status:

```go
type EyeColor string

const (
    EyeGreen     EyeColor = "green"
    EyeBlue      EyeColor = "blue"
    EyeBrown     EyeColor = "brown"
    EyeBlack     EyeColor = "black"
    EyeDarkBrown EyeColor = "dark_brown"
)

type HairColor string

const (
    HairBlack   HairColor = "black"
    HairBrown   HairColor = "brown"
    HairRedhead HairColor = "redhead"
    HairBlond   HairColor = "blond"
)

type Status string

const (
    StatusDisappeared Status = "disappeared"
    StatusFound       Status = "found"
)
```

#### Por que isso importa?

1. **Type safety** — uma função que recebe `Gender` não aceita uma `string` qualquer
2. **Documentação viva** — olhando os `const`, você sabe todos os valores válidos
3. **Comportamento no tipo** — `Label()` substitui aquele `basic_parse.py` do projeto legado que traduzia valores

No projeto legado, havia um arquivo `parser/basic_parse.py` com funções como `translate_hair()`, `translate_eyes()`. No Go, cada tipo traduz a si mesmo. A lógica fica **onde pertence**.

---

### 3. Entidade rica vs modelo anêmico — a diferença que define qualidade

No projeto legado, o "desaparecido" é um dicionário. Toda a lógica está espalhada no service:

```python
# Legado: service faz TUDO
def create(self, user_id, args):
    args['user_id'] = user_id
    args['registration_date'] = datetime.now()
    args['was_child'] = was_child(args['birth_date'], args['date_of_disappearance'])
    args['slug'] = create_slug(args)
    return self.repository.save(args)
```

No Go, a entidade **sabe fazer coisas**:

```go
// domain/missing/entity.go

type Missing struct {
    ID                    string
    UserID                string
    Name                  string
    Nickname              string
    BirthDate             time.Time
    DateOfDisappearance   time.Time
    Height                string
    Clothes               string
    Gender                Gender
    Eyes                  EyeColor
    Hair                  HairColor
    Skin                  SkinColor
    PhotoURL              string
    Location              GeoPoint
    Status                Status
    EventReport           string
    TattooDescription     string
    ScarDescription       string
    WasChild              bool
    Slug                  string
    Timestamps
}

type GeoPoint struct {
    Lat float64
    Lng float64
}

// A entidade calcula se era criança — a regra fica no domínio
func (m *Missing) CalculateWasChild() {
    if m.BirthDate.IsZero() || m.DateOfDisappearance.IsZero() {
        m.WasChild = false
        return
    }
    age := m.DateOfDisappearance.Year() - m.BirthDate.Year()
    if m.DateOfDisappearance.YearDay() < m.BirthDate.YearDay() {
        age--
    }
    m.WasChild = age < 18
}

// A entidade calcula a idade atual
func (m *Missing) Age() int {
    if m.BirthDate.IsZero() {
        return 0
    }
    now := time.Now()
    age := now.Year() - m.BirthDate.Year()
    if now.YearDay() < m.BirthDate.YearDay() {
        age--
    }
    return age
}

// A entidade gera o próprio slug
func (m *Missing) GenerateSlug() {
    base := strings.ToLower(m.Name)
    base = strings.ReplaceAll(base, " ", "-")
    // remove acentos, caracteres especiais...
    m.Slug = base + "-" + m.ID[:8]
}

// A entidade tem tatuagem?
func (m *Missing) HasTattoo() bool {
    return m.TattooDescription != ""
}

// A entidade tem cicatriz?
func (m *Missing) HasScar() bool {
    return m.ScarDescription != ""
}

// Validação completa
func (m *Missing) Validate() error {
    if m.Name == "" {
        return errors.New("name is required")
    }
    if m.UserID == "" {
        return errors.New("user_id is required")
    }
    if !m.Gender.IsValid() {
        return fmt.Errorf("invalid gender: %s", m.Gender)
    }
    if !m.Eyes.IsValid() {
        return fmt.Errorf("invalid eye color: %s", m.Eyes)
    }
    // ... outras validações
    return nil
}
```

#### O service fica enxuto

```go
// domain/missing/service.go

func (s *Service) Create(ctx context.Context, input CreateInput) (*Missing, error) {
    m := &Missing{
        ID:                  generateID(),
        UserID:              input.UserID,
        Name:                input.Name,
        // ... mapeia os campos do input
    }

    // A entidade faz seus cálculos
    m.CalculateWasChild()
    m.GenerateSlug()
    m.CreatedAt = time.Now()
    m.UpdatedAt = time.Now()

    // A entidade se valida
    if err := m.Validate(); err != nil {
        return nil, fmt.Errorf("invalid missing person: %w", err)
    }

    // Persiste
    if err := s.repo.Create(ctx, m); err != nil {
        return nil, fmt.Errorf("creating missing person: %w", err)
    }

    return m, nil
}
```

Percebe a diferença? O service **orquestra**, mas quem sabe as regras é a **entidade**. Se amanhã a regra de "era criança" mudar (ex: considerar 16 anos em vez de 18), você altera **um método** no entity.go, não procura em 5 services diferentes.

---

### 4. Slices e iteração — arrays dinâmicos do Go

Em Python: `list`. Em Go: **slice**.

```go
// Criar um slice
names := []string{"João", "Maria", "Pedro"}

// Adicionar elemento
names = append(names, "Ana")

// Iterar
for i, name := range names {
    fmt.Printf("%d: %s\n", i, name)
}

// Filtrar (não tem filter built-in — você faz um loop)
var children []*Missing
for _, m := range allMissing {
    if m.WasChild {
        children = append(children, m)
    }
}
```

Go é propositalmente **menos mágico** que Python. Não tem list comprehension, não tem `map()`, `filter()`, `reduce()`. Tem `for` loops. Isso torna o código mais explícito e legível, especialmente em times grandes.

#### Slices no nosso projeto

A listagem de desaparecidos retorna um slice:

```go
type MissingRepository interface {
    FindAll(ctx context.Context, opts ListOptions) ([]*Missing, error)
    // []*Missing → slice de ponteiros para Missing
}

type ListOptions struct {
    Page     int
    PageSize int
    UserID   string // filtro opcional
}
```

---

### 5. Time handling — datas em Go

O package `time` do Go tem uma peculiaridade famosa: o formato de referência é uma data específica:

```go
// Go usa uma "data de referência" para definir formatos
// A data de referência é: Mon Jan 2 15:04:05 MST 2006
// Ou seja: 01/02 03:04:05PM '06 -0700

// Formato brasileiro DD/MM/YYYY
const DateFormat = "02/01/2006"

// Parsear string para time.Time
date, err := time.Parse(DateFormat, "25/12/1990")

// Formatar time.Time para string
str := date.Format(DateFormat) // "25/12/1990"
```

**Por que essa data estranha (Jan 2, 2006)?**

Porque cada componente tem um número único:
- `01` = mês
- `02` = dia
- `03` = hora (12h) / `15` = hora (24h)
- `04` = minuto
- `05` = segundo
- `2006` = ano

Então para lembrar: **1, 2, 3, 4, 5, 6**. É estranho na primeira vez, mas depois fica intuitivo.

No nosso projeto, o legado usa o formato `DD/MM/YYYY`. Vamos manter a compatibilidade no input/output da API, mas internamente armazenar como `time.Time`.

---

### 6. Paginação com cursor no Firestore

O projeto legado usa paginação offset-based (página 1, 2, 3...). Firestore funciona melhor com **cursor-based pagination**.

#### Offset vs Cursor — qual a diferença?

**Offset** (legado):
```
GET /disappeared?page=3&size=10
→ "pule os primeiros 20, me dê 10"
→ Problema: Firestore precisa ler e descartar os 20 primeiros. Caro e lento.
```

**Cursor** (Firestore-friendly):
```
GET /missing?after=abc123&size=10
→ "a partir do documento abc123, me dê 10"
→ Firestore vai direto para o documento, sem ler os anteriores.
```

Na prática:

```go
func (r *firestoreMissingRepo) FindAll(ctx context.Context, opts ListOptions) ([]*Missing, string, error) {
    query := r.client.Collection("missing").
        OrderBy("CreatedAt", firestore.Desc).
        Limit(opts.PageSize)

    // Se tem cursor, começa a partir dele
    if opts.After != "" {
        doc, _ := r.client.Collection("missing").Doc(opts.After).Get(ctx)
        query = query.StartAfter(doc)
    }

    docs, err := query.Documents(ctx).GetAll()
    // ...

    // Retorna o ID do último documento como "next cursor"
    var nextCursor string
    if len(docs) == opts.PageSize {
        nextCursor = docs[len(docs)-1].Ref.ID
    }

    return results, nextCursor, nil
}
```

O frontend recebe o cursor e usa na próxima requisição:
```
GET /api/v1/missing?after=xyz789&size=10
```

**Trade-off**: com cursor, você não pode ir direto para a "página 5". Precisa navegar sequencialmente. Para o nosso caso (scroll infinito ou "próximo/anterior"), é perfeito.

---

## Tarefas Detalhadas

### Backend

#### Tarefa 2.1 — Value Objects (Gender, EyeColor, HairColor, SkinColor, Status)

Criar `internal/domain/missing/value_objects.go`:
- Cada tipo com constantes e método `IsValid()`
- Método `Label()` para tradução PT-BR
- Método `UnmarshalJSON()` para desserialização segura

#### Tarefa 2.2 — Entity Missing com comportamento

Criar `internal/domain/missing/entity.go`:
- Struct com todos os campos
- Métodos: `CalculateWasChild()`, `Age()`, `GenerateSlug()`, `HasTattoo()`, `HasScar()`, `Validate()`
- Erros sentinela: `ErrMissingNotFound`, `ErrInvalidMissing`

#### Tarefa 2.3 — Interface MissingRepository

Criar `internal/domain/missing/repository.go`:
```go
type Repository interface {
    Create(ctx context.Context, m *Missing) error
    FindByID(ctx context.Context, id string) (*Missing, error)
    Update(ctx context.Context, m *Missing) error
    Delete(ctx context.Context, id string) error
    FindByUserID(ctx context.Context, userID string) ([]*Missing, error)
    FindAll(ctx context.Context, opts ListOptions) ([]*Missing, string, error)
    Count(ctx context.Context) (int64, error)
}
```

#### Tarefa 2.4 — MissingService

Criar `internal/domain/missing/service.go`:
- Create, FindByID, Update, Delete, FindByUserID, List (paginado), Count
- Verificar que o userID existe antes de criar

#### Tarefa 2.5 — Repositório Firestore para Missing

Criar `internal/infrastructure/firestore/missing_repository.go`:
- Implementar todos os métodos
- Paginação cursor-based
- Mapeamento de erros

#### Tarefa 2.6 — Handlers REST de Missing

Criar `internal/handler/missing_handler.go`:
- `POST /api/v1/missing`
- `GET /api/v1/missing` (com paginação)
- `GET /api/v1/missing/{id}`
- `PUT /api/v1/missing/{id}`
- `DELETE /api/v1/missing/{id}`
- `GET /api/v1/users/{id}/missing`

#### Tarefa 2.7 — Upload de foto do desaparecido

Reutilizar o serviço de upload da Fase 1 (Cloud Storage).

#### Tarefa 2.8 — Geração de slug

Criar `pkg/slug/slug.go`:
- Remover acentos, caracteres especiais
- Substituir espaços por hífens
- Adicionar sufixo único (primeiros 8 chars do ID)

#### Tarefa 2.9 — Testes unitários

- Entity: CalculateWasChild, Age, Validate, HasTattoo, HasScar
- Value Objects: IsValid, Label
- Service: Create (sucesso + validação), FindByID, Update, Delete
- Slug: geração com acentos, espaços, caracteres especiais

### Frontend (React)

#### Tarefa 2.10 — Componente de mapa reutilizável

Usando `@vis.gl/react-google-maps` (biblioteca oficial do Google Maps para React):

```tsx
// src/shared/components/MapPicker.tsx
// — Mapa interativo para selecionar localização (click to place marker)
// — Usado em: cadastro de desaparecido, cadastro de homeless, registro de avistamento

// src/shared/components/MapView.tsx
// — Mapa de visualização (read-only, mostra marker numa posição)
// — Usado em: card de detalhes, listagem de homeless

// src/shared/components/MapHeatmap.tsx
// — Mapa de calor com áreas de risco de desaparecimento
// — Usado em: dashboard, página de mapa geral

// src/shared/components/MapClusters.tsx
// — Markers agrupados por proximidade
// — Usado em: listagem geral no mapa
```

Por que Google Maps em vez de Mapbox: integração nativa GCP (mesmo projeto, mesmo billing), heatmaps nativos, geocoding superior no Brasil, e `@vis.gl/react-google-maps` é mantida pelo Google com TypeScript nativo.

#### Tarefa 2.11 — Componente de upload de imagem

```tsx
// src/shared/components/ImageUpload.tsx
// — Drag & drop ou click para selecionar
// — Preview da imagem
// — Mostra tamanho do arquivo
// — Comprime no client antes de enviar (opcional)
```

#### Tarefa 2.12 — Página de listagem de desaparecidos

- Grid de cards responsivo
- Cada card: foto, nome, data de desaparecimento, status, idade, características
- Botão "Veja mais" abre modal de detalhes
- Scroll infinito ou botão "Carregar mais" (cursor-based)

#### Tarefa 2.13 — Modal de detalhes completo

- Foto grande
- Todas as características
- Mapa com marker no local de desaparecimento
- Botão "Informar avistamento" (link para Fase 4)

#### Tarefa 2.14 — Formulário de cadastro de desaparecido

- Campos: nome, apelido, nascimento, data desaparecimento, altura, roupas
- Selects: gênero, olhos, cabelo, pele (com labels PT-BR)
- Textarea: BO, tatuagens, cicatrizes
- Upload de foto
- Mapa (Google Maps) para selecionar local de desaparecimento
- Validação de datas (não pode ser futura)
- Máscara para data (DD/MM/YYYY)

#### Tarefa 2.15 — Formulário de edição de desaparecido

- Pré-preenchido com dados existentes
- Mesmos campos do cadastro
- Preview da foto atual
- Mapa (Google Maps) com marker na posição atual

#### Tarefa 2.16 — Selects de características com tradução

```tsx
// src/features/missing/components/PhysicalTraitsSelect.tsx
const GENDER_OPTIONS = [
    { value: 'male', label: 'Masculino' },
    { value: 'female', label: 'Feminino' },
]

const EYE_OPTIONS = [
    { value: 'green', label: 'Verde' },
    { value: 'blue', label: 'Azul' },
    { value: 'brown', label: 'Castanho' },
    { value: 'black', label: 'Pretos' },
    { value: 'dark_brown', label: 'Castanho Escuro' },
]
// ...
```

---

## Decisões Específicas desta Fase

### Por que value objects como tipos e não enums?

Go não tem `enum` como Java/TypeScript. Tem constantes tipadas. Na prática, funciona da mesma forma, mas com uma diferença: Go **não impede** você de criar um valor inválido em compile time:

```go
var g Gender = "banana" // compila sem erro
```

Por isso adicionamos `IsValid()` e validamos na entidade. Em runtime, o `Validate()` da entidade pega isso. Não é perfeito, mas é o idiomático em Go.

**Alternativa considerada**: usar `int` como enum (iota). Mas strings são melhores para serialização JSON e legibilidade nos logs.

### Por que cursor-based e não offset-based?

Além da performance no Firestore, cursor-based resolve o problema de **dados inconsistentes durante paginação**:

Com offset: se alguém cadastra um novo desaparecido enquanto você está na página 2, ao ir para a página 3 você pode ver um registro duplicado (que "deslocou" da página 2 para a 3).

Com cursor: a consulta parte de um documento específico. Novos documentos não afetam a posição do cursor.

### GeoPoint: struct customizada vs tipo do Firestore

Criamos nosso próprio `GeoPoint` no domínio em vez de usar `*latlng.LatLng` do Firestore SDK. Motivo: o domínio não pode depender de infraestrutura. Na camada de infraestrutura, convertemos:

```go
// No repositório Firestore
firestoreGeo := &latlng.LatLng{
    Latitude:  m.Location.Lat,
    Longitude: m.Location.Lng,
}
```

---

## Entregáveis da Fase 2

- [ ] Value Objects: Gender, EyeColor, HairColor, SkinColor, Status
- [ ] Entity Missing com comportamento completo
- [ ] Interface MissingRepository
- [ ] MissingService com todos os casos de uso
- [ ] Repositório Firestore com paginação cursor-based
- [ ] Handlers REST completos para Missing
- [ ] Geração de slug (package pkg/slug)
- [ ] Testes unitários (~80% cobertura)
- [ ] React: Listagem paginada com cards
- [ ] React: Modal de detalhes com mapa
- [ ] React: Formulário de cadastro com mapa e upload
- [ ] React: Formulário de edição
- [ ] React: Componentes reutilizáveis (MapPicker, MapView, ImageUpload)

---

## Próxima Fase

→ [FASE_03_BUSCA_DASHBOARD.md](./FASE_03_BUSCA_DASHBOARD.md) — Busca Textual + Dashboard de Estatísticas

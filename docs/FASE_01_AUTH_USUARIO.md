# Fase 1 — Autenticação & Usuário

> **Duração estimada**: 3 semanas
> **Pré-requisito**: Fase 0 concluída (health check rodando)

---

## Objetivo

Sistema de autenticação **completo** e CRUD de usuário funcionando end-to-end (Go ↔ React). Ao final desta fase:

- Usuário se registra, faz login, edita perfil, altera senha
- JWT via Firebase Auth protege as rotas
- Upload de avatar funciona (Cloud Storage)
- React tem layout base com sidebar + área de conteúdo

Esta é a fase **mais densa em conceitos Go**, porque precisamos construir a fundação inteira: interfaces, error handling, middleware, testes.

---

## Conceitos Go que você vai aprender nesta fase

### 1. Interfaces — o contrato sem herança

Em Java/C#, você define uma interface e uma classe **explicitamente** implementa:

```java
class UserRepository implements IUserRepository { ... }
```

Em Go, **a implementação é implícita**. Se uma struct tem os métodos que a interface exige, ela implementa a interface. Ponto.

```go
// No domínio: definimos O QUE precisamos
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id string) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}
```

```go
// Na infraestrutura: implementamos COMO
type firestoreUserRepo struct {
    client *firestore.Client
}

func (r *firestoreUserRepo) Create(ctx context.Context, user *User) error {
    _, err := r.client.Collection("users").Doc(user.ID).Set(ctx, user)
    return err
}

func (r *firestoreUserRepo) FindByID(ctx context.Context, id string) (*User, error) {
    // ...
}
// ... implementa todos os métodos da interface
```

Não tem `implements`, não tem `extends`, não tem anotação. Se `firestoreUserRepo` tem todos os métodos que `UserRepository` pede, **automaticamente satisfaz a interface**.

#### Por que isso é poderoso?

1. **Desacoplamento real** — o domínio define a interface. A infraestrutura implementa. O domínio **nunca importa** a infraestrutura.

2. **Testes fáceis** — nos testes, criamos um mock que satisfaz a mesma interface:
```go
type mockUserRepo struct {
    users map[string]*User
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*User, error) {
    u, ok := m.users[id]
    if !ok {
        return nil, ErrUserNotFound
    }
    return u, nil
}
```

3. **Troca de implementação** — se amanhã quisermos trocar Firestore por PostgreSQL, criamos um `postgresUserRepo` que satisfaz a mesma interface. O service não muda uma linha.

#### Onde a interface fica no nosso projeto?

A interface fica **no pacote do domínio**, não na infraestrutura. Isso é inversão de dependência:

```
domain/user/repository.go       ← interface UserRepository (contrato)
infrastructure/firestore/user.go ← firestoreUserRepo (implementação)
```

O domínio dita as regras. A infraestrutura obedece.

---

### 2. Context (`context.Context`) — o passaporte de toda requisição

Em Go, quase toda função que faz I/O recebe um `context.Context` como primeiro parâmetro. Isso é **obrigatório** — se você não passar context, o linter reclama.

```go
func (s *UserService) FindByID(ctx context.Context, id string) (*User, error) {
    return s.repo.FindByID(ctx, id)
}
```

#### O que é o context e para que serve?

O context carrega:
- **Deadline/Timeout** — "essa operação tem que terminar em 5 segundos"
- **Cancellation** — "o cliente desconectou, cancele tudo"
- **Values** — metadados da requisição (user ID autenticado, request ID para tracing)

Exemplo prático: quando o Go recebe uma requisição HTTP, ele cria um context. Se o cliente fechar o browser antes da resposta, o context é cancelado automaticamente. Todas as operações downstream (query no Firestore, chamada ao SendGrid) são canceladas juntas.

```go
// No handler, o context vem do request
func (h *UserHandler) FindByID(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context() // context da requisição HTTP
    id := chi.URLParam(r, "id")

    user, err := h.service.FindByID(ctx, id) // passa o context para baixo
    if err != nil {
        // ...
    }
}
```

#### Regra de ouro

**Sempre passe context como primeiro parâmetro.** Nunca armazene context em struct. Nunca crie context global.

```go
// ✅ Correto
func DoSomething(ctx context.Context, data string) error

// ❌ Errado — context dentro de struct
type Service struct {
    ctx context.Context
}
```

---

### 3. Error handling avançado — erros custom e wrapping

Na Fase 0, vimos que Go retorna erros como valores. Agora vamos aprofundar.

#### Erros sentinela (sentinel errors)

São erros pré-definidos que representam condições específicas do domínio:

```go
package user

import "errors"

var (
    ErrUserNotFound      = errors.New("user not found")
    ErrEmailAlreadyExists = errors.New("email already exists")
    ErrInvalidPassword   = errors.New("invalid password")
)
```

No handler, você checa qual erro aconteceu:

```go
user, err := h.service.FindByID(ctx, id)
if err != nil {
    if errors.Is(err, user.ErrUserNotFound) {
        http.Error(w, "User not found", http.StatusNotFound) // 404
        return
    }
    http.Error(w, "Internal error", http.StatusInternalServerError) // 500
    return
}
```

#### Error wrapping — adicionando contexto

Quando um erro acontece numa camada baixa (Firestore), queremos adicionar contexto sem perder o erro original:

```go
func (r *firestoreUserRepo) FindByID(ctx context.Context, id string) (*User, error) {
    doc, err := r.client.Collection("users").Doc(id).Get(ctx)
    if err != nil {
        if status.Code(err) == codes.NotFound {
            return nil, user.ErrUserNotFound // erro de domínio
        }
        return nil, fmt.Errorf("firestore: finding user %s: %w", id, err) // wrap o erro original
    }
    // ...
}
```

O `%w` faz **wrap** — o erro original fica dentro do novo erro. `errors.Is()` consegue "desembrulhar" e achar o erro original.

#### Comparação com Python

```python
# Python — exceção voa até alguém capturar
def find_user(id):
    user = db.users.find_one({"_id": id})
    if not user:
        raise UserNotFoundError()  # ← voa para cima
    return user

# Em algum lugar lá em cima...
try:
    user = find_user(id)
except UserNotFoundError:
    return 404
except Exception:
    return 500
```

```go
// Go — erro é retornado explicitamente em cada nível
user, err := service.FindByID(ctx, id)
if err != nil {
    // Eu sei EXATAMENTE que esse erro pode acontecer aqui
    // Não tem surpresa de uma exceção vindo de 10 camadas abaixo
}
```

A vantagem do Go: **zero surpresas**. Cada ponto de erro é visível.

---

### 4. Ponteiros — quando usar `*User` vs `User`

Em Go, structs são passadas **por valor** (cópia). Se você quer modificar o original ou evitar cópia de structs grandes, usa ponteiro.

```go
// Por valor — cria uma cópia
func updateName(u User, name string) {
    u.Name = name // modifica a CÓPIA, o original não muda
}

// Por ponteiro — modifica o original
func updateName(u *User, name string) {
    u.Name = name // modifica o ORIGINAL
}
```

#### Regras práticas

1. **Retornos de repository/service: use ponteiro** (`*User`)
   - Porque o User pode não existir (`nil` = não encontrado)
   - Porque evita cópia de struct

2. **Parâmetros que você não modifica: valor está OK** (`User`)
   - Mas se a struct for grande (muitos campos), ponteiro é mais eficiente

3. **Receivers de método: ponteiro se modifica o estado**
   ```go
   func (u *User) SetName(name string) { u.Name = name }  // ← ponteiro, modifica
   func (u User) FullName() string { return u.Name }       // ← valor, só lê
   ```

No nosso projeto, a convenção será:
- **Entity methods** que modificam estado → receiver de ponteiro (`*User`)
- **Repository/Service** retornam ponteiros (`*User, error`)
- **Handler** recebe ponteiros do service

---

### 5. Structs como DTOs — separação request/response/entity

Um erro comum em Go é usar a mesma struct para tudo: banco de dados, request HTTP e response HTTP. Isso acopla as camadas.

Nós vamos separar:

```go
// domain/user/entity.go — a entidade de domínio
type User struct {
    ID           string
    Name         string
    Email        string
    Phone        string
    CellPhone    string
    AvatarURL    string
    AcceptedTerms bool
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

```go
// handler/dto.go — o que o cliente envia
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,max=150"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=6,max=10"`
    Phone    string `json:"phone,omitempty"`
    CellPhone string `json:"cell_phone,omitempty"`
}
```

```go
// handler/dto.go — o que o cliente recebe
type UserResponse struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Email     string `json:"email"`
    AvatarURL string `json:"avatar_url,omitempty"`
    CreatedAt string `json:"created_at"`
}
```

#### Por que separar?

1. **Segurança** — o `CreateUserRequest` tem `Password`, mas o `UserResponse` **não** retorna a senha
2. **Evolução** — mudar o JSON da API não obriga a mudar a entidade de domínio
3. **Validação** — tags `validate` ficam no DTO, não na entidade
4. **Serialização** — tags `json` ficam no DTO. A entidade fica limpa.

#### Tags de struct — o que são aqueles \`json:"name"\`?

Em Go, struct fields podem ter **tags** — metadados entre crases:

```go
type UserResponse struct {
    ID   string `json:"id"`         // serializa como "id" no JSON
    Name string `json:"name"`       // serializa como "name"
    Age  int    `json:"age,omitempty"` // omite se for zero
}
```

Tags são usadas por bibliotecas para:
- `json:"..."` → serialização JSON (stdlib)
- `validate:"..."` → validação (go-playground/validator)
- `firestore:"..."` → mapeamento Firestore

Não confunda com annotations do Java — tags são strings simples, sem mágica de compilador.

---

### 6. Middleware — funções que envolvem handlers

Middleware em Go é uma **função que recebe um handler e retorna um handler**:

```go
func AuthMiddleware(authService AuthVerifier) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")
            if token == "" {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return // para aqui, não chama o próximo
            }

            userID, err := authService.VerifyToken(r.Context(), token)
            if err != nil {
                http.Error(w, "invalid token", http.StatusUnauthorized)
                return
            }

            // Coloca o userID no context para os handlers usarem
            ctx := context.WithValue(r.Context(), "userID", userID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

Parece complexo? Vamos destrinchar:

1. `AuthMiddleware` recebe o serviço de auth (para verificar tokens)
2. Retorna uma função que recebe o `next` handler (o que vem depois)
3. Dentro, verifica o token. Se válido, chama `next`. Se inválido, para.

No Chi, aplicar middleware:

```go
r.Route("/api/v1", func(r chi.Router) {
    r.Get("/health", healthHandler.Check)  // SEM auth

    r.Group(func(r chi.Router) {
        r.Use(AuthMiddleware(authService))  // COM auth
        r.Post("/users", userHandler.Create)
        r.Get("/users/{id}", userHandler.FindByID)
    })
})
```

#### Comparação com Flask

```python
# Flask — decorator
@app.route('/users', methods=['POST'])
@auth.login_required
def create_user():
    ...
```

```go
// Go/Chi — middleware na rota
r.With(AuthMiddleware(authService)).Post("/users", userHandler.Create)
```

A ideia é a mesma. A sintaxe é diferente.

---

### 7. Testes em Go — `testing` package

Go tem um framework de testes **built-in**. Não precisa instalar pytest, jest ou junit.

```go
// user_service_test.go
package user_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/seu-usuario/traceo-api/internal/domain/user"
)

func TestCreateUser_Success(t *testing.T) {
    // Arrange
    repo := &mockUserRepo{users: make(map[string]*user.User)}
    service := user.NewService(repo)

    // Act
    created, err := service.Create(context.Background(), &user.CreateInput{
        Name:  "João Silva",
        Email: "joao@email.com",
    })

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "João Silva", created.Name)
    assert.NotEmpty(t, created.ID)
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
    repo := &mockUserRepo{
        users: map[string]*user.User{
            "1": {ID: "1", Email: "joao@email.com"},
        },
    }
    service := user.NewService(repo)

    _, err := service.Create(context.Background(), &user.CreateInput{
        Name:  "Outro João",
        Email: "joao@email.com",
    })

    assert.ErrorIs(t, err, user.ErrEmailAlreadyExists)
}
```

Para rodar:
```bash
go test ./...                    # roda todos os testes
go test ./internal/domain/user/  # roda só os testes do package user
go test -v ./...                 # verbose (mostra nome de cada teste)
go test -cover ./...             # mostra cobertura
```

#### Convenções de teste em Go

- Arquivo de teste: `xxx_test.go` (mesmo diretório do código)
- Função de teste: `TestNomeDoTeste(t *testing.T)`
- Package de teste: pode ser `user` (white-box) ou `user_test` (black-box)

Nós vamos usar **testify** (`github.com/stretchr/testify`) para assertions mais legíveis. É a lib de teste mais usada no ecossistema Go.

---

## Tarefas Detalhadas

### Backend

#### Tarefa 1.1 — Entity User com validações

Criar `internal/domain/user/entity.go`:
- Struct `User` com todos os campos
- Método `Validate()` que verifica regras de negócio
- Erros sentinela: `ErrUserNotFound`, `ErrEmailAlreadyExists`, `ErrInvalidPassword`

**Decisão**: a entidade User **não** armazena senha. Firebase Auth gerencia isso. A entidade só tem dados de perfil.

#### Tarefa 1.2 — Interface UserRepository

Criar `internal/domain/user/repository.go`:
- Interface com: Create, FindByID, FindByEmail, Update, Delete, FindAll

#### Tarefa 1.3 — UserService com casos de uso

Criar `internal/domain/user/service.go`:
- `NewService(repo UserRepository, auth AuthService) *Service`
- Métodos: Create, FindByID, Update, Delete, ChangePassword

**Decisão sobre o Service**: o service recebe a interface do repository via construtor (dependency injection manual — Go não usa frameworks de DI como Spring).

```go
type Service struct {
    repo UserRepository
    auth AuthService
}

func NewService(repo UserRepository, auth AuthService) *Service {
    return &Service{repo: repo, auth: auth}
}
```

**Por que DI manual e não um framework?**

Em Java/C#, é comum usar Spring/ASP.NET para injeção de dependência automática. Em Go, a comunidade **rejeita** frameworks de DI por uma razão filosófica:

> Se as dependências não são óbvias lendo o código, o código está complexo demais.

DI manual em Go é: você cria as instâncias no `main.go` e passa para quem precisa. É explícito, rastreável, e funciona perfeitamente para projetos do nosso tamanho.

```go
// cmd/server/main.go
func main() {
    // Cria as dependências na ordem
    firestoreClient := firestore.NewClient(ctx)
    userRepo := firestoreRepo.NewUserRepository(firestoreClient)
    authService := firebaseAuth.NewService(authClient)
    userService := user.NewService(userRepo, authService)
    userHandler := handler.NewUserHandler(userService)

    // Monta o router
    r := chi.NewRouter()
    r.Post("/api/v1/users", userHandler.Create)
    // ...
}
```

Você lê o `main.go` e sabe **exatamente** quem depende de quem. Sem mágica, sem annotations, sem scanning de packages.

#### Tarefa 1.4 — Implementação Firestore do UserRepository

Criar `internal/infrastructure/firestore/user_repository.go`:
- Implementar cada método da interface
- Mapear erros do Firestore para erros de domínio (ex: NotFound → ErrUserNotFound)

#### Tarefa 1.5 — Firebase Auth integration

Criar `internal/infrastructure/auth/firebase_auth.go`:
- `CreateUser(email, password)` → cria no Firebase Auth
- `VerifyToken(token)` → verifica JWT e retorna userID
- `ChangePassword(userID, newPassword)`
- `SendPasswordResetEmail(email)`

#### Tarefa 1.6 — Middleware de autenticação

Criar `internal/handler/middleware/auth.go`:
- Extrair token do header `Authorization: Bearer <token>`
- Verificar via Firebase Auth
- Injetar userID no context

#### Tarefa 1.7 — Handlers REST de User

Criar `internal/handler/user_handler.go`:
- `POST /api/v1/users` → Create
- `GET /api/v1/users/{id}` → FindByID
- `PUT /api/v1/users/{id}` → Update
- `DELETE /api/v1/users/{id}` → Delete
- `PATCH /api/v1/users/{id}/password` → ChangePassword

Cada handler:
1. Decodifica o JSON do request
2. Valida os campos
3. Chama o service
4. Retorna o JSON da response

#### Tarefa 1.8 — Handler de Auth

Criar `internal/handler/auth_handler.go`:
- `POST /api/v1/auth/login` → Login (delega ao Firebase Auth)
- `POST /api/v1/auth/forgot-password` → ForgotPassword

#### Tarefa 1.9 — Upload de avatar (Cloud Storage)

Criar `internal/infrastructure/storage/gcs_client.go`:
- Receber arquivo multipart
- Upload para Cloud Storage
- Retornar URL pública

#### Tarefa 1.10 — Anotações Swagger em todos os handlers

Todo handler criado nesta fase deve ter anotações swaggo:
- `@Summary`, `@Description`, `@Tags`, `@Accept`, `@Produce`
- `@Param` para path params, query params e body
- `@Success` e `@Failure` com os DTOs de response
- `@Security BearerAuth` nas rotas protegidas
- Rodar `swag init` e verificar que a spec está atualizada

Exemplo para o Create User:

```go
// @Summary      Criar usuário
// @Description  Registra um novo usuário na plataforma
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body  CreateUserRequest  true  "Dados do usuário"
// @Success      201  {object}  UserResponse
// @Failure      400  {object}  ErrorResponse  "Dados inválidos"
// @Failure      409  {object}  ErrorResponse  "Email já existe"
// @Router       /api/v1/users [post]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
```

A partir desta fase, **nenhum handler é considerado pronto sem anotação Swagger**.

#### Tarefa 1.11 — Middlewares de segurança (Rate Limiting + Headers + Body Limit)

Rate limiting é **fundação**, não polimento. Deve existir desde o primeiro endpoint.

Criar `internal/handler/middleware/rate_limit.go`:
- Rate limiter global: 200 req/min por IP
- Rate limiter específico: 5 req/min no login, 3 req/min no cadastro
- Headers de resposta: `Retry-After`, `X-RateLimit-Remaining`
- Retornar 429 Too Many Requests quando exceder
- Limpeza periódica de IPs inativos (evitar memory leak)

Criar `internal/handler/middleware/security.go`:
- Security Headers: `X-Content-Type-Options`, `X-Frame-Options`, `HSTS`, `Referrer-Policy`, `Permissions-Policy`
- Body size limit: 1 MB para JSON, 10 MB para uploads
- Request timeout: 30 segundos

Criar `internal/handler/middleware/cors.go`:
- CORS configurado por ambiente (localhost em dev, domínio real em prod)
- Nunca `AllowedOrigins: ["*"]` em produção

> Detalhes completos: [SECURITY.md](./SECURITY.md) — seções 1.1 a 1.5

#### Tarefa 1.12 — Input validation e sanitização

- Integrar `go-playground/validator` para validação de DTOs
- Integrar `bluemonday` (StrictPolicy) para sanitização de texto livre
- Helper `httputil.DecodeAndValidate(r, &dto)` que faz decode JSON + validate + sanitize em um passo
- Testar: campos obrigatórios, formatos inválidos, HTML em texto livre

#### Tarefa 1.13 — Testes unitários do domínio

- Testar criação de User (sucesso + validações)
- Testar email duplicado
- Testar FindByID (encontrado + não encontrado)
- Mock do repository

#### Tarefa 1.14 — Testes de integração dos handlers

- Testar cada endpoint com `httptest` (stdlib do Go para testes HTTP)
- Verificar status codes, body, headers
- Testar rate limiting (burst requests → esperar 429)
- Testar auth middleware (sem token → 401, token inválido → 401)

### Frontend (React)

#### Tarefa 1.15 — Setup React Router

- Configurar rotas: `/login`, `/register`, `/profile`, `/password`
- Rotas protegidas (redirect para login se não autenticado)

#### Tarefa 1.16 — Firebase Auth no React

```tsx
// src/shared/lib/firebase.ts
import { initializeApp } from 'firebase/app'
import { getAuth } from 'firebase/auth'

const app = initializeApp({ /* config */ })
export const auth = getAuth(app)
```

- Hook `useAuth()` com context
- `signInWithEmailAndPassword`
- `createUserWithEmailAndPassword`
- `signOut`
- Interceptor para enviar JWT em toda requisição para a API

#### Tarefa 1.17 — Layout base

- Sidebar com navegação (mesmos itens do legado)
- Top bar com busca e menu do usuário
- Content area responsiva
- Usando shadcn/ui: Sidebar, Button, Input, Card, Avatar, DropdownMenu

#### Tarefa 1.18 — Página de Login

- Formulário: email + senha
- Link para "Recuperar Senha"
- Botão "Entrar"
- Feedback de erro (credenciais inválidas)

#### Tarefa 1.19 — Página de Cadastro

- Formulário: foto, nome, email, senha, telefone, celular
- Upload de foto com preview
- Checkbox de termos de uso
- Validação no cliente

#### Tarefa 1.20 — Página de Edição de Perfil

- Formulário pré-preenchido com dados do usuário
- Upload de nova foto com preview da atual
- Botão "Alterar"

#### Tarefa 1.21 — Página de Alteração de Senha

- Campos: nova senha + confirmação
- Validação de match entre senhas
- Feedback de sucesso/erro

#### Tarefa 1.22 — Página de Recuperação de Senha

- Campo: email
- Botão "Enviar"
- Firebase Auth envia o email automaticamente

#### Tarefa 1.23 — API client com interceptors

```tsx
// src/shared/lib/api.ts
const api = axios.create({
    baseURL: import.meta.env.VITE_API_URL
})

api.interceptors.request.use(async (config) => {
    const token = await auth.currentUser?.getIdToken()
    if (token) {
        config.headers.Authorization = `Bearer ${token}`
    }
    return config
})
```

---

## Decisões Específicas desta Fase

### Por que a entidade User não tem senha?

No projeto legado, a senha (MD5) ficava no documento do MongoDB junto com o restante dos dados do usuário. Isso é um anti-pattern:

1. Qualquer query que retorna um usuário também retorna a hash da senha
2. Se o banco vazar, as senhas vão junto
3. A lógica de crypto fica misturada com a lógica de domínio

Com Firebase Auth, as credenciais ficam em um **sistema separado e gerenciado**. A entidade User só tem dados de perfil. A senha nunca entra no Firestore.

### Por que `NewService()` e não `service.New()`?

Em Go, quando um package tem uma "coisa principal", você pode usar `New()`:

```go
// Se o package user tem só um service
service := user.New(repo)  // user.New → cria a coisa principal do package

// Se o package user tem múltiplas coisas (Service, Validator, etc)
service := user.NewService(repo)     // explícito
validator := user.NewValidator()     // explícito
```

No nosso caso, como o package `user` tem entity + repository interface + service, usamos `NewService()` para clareza.

### JSON naming convention: snake_case

A API legada usa `snake_case` no JSON (`cell_phone`, `birth_date`). Vamos manter isso por consistência e porque:
- É a convenção mais comum em APIs REST
- React/TypeScript trabalha bem com snake_case via transformação automática
- Firebase/Firestore usa camelCase internamente, mas serializamos para snake_case na API

---

## Entregáveis da Fase 1

- [ ] Entity User com validações
- [ ] Interface UserRepository definida no domínio
- [ ] Service com Create, FindByID, Update, Delete, ChangePassword
- [ ] Repositório Firestore implementando a interface
- [ ] Firebase Auth integrado (criar user, verificar token, reset password)
- [ ] Middleware de autenticação JWT
- [ ] **Middleware de rate limiting** (global 200/min + por endpoint)
- [ ] **Middleware de security headers** (HSTS, X-Frame-Options, etc.)
- [ ] **CORS configurado** por ambiente (nunca wildcard)
- [ ] **Body size limit** (1 MB JSON, 10 MB upload)
- [ ] **Input validation** (go-playground/validator)
- [ ] **Sanitização** de texto livre (bluemonday)
- [ ] Handlers REST para User e Auth
- [ ] Upload de avatar (Cloud Storage)
- [ ] Testes unitários do domínio (~80% cobertura)
- [ ] Testes de integração dos handlers (incluindo rate limit e auth)
- [ ] React: Login, Cadastro, Perfil, Senha, Recuperação
- [ ] React: Layout base (sidebar + top bar)
- [ ] React: API client com JWT automático

---

## Próxima Fase

→ [FASE_02_DESAPARECIDOS.md](./FASE_02_DESAPARECIDOS.md) — CRUD de Desaparecidos + Mapas

# Traceo — Roadmap de Reconstrução

> **De**: Flask + MongoDB → **Para**: Go + React + Firestore + Cloud Run

---

## O que é este documento?

Este é o **mapa mestre** da reconstrução do projeto. Ele contém:

1. O diagnóstico do sistema legado
2. A arquitetura proposta e as **razões profundas** de cada decisão
3. O índice das fases (cada uma é um arquivo separado)

Os detalhes de implementação, tarefas, e conceitos de Go estão nos arquivos de cada fase.

---

## 1. Diagnóstico do Sistema Atual

### 1.1 O que a plataforma faz

O **Traceo** é uma plataforma social para localização de pessoas desaparecidas no Brasil. Tem quatro funcionalidades principais:

- **Cadastro de desaparecidos** — um familiar se registra, cadastra dados da pessoa desaparecida (foto, características, local) e pode ser notificado se alguém avistar a pessoa.
- **Avistamento** — qualquer pessoa pode informar onde viu alguém desaparecido, marcando no mapa e adicionando uma observação.
- **"Quero ser encontrado"** — ONGs e voluntários podem cadastrar moradores de rua que querem ser encontrados pela família.
- **Dashboard** — estatísticas gerais, gráficos por gênero/ano e mapa com regiões de maior incidência.

### 1.2 Arquitetura Atual

Dois repositórios separados que se comunicam via HTTP com Basic Auth:

```
┌─────────────────────┐     HTTP/Basic Auth     ┌─────────────────────────┐
│   Portal Web        │ ──────────────────────→  │   API REST              │
│   (Flask + Jinja2)  │                          │   (Flask-RESTful)       │
└─────────────────────┘                          └────────┬────────────────┘
        │                                                 │
        ▼                                        ┌────────┴────────┐
   Cloudinary                               MongoDB          Redis/Celery
   (imagens)                              (4 coleções)     (filas async)
                                                               │
                                                      ┌────────┴────────┐
                                                  SendGrid          Telegram
```

### 1.3 Stack Legada

| Componente | Tecnologia | Versão |
|---|---|---|
| Backend | Python + Flask-RESTful | 3.6 / 1.0 |
| Frontend | Flask + Jinja2 + Material Dashboard | 1.0 |
| Banco | MongoDB | — |
| Fila async | Redis + Celery | — |
| Imagens | Cloudinary | — |
| Email | SendGrid | — |
| Notificação | Telegram Bot API | — |
| Mapas | Mapbox GL JS | 0.47 |
| Analytics | Google Analytics + Hotjar | — |

### 1.4 Problemas Identificados

**Segurança (Crítico)**
- Senhas armazenadas com **MD5 sem salt** — qualquer tabela rainbow quebra isso em segundos
- Secret keys e API keys **hardcoded** no código-fonte
- Basic Auth entre frontend e API — credenciais trafegam em texto base64 em toda requisição
- Tokens do Mapbox e Telegram expostos em código público

**Arquitetura**
- O frontend Flask funciona como um "BFF" (Backend for Frontend) que só repassa chamadas para a API — camada desnecessária que duplica toda a lógica de chamada
- Modelo **anêmico**: entidades são dicionários Python sem comportamento, regras espalhadas entre parser e service
- Sem testes robustos
- Código duplicado entre parsers dos dois repositórios

---

## 2. Domínios de Negócio (Bounded Contexts)

Antes de falar de tecnologia, precisamos entender **o que o sistema faz** em termos de domínio. Isso é DDD (Domain-Driven Design) — primeiro entendemos o negócio, depois escolhemos a tecnologia.

### 2.1 Usuário (User)

O **familiar** que registra o desaparecimento. É quem recebe notificações de avistamento.

**Dados**: nome, email, senha, telefone, celular, avatar (foto), aceite de termos
**Ações**: cadastrar-se, fazer login, editar perfil, alterar senha, recuperar senha

### 2.2 Desaparecido (Missing)

A **pessoa desaparecida** em si. É o registro central do sistema.

**Dados**: nome, apelido, data de nascimento, data do desaparecimento, altura, roupas, gênero, olhos, cabelo, pele, foto, localização (lat/lng), status (desaparecido/encontrado), BO policial, tatuagens, cicatrizes
**Calculados**: idade, se era criança no desaparecimento, slug
**Ações**: cadastrar, editar, deletar, listar (paginado), buscar por texto, ver estatísticas, ver no mapa

### 2.3 Avistamento (Sighting)

Quando alguém **vê** a pessoa desaparecida em algum lugar e quer informar.

**Dados**: localização (lat/lng), observação, data
**Vinculado a**: um desaparecido específico
**Efeito colateral**: envia notificação para o familiar (email + WhatsApp)

### 2.4 Homeless ("Quero ser encontrado")

Moradores de rua/desabrigados que querem ser encontrados pela família. Cadastro feito por **ONGs e voluntários**.

**Dados**: nome, apelido, data de nascimento, características físicas, foto, localização
**Efeito colateral**: envia mensagem no Telegram + dispara matching automático com base de desaparecidos

### 2.5 Inteligência Artificial (AI/Matching)

Camada inteligente que usa **Gemini + Imagen** para ajudar na identificação de pessoas. Dois fluxos principais:

**Age Progression**: quando um desaparecido é cadastrado, envia foto para o Gemini → gera projeções visuais de como a pessoa estaria hoje (+1, +3, +5, +10 anos). Útil para desaparecimentos antigos onde a aparência mudou significativamente.

**Face Matching (homeless ↔ missing)**: quando um homeless é cadastrado, compara foto e características físicas com a base de desaparecidos. Filtra por gênero, cor de pele, faixa etária. Usa Gemini Vision para comparação facial. Se score de similaridade > threshold → notifica o familiar automaticamente.

### 2.6 Notificações (Cross-cutting)

Não é um domínio, é um **serviço de infraestrutura** que atravessa todos os domínios:
- Email (Resend) → quando desaparecido é avistado, quando matching AI encontra candidato
- WhatsApp Business → canal principal para familiares (avistamento, matching)
- Telegram Bot → quando novo homeless é cadastrado (notificação para ONGs)
- Push Notification → futuro, mobile (Fase 9)

---

## 3. Arquitetura Proposta

### 3.1 Visão Macro

```
┌──────────────────────┐        ┌──────────────────────────────────┐
│   React SPA          │  REST  │   Go API (Cloud Run)             │
│   (Vite + TS)        │◄─────►│                                  │
│                      │  JSON  │  ┌────────┐  ┌────────────────┐  │
│  - React Router      │        │  │ Handler │→ │ Service/Domain │  │
│  - TanStack Query    │        │  └────────┘  └───────┬────────┘  │
│  - Tailwind + shadcn │        │                      │           │
│  - Mapbox GL JS      │        │              ┌───────┴────────┐  │
└──────────────────────┘        │              │  Repository    │  │
                                │              └───────┬────────┘  │
  ┌──────────────────────┐      │                      │           │
  │ React Native (futuro)│      └──────────────────────┼───────────┘
  └──────────────────────┘                             │
                                              ┌────────┴────────┐
                                          Firestore      Cloud Storage
                                         (documentos)    (imagens)
```

### 3.2 Stack Proposta

| Camada | Tecnologia | Justificativa |
|---|---|---|
| **Backend** | Go 1.22+ | Performance, tipagem forte, goroutines nativas |
| **HTTP Router** | Chi | Leve, idiomático, compatível com `net/http` |
| **Frontend Web** | React 18 + Vite + TypeScript | Ecossistema maduro, componentização |
| **UI** | Tailwind CSS + shadcn/ui + Lucide | Design system moderno, acessível |
| **Mapas** | Google Maps Platform (@vis.gl/react-google-maps) | Integração GCP nativa, heatmaps, geocoding superior no Brasil |
| **AI / Visão** | Gemini API + Imagen 3 | Age progression de desaparecidos, face matching homeless↔missing |
| **Banco** | Firestore | Serverless, real-time, escala automática |
| **Storage** | Cloud Storage (GCS) | Substitui Cloudinary |
| **Async** | Cloud Tasks / Pub/Sub | Substitui Redis+Celery |
| **Email** | Resend | Moderno, DX superior, React Email para templates |
| **WhatsApp** | WhatsApp Business API | Canal principal para familiares brasileiros |
| **Auth** | Firebase Auth | JWT, seguro, OAuth futuro |
| **i18n Backend** | go-i18n v2 (TOML, embed FS) | Mensagens da API traduzidas via `Accept-Language` |
| **i18n Frontend** | react-i18next + i18next-browser-languagedetector | Detecção automática do idioma do browser, fallback PT-BR |
| **Deploy** | Cloud Run + Firebase Hosting | Serverless |

**Idiomas suportados**: PT-BR (padrão) e EN — desde a Fase 0.

---

## 4. Decisões Arquiteturais — Explicadas em Profundidade

Cada decisão abaixo inclui o **contexto**, as **alternativas consideradas**, o **porquê** da escolha e os **trade-offs**.

### 4.1 Estrutura do Projeto Go: Pragmatic Clean Architecture

#### O Problema

No mundo Go, existe um debate grande sobre como organizar projetos. Você vai encontrar:

- **Flat structure** — tudo na raiz, sem pastas. Funciona para projetos pequenos, mas vira bagunça com 4 domínios como o nosso.
- **Clean Architecture "clássica"** (Uncle Bob) — camadas rígidas: entities, use cases, interfaces, frameworks. Funciona bem em Java/C#, mas em Go pode gerar **excesso de abstração**.
- **Domain-Driven com packages por contexto** — cada bounded context é um package Go. É o mais idiomático em Go.

#### Por que NÃO vamos fazer Clean Arch "puro" ao estilo Java/C#

Você tocou num ponto excelente. Em muitos projetos Go (e Java), as pessoas criam um diretório `usecases/` ou `commands/` onde cada caso de uso vira um arquivo separado:

```
usecases/
  create_user.go
  update_user.go
  delete_user.go
  authenticate_user.go
  change_password.go
  ...
```

**O problema disso:**

1. **Explosão de arquivos** — cada ação do sistema vira um arquivo. Com 4 domínios e ~5 ações cada, você tem 20+ arquivos que são basicamente funções com 10-20 linhas.
2. **Indireção desnecessária** — o handler chama o use case, que chama o service, que chama o repository. São 4 camadas para o que poderia ser 2.
3. **Em Go, não é idiomático** — Go valoriza simplicidade. A comunidade Go rejeita abstrações que existem "por princípio" sem resolver um problema real. Rob Pike (criador do Go) diz: *"A little copying is better than a little dependency."*
4. **Use Cases como você conhece no DDD do Java** são classes com um método `Execute()`. Em Go, **funções são cidadãs de primeira classe** — você não precisa de uma classe wrapper para cada ação.

#### O que vamos fazer: Pragmatic Domain Architecture

Em vez de camadas horizontais (use cases → services → repositories), vamos organizar por **domínio vertical**:

```
internal/
  domain/
    user/
      entity.go        ← struct User com validações e comportamento
      repository.go    ← interface (contrato)
      service.go       ← TODAS as ações do usuário (Create, Auth, Update, ChangePassword...)
    missing/
      entity.go
      repository.go
      service.go
    sighted/
      ...
    homeless/
      ...
  infrastructure/
    firestore/
      user_repository.go   ← implementação concreta
      ...
  handler/
    user_handler.go        ← HTTP handlers (recebe request, chama service, retorna response)
    ...
```

**Por que isso é melhor:**

1. **O service.go agrupa as ações do domínio num único lugar.** Se eu quero entender tudo que o sistema faz com usuário, abro UM arquivo. Não preciso navegar em 7 arquivos de use case.

2. **A interface do repository fica no pacote do domínio** (não na infraestrutura). Isso é o princípio de inversão de dependência — o domínio define O QUE ele precisa, a infraestrutura implementa COMO.

3. **São só 3 camadas**: handler → service → repository. Simples, rastreável, debugável.

4. **É o padrão mais adotado na comunidade Go profissional.** Empresas como Uber, Google, Cloudflare organizam seus projetos Go dessa forma. O guia oficial de estilo do Uber ([uber-go/guide](https://github.com/uber-go/guide)) recomenda essa estrutura.

#### Quando um service fica grande demais?

Se no futuro o `service.go` de um domínio passar de ~300 linhas, aí sim faz sentido dividir:

```
user/
  service.go              → Create, FindByID, Update, Delete
  auth_service.go         → Authenticate, ChangePassword, ForgotPassword
```

Mas essa é uma decisão que tomamos **quando a dor aparecer**, não preventivamente. Em Go, a regra é: **comece simples, refatore quando necessário**.

#### Resumo da decisão

| Abordagem | Prós | Contras | Veredicto |
|---|---|---|---|
| Flat (sem pastas) | Simples | Não escala com 4+ domínios | ❌ |
| Clean Arch puro (use cases) | Separação rígida | Excesso de arquivos, não idiomático Go | ❌ |
| Hexagonal / Ports & Adapters | Testável | Muita indireção para nosso tamanho | ❌ |
| **Domain packages + service** | Simples, Go idiomático, testável | Service pode crescer | ✅ |

---

### 4.2 Por que Go e não manter Python / ou usar Node / ou Rust?

#### Go vs Python (Flask)

O projeto atual é Python. Por que não reconstruir em Python com FastAPI?

| Aspecto | Python/FastAPI | Go |
|---|---|---|
| **Performance** | ~10-50x mais lento em CPU-bound | Compilado, extremamente rápido |
| **Concorrência** | asyncio (event loop, uma thread) | Goroutines (milhares de threads leves) |
| **Tipagem** | Opcional (type hints) | Obrigatória em compilação |
| **Deploy** | Requer runtime Python + dependências | Binário estático único (~15MB) |
| **Memory** | ~50-100MB por instância | ~5-15MB por instância |
| **Cloud Run** | Cold start ~2-5s | Cold start ~100-300ms |

Para o **Cloud Run** especificamente, Go é ideal porque:
- Cold start rápido = resposta imediata mesmo com scale-to-zero
- Binário pequeno = imagem Docker mínima
- Baixo consumo de memória = custo menor

#### Go vs Node.js (TypeScript)

| Aspecto | Node/TS | Go |
|---|---|---|
| **Ecossistema web** | Enorme (Express, Nest, etc) | Menor, mas suficiente |
| **Tipagem** | TypeScript (transpilado) | Nativa (compilada) |
| **Concorrência** | Event loop (single-threaded) | Goroutines (multi-threaded real) |
| **Deploy** | node_modules pesado | Binário único |
| **Aprendizado** | Você já sabe | Oportunidade de aprender algo novo |

Node seria uma escolha válida, mas **você quer aprender Go**, e Go é objetivamente superior para APIs em Cloud Run.

#### Go vs Rust

Rust tem performance ainda melhor, mas a curva de aprendizado é **brutal** (borrow checker, lifetimes). Go oferece 90% do benefício de performance com 30% da complexidade.

**Veredicto**: Go é o sweet spot entre performance, simplicidade e curva de aprendizado.

---

### 4.3 Por que Firestore e não PostgreSQL ou manter MongoDB?

Essa é uma decisão importante. Vamos analisar com calma.

#### MongoDB → Firestore (por que não manter Mongo?)

O projeto atual usa MongoDB. Por que não continuar?

1. **Infraestrutura**: MongoDB requer um servidor (Atlas ou self-hosted). Firestore é 100% serverless — zero administração.
2. **Integração GCP**: Se já estamos no Cloud Run, Firestore se conecta sem configuração especial. MongoDB Atlas é um serviço separado com billing separado.
3. **Real-time**: Firestore tem listeners em tempo real nativos. Se no futuro quisermos que o mapa atualize em tempo real quando alguém reportar um avistamento, já está pronto.
4. **Custo**: Para o volume deste projeto (centenas a milhares de documentos), Firestore é **gratuito** no free tier.

#### Por que não PostgreSQL?

PostgreSQL é excelente — seria uma escolha válida. Mas:

1. **Requer servidor** — precisa provisionar, configurar, manter (Cloud SQL ou similar)
2. **Schema rígido** — precisa de migrations, ORM ou query builder
3. **Custo fixo** — Cloud SQL cobra por hora, mesmo sem uso

Firestore é schemaless como o MongoDB que você já usava, mas com a vantagem de ser serverless.

#### Trade-offs do Firestore (honestidade)

Firestore **não é perfeito**:

| Limitação | Impacto no projeto | Solução |
|---|---|---|
| Sem full-text search | Busca de desaparecidos por nome | Algolia ou Typesense como índice externo |
| Sem aggregations nativas | Dashboard de estatísticas | Distributed counters + queries |
| 1 write/second por documento | Irrelevante para nosso volume | — |
| Vendor lock-in (Google) | Mudança futura seria trabalhosa | Interface de repositório isola isso |

O ponto-chave: a **interface de repositório no domínio** (`repository.go`) nos protege. Se no futuro você quiser trocar Firestore por PostgreSQL, só precisa criar uma nova implementação da mesma interface. O domínio não muda.

---

### 4.4 Por que Firebase Auth e não autenticação custom?

O projeto legado armazena senhas com **MD5 sem salt**. Isso é um risco crítico. Se esse banco vazasse, todas as senhas seriam quebradas em minutos.

**Firebase Auth resolve isso completamente:**

1. Senhas armazenadas com **bcrypt** (padrão da indústria)
2. JWT gerado e verificado pelo Firebase — você não implementa crypto
3. Recuperação de senha built-in (Firebase envia o email)
4. No futuro: login com Google, Facebook, Apple com **zero código adicional**
5. Proteção contra brute force built-in

**A alternativa seria** implementar auth manualmente em Go com bcrypt + JWT. É possível e muitos projetos fazem isso. Mas:
- Requer implementar geração e validação de JWT
- Requer implementar fluxo de refresh token
- Requer implementar rate limiting no login
- Requer implementar envio de email para reset de senha

Tudo isso o Firebase Auth faz com **2 linhas de código**. Para um projeto onde o objetivo é aprender Go (não aprender criptografia), delegar auth é a decisão correta.

---

### 4.5 REST vs gRPC — Quando cada um faz sentido?

Você mencionou que não tem domínio de gRPC. Boa decisão ficar com REST por agora. Mas é importante entender **quando** gRPC brilha:

| Critério | REST (JSON) | gRPC (Protobuf) |
|---|---|---|
| **Quem consome** | Browsers, apps mobile, qualquer HTTP client | Serviços backend entre si |
| **Performance** | Boa (JSON ~1-5ms de parse) | Excelente (Protobuf ~10x menor, typed) |
| **Debugging** | Fácil (curl, Postman, browser) | Difícil (precisa de tooling especial) |
| **Streaming** | Limitado (SSE, WebSocket) | Bidirecional nativo |
| **Contrato** | OpenAPI/Swagger (opcional) | .proto file (obrigatório) |

**Para o nosso projeto**: REST é a escolha certa porque:
- React consome REST nativamente
- React Native consome REST nativamente
- Temos **um** serviço backend (não microsserviços conversando entre si)
- Facilita seu aprendizado e debugging

**Quando considerar gRPC no futuro**: se um dia o projeto crescer e você criar microsserviços internos (ex: serviço de busca separado, serviço de notificações separado), a comunicação **entre eles** poderia ser gRPC. A comunicação com o frontend continuaria REST.

---

### 4.6 Assincronicidade: Goroutines vs Celery/Redis

No Python, quando você quer executar algo em background (enviar email, notificar Telegram), precisa de **infraestrutura extra**: Redis como broker + Celery como worker + processo separado rodando o worker.

Em Go, a história é completamente diferente.

#### Goroutines — a superpower do Go

Uma goroutine é como uma "thread leve" que o Go gerencia internamente. Criar uma goroutine custa ~2KB de memória. Criar uma thread no OS custa ~1MB.

Para enviar um email em background:

```go
// Python/Celery: precisa de Redis + Celery + worker process
send_email.delay(user_email, subject, body)

// Go: nativo, zero infraestrutura extra
go sendEmail(ctx, userEmail, subject, body)
```

Sim, é literalmente a palavra `go` antes da chamada de função. A função executa em paralelo enquanto o handler retorna a resposta.

#### Quando goroutine simples NÃO basta

Se o processo do Go **morrer** enquanto a goroutine está executando, a tarefa se perde. Para tarefas críticas (que **não podem ser perdidas**), usamos **Cloud Tasks** ou **Pub/Sub**:

| Cenário | Solução | Por quê |
|---|---|---|
| Enviar email de avistamento | Cloud Tasks | Não pode perder — o familiar precisa ser notificado |
| Notificação Telegram | Goroutine simples | Se perder, não é catastrófico |
| Indexar busca no Algolia | Cloud Tasks | Precisa de consistência |

Cloud Tasks e Pub/Sub são serviços do GCP que garantem **at-least-once delivery** — se falhar, retenta automaticamente.

---

### 4.7 Monorepo vs Multi-repo

**Decisão: Monorepo** — um único repositório Git com todos os projetos.

#### Visão completa da estrutura

```
desaparecidos/                          ← raiz do monorepo (1 repositório Git)
│
├── api/                                ← módulo Go (backend)
│   ├── cmd/
│   │   └── server/
│   │       └── main.go                 ← entry point do servidor
│   ├── internal/                       ← protegido pelo compilador Go
│   │   ├── config/                     ← leitura de variáveis de ambiente
│   │   ├── domain/                     ← regras de negócio (o coração)
│   │   │   ├── shared/                 ← value objects compartilhados
│   │   │   ├── user/                   ← entity + repository + service
│   │   │   ├── missing/                ← entity + repository + service
│   │   │   ├── sighted/                ← entity + repository + service
│   │   │   └── homeless/               ← entity + repository + service
│   │   ├── handler/                    ← HTTP handlers (recebem request, chamam service)
│   │   │   ├── middleware/             ← auth, cors, logging, rate limit
│   │   │   └── dto/                    ← request/response structs
│   │   └── infrastructure/             ← implementações concretas
│   │       ├── firestore/              ← repositórios Firestore
│   │       ├── storage/                ← upload de imagens (Cloud Storage)
│   │       ├── notification/           ← email, WhatsApp, Telegram, push
│   │       └── auth/                   ← Firebase Auth
│   ├── pkg/                            ← utilitários reutilizáveis
│   │   ├── slug/
│   │   ├── httputil/
│   │   └── dateutil/
│   ├── Dockerfile
│   ├── go.mod                          ← define o módulo Go
│   └── go.sum
│
├── web/                                ← módulo React (frontend web)
│   ├── src/
│   │   ├── app/                        ← rotas e providers
│   │   ├── features/                   ← funcionalidades por domínio
│   │   │   ├── auth/
│   │   │   ├── missing/
│   │   │   ├── homeless/
│   │   │   ├── sighted/
│   │   │   └── dashboard/
│   │   └── shared/                     ← componentes e hooks reutilizáveis
│   ├── public/
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   └── tsconfig.json
│
├── mobile/                             ← módulo React Native (futuro, Fase 8)
│   ├── app/
│   ├── package.json
│   └── app.json
│
├── shared/                             ← código TypeScript compartilhado (web + mobile)
│   ├── types/                          ← interfaces: User, Missing, Sighting...
│   ├── services/                       ← API client, chamadas ao backend
│   ├── constants/                      ← labels PT-BR, rotas da API
│   ├── validators/                     ← regras de validação de formulários
│   └── package.json
│
├── docs/                               ← documentação (estes arquivos)
│   ├── ROADMAP.md
│   ├── FASE_00_FUNDACAO.md
│   └── ...
│
├── .github/
│   └── workflows/                      ← CI/CD (GitHub Actions)
│       ├── api.yml                     ← build + deploy do Go no Cloud Run
│       └── web.yml                     ← build + deploy do React no Firebase Hosting
│
├── docker-compose.yml                  ← Firebase Emulators para dev local
├── .gitignore
├── LICENSE
└── README.md
```

#### O que cada diretório faz

| Diretório | Toolchain | Função |
|---|---|---|
| `api/` | Go (`go build`, `go test`) | API REST — toda a lógica de backend |
| `web/` | Node (`npm run dev/build`) | Frontend web React |
| `mobile/` | Expo (`npx expo start`) | App mobile (Fase 8) |
| `shared/` | TypeScript puro | Types, services e constantes compartilhados entre web e mobile |
| `docs/` | Markdown | Documentação do roadmap e fases |

**Nota importante**: `internal/` é um diretório que existe **dentro de `api/`**, não na raiz. É uma keyword do Go que o compilador protege — código dentro de `internal/` só pode ser importado por código do mesmo módulo.

#### Por que monorepo?

- Você é o único desenvolvedor — ter múltiplos repos é overhead sem benefício
- Facilita CI/CD (workflows separados por diretório, mas um só lugar para configurar)
- Facilita correlacionar mudanças (commit que muda API + frontend junto)
- `shared/` permite compartilhar código entre web e mobile sem publicar pacotes npm
- Se um dia tiver time, migrar para multi-repo é simples

---

### 4.8 Evolução dos Canais de Notificação

O cenário de comunicação mudou muito desde 2018. O projeto legado usava SendGrid (email) e Telegram Bot. Hoje temos opções muito melhores.

#### Comparação de canais

| Canal | 2018 (legado) | 2025+ (novo) | Para quê |
|---|---|---|---|
| **Email** | SendGrid | **Resend** | Recuperação de senha, avistamento registrado |
| **Telegram Bot** | ✅ Usado | Manter como secundário | Notificação interna para ONGs |
| **WhatsApp Business** | Não existia | ✅ **Adicionar** | Canal principal para familiares |
| **Push Notification** | Não existia | ✅ Fase 8 (mobile) | Alertas no celular |

#### Por que WhatsApp Business?

1. **96% dos smartphones brasileiros têm WhatsApp** — é o canal com maior alcance no Brasil
2. **Urgência** — avistamento é urgente. Email pode demorar horas. WhatsApp é lido em minutos
3. **Interação** — no futuro, o familiar pode responder ("Confirmo, é ele!")
4. **Familiaridade** — o familiar não precisa aprender nada novo

Custo: ~R$0,15-0,40 por mensagem (conversation-based pricing). Para o volume do projeto, muito acessível.

Requisitos: Meta Business Account + aprovação de templates de mensagem + número verificado.

#### Por que Resend em vez de SendGrid?

| Aspecto | SendGrid | Resend |
|---|---|---|
| **API** | REST, funcional mas verbosa | REST, moderna e clean |
| **Free tier** | 100 emails/dia | 3.000 emails/mês |
| **Templates** | Limitados | React Email (componentes React para email) |
| **DX** | OK | Excelente |

#### Como a interface protege essa evolução

A interface `Notifier` no domínio não sabe **como** a notificação é enviada:

```go
type Notifier interface {
    NotifySighting(ctx context.Context, params SightingNotification) error
    NotifyNewHomeless(ctx context.Context, params NewHomelessNotification) error
}
```

A implementação concreta decide quais canais usar:

```go
type MultiChannelNotifier struct {
    email    *ResendSender       // substitui SendGrid
    whatsapp *WhatsAppSender     // novo!
    telegram *TelegramSender     // mantém para ONGs
    push     *FCMPushSender      // futuro (mobile)
}
```

O service que registra o avistamento chama `notifier.NotifySighting()` e não sabe se vai por email, WhatsApp, Telegram ou push. A infraestrutura decide. Adicionar um canal novo = criar um sender + registrar no MultiChannelNotifier. Zero mudança no domínio.

---

### 4.9 O que NÃO migrar (decisões conscientes)

| Item Legado | Decisão | Motivo |
|---|---|---|
| MD5 para senhas | ❌ Eliminar | Firebase Auth usa bcrypt |
| Basic Auth | ❌ Eliminar | JWT via Firebase é superior |
| Cloudinary | ❌ Substituir por Cloud Storage | Integração GCP nativa |
| SendGrid | 🔄 Substituir por Resend | API moderna, React Email, melhor DX |
| Flask BFF | ❌ Eliminar | React consome API diretamente |
| Redis + Celery | ❌ Substituir | Goroutines + Cloud Tasks |
| Hotjar | ⏸️ Avaliar depois | Pode não ser necessário |

---

## 5. Modelo de Dados no Firestore

### 5.1 Coleções

```
users/{userId}
  ├── name: string
  ├── email: string
  ├── phone: string
  ├── cellPhone: string
  ├── avatarURL: string
  ├── acceptedTerms: boolean
  ├── role: string (user | volunteer | ong | admin)
  ├── alertRadius: number (km — raio para receber alertas de proximidade, default: 0 = desativado)
  ├── alertLocation: geopoint (centro do raio de alerta, null se desativado)
  ├── createdAt: timestamp
  └── updatedAt: timestamp

missing/{missingId}
  │
  │ ── Identificação ──
  ├── userId: string (referência ao familiar que cadastrou)
  ├── name: string
  ├── nickname: string
  ├── birthDate: timestamp
  ├── slug: string
  ├── status: string (disappeared | found)
  │
  │ ── Características Físicas ──
  ├── gender: string
  ├── eyes: string
  ├── hair: string
  ├── skin: string
  ├── height: string (ex: "1.72")
  ├── weight: string (ex: "68")
  ├── bodyType: string (slim | medium | heavy)
  ├── birthmarkDescription: string (marcas de nascença — imutáveis)
  ├── tattooDescription: string
  ├── scarDescription: string
  ├── prosthetics: string (óculos, aparelho auditivo, cadeira de rodas, prótese)
  │
  │ ── Saúde (define urgência) ──
  ├── medicalCondition: string (alzheimer | autism | epilepsy | intellectual_disability | none | other)
  ├── medicalConditionDetails: string (detalhes se other)
  ├── continuousMedication: string (medicação de uso contínuo, se houver)
  ├── bloodType: string (A+ | A- | B+ | B- | AB+ | AB- | O+ | O-)
  │
  │ ── Circunstância ──
  ├── dateOfDisappearance: timestamp
  ├── disappearanceLocation: geopoint (local do desaparecimento)
  ├── lastSeenLocation: geopoint (último local onde foi visto — pode diferir do anterior)
  ├── lastSeenClothes: string (roupa que usava quando desapareceu)
  ├── usualClothes: string (roupas que costuma usar)
  ├── circumstance: string (left_home | ran_away | abduction | hospital | disaster | unknown)
  ├── circumstanceDetails: string (detalhes adicionais)
  │
  │ ── Investigação ──
  ├── policeReportNumber: string (número do BO — estruturado)
  ├── policeStation: string (delegacia responsável)
  ├── investigatorContact: string (telefone do investigador)
  ├── riskLevel: string (critical | high | medium | low — calculado pelo sistema)
  │
  │ ── Mídia ──
  ├── photoURLs: array<string> (múltiplas fotos — frente, perfil, corpo inteiro)
  ├── ageProgressionURLs: map<string, string> ({"1y": url, "3y": url, "5y": url, "10y": url})
  │
  │ ── Legado ──
  ├── wasChild: boolean
  ├── createdAt: timestamp
  └── updatedAt: timestamp

matches/{matchId}
  ├── homelessId: string (referência)
  ├── missingId: string (referência)
  ├── score: number (0.0 - 1.0)
  ├── status: string (pending | confirmed | rejected)
  ├── reviewedBy: string (userId que confirmou/rejeitou)
  ├── geminiAnalysis: string (texto da análise do Gemini)
  ├── createdAt: timestamp
  └── reviewedAt: timestamp

sightings/{sightingId}
  ├── missingId: string (referência)
  ├── userId: string (quem registrou — null se anônimo)
  │
  │ ── Quando e onde ──
  ├── seenAt: timestamp (quando efetivamente viu a pessoa)
  ├── location: geopoint (onde viu)
  ├── movementDirection: string (north | south | east | west | unknown)
  │
  │ ── O que observou ──
  ├── observation: string (texto livre)
  ├── physicalState: string (apparently_well | injured | disoriented | substance_use | unknown)
  ├── accompanied: string (alone | with_adult | with_child | in_group)
  ├── companionDescription: string (se acompanhado)
  ├── confidenceLevel: string (certain | likely | uncertain)
  │
  │ ── Mídia ──
  ├── photoURLs: array<string> (fotos/prints do avistamento)
  │
  └── createdAt: timestamp (quando registrou na plataforma)

homeless/{homelessId}
  │
  │ ── Identificação ──
  ├── name: string (se souber)
  ├── nickname: string
  ├── estimatedAge: number (idade estimada pelo voluntário, se nascimento desconhecido)
  ├── birthDate: timestamp (se souber)
  ├── slug: string
  │
  │ ── Características Físicas ──
  ├── gender: string
  ├── eyes: string
  ├── hair: string
  ├── skin: string
  ├── height: string
  ├── weight: string
  ├── bodyType: string (slim | medium | heavy)
  ├── birthmarkDescription: string
  ├── tattooDescription: string
  ├── scarDescription: string
  ├── prosthetics: string
  │
  │ ── Contexto ──
  ├── location: geopoint
  ├── spokenLanguage: string (portuguese | spanish | english | unknown | other)
  ├── mentalState: string (oriented | disoriented | non_responsive | incoherent)
  ├── selfReportedInfo: string (cidade de origem, nomes de familiares, fragmentos de memória)
  ├── estimatedTimeOnStreet: string (days | weeks | months | years | unknown)
  ├── physicalCondition: string (apparently_well | malnourished | injured | other)
  │
  │ ── Mídia ──
  ├── photoURLs: array<string> (múltiplas fotos — frente, perfil, mãos, marcas)
  │
  ├── createdAt: timestamp
  └── updatedAt: timestamp

tips/{tipId}  ← NOVA COLEÇÃO — denúncias anônimas
  ├── missingId: string (referência, opcional — pode ser denúncia geral)
  ├── message: string
  ├── location: geopoint (opcional)
  ├── anonymousCode: string (código de protocolo para acompanhamento)
  ├── status: string (new | reviewed | actionable | dismissed)
  ├── reviewedBy: string (userId do admin/ONG que revisou)
  ├── reviewNote: string (nota interna da revisão)
  └── createdAt: timestamp

timeline/{eventId}  ← NOVA COLEÇÃO — histórico investigativo do caso
  ├── missingId: string (referência)
  ├── type: string (created | updated | sighting_added | tip_received |
  │                  ai_age_progression | ai_match_found | status_changed |
  │                  alert_sent | photo_added)
  ├── description: string (texto legível: "Avistamento registrado em São Paulo")
  ├── userId: string (quem gerou o evento — null se sistema)
  ├── metadata: map (dados específicos: {sightingId: "...", location: {...}})
  └── createdAt: timestamp
```

### 5.1.1 Campo calculado: riskLevel

O `riskLevel` do desaparecido é **calculado automaticamente** pelo sistema com base em regras:

| Nível | Critérios |
|---|---|
| **critical** | Criança < 12 anos, OU idoso > 70 anos, OU condição médica grave (alzheimer, epilepsy), OU medicação urgente, OU circunstância = abduction |
| **high** | Adolescente < 18, OU deficiência intelectual, OU desaparecido há > 72h sem avistamento |
| **medium** | Adulto sem condição de risco, desaparecido recentemente (< 72h) |
| **low** | Possível saída voluntária (ran_away), adulto sem vulnerabilidade |

O risco influencia:
- **Ordenação na listagem** — casos críticos aparecem primeiro
- **Badge visual** — vermelho, laranja, amarelo, verde
- **Frequência de alertas** — casos críticos disparam alertas regionais imediatos
- **Prioridade no AI** — age progression e matching processados primeiro

### 5.2 Diferenças-Chave vs MongoDB

| Aspecto | MongoDB (atual) | Firestore (novo) |
|---|---|---|
| IDs | ObjectId | Auto-generated string |
| Geo | Array [lng, lat] | GeoPoint nativo |
| Full-text search | $text index | Serviço externo (Algolia/Typesense) |
| Aggregations | Pipeline nativo | Distributed counters + queries |
| Senhas | MD5 no documento | Firebase Auth (separado) |

---

## 6. Endpoints da Nova API

| Método | Rota | Descrição |
|---|---|---|
| `GET` | `/api/v1/health` | Health check |
| `POST` | `/api/v1/auth/login` | Login |
| `POST` | `/api/v1/auth/forgot-password` | Recuperar senha |
| `POST` | `/api/v1/users` | Criar usuário |
| `GET` | `/api/v1/users/:id` | Buscar usuário |
| `PUT` | `/api/v1/users/:id` | Atualizar usuário |
| `DELETE` | `/api/v1/users/:id` | Deletar usuário |
| `PATCH` | `/api/v1/users/:id/password` | Alterar senha |
| `POST` | `/api/v1/missing` | Criar desaparecido |
| `GET` | `/api/v1/missing` | Listar (paginado) |
| `GET` | `/api/v1/missing/:id` | Buscar por ID |
| `PUT` | `/api/v1/missing/:id` | Atualizar |
| `DELETE` | `/api/v1/missing/:id` | Deletar |
| `GET` | `/api/v1/missing/search?q=` | Busca textual |
| `GET` | `/api/v1/missing/stats` | Estatísticas |
| `GET` | `/api/v1/missing/locations` | Geo-agrupamento |
| `GET` | `/api/v1/users/:id/missing` | Missing por usuário |
| `POST` | `/api/v1/missing/:id/sightings` | Registrar avistamento |
| `GET` | `/api/v1/missing/:id/sightings` | Listar avistamentos |
| `GET` | `/api/v1/sightings/:id` | Buscar avistamento |
| `PATCH` | `/api/v1/missing/:id/status` | Alterar status (disappeared ↔ found) |
| `GET` | `/api/v1/missing/:id/age-progression` | Obter imagens de age progression |
| `POST` | `/api/v1/missing/:id/age-progression` | Gerar age progression (trigger Gemini) |
| `POST` | `/api/v1/homeless` | Cadastrar homeless |
| `GET` | `/api/v1/homeless` | Listar homeless |
| `GET` | `/api/v1/homeless/:id` | Buscar por ID |
| `GET` | `/api/v1/homeless/stats` | Estatísticas |
| `GET` | `/api/v1/homeless/:id/matches` | Candidatos de matching |
| `PATCH` | `/api/v1/matches/:id` | Confirmar/rejeitar match |
| `POST` | `/api/v1/upload` | Upload de imagem |
| | | |
| | **Denúncias Anônimas (Tips)** | |
| `POST` | `/api/v1/tips` | Criar denúncia anônima (sem auth) |
| `GET` | `/api/v1/tips/:code` | Consultar status por código anônimo (sem auth) |
| `GET` | `/api/v1/tips` | Listar denúncias (admin/ONG) |
| `PATCH` | `/api/v1/tips/:id` | Revisar denúncia (admin/ONG) |
| | | |
| | **Timeline** | |
| `GET` | `/api/v1/missing/:id/timeline` | Histórico do caso |
| | | |
| | **Alertas de Proximidade** | |
| `PATCH` | `/api/v1/users/:id/alert-settings` | Configurar raio de alerta |
| | | |
| | **Cartaz Digital** | |
| `GET` | `/api/v1/missing/:id/poster` | Gerar cartaz PDF/imagem com QR Code |

---

## 7. Índice das Fases

Cada fase tem seu próprio arquivo com:
- Objetivos e entregáveis
- Conceitos de Go explicados em profundidade
- Tarefas detalhadas (backend + frontend)
- Decisões e trade-offs específicos da fase

| Fase | Arquivo | Duração | Tema |
|---|---|---|---|
| 0 | [FASE_00_FUNDACAO.md](./FASE_00_FUNDACAO.md) | 2 semanas | Setup, primeiros bytes de Go |
| 0B | [FASE_00B_DOCKER_LOCAL.md](./FASE_00B_DOCKER_LOCAL.md) | 1–2 dias | Ambiente local dockerizado (docker-compose, hot-reload, Makefile) |
| 1 | [FASE_01_AUTH_USUARIO.md](./FASE_01_AUTH_USUARIO.md) | 3 semanas | Auth completo + CRUD Usuário |
| 2 | [FASE_02_DESAPARECIDOS.md](./FASE_02_DESAPARECIDOS.md) | 4 semanas | CRUD Missing + Google Maps |
| 3 | [FASE_03_BUSCA_DASHBOARD.md](./FASE_03_BUSCA_DASHBOARD.md) | 3 semanas | Busca textual + Dashboard |
| 4 | [FASE_04_AVISTAMENTOS.md](./FASE_04_AVISTAMENTOS.md) | 3 semanas | Sightings + Notificações |
| 5 | [FASE_05_HOMELESS.md](./FASE_05_HOMELESS.md) | 2 semanas | Módulo Homeless |
| 6 | [FASE_06_INTELIGENCIA.md](./FASE_06_INTELIGENCIA.md) | 4 semanas | AI: Age Progression + Face Matching |
| 7 | [FASE_07_POLISH.md](./FASE_07_POLISH.md) | 2 semanas | FAQ, SEO, acessibilidade |
| 8 | [FASE_08_DEPLOY.md](./FASE_08_DEPLOY.md) | 2 semanas | Cloud Run + Observabilidade |
| 9 | [FASE_09_MOBILE.md](./FASE_09_MOBILE.md) | 4+ semanas | React Native |

**Total estimado (web): ~25 semanas** (dedicação parcial ~10-15h/semana)

### Diagramas

→ [DIAGRAMAS.md](./DIAGRAMAS.md) — Diagrama de classes (entidades + services), diagramas de sequência dos fluxos principais e visão geral da arquitetura de componentes.

---

## 8. Melhorias Investigativas — Distribuição nas Fases

As funcionalidades de inteligência investigativa foram distribuídas nas fases existentes para não criar fases extras desnecessárias:

### Fase 2 — Desaparecidos (absorve os novos campos)

- Novos campos na entidade Missing: `weight`, `bodyType`, `birthmarkDescription`, `prosthetics`, `medicalCondition`, `medicalConditionDetails`, `continuousMedication`, `bloodType`, `disappearanceLocation`, `lastSeenLocation`, `lastSeenClothes`, `usualClothes`, `circumstance`, `circumstanceDetails`, `policeReportNumber`, `policeStation`, `investigatorContact`
- **`photoURLs`** substitui `photoURL` — múltiplas fotos (frente, perfil, corpo inteiro)
- **`riskLevel`** calculado automaticamente no `Create()` e `Update()` do service
- Badge visual de risco no card do desaparecido (🔴🟡🟢)
- Formulário de cadastro com seções colapsáveis (identificação / físico / saúde / circunstância / investigação)
- Upload múltiplo de fotos com preview

### Fase 3 — Busca & Dashboard (absorve timeline + filtros de risco)

- **Timeline do caso** — `GET /api/v1/missing/:id/timeline`
- Coleção `timeline/` alimentada automaticamente em cada ação (create, update, sighting, match, status change)
- Componente React de timeline na página de detalhes do desaparecido
- Filtro por `riskLevel` na listagem (mostrar críticos primeiro)
- Dashboard com métricas de risco: quantos casos críticos, altos, médios, baixos
- Polyline no mapa conectando avistamentos por ordem cronológica (rastro de movimento)

### Fase 4 — Avistamentos (absorve avistamento enriquecido + denúncia anônima)

- Novos campos no avistamento: `seenAt`, `physicalState`, `accompanied`, `companionDescription`, `movementDirection`, `confidenceLevel`, `photoURLs`
- Upload de foto no avistamento
- **Denúncia anônima (Tips)** — nova coleção `tips/`, endpoints CRUD, código de protocolo anônimo
- Tela React de denúncia sem login
- Tela de consulta por código anônimo
- Painel de revisão de denúncias para admin/ONG

### Fase 5 — Homeless (absorve dados enriquecidos)

- Novos campos: `estimatedAge`, `height`, `weight`, `bodyType`, `birthmarkDescription`, `tattooDescription`, `scarDescription`, `prosthetics`, `spokenLanguage`, `mentalState`, `selfReportedInfo`, `estimatedTimeOnStreet`, `physicalCondition`
- **`photoURLs`** substitui `photoURL` — múltiplas fotos
- Formulário enriquecido com seções de contexto investigativo

### Fase 6 — AI (absorve prioridade por risco)

- Fila de processamento prioriza por `riskLevel` (critical → low)
- Face matching usa `bodyType`, `birthmarkDescription`, `prosthetics` como filtros adicionais

### Fase 7 — Polish (absorve cartaz digital + QR Code + compartilhamento)

- **Cartaz digital** — `GET /api/v1/missing/:id/poster` → gera PDF com foto + age progression + características + QR Code + telefone
- QR Code por desaparecido linkando para a página pública
- Open Graph otimizado para compartilhamento no WhatsApp/Facebook (foto + nome + "DESAPARECIDO" + cidade)
- Botão "Compartilhar" com link direto para redes sociais

### Fase 9 — Mobile (absorve Radar de Proximidade)

- **Alerta por região** — campo `alertRadius` e `alertLocation` no User
- `PATCH /api/v1/users/:id/alert-settings` — configurar raio
- Push notification via FCM quando novo desaparecido/avistamento ocorre dentro do raio
- Tela de configuração de alerta com mapa para selecionar centro e raio

### Futuro (pós-Fase 9)

- Integração com SINALID (Polícia Federal)
- Cross-reference com IML (corpos não identificados)
- Integração com CRAS/CREAS/abrigos
- Busca ativa coordenada com quadrantes no mapa
- Rede de voluntários verificados com badge

---

## 9. Segurança

A estratégia completa de segurança está documentada em:

→ [SECURITY.md](./SECURITY.md) — Proteção, risco e mitigação em todas as camadas

### Checklist resumido

**Perímetro (Fase 1)**
- [ ] Rate limiting global (200 req/min por IP)
- [ ] Rate limiting por endpoint (5 req/min no login, 2 req/min para AI)
- [ ] CORS para domínios específicos (nunca wildcard)
- [ ] Security Headers (HSTS, X-Frame-Options, Referrer-Policy)
- [ ] Body size limits (1 MB JSON, 10 MB upload)
- [ ] Request timeout (30s)

**Autenticação (Fase 1)**
- [ ] Firebase Auth com email/senha
- [ ] JWT com verificação + expiração + refresh automático
- [ ] Custom claims para roles (user/volunteer/ong/admin)
- [ ] Ownership check em toda operação de escrita

**Dados (toda fase)**
- [ ] Input validation com go-playground/validator
- [ ] Sanitização de texto livre com bluemonday
- [ ] Firestore Security Rules (produção)
- [ ] Cloud Storage rules (auth + size + content-type)
- [ ] Logs sem dados sensíveis

**APIs externas (Fase 2+)**
- [ ] Google Maps API key com HTTP referrer restriction
- [ ] API keys separadas dev/prod
- [ ] Quotas configuradas no GCP Console para cada API
- [ ] Proxy endpoint para Geocoding (key no backend)
- [ ] Gemini API key no Secret Manager + worker pool limitado

**Proteção de custo (Fase 8)**
- [ ] Budget alerts no GCP ($50/mês)
- [ ] Circuit breaker ao atingir 120% do budget

**Bot protection (Fase 4)**
- [ ] reCAPTCHA v3 em endpoints públicos sem auth

**Mobile (Fase 9)**
- [ ] Firebase App Check
- [ ] Google/Apple providers

---

## 10. Filosofia de Aprendizado

Em cada fase, o ciclo será:

1. **Eu explico** o conceito Go e o porquê da abordagem
2. **Você implementa** com minha orientação
3. **Eu reviso** e sugiro melhorias idiomáticas
4. **Refatoramos juntos** quando necessário

Não vou gerar código para você copiar e colar. Vou te ajudar a **entender** e **construir**.

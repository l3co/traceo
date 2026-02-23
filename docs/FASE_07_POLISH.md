# Fase 7 — Páginas Institucionais, SEO & Acessibilidade

> **Duração estimada**: 2 semanas
> **Pré-requisito**: Fase 6 concluída (AI + todos os domínios implementados)

---

## Objetivo

Polir a plataforma: criar as páginas estáticas, otimizar SEO para que desaparecidos sejam encontrados via Google, garantir acessibilidade e responsividade mobile. Ao final desta fase:

- FAQ, Termos de Uso e Política de Privacidade funcionando
- Páginas de erro personalizadas (404, 500)
- Meta tags de SEO e Open Graph (compartilhamento no Facebook/WhatsApp)
- PWA básico (funciona offline para consultas já carregadas)
- Responsividade mobile completa

Esta fase não introduz conceitos novos de Go. O foco é frontend e experiência do usuário.

---

## Por que SEO é crítico neste projeto?

Diferente de um SaaS B2B onde os usuários chegam por convite, o **desaparecidos.me** precisa ser encontrado por:

1. **Familiares desesperados** que buscam "pessoa desaparecida [nome] [cidade]" no Google
2. **Voluntários e ONGs** que buscam "plataforma para registrar moradores de rua"
3. **Compartilhamento social** — quando alguém compartilha um card de desaparecido no Facebook/WhatsApp, precisa ter preview bonito e informativo

Se o Google não indexar bem a plataforma, ela perde o propósito.

### O desafio: SPAs e SEO

React é uma SPA (Single Page Application) — o HTML inicial é vazio, o conteúdo é renderizado via JavaScript. Bots de busca **podem** executar JavaScript, mas não é garantido e é mais lento.

#### Soluções consideradas

| Solução | Prós | Contras | Veredicto |
|---|---|---|---|
| **Client-side rendering (CSR)** | Simples, já é o que temos | SEO ruim, bots podem não indexar | ❌ Para páginas públicas |
| **Server-side rendering (SSR)** com Next.js | SEO perfeito, performance | Muda o framework, complexidade | ⚠️ Possível futuro |
| **Pre-rendering / SSG** | Bom SEO, simples | Dados estáticos (OK para FAQ/termos) | ✅ Para estáticas |
| **Dynamic meta tags** via Go API | SEO para compartilhamento social | Não resolve indexação de conteúdo | ✅ Para cards de desaparecidos |

#### Nossa decisão: abordagem híbrida

1. **Páginas estáticas** (FAQ, Termos, Home) → React com meta tags estáticas. O Google indexa bem conteúdo estático mesmo em SPAs modernas.

2. **Cards de desaparecidos** (a parte mais importante para SEO) → O Go API serve uma **rota de meta tags** que retorna HTML mínimo com Open Graph tags quando detecta um bot:

```go
// Quando o user-agent é um bot (Googlebot, facebookexternalhit, etc.)
// O Go retorna HTML com meta tags
func (h *MetaHandler) ServeMissingMeta(w http.ResponseWriter, r *http.Request) {
    id := chi.URLParam(r, "id")
    m, err := h.service.FindByID(r.Context(), id)
    if err != nil {
        http.Error(w, "Not found", 404)
        return
    }

    html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta property="og:title" content="%s - Desaparecido" />
    <meta property="og:description" content="Desaparecido desde %s. Ajude a encontrar." />
    <meta property="og:image" content="%s" />
    <meta property="og:url" content="https://desaparecidos.me/missing/%s" />
    <meta property="og:type" content="article" />
</head>
<body>
    <script>window.location.href = 'https://desaparecidos.me/missing/%s';</script>
</body>
</html>`, m.Name, m.DateOfDisappearance.Format("02/01/2006"), m.PhotoURL, m.ID, m.ID)

    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(html))
}
```

Isso garante que ao compartilhar `desaparecidos.me/missing/abc123` no WhatsApp/Facebook, apareça:
- **Título**: "João Silva - Desaparecido"
- **Descrição**: "Desaparecido desde 15/03/2020. Ajude a encontrar."
- **Imagem**: Foto do desaparecido

3. **Futuro (se necessário)**: migrar para Next.js para SSR completo. A API Go não muda — Next.js consome a mesma API REST.

---

## Tarefas Detalhadas

### Backend (Go)

#### Tarefa 6.1 — Rota de meta tags para bots

Criar middleware que detecta user-agent de bots e serve HTML com Open Graph tags:
- Googlebot, Bingbot, facebookexternalhit, Twitterbot, WhatsApp, Telegram
- Para usuários normais, serve o SPA React

#### Tarefa 6.2 — Sitemap XML

Criar `GET /sitemap.xml`:
- Lista todas as páginas públicas
- Lista todos os desaparecidos (atualizado dinamicamente)
- Lista todos os homeless

```go
func (h *SitemapHandler) Serve(w http.ResponseWriter, r *http.Request) {
    missing, _ := h.missingService.FindAllPublic(r.Context())

    var urls []SitemapURL
    // Páginas estáticas
    urls = append(urls, SitemapURL{Loc: "https://desaparecidos.me/", Priority: "1.0"})
    urls = append(urls, SitemapURL{Loc: "https://desaparecidos.me/faq", Priority: "0.5"})
    urls = append(urls, SitemapURL{Loc: "https://desaparecidos.me/missing", Priority: "0.9"})

    // Desaparecidos individuais
    for _, m := range missing {
        urls = append(urls, SitemapURL{
            Loc:     fmt.Sprintf("https://desaparecidos.me/missing/%s", m.ID),
            LastMod: m.UpdatedAt.Format("2006-01-02"),
            Priority: "0.8",
        })
    }

    // Render XML
    w.Header().Set("Content-Type", "application/xml")
    // ... gera XML ...
}
```

#### Tarefa 6.3 — robots.txt

```
GET /robots.txt

User-agent: *
Allow: /
Sitemap: https://desaparecidos.me/sitemap.xml
```

### Frontend (React)

#### Tarefa 6.4 — Página FAQ

- Accordion com 5 perguntas (mesmas do legado, texto revisado)
- Componente shadcn/ui `Accordion`
- Conteúdo em PT-BR

#### Tarefa 6.5 — Página de Termos de Uso / Política de Privacidade

- Texto revisado e atualizado (LGPD compliance)
- Layout limpo e legível
- Link no rodapé do site e no formulário de cadastro

#### Tarefa 6.6 — Páginas de erro

- **404** — ilustração amigável + link para home + busca
- **403** — mensagem de acesso negado + link para login
- **500** — mensagem de erro genérica + "tente novamente"
- React Router catch-all para 404

#### Tarefa 6.7 — Meta tags no React

```tsx
// src/shared/components/Head.tsx — usando react-helmet-async
import { Helmet } from 'react-helmet-async'

export function Head({ title, description, image }: HeadProps) {
    return (
        <Helmet>
            <title>{title} | Desaparecidos.me</title>
            <meta name="description" content={description} />
            <meta property="og:title" content={title} />
            <meta property="og:description" content={description} />
            {image && <meta property="og:image" content={image} />}
            <meta property="og:type" content="website" />
        </Helmet>
    )
}
```

Cada página define seus meta tags:
```tsx
function MissingListPage() {
    return (
        <>
            <Head
                title="Desaparecidos"
                description="Encontre pessoas desaparecidas no Brasil. Registre um desaparecimento ou informe um avistamento."
            />
            {/* ... conteúdo ... */}
        </>
    )
}
```

#### Tarefa 6.8 — PWA básico

- `manifest.json` com nome, ícones, cores
- Service worker para cache de assets estáticos
- Banner "Adicionar à tela inicial" no mobile
- Funciona offline para páginas já visitadas (cache-first strategy)

```json
{
    "name": "Desaparecidos.me",
    "short_name": "Desaparecidos",
    "description": "Plataforma para encontrar pessoas desaparecidas",
    "start_url": "/",
    "display": "standalone",
    "background_color": "#ffffff",
    "theme_color": "#0097D6"
}
```

#### Tarefa 6.9 — Responsividade mobile

Verificar e ajustar:
- Sidebar colapsa em hamburger menu no mobile
- Cards de desaparecidos: grid 3 colunas (desktop) → 1 coluna (mobile)
- Mapas: full-width no mobile
- Formulários: campos empilhados no mobile
- Touch targets mínimo 44x44px (acessibilidade)

#### Tarefa 6.10 — Acessibilidade (a11y)

- Todas as imagens com `alt` descritivo
- Formulários com `label` associado ao `input`
- Contraste de cores WCAG AA (mínimo 4.5:1)
- Navegação por teclado (Tab, Enter, Escape)
- ARIA labels nos componentes interativos
- Skip-to-content link
- Anúncios de status com `aria-live`

Ferramentas para verificar:
- Chrome DevTools → Lighthouse → Accessibility
- `eslint-plugin-jsx-a11y` (lint de acessibilidade para React)

#### Tarefa 6.11 — Google Analytics

```tsx
// src/shared/lib/analytics.ts
// Usando gtag.js (GA4)
export function trackPageView(path: string) {
    window.gtag?.('event', 'page_view', { page_path: path })
}

export function trackEvent(action: string, category: string, label?: string) {
    window.gtag?.('event', action, {
        event_category: category,
        event_label: label,
    })
}
```

Eventos a rastrear:
- Visualização de card de desaparecido
- Registro de avistamento
- Cadastro de usuário
- Busca realizada

#### Tarefa 6.12 — Footer

- Links: FAQ, Termos, Política de Privacidade
- Créditos
- Links de redes sociais (se houver)

---

## Decisões Específicas desta Fase

### LGPD e Política de Privacidade

O projeto lida com **dados sensíveis**: fotos, nomes, localização de pessoas vulneráveis. A LGPD (Lei Geral de Proteção de Dados) exige:

1. **Consentimento explícito** para armazenar dados pessoais → checkbox no cadastro
2. **Direito de exclusão** → endpoint DELETE para remover conta e todos os dados associados
3. **Transparência** → página explicando quais dados são coletados e por quê
4. **Segurança** → já tratado nas fases anteriores (Firebase Auth, HTTPS, etc.)

A política de privacidade do legado é genérica e não menciona LGPD. Precisamos atualizar.

### Next.js no futuro?

Se o SEO se mostrar insuficiente com a abordagem de meta tags para bots, a migração para Next.js é viável:

- A API Go não muda nada
- O React components podem ser reutilizados
- Next.js adiciona SSR/SSG automaticamente
- É uma evolução, não uma reescrita

Por agora, a abordagem híbrida é suficiente e muito mais simples.

---

## Entregáveis da Fase 6

- [ ] Rota de meta tags para bots (Open Graph)
- [ ] Sitemap XML dinâmico
- [ ] robots.txt
- [ ] React: FAQ com accordion
- [ ] React: Termos de Uso / Política de Privacidade (atualizada LGPD)
- [ ] React: Páginas de erro (404, 403, 500)
- [ ] React: Meta tags em todas as páginas
- [ ] React: PWA (manifest, service worker)
- [ ] React: Responsividade mobile completa
- [ ] React: Acessibilidade WCAG AA
- [ ] React: Google Analytics
- [ ] React: Footer com links institucionais

---

## Próxima Fase

→ [FASE_08_DEPLOY.md](./FASE_08_DEPLOY.md) — Deploy no Cloud Run & Observabilidade

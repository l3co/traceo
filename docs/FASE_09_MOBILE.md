# Fase 9 — React Native (App Mobile)

> **Duração estimada**: 4+ semanas
> **Pré-requisito**: Fase 8 concluída (plataforma web em produção)

---

## Objetivo

Criar uma versão mobile da plataforma usando React Native, consumindo a **mesma API Go** que o frontend web. Ao final desta fase:

- App funcional para iOS e Android
- Funcionalidades principais: listagem, busca, detalhes, avistamento, login
- Push notifications para avistamentos
- Mapas nativos
- Disponível nas lojas (App Store / Google Play)

Esta fase é um projeto à parte — a API Go não muda. Todo o trabalho é mobile.

---

## Por que React Native e não Flutter / Native?

| Tecnologia | Prós | Contras | Veredicto |
|---|---|---|---|
| **React Native + Expo** | Mesmo ecossistema React, compartilha lógica com web, hot reload | Performance um pouco menor que nativo | ✅ |
| Flutter (Dart) | Performance excelente, UI consistente | Linguagem diferente (Dart), não compartilha com web | ❌ Para nosso caso |
| Native (Swift + Kotlin) | Performance máxima | Dois codebases, duas equipes, duas linguagens | ❌ Para dev solo |
| Kotlin Multiplatform | Compartilha lógica de negócio | Mais novo, ecossistema menor | ❌ Maturidade |

**React Native com Expo** é a escolha porque:

1. **Você já vai saber React** — da Fase 1-7, todo o frontend web é React + TypeScript. O conhecimento se transfere diretamente.
2. **Compartilhamento de código** — types, interfaces, serviços de API, validações podem ser compartilhados entre web e mobile.
3. **Expo** simplifica drasticamente o setup — build nativo sem Xcode/Android Studio (Expo Application Services).
4. **Uma pessoa, uma linguagem** — você é dev solo. Manter duas bases (Swift + Kotlin) é inviável.

---

## O que muda do React Web para React Native?

### Componentes

| Web (React) | Mobile (React Native) |
|---|---|
| `<div>` | `<View>` |
| `<p>`, `<span>` | `<Text>` |
| `<img>` | `<Image>` |
| `<input>` | `<TextInput>` |
| `<button>` | `<Pressable>` ou `<TouchableOpacity>` |
| `<a href>` | `<Link>` (Expo Router) |
| CSS / Tailwind | StyleSheet / NativeWind (Tailwind para RN) |
| React Router | Expo Router (file-based routing) |

### Navegação

Web usa URLs (`/missing/abc123`). Mobile usa **stacks e tabs**:

```
Tab Bar
├── Home (listagem de desaparecidos)
│   └── Detalhes do desaparecido (stack push)
│       └── Registrar avistamento (stack push)
├── Busca
├── Mapa
├── Perfil
│   ├── Editar perfil
│   ├── Alterar senha
│   └── Notificações
└── Mais
    ├── FAQ
    └── Termos
```

### Mapas

Web usa `react-map-gl` (Mapbox). Mobile usa `react-native-maps` (Google Maps / Apple Maps nativos):

```tsx
import MapView, { Marker } from 'react-native-maps'

function MissingMap({ location }: { location: GeoPoint }) {
    return (
        <MapView
            style={{ flex: 1 }}
            initialRegion={{
                latitude: location.lat,
                longitude: location.lng,
                latitudeDelta: 0.05,
                longitudeDelta: 0.05,
            }}
        >
            <Marker coordinate={{ latitude: location.lat, longitude: location.lng }} />
        </MapView>
    )
}
```

### Imagens

Web usa `<img src={url}>`. Mobile usa `<Image source={{ uri: url }}>` com cache automático.

Para upload de câmera (feature nova no mobile):

```tsx
import * as ImagePicker from 'expo-image-picker'

async function pickImage() {
    const result = await ImagePicker.launchCameraAsync({
        mediaTypes: ImagePicker.MediaTypeOptions.Images,
        allowsEditing: true,
        aspect: [4, 3],
        quality: 0.8,
    })

    if (!result.canceled) {
        return result.assets[0].uri
    }
}
```

---

## Push Notifications — a feature mais importante do mobile

No web, notificações de avistamento são por email. No mobile, podemos fazer **push notification** — alerta direto no celular do familiar.

### Fluxo

```
1. Familiar instala o app e faz login
2. App registra o device token no Firebase Cloud Messaging (FCM)
3. Alguém registra um avistamento na web ou no app
4. API Go envia push notification via FCM para o device do familiar
5. Familiar recebe notificação no celular, mesmo com app fechado
```

### Implementação no Go (API)

```go
// infrastructure/notification/push_sender.go

import "firebase.google.com/go/v4/messaging"

type PushSender struct {
    client *messaging.Client
}

func (p *PushSender) SendToDevice(ctx context.Context, deviceToken, title, body string) error {
    message := &messaging.Message{
        Notification: &messaging.Notification{
            Title: title,
            Body:  body,
        },
        Token: deviceToken,
    }

    _, err := p.client.Send(ctx, message)
    return err
}
```

### Implementação no React Native

```tsx
import * as Notifications from 'expo-notifications'

// Registrar para push notifications
async function registerForPush() {
    const { status } = await Notifications.requestPermissionsAsync()
    if (status !== 'granted') return

    const token = await Notifications.getExpoPushTokenAsync()
    // Enviar token para a API: POST /api/v1/users/:id/device-token
    await api.post(`/users/${userId}/device-token`, { token: token.data })
}

// Listener para notificações recebidas
Notifications.addNotificationReceivedListener(notification => {
    // Mostrar in-app notification
})

// Listener para quando o usuário toca na notificação
Notifications.addNotificationResponseReceivedListener(response => {
    const missingId = response.notification.request.content.data.missingId
    // Navegar para a tela de detalhes
    router.push(`/missing/${missingId}`)
})
```

---

## Código compartilhado entre Web e Mobile

### Estrutura do monorepo com mobile

```
desaparecidos/
├── api/              ← Go backend (não muda)
├── web/              ← React web
├── mobile/           ← React Native (Expo)
├── shared/           ← Código compartilhado (novo!)
│   ├── types/
│   │   ├── user.ts
│   │   ├── missing.ts
│   │   ├── sighting.ts
│   │   └── homeless.ts
│   ├── services/
│   │   ├── api-client.ts
│   │   ├── user-service.ts
│   │   ├── missing-service.ts
│   │   └── sighting-service.ts
│   ├── validators/
│   │   ├── user-validator.ts
│   │   └── missing-validator.ts
│   └── constants/
│       ├── physical-traits.ts   ← Gender, EyeColor labels PT-BR
│       └── routes.ts
└── docs/
```

### O que pode ser compartilhado

| Camada | Compartilhável? | Exemplo |
|---|---|---|
| **Types/Interfaces** | ✅ Sim | `Missing`, `User`, `Sighting` TypeScript types |
| **API client** | ✅ Sim | Chamadas ao backend (axios funciona em ambos) |
| **Validação** | ✅ Sim | Regras de validação de formulários |
| **Constantes** | ✅ Sim | Labels de características físicas, rotas da API |
| **Componentes UI** | ❌ Não | `<div>` vs `<View>` são diferentes |
| **Navegação** | ❌ Não | React Router vs Expo Router |
| **Mapas** | ❌ Não | react-map-gl vs react-native-maps |

### Configurar como package compartilhado

No `package.json` do monorepo raiz:

```json
{
    "workspaces": ["web", "mobile", "shared"]
}
```

No `shared/package.json`:
```json
{
    "name": "@desaparecidos/shared",
    "main": "index.ts"
}
```

Uso no web e mobile:
```tsx
import { Missing, User } from '@desaparecidos/shared/types'
import { missingService } from '@desaparecidos/shared/services'
import { GENDER_OPTIONS } from '@desaparecidos/shared/constants'
```

---

## Tarefas Detalhadas

### Setup

#### Tarefa 8.1 — Inicializar projeto Expo

```bash
npx create-expo-app@latest mobile --template blank-typescript
cd mobile
npx expo install expo-router expo-notifications expo-image-picker
npx expo install react-native-maps
```

#### Tarefa 8.2 — Configurar NativeWind (Tailwind para RN)

Para manter consistência visual com o web:
```bash
npm install nativewind tailwindcss
```

#### Tarefa 8.3 — Extrair código compartilhado

Mover types, services e constantes do `web/src/shared/` para `shared/`.

#### Tarefa 8.4 — Configurar Firebase no mobile

```bash
npx expo install @react-native-firebase/app @react-native-firebase/auth
```

### Telas

#### Tarefa 8.5 — Tab navigation

- Home (listagem)
- Busca
- Mapa
- Perfil

#### Tarefa 8.6 — Tela de Login/Cadastro

- Email + senha
- Link para recuperação de senha
- Upload de avatar (câmera ou galeria)

#### Tarefa 8.7 — Tela de Listagem (Home)

- FlatList com cards de desaparecidos
- Pull-to-refresh
- Infinite scroll (cursor-based)
- Skeleton loading

#### Tarefa 8.8 — Tela de Detalhes

- Foto fullscreen com zoom (pinch)
- Todas as características
- Mapa nativo com marker
- Botão "Informar avistamento"

#### Tarefa 8.9 — Tela de Busca

- Input com busca em tempo real
- Lista de resultados
- Histórico de buscas recentes (AsyncStorage)

#### Tarefa 8.10 — Tela de Mapa

- Mapa fullscreen com clusters
- Geolocalização do usuário (com permissão)
- Filtro por raio de distância

#### Tarefa 8.11 — Tela de Avistamento

- Mapa para selecionar localização
- Campo de observação
- Botão "Enviar"
- Usar localização atual como default

#### Tarefa 8.12 — Tela de Perfil

- Dados do usuário
- Editar perfil
- Alterar senha
- Ver notificações (avistamentos)
- Logout

#### Tarefa 8.13 — Push Notifications

- Registrar device token no login
- Enviar push quando avistamento é registrado
- Deep link: tocar na notificação abre a tela de detalhes

### Backend (adições)

#### Tarefa 8.14 — Endpoint para device token

`POST /api/v1/users/:id/device-token`
- Salva o token FCM associado ao usuário no Firestore

#### Tarefa 8.15 — Push notification no fluxo de avistamento

Quando um avistamento é registrado:
1. Busca o userId do desaparecido
2. Busca os device tokens do usuário
3. Envia push notification via FCM

### Publicação

#### Tarefa 8.16 — Build com EAS (Expo Application Services)

```bash
eas build --platform ios
eas build --platform android
```

#### Tarefa 8.17 — Publicar nas lojas

- **App Store**: Apple Developer Account ($99/ano), review process (~1-3 dias)
- **Google Play**: Google Developer Account ($25 one-time), review process (~1-2 dias)

Requer: screenshots, descrição, política de privacidade, ícone, splash screen.

---

## Funcionalidades exclusivas do mobile (futuro)

Features que fazem sentido **só no mobile**:

| Feature | Descrição | Complexidade |
|---|---|---|
| **Câmera para avistamento** | Tirar foto no momento do avistamento | Baixa |
| **Geolocalização automática** | Usar GPS do celular para marcar localização | Baixa |
| **Modo offline** | Consultar desaparecidos sem internet | Média |
| **Compartilhar card** | Compartilhar card de desaparecido via WhatsApp/Instagram | Baixa |
| **Reconhecimento facial** | Comparar foto tirada com banco de desaparecidos | Alta (ML) |
| **Alertas por proximidade** | Notificar se o usuário está perto de uma região de desaparecimento | Média |

Essas features são para **depois** de ter o MVP mobile funcionando.

---

## Entregáveis da Fase 8

- [ ] Projeto Expo inicializado e configurado
- [ ] Código compartilhado extraído (`shared/`)
- [ ] Tab navigation com 4 abas
- [ ] Login/Cadastro com Firebase Auth
- [ ] Listagem com infinite scroll
- [ ] Detalhes com mapa nativo
- [ ] Busca em tempo real
- [ ] Mapa fullscreen com clusters
- [ ] Formulário de avistamento com GPS
- [ ] Push notifications funcionando
- [ ] Endpoint de device token na API
- [ ] Build para iOS e Android
- [ ] Publicação nas lojas

---

## Fim do Roadmap

Ao completar todas as 8 fases, você terá:

- ✅ API Go em produção (Cloud Run)
- ✅ Frontend React em produção (Firebase Hosting)
- ✅ App mobile em produção (App Store + Google Play)
- ✅ Firestore como banco de dados
- ✅ Autenticação segura (Firebase Auth)
- ✅ Notificações (email + Telegram + push)
- ✅ CI/CD automatizado
- ✅ Observabilidade e monitoramento
- ✅ **Conhecimento sólido de Go**

E o mais importante: uma plataforma que ajuda famílias a encontrar pessoas queridas.

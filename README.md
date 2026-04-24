# Arquitetura de Microserviços — Documentação Técnica

## Visão Geral

Sistema distribuído composto por 4 microserviços independentes, comunicando-se via HTTP (síncrono) e RabbitMQ (assíncrono).

```
Cliente (Insomnia/App)
        │
        ▼
  ┌─────────────┐
  │ API Gateway │  :8080  ← único ponto de entrada
  └──────┬──────┘
         │  reverse proxy + JWT auth
    ┌────┴─────┐
    │          │
    ▼          ▼
┌────────┐  ┌─────────┐
│  User  │  │  Order  │
│Service │  │ Service │
│ :8081  │  │  :8082  │
└───┬────┘  └────┬────┘
    │             │
    └──────┬──────┘
           │ publica eventos
           ▼
      ┌──────────┐
      │ RabbitMQ │
      └────┬─────┘
           │ consome eventos
           ▼
  ┌──────────────────┐
  │ Notification Svc │  (sem HTTP, só consumer)
  └──────────────────┘

Todos os serviços com DB ──► SQL Server :1433
```

---

## Estrutura de Pastas

```
microservices/
│
├── go.work                        ← Go Workspace (une todos os módulos localmente)
├── docker-compose.yml             ← Infraestrutura completa
├── Makefile                       ← Comandos de build, migrate, logs
├── insomnia-collection.json       ← Coleção de testes
│
├── shared/                        ← Código compartilhado entre serviços
│   ├── config/config.go           ← Leitura de variáveis de ambiente
│   ├── database/sqlserver.go      ← Pool de conexões SQL Server
│   ├── messaging/rabbitmq.go      ← Cliente RabbitMQ (publish/consume)
│   ├── middleware/auth.go         ← Geração e validação de JWT
│   ├── logger/logger.go           ← Logger JSON estruturado (slog)
│   └── errors/errors.go           ← Erros tipados com HTTP status code
│
├── api-gateway/                   ← Porta :8080
│   ├── cmd/main.go                ← Entry point, roteamento, graceful shutdown
│   └── internal/
│       ├── middleware/auth.go     ← Validação JWT, extração de claims
│       └── proxy/proxy.go        ← Reverse proxy para os serviços
│
├── user-service/                  ← Porta :8081
│   ├── cmd/main.go                ← Entry point
│   ├── migrations/
│   │   └── 001_create_users.sql
│   └── internal/
│       ├── domain/user.go         ← Entidade User, regras de negócio
│       ├── repository/            ← Acesso ao SQL Server
│       ├── service/               ← Orquestração, lógica de negócio
│       ├── handler/               ← HTTP handlers (Gin)
│       └── events/publisher.go   ← Publica eventos no RabbitMQ
│
├── order-service/                 ← Porta :8082
│   ├── cmd/main.go
│   ├── migrations/
│   │   └── 001_create_orders.sql
│   └── internal/
│       ├── domain/order.go        ← Entidade Order + state machine de status
│       ├── repository/            ← Acesso ao SQL Server
│       ├── service/               ← Orquestração, validação de transições
│       ├── handler/               ← HTTP handlers (Gin)
│       └── events/publisher.go   ← Publica eventos no RabbitMQ
│
└── notification-service/          ← Sem HTTP (só worker)
    ├── cmd/main.go
    └── internal/
        └── consumer/consumer.go  ← Consome filas do RabbitMQ e envia notificações
```

---

## Fluxo HTTP — Requisição do Cliente

```
1. Cliente envia requisição para http://localhost:8080

2. API Gateway recebe e aplica middlewares:
   ├── CORS (libera origens)
   ├── Request Logger (loga método, path, status, duração)
   └── JWT Auth (rotas protegidas)
       ├── Extrai token do header: Authorization: Bearer <token>
       ├── Valida assinatura e expiração
       ├── Injeta headers internos: X-User-ID, X-User-Email, X-User-Role
       └── Rejeita com 401 se inválido

3. Gateway faz reverse proxy para o serviço correto:
   ├── /api/v1/auth/*    → user-service (sem auth)
   ├── /api/v1/users     POST → user-service (sem auth, registro)
   ├── /api/v1/users/*   → user-service (com auth)
   └── /api/v1/orders/*  → order-service (com auth)

4. Serviço processa e responde → Gateway repassa ao cliente
```

---

## Fluxo de Autenticação

```
POST /api/v1/auth/login
        │
        ▼
   user-service
        │
        ├── Busca usuário por email no SQL Server
        ├── Verifica senha com bcrypt.CompareHashAndPassword
        ├── Gera JWT com claims: { user_id, email, role, exp }
        └── Retorna { token, user }

Próximas requisições:
        │
        ▼
   Authorization: Bearer <jwt>
        │
   API Gateway valida o token
        │
   Injeta X-User-ID no header interno
        │
   Serviço downstream pode ler o user_id sem re-validar
```

---

## Fluxo RabbitMQ — Publish / Subscribe

### Topologia de Exchanges e Filas

```
                     EXCHANGES (topic)
    ┌─────────────────────────────────────────────┐
    │              users.events                    │
    │                                             │
    │  routing keys:                              │
    │  ● user.created                             │
    │  ● user.updated                             │
    │  ● user.deleted                             │
    └───────────────────┬─────────────────────────┘
                        │ bind: user.*
                        ▼
            ┌───────────────────────┐
            │ notifications.        │
            │ user.events  (queue)  │
            └───────────┬───────────┘
                        │ consume
                        ▼
              notification-service


    ┌─────────────────────────────────────────────┐
    │              orders.events                   │
    │                                             │
    │  routing keys:                              │
    │  ● order.created                            │
    │  ● order.status_changed                     │
    │  ● order.cancelled                          │
    └───────────────────┬─────────────────────────┘
                        │ bind: order.*
                        ▼
            ┌───────────────────────┐
            │ notifications.        │
            │ order.events (queue)  │
            └───────────┬───────────┘
                        │ consume
                        ▼
              notification-service
```

### Fluxo Completo — Criação de Usuário

```
Cliente
  │
  │  POST /api/v1/users
  ▼
API Gateway ──► user-service
                    │
                    ├── 1. Valida input (nome, email, senha)
                    ├── 2. Verifica email duplicado no SQL Server
                    ├── 3. Gera hash da senha (bcrypt)
                    ├── 4. Salva usuário no SQL Server
                    ├── 5. Retorna 201 Created ao cliente  ← resposta síncrona
                    │
                    └── 6. go func() → publica no RabbitMQ (assíncrono)
                                │
                                │  Exchange: users.events
                                │  Routing Key: user.created
                                │  Payload: { id, name, email, role }
                                ▼
                          RabbitMQ
                                │
                                │  Queue: notifications.user.events
                                ▼
                    notification-service
                                │
                                ├── Deserializa a mensagem
                                ├── Identifica tipo: user.created
                                ├── Envia e-mail de boas-vindas (TODO: integrar SendGrid/SES)
                                └── ACK → mensagem removida da fila
```

### Fluxo Completo — Atualização de Status do Pedido

```
Cliente
  │
  │  PATCH /api/v1/orders/{id}/status  { "status": "confirmed" }
  ▼
API Gateway ──► order-service
                    │
                    ├── 1. Busca pedido no SQL Server
                    ├── 2. Valida transição de status (state machine)
                    │       pending → confirmed ✔
                    │       delivered → confirmed ✖ (422)
                    ├── 3. Atualiza status no SQL Server
                    ├── 4. Retorna 200 OK ao cliente  ← resposta síncrona
                    │
                    └── 5. go func() → publica no RabbitMQ (assíncrono)
                                │
                                │  Exchange: orders.events
                                │  Routing Key: order.status_changed
                                │  Payload: { order_id, user_id, old_status, new_status }
                                ▼
                          RabbitMQ
                                │
                                │  Queue: notifications.order.events
                                ▼
                    notification-service
                                │
                                ├── Identifica tipo: order.status_changed
                                ├── Envia notificação push / e-mail ao usuário
                                └── ACK
```

### Garantias de Entrega

| Mecanismo | Configuração | Efeito |
|-----------|-------------|--------|
| Exchange durável | `durable: true` | Sobrevive a restart do RabbitMQ |
| Fila durável | `durable: true` | Mensagens não perdidas em restart |
| Mensagem persistente | `DeliveryMode: Persistent` | Salva em disco |
| Publisher Confirms | `ch.Confirm(false)` | Garante que o broker recebeu |
| Nack + Requeue | `d.Nack(false, true)` | Reprocessa em caso de erro no consumer |
| Ack manual | `d.Ack(false)` | Só remove após processamento bem-sucedido |

---

## State Machine — Status do Pedido

```
                ┌──────────┐
    criado ────►│ pending  │────────────────────────────┐
                └────┬─────┘                            │
                     │ confirmed                        │
                     ▼                                  │
                ┌──────────┐                            │
                │confirmed │────────────────────────────┤
                └────┬─────┘                            │
                     │ processing                       │ cancelled
                     ▼                                  │
                ┌──────────┐                            │
                │processing│────────────────────────────┤
                └────┬─────┘                            │
                     │ shipped                          │
                     ▼                                  ▼
                ┌──────────┐                    ┌────────────┐
                │ shipped  │                    │ cancelled  │ (estado final)
                └────┬─────┘                    └────────────┘
                     │ delivered
                     ▼
                ┌──────────┐
                │delivered │ (estado final)
                └──────────┘
```

---

## Variáveis de Ambiente

| Variável | Padrão | Serviço |
|----------|--------|---------|
| `DB_HOST` | localhost | user, order |
| `DB_PORT` | 1433 | user, order |
| `DB_USER` | sa | user, order |
| `DB_PASSWORD` | — | user, order |
| `DB_NAME` | microservices | user, order |
| `RABBITMQ_HOST` | localhost | user, order, notification |
| `RABBITMQ_PORT` | 5672 | user, order, notification |
| `RABBITMQ_USER` | admin | user, order, notification |
| `RABBITMQ_PASSWORD` | — | user, order, notification |
| `JWT_SECRET` | — | api-gateway, user |
| `JWT_EXPIRATION_HOURS` | 24 | user |
| `USER_SERVICE_URL` | http://user-service:8081 | api-gateway |
| `ORDER_SERVICE_URL` | http://order-service:8082 | api-gateway |
| `LOG_LEVEL` | info | todos |

---

## Tecnologias

| Tecnologia | Uso |
|------------|-----|
| Go 1.22 | Linguagem principal |
| Gin | Framework HTTP (user-service, order-service) |
| net/http | HTTP server (api-gateway) |
| go-mssqldb | Driver SQL Server |
| amqp091-go | Cliente RabbitMQ |
| golang-jwt | Geração e validação JWT |
| bcrypt | Hash de senhas |
| slog | Logging estruturado JSON |
| uuid | Geração de IDs únicos |
| Docker + Compose | Containerização |
| SQL Server 2022 | Banco de dados |
| RabbitMQ 3.13 | Message broker |

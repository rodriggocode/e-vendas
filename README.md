# Loja

API de loja escrita em Go, usando `chi` como router HTTP.

> **Status:** esqueleto inicial. Neste momento existe apenas um servidor HTTP
> com uma rota de exemplo; as camadas de domínio, aplicação e persistência
> ainda não foram implementadas.

## Documentação

- [Arquitetura](docs/ARQUITETURA.md) — diagnóstico, arquitetura alvo, camadas,
  contratos, decisões e roadmap.
- [Banco de dados](docs/BANCO-DE-DADOS.md) — diagrama ER, tabelas,
  relacionamentos e índices.
- [Diagrama ER em imagem](docs/diagrama-banco.png) — versão PNG do diagrama.
- [Migrations](docs/MIGRATIONS.md) — comandos do golang-migrate, criação,
  aplicação e reversão de migrations.

## Stack

- Go 1.27
- `github.com/go-chi/chi/v5`

## Como executar o esqueleto atual

```bash
go run ./cmd/api
```

O servidor sobe em `http://localhost:3000/`.

## Banco de dados local (PostgreSQL)

O arquivo `compose.yaml` sobe um PostgreSQL para desenvolvimento usando Podman
rootless (sem `sudo`):

```bash
podman compose up -d
podman compose ps
```

Conexão padrão:

```
postgres://loja:loja@localhost:5433/loja?sslmode=disable
```

As credenciais e a porta estão definidas diretamente no `compose.yaml`:
banco `loja`, usuário `loja`, senha `loja` e porta `5433` no host
(`5433` → `5432` dentro do container).

Para parar o banco:

```bash
podman compose down
```

Para parar e apagar os dados do volume:

```bash
podman compose down -v
```

> Configuração simples para desenvolvimento local. Não use essas credenciais em
> produção.

## Estrutura de pastas

```
.
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   ├── domain/
│   ├── usecase/
│   ├── repository/
│   │   ├── memory/
│   │   └── postgres/
│   ├── handler/
│   ├── middleware/
│   └── router/
├── migrations/
├── docs/
│   └── ARQUITETURA.md
├── .env
├── compose.yaml
├── go.mod
├── go.sum
└── README.md
```

### O que vai em cada pasta

| Pasta | Para que serve | O que deve conter |
|---|---|---|
| `cmd/api/` | Ponto de entrada da aplicação e **composition root**. É o único lugar que monta as dependências concretas. | `main.go` com carregamento de config, criação do repositório, casos de uso, handlers, router, `http.Server` com timeouts e shutdown gracioso. Sem regra de negócio. |
| `internal/config/` | Centralizar a configuração da aplicação. | Leitura e validação de variáveis de ambiente, valores padrão, timeouts, portas e origens de CORS. Deve falhar rápido se algo estiver inválido. Sem regra de negócio. |
| `internal/domain/` | Coração do sistema: modelo de negócio independente de tecnologia. | Entidades (`Product`, `Order`, `Customer`), value objects, invariantes, erros de domínio e **interfaces de repositório**. Não importa `chi`, `net/http`, SQL ou JSON. |
| `internal/usecase/` | Regras de aplicação e orquestração dos fluxos. | Casos de uso como criar produto, listar produtos, criar pedido e baixar estoque. Coordena transações e depende apenas das interfaces do domínio. Não conhece HTTP nem banco. |
| `internal/repository/` | Implementações concretas de persistência. | Adaptadores que implementam as interfaces do domínio e traduzem erros do banco para erros de domínio. |
| `internal/repository/memory/` | Persistência em memória para desenvolvimento e testes. | Implementação rápida, sem infraestrutura, usada para rodar e testar a aplicação localmente. |
| `internal/repository/postgres/` | Persistência real em PostgreSQL. | Queries SQL, uso do pool `pgx`, mapeamento entre linhas e entidades e tradução de erros (`no rows`, `unique violation`). |
| `internal/handler/` | Camada de entrega HTTP. | Handlers, DTOs de request/response, decodificação e validação de formato, paginação e tradução de erros de domínio para status HTTP. Não acessa banco nem contém regra de negócio. |
| `internal/middleware/` | Comportamentos transversais a todas as requisições. | Request ID, log estruturado, recover de panic, CORS, timeout, autenticação, autorização e rate limiting. |
| `internal/router/` | Montagem das rotas e da cadeia de middleware. | Registro dos endpoints versionados (`/api/v1`), health checks, grupos de rotas e ordem dos middlewares. Sem regra de negócio. |
| `migrations/` | Versionar a evolução do banco de dados. | Arquivos SQL de migração `up` e `down` (ex.: criação da tabela `products`), aplicados por ferramenta como `golang-migrate`. |
| `docs/` | Documentação técnica do projeto. | Arquitetura, decisões (ADRs), contratos da API e guias operacionais. |

### Regra de dependência

```
cmd/api ──> router ──> handler ──> usecase ──> domain
               │                      ▲
               └──> middleware        │
repository ────────────────────────────┘
```

As dependências apontam para dentro: `domain` não importa ninguém do projeto;
`usecase` depende de `domain`; `repository` implementa interfaces de `domain`;
`handler` depende de `usecase`; e `cmd/api` conhece todas as camadas para fazer
o wiring.

As pastas já existem, mas ainda estão vazias — o único código atual é
`cmd/api/main.go`. Os detalhes de cada camada, contratos e roadmap estão em
[`docs/ARQUITETURA.md`](docs/ARQUITETURA.md).

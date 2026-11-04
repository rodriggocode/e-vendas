# Arquitetura — Loja

Documento de referência da arquitetura do projeto. Ele descreve o estado atual,
o diagnóstico, a arquitetura alvo, as convenções e o que falta para evoluir.

**Status:** esqueleto inicial — servidor HTTP com rota de exemplo, sem camadas
de domínio, aplicação ou persistência implementadas.

---

## 1. Resumo executivo

A aplicação será uma **API REST de loja**, inicialmente um **monólito modular**
em Go. A arquitetura adotada é em camadas, com dependências apontando para
dentro (em direção ao domínio), no estilo ports & adapters simplificado:

- O **domínio** define entidades, invariantes e interfaces de repositório.
- Os **casos de uso** orquestram regras de aplicação.
- Os **handlers HTTP** traduzem requisições e respostas.
- Os **repositórios** implementam as interfaces do domínio.
- O **composition root** (`cmd/api`) é o único lugar que conhece todas as
  implementações concretas.

Microserviços não são necessários neste estágio. O monólito modular permite
extrair um contexto (catálogo, pedidos, pagamentos) no futuro sem reescrever as
regras de negócio, desde que as fronteiras entre camadas e contextos sejam
respeitadas desde o início.

O **frontend será uma aplicação separada em React + Vite + TypeScript**, criada
em `Www/web` (fora da pasta `loja`), consumindo a API Go via HTTP. A API
continua sendo a fonte da verdade das regras de negócio; o React cuida apenas de
interface, navegação e estado de tela. Next.js fica como opção futura caso SEO
de loja pública passe a ser prioridade.

---

## 2. Estado atual

| Elemento | Situação |
|---|---|
| `cmd/api/main.go` | Servidor HTTP funcional com uma rota `/` que responde texto |
| `internal/config/` | Pasta vazia, não participa do build |
| `internal/domain/` | Pasta vazia, não participa do build |
| `internal/usecase/` | Pasta vazia, não participa do build |
| `internal/repository/memory/` | Pasta vazia, não participa do build |
| `internal/repository/postgres/` | Pasta vazia, não participa do build |
| `internal/handler/` | Pasta vazia, não participa do build |
| `internal/middleware/` | Pasta vazia, não participa do build |
| `internal/router/` | Pasta vazia, não participa do build |
| `migrations/` | Pasta vazia, sem migrações |
| `.env` | Vazio e não carregado pela aplicação |
| `go.mod` / `go.sum` | `chi` v5 declarado |
| Testes | Nenhum |
| Git / CI | Projeto não é um repositório Git; sem pipeline |
| README / docs | Criados junto com este documento |

Conclusão: **a estrutura de pastas da arquitetura alvo já existe, mas ainda não
há código nas camadas**; o único código é o servidor HTTP em `cmd/api`. Este
documento define o comportamento esperado de cada camada antes de o código
crescer.

---

## 3. Diagnóstico do esqueleto atual

### Pontos positivos

- `chi` é uma escolha adequada: leve, compatível com `net/http`, com middleware
  padrão e subrouters.
- A intenção de separar `config`, `domain`, `handler` e `router` está correta.
- O projeto compila e passa no `go vet`.

### Problemas a corrigir antes de adicionar funcionalidades

| Local | Problema | Impacto |
|---|---|---|
| Rota `/` | Declara `Content-Type: application/json` mas responde texto puro | Contrato HTTP incorreto; clientes podem falhar ao fazer parse |
| `main.go` | Porta, timeouts e rotas hardcoded | Difícil configurar por ambiente e testar |
| `main.go` | `log.Fatal` no `ListenAndServe` | Encerra sem shutdown gracioso; requisições em andamento são cortadas |
| `main.go` | Sem timeouts de leitura/escrita/idle | Vulnerável a conexões lentas (slowloris) e recursos presos |
| `main.go` | Rota inline, sem composition root | Acoplamento e ausência de injeção de dependência |
| `go.mod` | `chi` marcado como `// indirect` mesmo sendo importado direto | Metadado incorreto; `go mod tidy` corrige |
| `main.go` | Arquivo fora do `gofmt` | Inconsistência de estilo e ruído em diffs |
| `.env` | Vazio, sem `.gitignore` e sem `.env.example` | Risco de commitar segredos; configuração não documentada |
| Raiz | Sem `.gitignore`, Makefile, Dockerfile, CI ou testes | Falta de padronização e de verificação automatizada |

---

## 4. Arquitetura alvo

### 4.1 Princípios

1. **Dependências apontam para dentro.** Domínio não importa HTTP, banco ou
   framework. Casos de uso dependem de interfaces do domínio.
2. **Domínio puro.** Entidades e regras de negócio não conhecem `chi`,
   `net/http`, SQL nem JSON.
3. **Composition root único.** A montagem de dependências acontece em
   `cmd/api/main.go`; nenhuma outra camada instancia implementações concretas.
4. **Contratos explícitos.** DTOs de entrada/saída são separados das entidades;
   erros de domínio são traduzidos na borda HTTP.
5. **Configuração por ambiente (12-factor).** Nada de segredo ou valor de
   ambiente hardcoded; falhar rápido se a configuração for inválida.
6. **Monólito modular.** Um deploy, mas com fronteiras claras por camada e,
   quando necessário, por contexto de negócio.
7. **Observabilidade desde o início.** Logs estruturados, request ID, health
   checks e métricas antes de o sistema crescer.

### 4.2 Estrutura de pastas alvo

```
loja/
├── cmd/
│   └── api/
│       └── main.go              # composition root: monta config, repo, usecase, router
├── internal/
│   ├── config/
│   │   └── config.go            # leitura e validação de variáveis de ambiente
│   ├── domain/
│   │   ├── product.go           # entidade + invariantes
│   │   ├── order.go             # entidade + máquina de estados
│   │   ├── errors.go            # erros de domínio
│   │   └── repository.go        # interfaces de persistência
│   ├── usecase/
│   │   ├── product.go           # regras de aplicação de produto
│   │   └── order.go             # criação de pedido, baixa de estoque
│   ├── repository/
│   │   ├── memory/              # implementação em memória (dev/testes)
│   │   └── postgres/            # implementação real
│   ├── handler/
│   │   ├── product.go           # handlers HTTP de produto
│   │   ├── order.go             # handlers HTTP de pedido
│   │   ├── health.go            # /healthz e /readyz
│   │   ├── dto.go               # request/response
│   │   └── response.go          # helpers de JSON e erro
│   ├── middleware/
│   │   ├── request_id.go
│   │   ├── logger.go
│   │   ├── recoverer.go
│   │   └── cors.go
│   └── router/
│       └── router.go            # registro de rotas e middleware
├── migrations/                  # SQL versionado
├── docs/
│   └── ARQUITETURA.md
├── .env.example
├── .gitignore
├── Makefile
├── Dockerfile
├── README.md
├── go.mod
└── go.sum
```

**Por que `internal/`:** impede que outros módulos importem detalhes internos,
deixa explícito o que é API pública e força a disciplina de fronteiras. As
pastas de camadas já foram criadas dentro de `internal/`; resta implementar o
conteúdo de cada uma.

O frontend **não** fica dentro de `loja/`: ele mora em `Www/web`, como projeto
irmão. Isso mantém o backend Go independente de Node/npm e permite deploy
separado. Detalhes na seção 4.7.

### 4.3 Responsabilidades por camada

| Camada | Responsabilidade | Pode importar |
|---|---|---|
| `cmd/api` | Ler config, montar dependências, iniciar e encerrar o servidor | Todas |
| `internal/config` | Ler e validar env; defaults seguros | Biblioteca padrão |
| `internal/domain` | Entidades, invariantes, erros e interfaces de repositório | Biblioteca padrão |
| `internal/usecase` | Orquestrar regras de aplicação; transações; autorização de negócio | `domain` |
| `internal/repository` | Persistência; traduzir erros do banco para erros de domínio | `domain` |
| `internal/handler` | Decodificar request, validar formato, chamar usecase, montar response | `usecase`, `domain` (erros) |
| `internal/middleware` | Request ID, log, recover, CORS, timeout, auth | Biblioteca padrão e router |
| `internal/router` | Compor rotas e middleware; nada de regra de negócio | `handler`, `middleware`, `config` |
| `migrations` | Esquema do banco versionado | — |

### 4.4 Regra de dependência

```
                    cmd/api
                       │
         ┌─────────────┼──────────────┐
         ▼             ▼              ▼
     config        router         repository
                       │              │
                       ▼              │
                    handler           │
                       │              │
                       ▼              │
                    usecase ◄─────────┘
                       │
                       ▼
                    domain
```

- `domain` não importa ninguém do projeto.
- `usecase` importa `domain`.
- `repository` importa `domain` e implementa suas interfaces.
- `handler` importa `usecase` e os erros de `domain`.
- `router` importa `handler`, `middleware` e `config`.
- `cmd/api` importa tudo e faz o wiring.

**Proibido:** handler acessar banco direto, domínio importar `chi`/SQL, usecase
depender de `*http.Request` ou de detalhes de JSON.

### 4.5 Fluxo de uma requisição

1. O servidor recebe a requisição e aplica a cadeia de middleware
   (request ID → real IP → CORS → log → recover → timeout).
2. O router seleciona o handler pelo método e caminho.
3. O handler decodifica o JSON, valida o formato e converte para um input de
   caso de uso.
4. O caso de uso aplica as regras de negócio, chama o repositório e devolve
   entidade ou erro de domínio.
5. O handler converte o resultado em DTO e escreve a resposta JSON.
6. O middleware de log registra método, rota, status, duração e request ID.

### 4.6 Composition root

`cmd/api/main.go` é o único lugar que:

- carrega a configuração;
- escolhe a implementação de repositório (memória em dev, Postgres em produção);
- instancia casos de uso e handlers;
- injeta dependências via construtor;
- cria o `http.Server` com timeouts;
- trata `SIGINT`/`SIGTERM` e faz shutdown gracioso.

### 4.7 Frontend (React + Vite)

O frontend será uma aplicação **separada** em React + Vite + TypeScript, localizada
em `Www/web`, como pasta irmã de `loja`:

```
Www/
├── loja/          # API Go
└── web/           # frontend React (por enquanto apenas a pasta, vazia)
```

**Por que separado:**

- deploy independente do backend;
- evita misturar `node_modules`, build de frontend e código Go;
- mantém a fronteira clara: o front consome a API, não importa código interno;
- permite trocar ou reescrever o front sem tocar no domínio.

**Stack prevista:**

- React + Vite + TypeScript;
- React Router para navegação;
- TanStack Query para dados da API (cache, loading e erro);
- React Hook Form + Zod para formulários e validação;
- Tailwind CSS (opcional) para estilo;
- Vitest + Testing Library para testes; Playwright depois, se necessário.

**Responsabilidades:**

- renderizar interface, navegação, formulários e estados de loading/erro;
- consumir os endpoints `/api/v1`;
- **não** conter regra de negócio, cálculo de preço, estoque ou pedido;
- **não** acessar banco de dados diretamente.

**Integração com a API:**

- em desenvolvimento, usar o proxy do Vite para encaminhar `/api` para
  `http://localhost:3000`, evitando CORS;
- em produção, servir front e API no mesmo domínio com proxy reverso
  (`/api` → API Go);
- documentar a API com OpenAPI e gerar os tipos do front a partir dela;
- autenticação preferencialmente com cookie `HttpOnly`; CORS explícito apenas
  se front e API ficarem em domínios diferentes.

**Estrutura futura (ainda não criada):**

```
web/
├── src/
│   ├── pages/
│   ├── components/
│   ├── services/       # chamadas à API
│   ├── hooks/
│   └── types/
├── package.json
└── vite.config.ts
```

Neste momento existe apenas a pasta `Www/web`, vazia, sem projeto Vite e sem
estrutura base.

---

## 5. Modelo de domínio inicial

Não modelar tudo de uma vez. A ordem sugerida é **Produto → Pedido →
Cliente/Autenticação → Pagamento**.

### 5.1 Produto

| Campo | Regra |
|---|---|
| `ID` | Identificador opaco, gerado na criação |
| `Name` | Obrigatório, até 120 caracteres, sem espaços nas pontas |
| `Description` | Opcional, até 2000 caracteres |
| `PriceCents` | Inteiro em centavos, nunca negativo |
| `Stock` | Inteiro, nunca negativo |
| `CreatedAt` / `UpdatedAt` | UTC, controlados pela aplicação |

### 5.2 Pedido

- Um pedido pertence a um cliente.
- Possui itens com produto, quantidade e preço no momento da compra (snapshot).
- Status inicial: `pending`. Transições previstas: `pending → paid → shipped →
  delivered`, com `cancelled` permitido antes do envio.
- A criação do pedido e a baixa de estoque devem ocorrer na **mesma transação**.
- O total é calculado no servidor; nunca confiar no valor enviado pelo cliente.

### 5.3 Convenções de domínio

- **Dinheiro em centavos (`int64`)**, nunca `float`.
- **Datas em UTC** (`timestamptz` no banco).
- **IDs opacos**: o cliente não deve depender do formato. Se for usar UUID,
  preferir UUID v7 (ordenável por tempo).
- **Erros de domínio** (`not found`, `conflict`, `validation`) são traduzidos
  para HTTP apenas na camada de handler.

---

## 6. Contratos HTTP

### 6.1 Versionamento e recursos

- Base: `/api/v1`.
- Recursos iniciais: `/products`, `/orders`, `/customers`.
- Health checks fora do versionamento: `/healthz` e `/readyz`.

### 6.2 Endpoints previstos para produto

| Método | Caminho | Descrição | Sucesso |
|---|---|---|---|
| `POST` | `/api/v1/products` | Cria produto | `201` |
| `GET` | `/api/v1/products` | Lista com paginação | `200` |
| `GET` | `/api/v1/products/{id}` | Busca por ID | `200` |
| `PUT` | `/api/v1/products/{id}` | Atualiza produto | `200` |
| `DELETE` | `/api/v1/products/{id}` | Remove produto | `204` |

### 6.3 Status codes

| Situação | Status |
|---|---|
| JSON malformado ou campo desconhecido | `400` |
| Não autenticado | `401` |
| Sem permissão | `403` |
| Recurso inexistente | `404` |
| Conflito (ex.: SKU duplicado) | `409` |
| Regra de negócio inválida | `422` |
| Erro inesperado | `500` |
| Timeout interno | `504` |

### 6.4 Formato de erro

Todas as respostas de erro devem seguir o mesmo envelope:

```json
{
  "error": {
    "code": "validation_error",
    "message": "nome é obrigatório"
  }
}
```

Códigos previstos: `invalid_json`, `validation_error`, `not_found`, `conflict`,
`unauthorized`, `forbidden`, `internal_error`.

### 6.5 Paginação

- Query params: `page` (padrão `1`) e `per_page` (padrão `20`, máximo `100`).
- Resposta: `items`, `total`, `page`, `per_page`.
- Ordenação padrão: mais recentes primeiro.

### 6.6 Boas práticas de borda

- Limitar o tamanho do corpo da requisição.
- Rejeitar campos desconhecidos.
- Separar DTO de entidade: o cliente não deve conseguir alterar `ID`,
  `CreatedAt` ou `UpdatedAt`.
- Validação de formato na borda; validação de negócio no domínio/caso de uso.
- Para endpoints de pagamento, prever chave de idempotência.

---

## 7. Configuração e ambientes

### 7.1 Variáveis previstas

| Variável | Padrão | Uso |
|---|---|---|
| `APP_ENV` | `development` | Ambiente lógico |
| `PORT` | `3000` | Porta HTTP |
| `READ_TIMEOUT` | `10s` | Leitura da requisição |
| `WRITE_TIMEOUT` | `15s` | Escrita da resposta |
| `IDLE_TIMEOUT` | `60s` | Conexões ociosas |
| `SHUTDOWN_TIMEOUT` | `10s` | Tempo máximo de shutdown gracioso |
| `REQUEST_TIMEOUT` | `30s` | Timeout por requisição |
| `CORS_ALLOWED_ORIGINS` | `*` em dev | Origens permitidas, separadas por vírgula |
| `DATABASE_URL` | — | Conexão Postgres (quando existir) |
| `LOG_LEVEL` | `info` | Nível de log |

### 7.2 Regras

- `.env` é **apenas para desenvolvimento** e deve estar no `.gitignore`.
- `.env.example` é versionado e documenta todas as variáveis.
- Em produção, usar variáveis do ambiente ou secret manager; nunca commitar
  segredos.
- A aplicação deve **falhar na inicialização** se uma variável obrigatória
  estiver ausente ou inválida.
- Timeouts e portas devem ser configuráveis, não hardcoded.

---

## 8. Persistência

### 8.1 Escolha

- **PostgreSQL** como banco principal.
- Driver `pgx` com pool de conexões.
- SQL explícito, sem ORM no início: menos mágica, mais controle sobre índices e
  transações.
- Migrações versionadas em `migrations/` (up e down), aplicadas por ferramenta
  dedicada (ex.: `golang-migrate`).

### 8.2 Regras

- O domínio define `ProductRepository`; a implementação Postgres fica em
  `internal/repository/postgres`.
- O repositório traduz erros do banco (`no rows`, violação de unique) para
  `domain.ErrNotFound` e `domain.ErrConflict`.
- Transações que envolvem mais de um agregado (pedido + estoque) são
  controladas pelo caso de uso, não pelo handler.
- Constraints no banco são a última linha de defesa: `NOT NULL`, `CHECK`,
  `UNIQUE`, chaves estrangeiras.
- Em desenvolvimento e testes, usar implementação em memória para feedback
  rápido; testes de integração rodam contra Postgres real (ex.: testcontainers).

---

## 9. Segurança

Itens a implementar conforme o escopo crescer:

- **Autenticação:** JWT de acesso curto + refresh token, ou sessão com cookie
  `HttpOnly`; senha com `argon2id` ou `bcrypt`.
- **Autorização:** papéis `customer` e `admin`; middleware de autorização por
  rota.
- **Rate limiting:** por IP e por usuário nas rotas públicas e de login.
- **CORS:** lista explícita de origens em produção; nunca `*` com credenciais.
- **Limite de corpo** e validação de `Content-Type`.
- **Security headers** (`X-Content-Type-Options`, `X-Frame-Options`, HSTS).
- **TLS** terminado no proxy/load balancer.
- **Segredos** fora do repositório, com rotação.
- **Logs sem dados sensíveis** (senha, token, cartão).

---

## 10. Observabilidade e resiliência

- **Logs estruturados** em JSON com `log/slog`, incluindo request ID, método,
  rota, status e duração.
- **Request ID** aceito do cliente ou gerado no servidor e devolvido no header.
- **Health checks:** `/healthz` (processo vivo) e `/readyz` (dependências
  prontas, incluindo banco).
- **Métricas:** latência, taxa de erro, throughput e saturação; expor em
  `/metrics` quando houver Prometheus/OpenTelemetry.
- **Tracing distribuído** apenas se houver múltiplos serviços.
- **Timeouts** em servidor, requisição, banco e chamadas externas.
- **Graceful shutdown** com `SIGINT`/`SIGTERM`.
- **Panic recovery** com log e resposta `500` padronizada.
- **Retries** somente em operações idempotentes, com backoff e limite.

---

## 11. Testes

| Nível | O que testar | Ferramenta |
|---|---|---|
| Domínio | Invariantes e transições de estado | `testing` |
| Caso de uso | Regras de aplicação com repositório fake/em memória | `testing` |
| Handler | Status, contrato JSON e erros | `net/http/httptest` |
| Repositório | SQL, constraints e tradução de erros | Postgres real/testcontainers |
| Integração | Fluxo completo de um recurso | `httptest` + banco de teste |

Regras:

- Testes de caso de uso não devem precisar de banco.
- Usar `-race` no CI.
- Testes de tabela para validação e casos de erro.
- Cobertura é indicador, não meta absoluta; priorizar caminhos de negócio.

---

## 12. Ferramentas e qualidade

- `gofmt` em todo o código.
- `go vet ./...`.
- `golangci-lint` com configuração versionada.
- `Makefile` com alvos: `run`, `build`, `test`, `fmt`, `vet`, `lint`, `tidy`,
  `docker-build`.
- `Dockerfile` multi-stage, imagem final mínima, execução como usuário não-root.
- `.gitignore` incluindo `.env`, binários, `coverage.out` e artefatos locais.
- CI com etapas: format check → vet → lint → test → build.
- ADRs curtos em `docs/adr/` para decisões relevantes.

---

## 13. Roadmap priorizado

### P0 — Fundação (antes de qualquer feature)

1. Inicializar repositório Git com `.gitignore` e `.env.example`.
2. Criar `cmd/api/main.go` como composition root.
3. Implementar `internal/config` com validação e fail-fast.
4. Configurar `http.Server` com timeouts e shutdown gracioso.
5. Criar `internal/router` e middleware: request ID, log, recover, CORS,
   timeout.
6. Adicionar `/healthz` e `/readyz`.
7. Corrigir `go.mod` com `go mod tidy` e formatar com `gofmt`.

### P1 — Primeiro recurso vertical

1. Modelar `Product` no domínio com invariantes.
2. Definir `ProductRepository` e implementação em memória.
3. Implementar `ProductUseCase`.
4. Implementar handlers HTTP de produto com contrato de erro e paginação.
5. Escrever testes de domínio, caso de uso e handler.

### P2 — Persistência real

1. Adicionar Postgres e `pgx`.
2. Criar migrações de produto.
3. Implementar `repository/postgres`.
4. Testes de integração com banco real.
5. Implementar `/readyz` verificando o banco.

### P2.5 — Frontend inicial (React + Vite)

1. Criar o projeto Vite + React + TypeScript em `Www/web`.
2. Configurar proxy do Vite para `/api` → `http://localhost:3000`.
3. Implementar listagem, criação e edição de produtos.
4. Integrar TanStack Query para cache, loading e erro.
5. Usar React Hook Form + Zod nos formulários.
6. Deixar autenticação e pedidos para depois da API estar pronta.

### P3 — Pedidos e autenticação

1. Modelar `Customer`, `Order`, `OrderItem` e máquina de estados.
2. Criar pedido com transação e baixa de estoque.
3. Autenticação e autorização.
4. Rate limiting nas rotas sensíveis.

### P4 — Produção

1. Métricas, tracing e alertas.
2. CI/CD e Dockerfile.
3. Revisão de segurança e gestão de segredos.
4. Documentação de operação e runbooks.

---

## 14. Decisões arquiteturais (ADRs)

| ID | Decisão | Motivo |
|---|---|---|
| ADR-001 | Monólito modular em vez de microserviços | Simplicidade operacional; fronteiras permitem extração futura |
| ADR-002 | Camadas com dependência para o domínio | Testabilidade e independência de framework/banco |
| ADR-003 | Dinheiro em centavos (`int64`) | Evita erro de ponto flutuante |
| ADR-004 | PostgreSQL + SQL explícito, sem ORM | Controle de queries, índices e transações |
| ADR-005 | API versionada em `/api/v1` | Evolução sem quebrar clientes |
| ADR-006 | Erros de domínio traduzidos na borda | Baixo acoplamento entre negócio e HTTP |
| ADR-007 | Configuração por ambiente com fail-fast | Evita subir com valores inválidos |
| ADR-008 | Repositório em memória para dev/testes | Feedback rápido sem infraestrutura |
| ADR-009 | Frontend separado em React + Vite + TypeScript | Ecossistema e mercado; simplicidade em relação a Next.js; API Go como fonte da verdade |
| ADR-010 | Next.js apenas se SEO público exigir SSR/SSG | Evita complexidade de SSR no início do projeto |

---

## 15. Definição de pronto por camada

- **Config:** todas as variáveis documentadas, validadas e com defaults seguros.
- **Domínio:** entidades com invariantes testadas e sem imports de infraestrutura.
- **Caso de uso:** regras cobertas por testes e dependente apenas de interfaces.
- **Repositório:** implementa a interface do domínio e traduz erros corretamente.
- **Handler:** contrato HTTP estável, DTO separado da entidade, erros padronizados.
- **Router:** rotas versionadas, middleware aplicado e health checks disponíveis.
- **Main:** composition root sem regra de negócio, com shutdown gracioso.
- **Frontend:** consome a API, sem regra de negócio, com loading/erro tratados e
  tipos alinhados ao contrato.
- **Projeto:** `gofmt`, `go vet`, testes e lint passando no CI.

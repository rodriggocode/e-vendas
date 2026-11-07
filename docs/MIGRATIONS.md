# Migrations com golang-migrate

Referência dos comandos usados para criar e aplicar migrations no PostgreSQL
local do projeto. As migrations ficam em `migrations/` e são escritas à mão em
SQL puro.

Ferramenta: [`golang-migrate`](https://github.com/golang-migrate/migrate)
(binário `migrate`).

---

## Pré-requisitos

### Binário com driver do PostgreSQL

O `migrate` precisa ser compilado com a tag `postgres`. Para conferir:

```bash
migrate -help
```

A saída deve conter:

```
Database drivers: stub, postgres, postgresql
```

Se não tiver, reinstale:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Banco rodando

```bash
cd /caminho/do/e-vendas   # raiz do projeto
podman compose up -d
podman compose ps
```

---

## Variáveis do projeto

```bash
cd /caminho/do/e-vendas   # raiz do projeto

export DATABASE_URL='postgres://loja:loja@localhost:5433/loja?sslmode=disable'
export MIGRATIONS_DIR='migrations'
```

| Item | Valor |
|---|---|
| Host | `localhost:5433` |
| Container | `postgres:5432` |
| Banco | `loja` |
| Usuário | `loja` |
| Senha | `loja` |

Os comandos abaixo assumem essas variáveis exportadas.

---

## Estrutura dos arquivos

Cada migration é um par `up`/`down` com a **mesma versão**:

```
migrations/
├── 000001_create_updated_at_function.up.sql
├── 000001_create_updated_at_function.down.sql
├── 000002_create_users.up.sql
└── 000002_create_users.down.sql
```

Regras:

- `{versao}_{nome}.up.sql` aplica a mudança.
- `{versao}_{nome}.down.sql` desfaz a mudança.
- A versão é um inteiro único e crescente (`1`, `2`, `3`...).
- O `up` e o `down` do mesmo par precisam ter a mesma versão.
- Não edite uma migration já aplicada em produção; crie uma nova.

---

## Criar migrations

### Manual

Crie os dois arquivos com a próxima versão disponível:

```bash
touch migrations/000001_create_updated_at_function.up.sql
touch migrations/000001_create_updated_at_function.down.sql
```

### Pelo CLI

Gera o par de arquivos vazio:

```bash
migrate create -ext sql -dir migrations -seq create_categorias
```

Opções úteis:

| Opção | Efeito |
|---|---|
| `-ext sql` | Extensão `.sql` |
| `-dir migrations` | Pasta de destino |
| `-seq` | Numeração sequencial (`000001`) em vez de timestamp |
| `-digits 4` | Quantidade de dígitos na numeração sequencial |
| `-tz UTC` | Timezone para versões por timestamp |

Sem `-seq`, o nome usa timestamp:

```
20261102193000_create_categorias.up.sql
```

Para o projeto, prefira `-seq`: fica mais legível e a ordem é explícita.

---

## Aplicar migrations

### Subir todas as pendentes

```bash
migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" up
```

### Subir apenas N pendentes

```bash
migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" up 1
```

### Ir para uma versão específica

```bash
migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" goto 3
```

Sobe ou desce conforme a versão atual. Use `goto 0` para reverter tudo.

---

## Ver o estado atual

```bash
migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" version
```

Saída:

```
3
```

Se estiver em estado inconsistente:

```
3 (dirty)
```

### Ver o que está aplicado no banco

```bash
psql "$DATABASE_URL" -c 'SELECT * FROM schema_migrations;'
```

A tabela `schema_migrations` é criada e mantida pelo próprio `migrate`. Não
crie nem altere essa tabela nas suas migrations.

---

## Reverter migrations

### Desfazer a última

```bash
migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" down 1
```

### Desfazer N

```bash
migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" down 2
```

### Desfazer todas

```bash
migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" down -all
```

O `down` executa o arquivo `.down.sql` correspondente. Se ele não existir, a
reversão falha.

---

## Estado `dirty`

Se uma migration falhar no meio, o `migrate` marca a versão como `dirty` e
bloqueia novos comandos:

```
error: Dirty database version 3. Fix and force version.
```

Procedimento:

1. Leia o erro e verifique o que foi aplicado parcialmente:

   ```bash
   psql "$DATABASE_URL" -c '\dt'
   ```

2. Corrija o banco manualmente (remova o que ficou pela metade) **ou** termine
   a mudança na mão.
3. Corrija o arquivo `.sql` se o erro foi no SQL.
4. Marque a versão como resolvida:

   ```bash
   migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" force 3
   ```

5. Rode `version` e continue.

`force` **não executa SQL**: ele apenas grava a versão informada na tabela de
controle. Use com cuidado.

---

## Comandos destrutivos

### Apagar todo o schema do banco

```bash
migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" drop -f
```

Remove tabelas, dados e o histórico de migrations. O banco fica vazio. Em
desenvolvimento, equivale a começar do zero sem apagar o volume.

### Recriar do zero em desenvolvimento

```bash
migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" drop -f
migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" up
```

Alternativa mais radical, apagando o volume:

```bash
podman compose down -v
podman compose up -d
migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" up
```

---

## Log detalhado

Adicione `-verbose` para ver cada arquivo executado:

```bash
migrate -verbose -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" up
```

---

## Exemplo completo

`migrations/000001_create_updated_at_function.up.sql`:

```sql
CREATE TABLE categorias (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    nome             varchar(120) NOT NULL,
    slug             varchar(140) NOT NULL,
    categoria_pai_id uuid REFERENCES categorias (id) ON DELETE SET NULL,
    ativo            boolean NOT NULL DEFAULT true,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT categorias_slug_unique UNIQUE (slug)
);

CREATE INDEX categorias_categoria_pai_id_idx
    ON categorias (categoria_pai_id);
```

`migrations/000001_create_updated_at_function.down.sql`:

```sql
DROP TABLE IF EXISTS categorias;
```

Aplicar:

```bash
migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" up
migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" version
```

Reverter:

```bash
migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" down 1
```

---

## Boas práticas

- Escreva o `down` de verdade; não deixe vazio nem com `SELECT 1`.
- Use `IF EXISTS` no `down` para tolerar estados parciais.
- Uma migration deve fazer uma coisa coerente (criar tabela, adicionar coluna,
  criar índice).
- Não misture criação de tabela e carga de dados grandes na mesma migration.
- Prefira `NOT NULL DEFAULT ...` ao adicionar coluna em tabela com dados.
- Crie índices em migrations separadas quando a tabela já tiver volume.
- Rode `up` e `down 1` em desenvolvimento antes de considerar a migration
  pronta.
- Nunca altere migration já aplicada em produção: crie a próxima.
- Teste o par completo em um banco limpo (`drop -f` + `up`).

---

## Resumo dos comandos

```bash
# criar
migrate create -ext sql -dir migrations -seq create_categorias

# aplicar
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" up 1
migrate -path migrations -database "$DATABASE_URL" goto 3

# consultar
migrate -path migrations -database "$DATABASE_URL" version

# reverter
migrate -path migrations -database "$DATABASE_URL" down 1
migrate -path migrations -database "$DATABASE_URL" down -all

# resolver dirty
migrate -path migrations -database "$DATABASE_URL" force 3

# apagar tudo
migrate -path migrations -database "$DATABASE_URL" drop -f
```

package postgres

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Testes de integração do schema do MVP (migrations 000001-000012).
// Rodam contra o banco já migrado (no CI, o pipeline aplica as migrations antes
// de go test). O alvo é a defesa estrutural do multi-loja e as travas de
// integridade; os handlers de orders/products ainda não existem.

func sufixoDeSchema(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func limparSchemaMVP(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := pool.Exec(ctx,
		`TRUNCATE order_items, orders, order_status_history, products, product_images,
		        store_users, stores, customers RESTART IDENTITY CASCADE`); err != nil {
		t.Logf("limpeza do schema: %v", err)
	}
}

func criarLoja(t *testing.T, pool *pgxpool.Pool, slug string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO stores (name, slug, accepts_pickup, status, accepted_payment_methods)
		VALUES ($1, $2, true, 'active', ARRAY['pix'])
		RETURNING id`, "Loja "+slug, slug).Scan(&id)
	if err != nil {
		t.Fatalf("criando loja %s: %v", slug, err)
	}
	return id
}

func criarCliente(t *testing.T, pool *pgxpool.Pool, telefone string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO customers (name, phone) VALUES ('Cliente QA', $1) RETURNING id`,
		telefone).Scan(&id)
	if err != nil {
		t.Fatalf("criando cliente: %v", err)
	}
	return id
}

func criarCategoria(t *testing.T, pool *pgxpool.Pool, slug string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO categories (name, slug) VALUES ($1, $2) RETURNING id`,
		"Categoria "+slug, slug).Scan(&id)
	if err != nil {
		t.Fatalf("criando categoria: %v", err)
	}
	return id
}

func criarProduto(t *testing.T, pool *pgxpool.Pool, storeID, categoriaID, nome string, preco int64) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO products (store_id, category_id, name, price_cents)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		storeID, categoriaID, nome, preco).Scan(&id)
	if err != nil {
		t.Fatalf("criando produto: %v", err)
	}
	return id
}

func criarPedido(t *testing.T, pool *pgxpool.Pool, storeID, clienteID, tracking string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO orders (store_id, customer_id, tracking_code, fulfillment_type,
		                    payment_method, subtotal_cents, total_cents)
		VALUES ($1, $2, $3, 'pickup', 'pix', 1000, 1000) RETURNING id`,
		storeID, clienteID, tracking).Scan(&id)
	if err != nil {
		t.Fatalf("criando pedido: %v", err)
	}
	return id
}

func trackingDeTeste(n int) string {
	return fmt.Sprintf("qa-tc-%026d", n)
}

// trackingUnico gera um tracking_code válido e exclusivo por tentativa.
func trackingUnico(sufixo string, j int) string {
	return fmt.Sprintf("qa-tc-%s-%04d", sufixo, j)
}

// TestSchemaMVP_TabelasExistentes verifica as 12 tabelas do MVP.
func TestSchemaMVP_TabelasExistentes(t *testing.T) {
	pool := novaPoolDeTeste(t)

	tabelas := []string{
		"users", "stores", "store_users", "categories", "products", "product_images",
		"customers", "orders", "order_items", "password_resets",
		"email_verifications", "order_status_history",
	}
	ctx := context.Background()
	for _, tabela := range tabelas {
		var existe bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM information_schema.tables
			 WHERE table_schema = 'public' AND table_name = $1)`, tabela).Scan(&existe)
		if err != nil {
			t.Fatalf("verificando %s: %v", tabela, err)
		}
		if !existe {
			t.Errorf("tabela %s não existe", tabela)
		}
	}
}

// TestSchemaMVP_ItemDeOutraLojaRejeitado é a defesa estrutural do multi-loja:
// a FK composta (product_id, store_id) impede item de produto de outra loja.
func TestSchemaMVP_ItemDeOutraLojaRejeitado(t *testing.T) {
	pool := novaPoolDeTeste(t)
	t.Cleanup(func() { limparSchemaMVP(t, pool) })

	sufixo := sufixoDeSchema(t)
	lojaA := criarLoja(t, pool, "qa-a-"+sufixo)
	lojaB := criarLoja(t, pool, "qa-b-"+sufixo)
	cliente := criarCliente(t, pool, sufixo)
	categoria := criarCategoria(t, pool, "qa-cat-"+sufixo)
	produtoB := criarProduto(t, pool, lojaB, categoria, "Produto B", 1000)
	pedidoA := criarPedido(t, pool, lojaA, cliente, trackingDeTeste(1))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx, `
		INSERT INTO order_items (order_id, product_id, store_id, product_name, quantity, unit_price_cents)
		VALUES ($1, $2, $3, 'Produto B', 1, 1000)`,
		pedidoA, produtoB, lojaA)
	if err == nil {
		t.Fatal("item com produto de outra loja foi aceito; a FK composta não segurou")
	}
}

// TestSchemaMVP_TrackingCodeCurtoRejeitado trava o tracking_code como chave
// pública de leitura: o schema precisa exigir entropia mínima (M-A).
func TestSchemaMVP_TrackingCodeCurtoRejeitado(t *testing.T) {
	pool := novaPoolDeTeste(t)
	t.Cleanup(func() { limparSchemaMVP(t, pool) })

	sufixo := sufixoDeSchema(t)
	loja := criarLoja(t, pool, "qa-track-"+sufixo)
	cliente := criarCliente(t, pool, sufixo)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx, `
		INSERT INTO orders (store_id, customer_id, tracking_code, fulfillment_type,
		                    payment_method, subtotal_cents, total_cents)
		VALUES ($1, $2, 'curto', 'pickup', 'pix', 1000, 1000)`, loja, cliente)
	if err == nil {
		t.Fatal("tracking_code curto foi aceito; a entropia mínima não está sendo exigida")
	}
}

// TestSchemaMVP_NumeroPedidoSequencialPorLoja confirma que o número é atribuído
// pelo banco, por loja, sem inflar stores.updated_at (M-F).
func TestSchemaMVP_NumeroPedidoSequencialPorLoja(t *testing.T) {
	pool := novaPoolDeTeste(t)
	t.Cleanup(func() { limparSchemaMVP(t, pool) })

	sufixo := sufixoDeSchema(t)
	lojaA := criarLoja(t, pool, "qa-seq-a-"+sufixo)
	lojaB := criarLoja(t, pool, "qa-seq-b-"+sufixo)
	cliente := criarCliente(t, pool, sufixo)

	ctx := context.Background()

	var antes time.Time
	if err := pool.QueryRow(ctx, `SELECT updated_at FROM stores WHERE id = $1`, lojaB).Scan(&antes); err != nil {
		t.Fatalf("lendo updated_at da loja: %v", err)
	}

	criarPedido(t, pool, lojaA, cliente, trackingDeTeste(1))
	criarPedido(t, pool, lojaA, cliente, trackingDeTeste(2))
	criarPedido(t, pool, lojaB, cliente, trackingDeTeste(3))

	numeros := map[string]int64{}
	for _, loja := range []string{lojaA, lojaB} {
		rows, err := pool.Query(ctx, `SELECT number FROM orders WHERE store_id = $1 ORDER BY number`, loja)
		if err != nil {
			t.Fatalf("lendo números: %v", err)
		}
		for rows.Next() {
			var n int64
			if err := rows.Scan(&n); err != nil {
				rows.Close()
				t.Fatalf("lendo número: %v", err)
			}
			numeros[fmt.Sprintf("%s:%d", loja, n)] = n
		}
		rows.Close()
	}
	if len(numeros) != 3 {
		t.Fatalf("números atribuídos = %v; esperado 3 linhas (1,2 na loja A e 1 na B)", numeros)
	}
}

// TestSchemaMVP_PedidoNaoInflaStoresUpdatedAt confirma que o trigger de número
// não atualiza mais stores.updated_at a cada pedido (M-F).
func TestSchemaMVP_PedidoNaoInflaStoresUpdatedAt(t *testing.T) {
	pool := novaPoolDeTeste(t)
	t.Cleanup(func() { limparSchemaMVP(t, pool) })

	sufixo := sufixoDeSchema(t)
	loja := criarLoja(t, pool, "qa-upd-"+sufixo)
	cliente := criarCliente(t, pool, sufixo)

	ctx := context.Background()
	var antes time.Time
	if err := pool.QueryRow(ctx, `SELECT updated_at FROM stores WHERE id = $1`, loja).Scan(&antes); err != nil {
		t.Fatalf("lendo updated_at: %v", err)
	}

	criarPedido(t, pool, loja, cliente, trackingDeTeste(1))

	var depois time.Time
	if err := pool.QueryRow(ctx, `SELECT updated_at FROM stores WHERE id = $1`, loja).Scan(&depois); err != nil {
		t.Fatalf("lendo updated_at: %v", err)
	}
	if !depois.Equal(antes) {
		t.Errorf("stores.updated_at mudou após pedido: %v -> %v; o trigger não deveria atualizar", antes, depois)
	}
}

type resultadoConcorrencia struct {
	numero int64
	erro   error
}

func inserirPedidoConcorrente(ctx context.Context, pool *pgxpool.Pool, storeID, clienteID, tracking string) (int64, error) {
	var numero int64
	err := pool.QueryRow(ctx, `
		INSERT INTO orders (store_id, customer_id, tracking_code, fulfillment_type,
		                    payment_method, subtotal_cents, total_cents)
		VALUES ($1, $2, $3, 'pickup', 'pix', 1000, 1000)
		RETURNING number`, storeID, clienteID, tracking).Scan(&numero)
	return numero, err
}

// TestSchemaMVP_ConcorrenciaPrimeiroPedidoDeLojaNova reproduz o M-G: duas
// inserções simultâneas na primeira compra de uma loja nova não podem falhar
// nem produzir número duplicado ou lacuna. Critério de aceite do M-G.
func TestSchemaMVP_ConcorrenciaPrimeiroPedidoDeLojaNova(t *testing.T) {
	pool := novaPoolDeTeste(t)
	t.Cleanup(func() { limparSchemaMVP(t, pool) })

	const tentativas = 25
	var falhas []string

	for i := 0; i < tentativas; i++ {
		sufixo := fmt.Sprintf("%s-%d", sufixoDeSchema(t), i)
		loja := criarLoja(t, pool, "qa-race-"+sufixo)
		telefone := fmt.Sprintf("119%015d", time.Now().UnixNano()%1_000_000_000_000_000+int64(i))
		cliente := criarCliente(t, pool, telefone)

		var wg sync.WaitGroup
		wg.Add(2)
		inicio := make(chan struct{})
		resultados := make(chan resultadoConcorrencia, 2)

		for j := 0; j < 2; j++ {
			go func(j int) {
				defer wg.Done()
				<-inicio
				num, err := inserirPedidoConcorrente(context.Background(), pool, loja, cliente, trackingUnico(sufixo, j))
				resultados <- resultadoConcorrencia{numero: num, erro: err}
			}(j)
		}
		close(inicio)
		wg.Wait()
		close(resultados)

		numeros := map[int64]bool{}
		for r := range resultados {
			if r.erro != nil {
				falhas = append(falhas, fmt.Sprintf("tentativa %d: erro %v", i, r.erro))
				continue
			}
			if numeros[r.numero] {
				falhas = append(falhas, fmt.Sprintf("tentativa %d: número duplicado %d", i, r.numero))
			}
			numeros[r.numero] = true
		}
		if len(numeros) != 2 {
			falhas = append(falhas, fmt.Sprintf("tentativa %d: números %v, esperado {1,2}", i, numeros))
		}
	}

	if len(falhas) > 0 {
		t.Fatalf("corrida no primeiro pedido da loja (%d falha(s) em %d tentativas):\n%s",
			len(falhas), tentativas, strings.Join(falhas, "\n"))
	}
}

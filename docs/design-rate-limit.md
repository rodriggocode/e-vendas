# Design — Rate limit em `POST /api/v1/auth/register`

**Autor:** @security · **Estado:** pronto para implementação · **Dono da implementação:** @devops
**Baseline:** `e-vendas` @ `0991490` (CI verde, `govulncheck` limpo, `gosec` 0 issues)

---

## 1. Ameaça que o controle mata

Rota pública sem autenticação. Sem controle, um script faz três coisas de graça:

1. **Enumera e-mail** (S2): o `409 "e-mail já cadastrado"` é um oráculo — percorrendo uma
   lista de e-mails o atacante descobre quem é cliente.
2. **Massa de cadastro**: gera usuários/lojas falsas, spam e lixo no banco.
3. **Queima recurso**: cada registro que passa custa **bcrypt** (~60 ms de CPU) + 1 escrita +
   1 conexão do pool (`MaxConns = 10`). O recurso caro não é o HTTP, é o bcrypt e o pool.

**Severidade:** média hoje, alta quando existir login. **CWE:** CWE-307 / CWE-799.

---

## 2. Chaves

| Balde | Derivação | Observação |
|---|---|---|
| **IP** | `RemoteAddr` (host) **ou** `X-Forwarded-For[0]` | XFF só é honrado **se o peer estiver em `TRUSTED_PROXIES`** (ver §6) |
| **E-mail** | `sha256(lower(trim(email)))` | normalizado; hash para não reter PII em memória/log |

Ambas as chaves entram no store **como hash SHA-256** (não é segredo, mas evita dado pessoal
em log/métrica e em dump de memória).

---

## 3. Cotas (defaults propostos)

| Balde | Taxa | Burst | Env |
|---|---|---|---|
| IP | 10/min | 5 | `RATE_LIMIT_IP_PER_MIN`, `RATE_LIMIT_IP_BURST` |
| E-mail | 5/min | 3 | `RATE_LIMIT_EMAIL_PER_MIN`, `RATE_LIMIT_EMAIL_BURST` |

Burst **menor** que a taxa: para cadastro, interessa negar rajada, não só o excesso de longo
prazo. `0` ou negativo numa variável = **falha na inicialização** (mesmo padrão de
`config.Load`), não "desliga silenciosamente".

---

## 4. Resposta

```
HTTP/1.1 429 Too Many Requests
Retry-After: <ceil(1/taxa)>          # segundos
Content-Type: application/json

{"error":{"code":"rate_limited","message":"muitas requisições, tente novamente em instantes"}}
```

Opcional (recomendado, barato): `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`.

---

## 5. Custo por requisição

- **Lookup + token:** O(1), ~centenas de ns. Desprezível.
- **Memória:** ~1 entrada por chave ativa; com GC de ociosos (§7) fica estável.
- **O alvo real:** bcrypt (~60 ms) dos pedidos que **passam** — o rate limit existe para
  reduzir esse número.

---

## 6. `X-Forwarded-For` — ponto que muda a segurança (não é detalhe)

Honrar XFF **incondicionalmente** permite burla: sem proxy na frente, o atacante manda
`X-Forwarded-For: 1.2.3.4` falso e troca de balde a cada requisição. Regra:

- `TRUSTED_PROXIES` (CIDRs, ex.: `10.0.0.0/8`) configurado e o peer (`RemoteAddr`) estiver
  dentro → usa `XFF[0]`.
- Caso contrário → **ignora XFF** e usa `RemoteAddr`.

Isso mantém a intenção do @hermes (atrás de proxy, ninguém cai no mesmo balde) sem criar o
buraco (sem proxy, XFF de cliente não engana o controle). Default `TRUSTED_PROXIES` **vazio**.

---

## 7. Store

**Em memória, por instância**, com `golang.org/x/time/rate`. Monólito de 1 réplica hoje;
trade-off assumido: com N réplicas a cota efetiva vira `N ×` — trocar por Redis (ou
`Ratelimit` do proxy) quando escalar. GC de entradas ociosas (>10 min) para não crescer sem fim.

---

## 8. Ordem dos middlewares

```
recover → headers de segurança → CORS → rate limit → mux
```

Rate limit **só** na rota de registro. `/healthz` não passa por ele (senão o healthcheck
queima cota e derruba o deploy).

---

## 9. Esqueleto (para o @devops)

```go
// internal/middleware/ratelimit.go
package middleware

import (
    "crypto/sha256"
    "encoding/hex"
    "math"
    "net"
    "net/http"
    "strings"
    "sync"
    "time"

    "golang.org/x/time/rate"
)

type store interface {
    allow(key string, r rate.Limit, burst int) (bool, time.Duration)
}

type memStore struct {
    mu      sync.Mutex
    entries map[string]*entry
}

type entry struct {
    lim  *rate.Limiter
    last time.Time
}

func (s *memStore) allow(key string, r rate.Limit, burst int) (bool, time.Duration) {
    s.mu.Lock()
    defer s.mu.Unlock()
    e, ok := s.entries[key]
    if !ok {
        e = &entry{lim: rate.NewLimiter(r, burst)}
        s.entries[key] = e
    }
    e.last = time.Now()
    if e.lim.Allow() {
        return true, 0
    }
    return false, time.Duration(math.Ceil(1 / float64(r) * float64(time.Second)))
}

func (s *memStore) gc(maxIdle time.Duration) {
    s.mu.Lock()
    defer s.mu.Unlock()
    for k, e := range s.entries {
        if time.Since(e.last) > maxIdle {
            delete(s.entries, k)
        }
    }
}

func hashKey(s string) string {
    sum := sha256.Sum256([]byte(s))
    return hex.EncodeToString(sum[:])
}

func clientIP(r *http.Request, trusted []*net.IPNet) string {
    host, _, err := net.SplitHostPort(r.RemoteAddr)
    if err != nil {
        host = r.RemoteAddr
    }
    ip := net.ParseIP(host)
    if ip == nil {
        return "unknown"
    }
    for _, n := range trusted {
        if n.Contains(ip) {
            if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
                first := strings.TrimSpace(strings.Split(xff, ",")[0])
                if real := net.ParseIP(first); real != nil {
                    return real.String()
                }
            }
            break
        }
    }
    return ip.String()
}
```

Config e wiring seguem o padrão existente de `config.Load` (fail-fast) + `router.Router`
(composition root). Logger de 429 **sem** IP/e-mail em claro.

---

## 10. Testes (@qa)

- 429 após `burst+1` do mesmo IP; `Retry-After` presente e coerente; envelope `rate_limited`.
- 429 por e-mail independente do IP (mesmo IP, outro e-mail continua com cota).
- IPs distintos não colapsam entre si.
- **`TRUSTED_PROXIES` vazio → XFF ignorado** (o ataque de burla não funcionando é o teste).
- `TRUSTED_PROXIES` com o peer dentro → XFF[0] honrado.
- `/healthz` nunca leva 429.
- Rodar com teto alto via env (ou IP variado), senão o próprio probe esgota a cota — 429 no
  teste é o limite funcionando.

---

## 11. Risco residual

- N réplicas → cota multiplicada (mitigar com Redis/proxy na escalada).
- Botnet com muitos IPs contorna por design; CAPTCHA/anti-bot é a camada seguinte.
- E-mail comum a muitas contas (ex.: `qa.n256@exemplo.com`) pode receber cota curta — o
  balde por e-mail é proposital, mas observar em monitoramento.

// IP do cliente para rate limit.
//
// Copiado de isofthub/internal/httputil/clientip.go em 2026-11-23; rever contra
// ips-v4/ips-v6 da Cloudflare. A unica diferenca e a inclusao de TRUSTED_PROXIES,
// que o e-vendas usa para proxies confiaveis configurados por env.
package middleware

import (
	"net"
	"net/http"
	"strings"
)

// rangesCloudflare vem de https://www.cloudflare.com/ips-v4 e ips-v6 (a
// Cloudflare publica e altera esses blocos; atualizar aqui quando mudar).
// O header só é confiável quando o peer (RemoteAddr) está em um desses blocos.
var rangesCloudflare = mustRanges(
	"173.245.48.0/20",
	"103.21.244.0/22",
	"103.22.200.0/22",
	"103.31.4.0/22",
	"141.101.64.0/18",
	"108.162.192.0/18",
	"190.93.240.0/20",
	"188.114.96.0/20",
	"197.234.240.0/22",
	"198.41.128.0/17",
	"162.158.0.0/15",
	"104.16.0.0/13",
	"104.24.0.0/14",
	"172.64.0.0/13",
	"131.0.72.0/22",
	"2400:cb00::/32",
	"2606:4700::/32",
	"2803:f800::/32",
	"2405:b500::/32",
	"2405:8100::/32",
	"2a06:98c0::/29",
	"2c0f:f248::/32",
)

func mustRanges(blocos ...string) []*net.IPNet {
	var redes []*net.IPNet
	for _, bloco := range blocos {
		_, rede, err := net.ParseCIDR(bloco)
		if err != nil {
			panic("middleware: CIDR inválido: " + bloco)
		}
		redes = append(redes, rede)
	}
	return redes
}

// ClientIP devolve o IP do cliente para fins de rate limit.
//
// A borda da Cloudflare (onde o Render entrega o tráfego de *.onrender.com) é
// um peer público, então o header só pode ser lido quando o peer está nos ranges
// oficiais, num proxy confiável configurado ou em rede privada (origem direta).
// Fora disso o atacante forja o header e escolhe o próprio balde; o sintoma da
// configuração errada é balde único (429 coletivo), nunca bypass.
func ClientIP(r *http.Request, trusted []*net.IPNet) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return host
	}

	if ehProxyLocal(ip) || ipEmRangesCloudflare(ip) || ipEmRedes(ip, trusted) {
		if cf := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); cf != "" {
			if real := net.ParseIP(cf); real != nil {
				return real.String()
			}
		}
		if xff := primeiroIPValido(r.Header.Get("X-Forwarded-For")); xff != "" {
			return xff
		}
	}

	return ip.String()
}

// primeiroIPValido devolve o primeiro valor de X-Forwarded-For que é um IP.
func primeiroIPValido(xff string) string {
	for _, candidato := range strings.Split(xff, ",") {
		candidato = strings.TrimSpace(candidato)
		if candidato == "" {
			continue
		}
		if real := net.ParseIP(candidato); real != nil {
			return real.String()
		}
	}
	return ""
}

func ipEmRangesCloudflare(ip net.IP) bool {
	return ipEmRedes(ip, rangesCloudflare)
}

func ipEmRedes(ip net.IP, redes []*net.IPNet) bool {
	for _, rede := range redes {
		if rede.Contains(ip) {
			return true
		}
	}
	return false
}

// ehProxyLocal cobre RFC1918/loopback/link-local. Se o provedor usar CGNAT
// (100.64.0.0/10), adicione o bloco aqui.
func ehProxyLocal(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()
}

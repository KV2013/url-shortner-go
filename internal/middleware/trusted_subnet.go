package middleware

import (
	"net"
	"net/http"

	"github.com/KV2013/url-shortner-go/internal/config"
	"go.uber.org/zap"
)

func TrustedSubnet(cfg *config.Config, logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.TrustedSubnet == "" {
				next.ServeHTTP(w, r)
				return
			}

			if r.URL.Path != "/api/internal/stats" {
				next.ServeHTTP(w, r)
				return
			}

			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				logger.Warn("запрос к /api/internal/stats без заголовка X-Real-IP")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			_, cidr, err := net.ParseCIDR(cfg.TrustedSubnet)
			if err != nil {
				logger.Error("некорректный trusted_subnet в конфигурации",
					zap.String("trusted_subnet", cfg.TrustedSubnet),
					zap.Error(err),
				)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			ip := net.ParseIP(realIP)
			if ip == nil {
				logger.Warn("некорректный IP в заголовке X-Real-IP",
					zap.String("x_real_ip", realIP),
				)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if !cidr.Contains(ip) {
				logger.Warn("IP не входит в доверенную подсеть",
					zap.String("ip", realIP),
					zap.String("trusted_subnet", cfg.TrustedSubnet),
				)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

package httpapi

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Limits are per API process and socket IP; forwarded headers are not trusted.
func security(next http.Handler) http.Handler {
	type bucket struct {
		count int
		until time.Time
	}
	var mu sync.Mutex
	buckets := make(map[string]bucket)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/api/v1/auth/") {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}
			now := time.Now()
			mu.Lock()
			for key, item := range buckets {
				if !now.Before(item.until) {
					delete(buckets, key)
				}
			}
			item := buckets[ip]
			blocked := item.count >= 30 || (item.count == 0 && len(buckets) >= 10000)
			if !blocked {
				if item.count == 0 {
					item.until = now.Add(time.Minute)
				}
				item.count++
				buckets[ip] = item
			}
			mu.Unlock()
			if blocked {
				w.Header().Set("Retry-After", "60")
				writeError(w, 429, "rate_limited", "Muitas tentativas. Aguarde um minuto.", nil)
				return
			}
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

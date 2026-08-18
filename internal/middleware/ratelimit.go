package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimit 返回基于固定窗口的限流中间件。
// limit 为每客户端每分钟允许的最大请求数；limit<=0 时不限流。
func RateLimit(limit int) func(http.Handler) http.Handler {
	if limit <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}

	var mu sync.Mutex
	window := time.Now().Truncate(time.Minute)
	counters := map[string]int{}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			now := time.Now()
			currentWindow := now.Truncate(time.Minute)
			if !currentWindow.Equal(window) {
				window = currentWindow
				counters = map[string]int{}
			}
			key := clientKey(r)
			counters[key]++
			allowed := counters[key] <= limit
			mu.Unlock()

			if !allowed {
				w.Header().Set("Retry-After", "60")
				http.Error(w, `{"code":429,"message":"请求过于频繁，请稍后再试"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// clientKey 基于客户端 IP 生成限流键。
func clientKey(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	if idx := indexByte(ip, ','); idx >= 0 {
		ip = ip[:idx]
	}
	return ip
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

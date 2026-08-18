// Package middleware 提供 HTTP 中间件：鉴权、限流、请求日志。
package middleware

import (
	"net/http"
	"strings"
)

// Auth 返回鉴权中间件。token 为空时放行；否则要求 Authorization 头携带
// "Bearer <token>"，校验通过才放行。
func Auth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, `{"code":401,"message":"缺少访问令牌"}`, http.StatusUnauthorized)
				return
			}
			if strings.TrimPrefix(auth, "Bearer ") != token {
				http.Error(w, `{"code":401,"message":"访问令牌无效"}`, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

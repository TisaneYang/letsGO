// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type AdminAuthMiddleware struct {
	jwtSecret string
}

func NewAdminAuthMiddleware(jwtSecret string) *AdminAuthMiddleware {
	return &AdminAuthMiddleware{
		jwtSecret: jwtSecret,
	}
}

func (m *AdminAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取 Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		// 解析 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		// 验证 JWT token
		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			return []byte(m.jwtSecret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// 获取 claims
		claims := token.Claims.(jwt.MapClaims)

		// 检查是否是管理员角色
		role, ok := claims["role"].(string)
		if !ok || role != "admin" {
			http.Error(w, "permission denied: admin access required", http.StatusForbidden)
			return
		}

		// 将用户信息存入 context
		userId := int64(claims["userId"].(float64))
		ctx := context.WithValue(r.Context(), "userId", userId)
		ctx = context.WithValue(ctx, "role", role)

		// Passthrough to next handler if need
		next(w, r.WithContext(ctx))
	}
}

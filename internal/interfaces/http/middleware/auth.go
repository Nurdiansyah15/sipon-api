package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/port"
	"sipon-api/internal/app/service/principal"
	"sipon-api/internal/interfaces/http/httperror"
)

const principalTTL = 5 * time.Minute

// JWTAuth validates the bearer token and sets "user_id" + "session_id" in context.
// Auth data (roles, permissions) is NOT embedded in the token — it is loaded
// from Redis/DB by PrincipalLoader.
//
// revocationStore memberi enforcement nyata untuk logout (per-session) dan
// logout-all (semua token diterbitkan sebelum suatu waktu) meski JWT sendiri
// stateless — lihat port.SessionRevocationStore. Best-effort: kalau store gagal
// dibaca (mis. Redis sempat down), request tidak diblokir (fail-open), supaya
// gangguan infra Redis tidak mengunci semua user keluar.
func JWTAuth(tokenGen port.TokenGenerator, revocationStore port.SessionRevocationStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))

		// fallback ke query param ?token= untuk WebSocket
		if authHeader == "" {
			if t := strings.TrimSpace(c.Query("token")); t != "" {
				authHeader = t
			}
		}

		if authHeader == "" {
			httperror.Handle(c, apperror.Unauthorized(string(apperror.CodeUnauthorized)))
			return
		}

		tokenStr := authHeader
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			tokenStr = strings.TrimSpace(authHeader[7:])
		}

		claims, err := tokenGen.ParseAccessToken(tokenStr)
		if err != nil {
			httperror.Handle(c, apperror.Unauthorized(string(apperror.CodeUnauthorized), err))
			return
		}

		if revocationStore != nil {
			ctx := c.Request.Context()
			if revoked, revErr := revocationStore.IsSessionRevoked(ctx, claims.SessionID); revErr == nil && revoked {
				httperror.Handle(c, apperror.Unauthorized(string(apperror.CodeUnauthorized)))
				return
			}
			if revokedBefore, revErr := revocationStore.RevokedBefore(ctx, claims.UserID); revErr == nil && revokedBefore != nil {
				if claims.IssuedAt.Before(*revokedBefore) {
					httperror.Handle(c, apperror.Unauthorized(string(apperror.CodeUnauthorized)))
					return
				}
			}
			// Device-scoped: enforce "logout device lain" — hanya berlaku kalau
			// token ini membawa device_id (client mengirimnya saat login).
			if claims.DeviceID != "" {
				if revokedBefore, revErr := revocationStore.DeviceRevokedBefore(ctx, claims.UserID, claims.DeviceID); revErr == nil && revokedBefore != nil {
					if claims.IssuedAt.Before(*revokedBefore) {
						httperror.Handle(c, apperror.Unauthorized(string(apperror.CodeUnauthorized)))
						return
					}
				}
			}
		}

		c.Set("user_id", claims.UserID)
		c.Set("session_id", claims.SessionID)
		c.Set("device_id", claims.DeviceID)
		c.Next()
	}
}

// OptionalJWTAuth attempts token validation but continues if missing/invalid.
func OptionalJWTAuth(tokenGen port.TokenGenerator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			c.Next()
			return
		}

		tokenStr := authHeader
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			tokenStr = strings.TrimSpace(authHeader[7:])
		}

		claims, err := tokenGen.ParseAccessToken(tokenStr)
		if err != nil {
			c.Next()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("session_id", claims.SessionID)
		c.Next()
	}
}

// PrincipalLoader loads the Principal for the authenticated user.
// Flow: request-cache → Redis → DB → store back.
// Must be placed after JWTAuth.
func PrincipalLoader(builder *principal.Builder, cache port.PrincipalCachePort) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			httperror.Handle(c, apperror.Unauthorized(string(apperror.CodeUnauthorized)))
			return
		}

		sessionID := c.GetString("session_id")

		// 1. request-scoped cache (already loaded in this request)
		if existing, ok := c.Get("principal"); ok {
			if p, ok := existing.(*principal.Principal); ok && p != nil {
				c.Next()
				_ = p
				return
			}
		}

		ctx := c.Request.Context()

		// 2. Redis cache
		var p *principal.Principal
		if cache != nil {
			cached, err := cache.Get(ctx, userID)
			if err == nil && cached != nil {
				p = cached
			}
		}

		// 3. build from DB
		if p == nil {
			built, err := builder.Build(ctx, userID, sessionID)
			if err != nil {
				httperror.Handle(c, apperror.Unauthorized(string(apperror.CodeUnauthorized), err))
				return
			}
			p = built

			if cache != nil {
				_ = cache.Set(ctx, userID, p, principalTTL)
			}
		}

		c.Set("principal", p)
		c.Next()
	}
}

// GetPrincipal extracts the Principal from gin context (exported for handlers/usecases).
func GetPrincipal(c *gin.Context) *principal.Principal {
	v, ok := c.Get("principal")
	if !ok {
		return nil
	}
	p, _ := v.(*principal.Principal)
	return p
}

// ── Guard Middleware ──────────────────────────────────────────────────────────

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := GetPrincipal(c)
		if p != nil {
			for _, role := range roles {
				if p.HasRole(role) {
					c.Next()
					return
				}
			}
		}
		httperror.Handle(c, apperror.Forbidden(string(apperror.CodeForbidden)))
	}
}

func RequirePermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := GetPrincipal(c)
		if p != nil {
			for _, perm := range permissions {
				if p.HasPermission(perm) {
					c.Next()
					return
				}
			}
		}
		httperror.Handle(c, apperror.Forbidden(string(apperror.CodeForbidden)))
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		p := GetPrincipal(c)
		if p != nil && p.IsAdmin() {
			c.Next()
			return
		}
		httperror.Handle(c, apperror.Forbidden(string(apperror.CodeForbidden)))
	}
}

func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		p := GetPrincipal(c)
		if p != nil && p.IsSuperAdmin() {
			c.Next()
			return
		}
		httperror.Handle(c, apperror.Forbidden(string(apperror.CodeForbidden)))
	}
}

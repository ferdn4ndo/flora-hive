package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"flora-hive/internal/controllers/authctx"
	"flora-hive/internal/infrastructure/userver"
	"flora-hive/internal/services"
	"flora-hive/lib"
)

// AttachAuthOptional wires JWT (uServer) and API key authentication for /v1.
func AttachAuthOptional(env lib.Env, uv *userver.Client, users *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if key := strings.TrimSpace(c.GetHeader("X-API-Key")); key != "" {
			for _, k := range env.HiveAPIKeys() {
				if k == key {
					authctx.Set(c, &authctx.Principal{Kind: authctx.KindAPIKey})
					c.Next()
					return
				}
			}
		}
		bearer := extractBearer(c.GetHeader("Authorization"))
		if bearer == "" || !env.UserverConfigured() {
			c.Next()
			return
		}
		me, status, _, err := uv.Me(bearer)
		if err != nil || status < 200 || status >= 300 || me == nil {
			c.Next()
			return
		}
		hiveID, err := users.UpsertFromMe(me)
		if err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}
		authctx.Set(c, &authctx.Principal{
			Kind:        authctx.KindJWT,
			HiveUserID:  hiveID,
			AccessToken: bearer,
		})
		c.Next()
	}
}

func extractBearer(h string) string {
	h = strings.TrimSpace(h)
	const p = "Bearer "
	if len(h) <= len(p) || !strings.EqualFold(h[:len(p)], p) {
		return ""
	}
	return strings.TrimSpace(h[len(p):])
}

// RequireAuth rejects requests with no principal.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		p, ok := authctx.Get(c)
		if !ok || p == nil || p.Kind == authctx.KindNone {
			c.JSON(401, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required: Bearer access token (uServer-Auth) or X-API-Key",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

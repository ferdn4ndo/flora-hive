package authctx

import "github.com/gin-gonic/gin"

// Kind describes how the request was authenticated.
type Kind string

const (
	// KindNone means no principal was attached (optional auth).
	KindNone Kind = ""
	// KindAPIKey is service authentication via X-API-Key.
	KindAPIKey Kind = "api_key"
	// KindJWT is a uServer-Auth bearer session synced to hive_users.
	KindJWT Kind = "jwt"
)

// Principal is stored in the Gin context after optional auth middleware.
type Principal struct {
	Kind        Kind
	HiveUserID  string
	AccessToken string
}

const ctxKey = "flora_hive_auth"

// Set stores the principal on the context.
func Set(c *gin.Context, p *Principal) {
	c.Set(ctxKey, p)
}

// Get returns the principal if present.
func Get(c *gin.Context) (*Principal, bool) {
	v, ok := c.Get(ctxKey)
	if !ok {
		return nil, false
	}
	p, ok := v.(*Principal)
	return p, ok
}

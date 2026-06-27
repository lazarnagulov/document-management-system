package middleware

import (
	"net/http"

	"bitbucket.org/HyperEther/document-management-system/internal/platform/httputil"
	"bitbucket.org/HyperEther/document-management-system/internal/platform/token"
	"github.com/gin-gonic/gin"
)

const ContextUserKey = "user_claims"

func Authenticate(cfg token.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		t, err := c.Cookie("access_token")
		if err != nil {
			httputil.WriteError(c, http.StatusUnauthorized, "missing access token")
			c.Abort()
			return
		}

		claims, err := token.ParseAccessToken(t, cfg.AccessSecret)
		if err != nil {
			httputil.WriteError(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(ContextUserKey, claims)
		c.Next()
	}
}

func ClaimsFromContext(c *gin.Context) (*token.Claims, bool) {
	val, exists := c.Get(ContextUserKey)
	if !exists {
		return nil, false
	}
	claims, ok := val.(*token.Claims)
	return claims, ok
}

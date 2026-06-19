package api

import (
	"net/http"
	"os"

	"github.com/ghpages/mobagent/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

const accountIDKey = "accountId"

func AccountMiddleware(authStore *auth.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := auth.ExtractToken(
			c.GetHeader("Authorization"),
			c.Query("token"),
			c.GetHeader("X-MobAgent-Token"),
		)
		if token != "" {
			if acct, ok := authStore.ValidateToken(token); ok {
				c.Set(accountIDKey, acct.ID)
				c.Next()
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		if os.Getenv("MOBAGENT_REQUIRE_AUTH") == "1" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token required"})
			return
		}
		c.Set(accountIDKey, authStore.DefaultAccountID())
		c.Next()
	}
}

func accountID(c *gin.Context) string {
	if v, ok := c.Get(accountIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return "demo"
}

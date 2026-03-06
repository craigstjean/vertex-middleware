package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/craigstjean/vertex-middleware/config"
	"github.com/craigstjean/vertex-middleware/types"
)

// KeyConfigCtxKey is the context key under which the matched KeyConfig is stored.
const KeyConfigCtxKey = "key_config"

// APIKeyAuth validates the Bearer token against the configured API keys.
func APIKeyAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			abortWithError(c, http.StatusUnauthorized, "missing_api_key", "Missing Authorization header")
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			abortWithError(c, http.StatusUnauthorized, "invalid_api_key", "Authorization header must use Bearer scheme")
			return
		}

		apiKey := strings.TrimPrefix(authHeader, "Bearer ")
		if apiKey == "" {
			abortWithError(c, http.StatusUnauthorized, "missing_api_key", "API key must not be empty")
			return
		}

		keyConfig, ok := cfg.APIKeys[apiKey]
		if !ok {
			abortWithError(c, http.StatusUnauthorized, "invalid_api_key", "Invalid API key")
			return
		}

		c.Set(KeyConfigCtxKey, keyConfig)
		c.Next()
	}
}

func abortWithError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, types.ErrorResponse{
		Error: types.ErrorDetail{
			Message: message,
			Type:    "invalid_request_error",
			Code:    code,
		},
	})
}

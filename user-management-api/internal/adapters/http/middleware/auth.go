package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"sevensolutions-backend/pkg/jwt"
	"sevensolutions-backend/pkg/response"
)

func Auth(jwtService *jwt.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")

		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			response.Unauthorized(c, "missing or invalid token")
			c.Abort()
			return
		}

		token := strings.TrimPrefix(header, prefix)

		userID, err := jwtService.Validate(token)
		if err != nil {
			response.Unauthorized(c, "missing or invalid token")
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

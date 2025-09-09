package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/lanyi1998/DNSlog-GO/internal/config"
	"net/http"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("token")
		_, ok := config.Config.User[token]
		if token == "" || !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": "401",
				"msg":  "Invalid token",
			})
			c.Abort()
			return
		}
		c.Set("token", token)
		c.Next()
	}
}
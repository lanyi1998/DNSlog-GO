package middleware

import (
	"DnsLog/internal/config"
	"github.com/gin-gonic/gin"
	"net/http"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("token")
		_, ok := config.Config.User[token]
		if token == "" || !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"HTTPStatusCode": "401",
				"Msg":            "Invalid token",
			})
			c.Abort()
			return
		}
		c.Set("token", token)
		c.Next()
	}
}

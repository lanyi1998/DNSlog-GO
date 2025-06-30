package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"HTTPStatusCode": "401",
				"Msg":            "token is empty",
			})
			c.Abort()
			return
		}
		c.Set("token", token)
		c.Next()
	}
}

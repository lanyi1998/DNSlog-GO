package handler

import (
	"DnsLog/internal/config"
	"github.com/gin-gonic/gin"
	"net/http"
)

type VerifyTokenReq struct {
	Token string `json:"token"`
}

// VerifyToken 验证Token
func VerifyToken(c *gin.Context) {
	var req VerifyTokenReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "Invalid request body",
		})
		return
	}
	if req.Token == "" {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "Token cannot be empty",
		})
		return
	}
	_, ok := config.Config.User[req.Token]
	if !ok {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusUnauthorized,
			Msg:  "Invalid token",
		})
		return
	}
	c.JSON(http.StatusOK, Response{
		Code: http.StatusOK,
		Msg:  SUCCESS,
		Data: gin.H{
			"subdomain": config.Config.User[req.Token] + "." + config.Config.Dns.Domain,
			"token":     req.Token,
		},
	})
}
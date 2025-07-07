package handler

import (
	"DnsLog/internal/ipwry"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Query(c *gin.Context) {
	ip := c.Param("ip")
	// 调用查询逻辑
	result, err := ipwry.Query(ip)
	if err != nil {
		c.JSON(200, gin.H{
			"code":    http.StatusBadRequest,
			"success": false,
			"data":    err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"code":    http.StatusBadRequest,
		"success": true,
		"data":    result,
	})
}

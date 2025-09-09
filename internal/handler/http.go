package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/lanyi1998/DNSlog-GO/internal/model"
	"net/http"
)

type VerifyHttpReq struct {
	Query string `json:"query"`
}

func VerifyHttp(c *gin.Context) {
	var req VerifyHttpReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "Invalid request body",
		})
	}
	token := c.GetString("token")
	for _, v := range model.UserDnsDataMap.Get(token) {
		if v.Subdomain == req.Query && v.Type == "HTTP" {
			c.JSON(http.StatusOK, Response{
				Code: http.StatusOK,
				Msg:  SUCCESS,
				Data: map[string]interface{}{
					"subdomain": v.Subdomain,
					"ipaddress": v.Ipaddress,
					"time":      v.Time,
					"type":      v.Type,
				},
			})
			return
		}
	}
	c.JSON(http.StatusOK, Response{
		Code: http.StatusOK,
		Msg:  "Not Found",
	})
}

type BulkVerifyHttpReq struct {
	Query []string `json:"query"`
}

func BulkVerifyHttp(c *gin.Context) {
	var req BulkVerifyHttpReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "Invalid request body",
		})
	}
	token := c.GetString("token")
	var respData []string
	for _, v := range model.UserDnsDataMap.Get(token) {
		for _, s := range req.Query {
			if v.Subdomain == s && v.Type == "HTTP" {
				respData = append(respData, s)
			}
		}
	}
	c.JSON(http.StatusOK, Response{
		Code: http.StatusOK,
		Msg:  SUCCESS,
		Data: respData,
	})
}
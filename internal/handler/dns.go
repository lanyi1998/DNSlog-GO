package handler

import (
	"DnsLog/internal/dns"
	"DnsLog/internal/model"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetDnsData 获取DNS数据
func GetDnsData(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: http.StatusOK,
		Msg:  SUCCESS,
		Data: model.UserDnsDataMap.Get(c.GetString("token")),
	})
}

func GetDnsDataAndClean(c *gin.Context) {
	model.UserDnsDataMap.Mu.Lock()
	defer model.UserDnsDataMap.Mu.Unlock()
	c.JSON(http.StatusOK, Response{
		Code: http.StatusOK,
		Msg:  SUCCESS,
		Data: model.UserDnsDataMap.Get(c.GetString("token")),
	})
	model.UserDnsDataMap.Clear(c.GetString("token"))
}

func Clean(c *gin.Context) {
	model.UserDnsDataMap.Clear(c.GetString("token"))
	c.JSON(http.StatusOK, Response{
		Code: http.StatusOK,
		Msg:  SUCCESS,
	})
}

type VerifyDnsReq struct {
	Query string `json:"query"`
}

func VerifyDns(c *gin.Context) {
	var req VerifyDnsReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "Invalid request body",
		})
		return
	}
	data := model.UserDnsDataMap.Get(c.GetString("token"))
	for _, v := range data {
		if v.Subdomain == req.Query {
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

type BulkVerifyDnsReq struct {
	Subdomain []string `json:"subdomain"`
}

func BulkVerifyDns(c *gin.Context) {
	var req BulkVerifyDnsReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "Invalid request body",
		})
	}
	data := model.UserDnsDataMap.Get(c.GetString("token"))
	var verifyData []string
	for _, s := range req.Subdomain {
		for _, v := range data {
			if s == v.Subdomain {
				verifyData = append(verifyData, v.Subdomain)
			}
		}
	}
	c.JSON(http.StatusOK, Response{
		Code: http.StatusOK,
		Msg:  SUCCESS,
		Data: verifyData,
	})
}

type SetARecordReq struct {
	Domain string `json:"domain"`
	Ip     string `json:"ip"`
}

func SetARecord(c *gin.Context) {
	var req SetARecordReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "Invalid request body",
		})
	}
	if err := dns.SetARecord(c.GetString("token"), req.Domain, req.Ip); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusOK,
			Msg:  SUCCESS,
		})
	}
}

type SetTXTRecordReq struct {
	Domain string `json:"domain"`
	Txt    string `json:"txt"`
}

func SetTXTRecord(c *gin.Context) {
	var req SetTXTRecordReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "Invalid request body",
		})
		return
	}
	if req.Domain == "" || req.Txt == "" {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  "Domain and TXT must not be empty",
		})
		return
	}
	dns.SetTXTRecord(c.GetString("token"), req.Domain, req.Txt)
	c.JSON(http.StatusOK, Response{
		Code: http.StatusOK,
		Msg:  SUCCESS,
	})
}
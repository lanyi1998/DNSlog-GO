package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lanyi1998/DNSlog-GO/internal/dns"
	"github.com/lanyi1998/DNSlog-GO/internal/model"
)

// GetDnsData 获取DNS数据
func GetDnsData(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code: http.StatusOK,
		Msg:  SUCCESS,
		Data: model.UserDnsDataMap.Get(c.GetString("token")),
	})
	return
}

func GetDnsDataAndClean(c *gin.Context) {
	model.UserDnsDataMap.Clear(c.GetString("token"))
	c.JSON(http.StatusOK, Response{
		Code: http.StatusOK,
		Msg:  SUCCESS,
		Data: model.UserDnsDataMap.Get(c.GetString("token")),
	})
	return
}

func Clean(c *gin.Context) {
	model.UserDnsDataMap.Clear(c.GetString("token"))
	c.JSON(http.StatusOK, Response{
		Code: http.StatusOK,
		Msg:  SUCCESS,
	})
	return
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
	return
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
		return
	}
	data := model.UserDnsDataMap.Get(c.GetString("token"))
	var verifyData map[string]struct{}
	verifyData = make(map[string]struct{})
	for _, s := range req.Subdomain {
		for _, v := range data {
			if s == v.Subdomain {
				verifyData[v.Subdomain] = struct{}{}
			}
		}
	}
	var dataList []string
	for key := range verifyData {
		dataList = append(dataList, key)
	}
	c.JSON(http.StatusOK, Response{
		Code: http.StatusOK,
		Msg:  SUCCESS,
		Data: dataList,
	})
	return
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
		return
	}
	if err := dns.SetARecord(c.GetString("token"), req.Domain, req.Ip); err != nil {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusInternalServerError,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: http.StatusOK,
		Msg:  SUCCESS,
	})
	return
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
	return
}
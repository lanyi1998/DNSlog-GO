package handler

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/lanyi1998/DNSlog-GO/internal/config"
	"github.com/lanyi1998/DNSlog-GO/internal/ipwry"
	"github.com/lanyi1998/DNSlog-GO/internal/model"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

var SUCCESS = "success"

func SafeDumpRequest(c *gin.Context) ([]byte, error) {
	// 读取原始请求体
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}

	// 恢复请求体,供后续使用
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// 执行 dump,再次恢复请求体
	dump, err := httputil.DumpRequest(c.Request, true)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return dump, err
}

func NoRoute(c *gin.Context) {
	for k, v := range config.Config.User {
		if strings.HasPrefix(c.Request.URL.Path, "/"+v+"/") {
			dump, err := SafeDumpRequest(c)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Server Error"})
				return
			}
			IpLocation, _ := ipwry.Query(c.ClientIP())
			model.UserDnsDataMap.Set(k, model.DnsInfo{
				Type:       "HTTP",
				Subdomain:  c.Request.URL.Path,
				Ipaddress:  c.ClientIP(),
				Time:       time.Now().Unix(),
				IpLocation: IpLocation,
				Request:    string(dump),
			})
			break
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"message": "404 Not Found"})
}
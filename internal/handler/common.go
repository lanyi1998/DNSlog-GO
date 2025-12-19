package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/lanyi1998/DNSlog-GO/internal/config"
	"github.com/lanyi1998/DNSlog-GO/internal/ipwry"
	"github.com/lanyi1998/DNSlog-GO/internal/model"
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

func NoRoute(c *gin.Context) {
	for k, v := range config.Config.User {
		if strings.HasPrefix(c.Request.URL.Path, "/"+v+"/") {
			dump, err := httputil.DumpRequest(c.Request, false)
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
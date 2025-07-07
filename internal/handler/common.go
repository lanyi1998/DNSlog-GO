package handler

import (
	"DnsLog/internal/config"
	"DnsLog/internal/ipwry"
	"DnsLog/internal/model"
	"github.com/gin-gonic/gin"
	"net/http"
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
			IpLocation, _ := ipwry.Query(c.ClientIP())
			model.UserDnsDataMap.Set(k, model.DnsInfo{
				Type:       "HTTP",
				Subdomain:  c.Request.URL.Path,
				Ipaddress:  c.ClientIP(),
				Time:       time.Now().Unix(),
				IpLocation: IpLocation,
			})
			break
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"message": "404 Not Found"})
}
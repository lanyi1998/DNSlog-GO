package router

import (
	"DnsLog/internal/config"
	"DnsLog/internal/handler"
	"DnsLog/internal/middleware"
	"embed"
	"github.com/gin-gonic/gin"
	"html/template"
	"io/fs"
	"net/http"
)

//go:embed resources/template/*.html
var tmplFS embed.FS

//go:embed resources/js/*.js
var jsFS embed.FS

func SetupRouter() *gin.Engine {
	r := gin.Default()

	if !config.Config.Http.ConsoleDisable {
		tmpl := template.Must(template.New("").Delims("[[", "]]").ParseFS(tmplFS, "resources/template/*.html"))
		r.SetHTMLTemplate(tmpl)
		// 修改静态文件服务配置，使用子文件系统
		jsFiles, _ := fs.Sub(jsFS, "resources/js")
		r.StaticFS("/js", http.FS(jsFiles))
		r.GET("/", func(c *gin.Context) {
			c.HTML(200, "index.html", gin.H{})
		})
	}

	api := r.Group("/api")

	// 公开接口
	api.Any("/verifyToken", handler.VerifyToken)

	// IP查询接口
	api.POST("/ip/query", handler.Query)
	api.GET("/ip/:ip", handler.Query)

	// 需要鉴权的接口
	authApi := api.Group("/", middleware.AuthMiddleware())
	{
		authApi.Any("/getDnsData", handler.GetDnsData)
		authApi.Any("/clean", handler.Clean)
		authApi.Any("/getDnsData_clear", handler.GetDnsDataAndClean)
		authApi.Any("/verifyDns", handler.VerifyDns)
		authApi.Any("/bulkVerifyDns", handler.BulkVerifyDns)
		authApi.Any("/verifyHttp", handler.VerifyHttp)
		authApi.Any("/bulkVerifyHttp", handler.BulkVerifyHttp)
		authApi.Any("/setARecord", handler.SetARecord)
		authApi.Any("/setTXTRecord", handler.SetTXTRecord)
	}

	r.NoRoute(handler.NoRoute)
	return r
}
package router

import (
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

	tmpl := template.Must(template.New("").Delims("[[", "]]").ParseFS(tmplFS, "resources/template/*.html"))
	r.SetHTMLTemplate(tmpl)

	// 修改静态文件服务配置，使用子文件系统
	jsFiles, _ := fs.Sub(jsFS, "resources/js")
	r.StaticFS("/js", http.FS(jsFiles))
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"title": "首页",
		})
	})

	api := r.Group("/api")
	api.Any("/verifyToken", handler.VerifyToken)
	api.Any("/getDnsData", middleware.AuthMiddleware(), handler.GetDnsData)
	api.Any("/clean", middleware.AuthMiddleware(), handler.Clean)
	api.Any("/getDnsData_clear", middleware.AuthMiddleware(), handler.GetDnsDataAndClean)
	api.Any("/verifyDns", middleware.AuthMiddleware(), handler.VerifyDns)
	api.Any("/bulkVerifyDns", middleware.AuthMiddleware(), handler.BulkVerifyDns)
	api.Any("/verifyHttp", middleware.AuthMiddleware(), handler.VerifyHttp)
	api.Any("/bulkVerifyHttp", middleware.AuthMiddleware(), handler.BulkVerifyHttp)

	r.NoRoute(handler.NoRoute)
	return r
}
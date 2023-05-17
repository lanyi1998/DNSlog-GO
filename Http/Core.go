package Http

import (
	"DnsLog/Core"
	"embed"
	"fmt"
	"log"
	"net/http"
)

//go:embed template
var template embed.FS

func ListingHttpManagementServer() {
	mux := http.NewServeMux()
	if !Core.Config.HTTP.ConsoleDisable {
		mux.Handle("/template/", http.FileServer(http.FS(template)))
		mux.HandleFunc("/", index)
	}
	mux.HandleFunc("/api/verifyToken", verifyTokenApi)
	mux.HandleFunc("/api/getDnsData", GetDnsData)
	mux.HandleFunc("/api/Clean", Clean)
	mux.HandleFunc("/api/verifyDns", verifyDns)
	mux.HandleFunc("/api/bulkVerifyDns", BulkVerifyDns)
	mux.HandleFunc("/api/verifyHttp", verifyHttp)
	mux.HandleFunc("/api/BulkVerifyHttp", BulkVerifyHttp)
	for key, _ := range Core.Config.HTTP.User {
		mux.HandleFunc("/"+key+"/", HttpRequestLog)
	}

	log.Println("Http Listing Start...")
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", Core.Config.HTTP.Port),
		Handler: mux,
	}
	log.Println("Http address: http://" + "0.0.0.0:" + Core.Config.HTTP.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

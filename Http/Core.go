package Http

import (
	"../Core"
	"log"
	"net/http"
)

var DnsData = make(map[string]string)

func ListingHttpManagementServer() {
	mux := http.NewServeMux()
	if !Core.Config.HTTP.ConsoleDisable {
		mux.Handle("/template/", http.FileServer(http.Dir("")))
		mux.HandleFunc("/", index)
	}
	mux.HandleFunc("/api/verifyToken", verifyToken)
	mux.HandleFunc("/api/getDnsData", GetDnsData)
	mux.HandleFunc("/api/Clean", Clean)
	mux.HandleFunc("/api/verifyDns", verifyDns)
	println("Http Listing Start...")
	server := &http.Server{
		Addr:    ":" + Core.Config.HTTP.Port,
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

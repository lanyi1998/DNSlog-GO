package main

import (
	"DnsLog/Core"
	"DnsLog/Dns"
	"DnsLog/Http"
	"gopkg.in/gcfg.v1"
)



func main() {
	var _ = gcfg.ReadFileInto(&Core.Config, "./config.ini")
	go Dns.ListingDnsServer()
	Http.ListingHttpManagementServer()
}

//GOOS=windows GOARCH=amd64 go build main.go
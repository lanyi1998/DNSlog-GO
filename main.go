package main

import (
	"DnsLog/Core"
	"DnsLog/Dns"
	"DnsLog/Http"
	"github.com/gogf/gf/encoding/gyaml"
	"io/ioutil"
	"log"
)

//GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-w -s" main.go

func main() {
	//var err = gcfg.ReadFileInto(&Core.Config, "./config.ini")
	var ConfigBody, err = ioutil.ReadFile("./config.yaml")
	if err != nil {
		log.Fatalln(err.Error())
	}
	err = gyaml.DecodeTo(ConfigBody, &Core.Config)
	if err != nil {
		log.Fatalln(err.Error())
	}
	go Dns.ListingDnsServer()
	Http.ListingHttpManagementServer()
}

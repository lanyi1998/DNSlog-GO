package Dns

import (
	"DnsLog/Core"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

var DnsData = make(map[string][]DnsInfo)

var DnsDataRwLock sync.RWMutex

type DnsInfo struct {
	Type      string
	Subdomain string
	Ipaddress string
	Time      int64
}

var D DnsInfo

// ListingDnsServer 监听dns端口
func ListingDnsServer() {
	//if runtime.GOOS != "windows" && os.Geteuid() != 0 {
	//	log.Fatal("Please run as root")
	//}
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 53})
	if err != nil {
		log.Fatal(err.Error())
	}
	defer conn.Close()
	log.Println("DNS Listing Start...")
	for {
		buf := make([]byte, 512)
		_, addr, _ := conn.ReadFromUDP(buf)
		var msg dnsmessage.Message
		if err := msg.Unpack(buf); err != nil {
			fmt.Println(err)
			continue
		}
		go serverDNS(addr, conn, msg)
	}
}

func serverDNS(addr *net.UDPAddr, conn *net.UDPConn, msg dnsmessage.Message) {
	if len(msg.Questions) < 1 {
		return
	}
	question := msg.Questions[0]
	var (
		queryNameStr = strings.ToLower(question.Name.String())
		queryType    = question.Type
		queryName, _ = dnsmessage.NewName(queryNameStr)
		resource     dnsmessage.Resource
		queryDoamin  = strings.Split(strings.Replace(queryNameStr, fmt.Sprintf(".%s.", Core.Config.DNS.Domain), "", 1), ".")
	)

	//域名过滤
	if strings.Contains(queryNameStr, Core.Config.DNS.Domain) {
		user := Core.GetUser(queryDoamin[len(queryDoamin)-1])
		D.Set(user, DnsInfo{
			Type:      "DNS",
			Subdomain: queryNameStr[:len(queryNameStr)-1],
			Ipaddress: addr.IP.String(),
			Time:      time.Now().Unix(),
		})
	}
	switch queryType {
	case dnsmessage.TypeA:
		resource = NewAResource(queryName, [4]byte{127, 0, 0, 1})
	default:
		resource = NewAResource(queryName, [4]byte{127, 0, 0, 1})
	}
	// send response
	msg.Response = true
	msg.Answers = append(msg.Answers, resource)
	Response(addr, conn, msg)
}

// Response return
func Response(addr *net.UDPAddr, conn *net.UDPConn, msg dnsmessage.Message) {
	packed, err := msg.Pack()
	if err != nil {
		fmt.Println(err)
		return
	}
	if _, err := conn.WriteToUDP(packed, addr); err != nil {
		fmt.Println(err)
	}
}

func NewAResource(query dnsmessage.Name, a [4]byte) dnsmessage.Resource {
	return dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  query,
			Class: dnsmessage.ClassINET,
			TTL:   0,
		},
		Body: &dnsmessage.AResource{
			A: a,
		},
	}
}

func (d *DnsInfo) Set(token string, data DnsInfo) {
	DnsDataRwLock.Lock()
	defer DnsDataRwLock.Unlock()
	if DnsData[token] == nil {
		DnsData[token] = []DnsInfo{data}
	} else {
		DnsData[token] = append(DnsData[token], data)
	}
}

func (d *DnsInfo) Get(token string) string {
	DnsDataRwLock.RLock()
	defer DnsDataRwLock.RUnlock()
	res := ""
	if DnsData[token] != nil {
		v, _ := json.Marshal(DnsData[token])
		res = string(v)
	}
	if res == "" {
		res = "null"
	}
	return res
}

func (d *DnsInfo) Clear(token string) {
	DnsData[token] = []DnsInfo{}
	DnsData["other"] = []DnsInfo{}
}

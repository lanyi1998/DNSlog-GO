package Dns

import (
	"DnsLog/Core"
	"encoding/json"
	"fmt"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var DnsData []DnsInfo

var rw sync.RWMutex

type DnsInfo struct {
	Subdomain string
	Ipaddress string
	Time      int64
}

var D DnsInfo

//监听dns端口
func ListingDnsServer() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 53})
	if err != nil {
		println("DNS port listing error,Please run as root")
		os.Exit(0)
	}
	defer conn.Close()
	fmt.Println("DNS Listing Start...")
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
		//queryTypeStr = question.Type.String()
		queryNameStr = question.Name.String()
		queryType    = question.Type
		queryName, _ = dnsmessage.NewName(queryNameStr)
	)
	//域名过滤，避免网络扫描
	if strings.Contains(queryNameStr, Core.Config.Dns.Domain) {
		D.Set(DnsInfo{
			Subdomain: queryNameStr[:len(queryNameStr)-1],
			Ipaddress: addr.IP.String(),
			Time:      time.Now().Unix(),
		})
	}else{
		return
	}
	var resource dnsmessage.Resource
	switch queryType {
	case dnsmessage.TypeA:
		resource = NewAResource(queryName, [4]byte{127, 0, 0, 1})
	default:
		//fmt.Printf("not support dns queryType: [%s] \n", queryTypeStr)
		return
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
			TTL:   600,
		},
		Body: &dnsmessage.AResource{
			A: a,
		},
	}
}

func (d *DnsInfo) Set(data DnsInfo) {
	rw.Lock()
	DnsData = append(DnsData, data)
	rw.Unlock()
}

func (d *DnsInfo) Get() string {
	rw.RLock()
	v, _ := json.Marshal(DnsData)
	rw.RUnlock()
	return string(v)
}

func (d *DnsInfo) Clear() {
	DnsData = (DnsData)[0:0]
}

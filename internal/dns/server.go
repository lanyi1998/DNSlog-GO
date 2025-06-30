package dns

import (
	"DnsLog/internal/config"
	"DnsLog/internal/logger"
	"DnsLog/internal/model"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"strings"
	"sync"
	"time"
)

var DnsARecordMap = sync.Map{}
var DnsTXTRecordMap = sync.Map{}

// ListingDnsServer 监听dns端口
func ListingDnsServer() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 53})
	if err != nil {
		logger.Logger.Error("Dns port Listen error", zap.Error(err))
	}
	defer conn.Close()
	logger.Logger.Info("DNS Listing start success...")
	for {
		buf := make([]byte, 512)
		_, addr, _ := conn.ReadFromUDP(buf)
		var msg dnsmessage.Message
		if err := msg.Unpack(buf); err != nil {
			logger.Logger.Error("DNS unpack error", zap.Error(err))
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
		queryDomain  = strings.Split(strings.Replace(queryNameStr, fmt.Sprintf(".%s.", config.Config.DNS.Domain), "", 1), ".")
	)

	//过滤非绑定域名请求
	if strings.Contains(queryNameStr, config.Config.DNS.Domain) {
		user := config.Config.GetUserByDomain(queryDomain[len(queryDomain)-1])
		model.UserDnsDataMap.Set(user, model.DnsInfo{
			Type:      "DNS",
			Subdomain: queryNameStr[:len(queryNameStr)-1],
			Ipaddress: addr.IP.String(),
			Time:      time.Now().Unix(),
		})
	}

	switch queryType {
	case dnsmessage.TypeA:
		var DnsValue interface{}
		DnsARecordMap.Range(func(key, value interface{}) bool {
			if strings.HasSuffix(queryNameStr, key.(string)) {
				DnsValue = value.([4]byte)
				return false
			}
			return true
		})
		if DnsValue != nil {
			resource = NewAResource(queryName, DnsValue.([4]byte))
		} else {
			resource = NewAResource(queryName, [4]byte{127, 0, 0, 1})
		}
	case dnsmessage.TypeTXT:
		var txtValue interface{}
		DnsTXTRecordMap.Range(func(key, value interface{}) bool {
			if strings.HasSuffix(queryNameStr, key.(string)) {
				txtValue = value.(string)
				return false
			}
			return true
		})
		if txtValue != nil {
			resource = dnsmessage.Resource{
				Header: dnsmessage.ResourceHeader{
					Name:  queryName,
					Class: dnsmessage.ClassINET,
					TTL:   0,
				},
				Body: &dnsmessage.TXTResource{
					TXT: []string{txtValue.(string)},
				},
			}
		} else {
			resource = dnsmessage.Resource{
				Header: dnsmessage.ResourceHeader{
					Name:  queryName,
					Class: dnsmessage.ClassINET,
					TTL:   0,
				},
				Body: &dnsmessage.TXTResource{
					TXT: []string{""},
				},
			}
		}
	default:
		resource = NewAResource(queryName, [4]byte{127, 0, 0, 1})
	}

	// send response
	msg.Response = true
	msg.Answers = append(msg.Answers, resource)
	Response(addr, conn, msg)
}

func Response(addr *net.UDPAddr, conn *net.UDPConn, msg dnsmessage.Message) {
	packed, err := msg.Pack()
	if err != nil {
		logger.Logger.Error("DNS pack error", zap.Error(err))
		return
	}
	if _, err := conn.WriteToUDP(packed, addr); err != nil {
		logger.Logger.Error("DNS write error", zap.Error(err))
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
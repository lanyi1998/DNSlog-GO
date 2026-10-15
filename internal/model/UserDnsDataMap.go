package model

import (
	"sync"

	"github.com/lanyi1998/DNSlog-GO/internal/config"
)

var UserDnsDataMap *DnsDataMap

type DnsDataMap struct {
	userDnsData       sync.Map
	Mu                sync.RWMutex
	TokenSubDomainMap map[string]string
	SubDomainTokenMap map[string]string
}

func NewDnsDataMap(conf config.UserConfig) *DnsDataMap {
	TokenSubDomainMap := make(map[string]string)
	SubDomainTokenMap := make(map[string]string)
	for token, subdomain := range conf {
		TokenSubDomainMap[token] = subdomain
		SubDomainTokenMap[subdomain] = token
	}
	return &DnsDataMap{
		TokenSubDomainMap: TokenSubDomainMap,
		SubDomainTokenMap: SubDomainTokenMap,
	}
}

func (d *DnsDataMap) Get(token string) []DnsInfo {
	value, ok := d.userDnsData.Load(token)
	if ok {
		return value.([]DnsInfo)
	}
	return []DnsInfo{}
}

func (d *DnsDataMap) Set(token string, data DnsInfo) {
	d.Mu.Lock()
	defer d.Mu.Unlock()
	value, ok := d.userDnsData.Load(token)
	if ok {
		d.userDnsData.Store(token, append(value.([]DnsInfo), data))
	} else {
		d.userDnsData.Store(token, []DnsInfo{data})
	}

}

func (d *DnsDataMap) Clear(token string) {
	d.Mu.Lock()
	defer d.Mu.Unlock()
	d.userDnsData.Delete(token)
}

// GetAndClear atomically removes and returns the token's current records.
func (d *DnsDataMap) GetAndClear(token string) []DnsInfo {
	d.Mu.Lock()
	defer d.Mu.Unlock()
	data := d.Get(token)
	d.userDnsData.Delete(token)
	return data
}

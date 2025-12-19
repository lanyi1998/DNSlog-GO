package model

import "sync"

type DnsInfo struct {
	Type       string
	Subdomain  string
	Ipaddress  string
	IpLocation string
	Time       int64
	Request    string
}

var UserDnsDataMap = &userDnsDataMap{}

type userDnsDataMap struct {
	userDnsData sync.Map
	Mu          sync.Mutex
}

func (u *userDnsDataMap) Get(token string) []DnsInfo {
	value, ok := u.userDnsData.Load(token)
	if ok {
		return value.([]DnsInfo)
	} else {
		return []DnsInfo{}
	}
}

func (u *userDnsDataMap) Set(token string, data DnsInfo) {
	u.Mu.Lock()
	defer u.Mu.Unlock()
	value, ok := u.userDnsData.Load(token)
	if ok {
		u.userDnsData.Store(token, append(value.([]DnsInfo), data))
	} else {
		u.userDnsData.Store(token, []DnsInfo{data})
	}
}

func (u *userDnsDataMap) Clear(token string) {
	u.Mu.Lock()
	defer u.Mu.Unlock()
	u.userDnsData.Delete(token)
}
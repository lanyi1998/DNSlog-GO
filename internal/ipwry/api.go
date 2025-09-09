package ipwry

import (
	"errors"
	"net"
	"sync"
)

var qqWry = InitFile()

func InitFile() QQwry {
	IPData.InitIPData()
	return NewQQwry()
}

var lock sync.Mutex

func Query(ip string) (string, error) {
	lock.Lock()
	defer lock.Unlock()
	i := net.ParseIP(ip)
	if i == nil || i.To4() == nil {
		return "", errors.New("invalid IP address")
	}
	result := qqWry.Find(ip)
	return result.Country + "/" + result.Area, nil
}
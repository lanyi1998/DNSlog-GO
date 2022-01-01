package Core

import (
	"math/rand"
	"strings"
	"time"
)

var Config = struct {
	HTTP struct {
		Port           string
		Token          string
		ConsoleDisable bool
	}
	Dns struct {
		Domain string
	}
}{}

var User = make(map[string]string)

func VerifyToken(token string) bool {
	flag := false
	for v := range User {
		if v == token {
			flag = true
		}
	}
	return flag
}

func GetUser(domain string) string {
	user := "other"
	for i, v := range User {
		if strings.Contains(domain, v) {
			user = i
			break
		}
	}
	return user
}

func GetRandStr() string {
	bytes := []byte(`abcdefghijklmnopqrstuvwxyz1234567890`)
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 6; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

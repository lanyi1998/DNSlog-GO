package Core

import (
	"strings"
)

var Config = struct {
	HTTP struct {
		Port           string            `yaml:"port"`
		User           map[string]string `yaml:"user"`
		ConsoleDisable bool              `yaml:"consoleDisable"`
	} `yaml:"HTTP"`
	DNS struct {
		Domain string `yaml:"domain"`
	} `yaml:"Dns"`
}{}

func VerifyToken(token string) bool {
	flag := false
	for v := range Config.HTTP.User {
		if v == token {
			flag = true
		}
	}
	return flag
}

func GetUser(domain string) string {
	user := "other"
	for i, v := range Config.HTTP.User {
		if strings.Contains(domain, v) {
			user = i
			break
		}
	}
	return user
}

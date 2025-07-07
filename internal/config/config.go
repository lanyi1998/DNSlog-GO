package config

import "strings"

var Config = &config{}

type config struct {
	Http HttpConfig `yaml:"Http"`
	User UserConfig `yaml:"User"`
	Dns  DnsConfig  `yaml:"Dns"`
}

type HttpConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	ConsoleDisable bool   `yaml:"consoleDisable"`
}

type UserConfig map[string]string

type DnsConfig struct {
	Domain  string            `yaml:"domain"`
	ARecord map[string]string `yaml:"ARecord"` // ARecord 也是键值对，使用 map[string]string
}

func (c *config) GetUserByDomain(domain string) string {
	for i, v := range c.User {
		if strings.Contains(domain, v) {
			return i
		}
	}
	return "other"
}
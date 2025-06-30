package config

import "strings"

var Config = &config{}

type config struct {
	HTTP struct {
		Host           string            `yaml:"host"`
		Port           string            `yaml:"port"`
		User           map[string]string `yaml:"user"`
		ConsoleDisable bool              `yaml:"consoleDisable"`
	} `yaml:"HTTP"`
	DNS struct {
		Domain string `yaml:"domain"`
	} `yaml:"Dns"`
}

func (c *config) GetUserByDomain(domain string) string {
	for i, v := range c.HTTP.User {
		if strings.Contains(domain, v) {
			return i
		}
	}
	return "other"
}
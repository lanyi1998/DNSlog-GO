package main

import (
	"DnsLog/internal/config"
	"DnsLog/internal/dns"
	"DnsLog/internal/handler"
	"DnsLog/internal/logger"
	"DnsLog/internal/router"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"os"
)

func readConfig() {
	file, err := os.ReadFile("config.yaml")
	if err != nil {
		logger.Logger.Panic("Failed to read config file", zap.Error(err))
	}
	err = yaml.Unmarshal(file, &config.Config)
	if err != nil {
		logger.Logger.Panic("Failed to unmarshal config", zap.Error(err))
	}
}

func main() {
	logger.InitLogger()
	defer logger.Sync()
	readConfig()

	go dns.ListingDnsServer()
	gin.SetMode(gin.ReleaseMode)
	r := router.SetupRouter()
	logger.Logger.Info("Http Server start...")
	err := handler.MultiProtocolListener(r, config.Config.Http.Port)
	if err != nil {
		logger.Logger.Panic("Failed to start server", zap.Error(err))
	}
}

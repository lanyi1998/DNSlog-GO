package main

import (
	"github.com/gin-gonic/gin"
	"github.com/lanyi1998/DNSlog-GO/internal/config"
	"github.com/lanyi1998/DNSlog-GO/internal/dns"
	"github.com/lanyi1998/DNSlog-GO/internal/handler"
	"github.com/lanyi1998/DNSlog-GO/internal/logger"
	"github.com/lanyi1998/DNSlog-GO/internal/router"
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
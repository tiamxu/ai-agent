package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/tiamxu/ai-agent/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	if err := cfg.Initial(); err != nil {
		log.Fatalf("Config initialization failed: %v", err)
	}
	r := gin.Default()
	if err := r.Run(cfg.HttpSrv.Address); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}

}

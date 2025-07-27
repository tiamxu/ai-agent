package main

import (
	"context"
	"os/signal"
	"syscall"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/tiamxu/ai-agent/api"
	"github.com/tiamxu/ai-agent/config"
	"github.com/tiamxu/ai-agent/llm"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	if err := cfg.Initial(); err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}

	llmConfig := cfg.GetActiveLLMconfig()
	chatModel := llm.NewOpenAIChatModel(ctx, llmConfig)

	llmService := llm.NewLLMService(chatModel)

	handler := api.NewChatHandler(cfg, llmService)
	r := gin.Default()

	// 添加路由
	r.GET("/health", handler.Health)
	r.POST("/api/chat", handler.Chat)
	r.POST("/api/stream", handler.Stream)
	// 在goroutine中启动HTTP服务
	go func() {
		log.Printf("启动HTTP服务，监听地址: %s", cfg.HttpSrv.Address)
		if err := r.Run(cfg.HttpSrv.Address); err != nil {
			log.Fatalf("HTTP服务启动失败: %v", err)
		}
	}()
	<-ctx.Done()
	log.Println("接收到终止信号，开始关闭服务...")

	log.Println("服务已关闭")

}

package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
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
		hlog.Fatalf("加载配置失败: %v", err)
	}
	if err := cfg.Initial(); err != nil {
		hlog.Fatalf("配置初始化失败: %v", err)
	}

	llmConfig := cfg.GetActiveLLMconfig()
	chatModel := llm.NewOpenAIChatModel(ctx, llmConfig)

	llmService := llm.NewLLMService(chatModel)

	handler := api.NewChatHandler(cfg, llmService)
	h := server.Default(
		server.WithHostPorts(cfg.HttpSrv.Address),
		server.WithSenseClientDisconnection(true),
	)

	// 添加路由
	h.GET("/health", func(c context.Context, ctx *app.RequestContext) {
		handler.Health(c, ctx)
	})
	h.POST("/api/chat", func(c context.Context, ctx *app.RequestContext) {
		handler.Chat(c, ctx)
	})

	// 在goroutine中启动HTTP服务
	go func() {
		hlog.Infof("启动HTTP服务，监听地址: %s", cfg.HttpSrv.Address)
		if err := h.Run(); err != nil {
			hlog.Fatalf("HTTP服务启动失败: %v", err)
		}
	}()

	<-ctx.Done()
	hlog.Info("接收到终止信号，开始关闭服务...")

	hlog.Info("服务已关闭")
}

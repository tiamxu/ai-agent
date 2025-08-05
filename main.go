package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/tiamxu/ai-agent/agent"
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
	fmt.Println("llmconfig:", llmConfig)
	// chatModel := llm.NewOpenAIChatModel(ctx, llmConfig)
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: llmConfig.BaseURL,
		Model:   llmConfig.LLMModel,
		APIKey:  llmConfig.APIKey,
	})
	if err != nil {
		hlog.Fatalf("创建ChatModel失败: %v", err)
	}
	hlog.Info("chatModel创建成功")

	dnsTool := agent.NewDNSTool("http://localhost:8801")
	aiAgent, err := agent.NewChainAgent(ctx, chatModel, dnsTool)
	if err != nil {
		hlog.Fatalf("创建Agent失败: %v", err)

	}

	hlog.Info("agent初始化成功")

	llmService := llm.NewLLMService(chatModel)

	handler := api.NewChatApi(cfg, llmService, aiAgent)
	h := server.Default(
		server.WithHostPorts(cfg.HttpSrv.Address),
		server.WithSenseClientDisconnection(true),
		server.WithReadTimeout(10*time.Second),
		server.WithWriteTimeout(30*time.Second),
		server.WithMaxRequestBodySize(10<<20),
	)

	// 添加路由
	h.GET("/ping", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, utils.H{"message": "pong"})
	})
	h.POST("/api/chat", func(c context.Context, ctx *app.RequestContext) {
		handler.Chat(c, ctx)
	})

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

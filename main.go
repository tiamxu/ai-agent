package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/tiamxu/ai-agent/api"
	"github.com/tiamxu/ai-agent/config"
	"github.com/tiamxu/ai-agent/llm"
	"github.com/tiamxu/ai-agent/tools"
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
	// chatModel := llm.NewOpenAIChatModel(ctx, llmConfig)
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: llmConfig.BaseURL,
		Model:   llmConfig.LLMModel,
		APIKey:  llmConfig.APIKey,
	})
	if err != nil {
		hlog.Fatalf("创建ChatModel失败: %v", err)
	}
	dnsTool := tools.NewDNSTool("http://localhost:8800/")
	toolInfos := make([]*schema.ToolInfo, 0)
	info, err := dnsTool.Info(ctx)
	hlog.Infof("info:%s", info)
	if err != nil {
		hlog.Fatalf("获取工具信息失败: %v", err)
	}
	toolInfos = append(toolInfos, info)

	if err := chatModel.BindTools(toolInfos); err != nil {
		hlog.Fatalf("绑定工具失败: %v", err)
	}
	// 创建工具节点
	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{dnsTool},
	})
	if err != nil {
		hlog.Fatalf("创建工具节点失败: %v", err)
	}

	// 构建处理链
	chain := compose.NewChain[[]*schema.Message, []*schema.Message]()
	chain.
		AppendChatModel(chatModel, compose.WithNodeName("chat_model")).
		AppendToolsNode(toolsNode, compose.WithNodeName("tools"))

	// 编译Agent
	agent, err := chain.Compile(ctx)
	if err != nil {
		hlog.Fatalf("编译Agent失败: %v", err)
	}
	resp, err := agent.Invoke(ctx, []*schema.Message{
		{
			Role:    schema.User,
			Content: "帮我添加一个test域名，解析到192.168.1.100",
		},
	})
	if err != nil {
		hlog.Errorf("agent.Invoke failed, err=%v", err)
		return

	}

	for idx, msg := range resp {
		hlog.Infof("\n")
		hlog.Infof("message %d: %s: %s", idx, msg.Role, msg.Content)
	}
	llmService := llm.NewLLMService(chatModel)

	handler := api.NewChatHandler(cfg, llmService)
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

package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type Agent struct {
	chatModel *openai.ChatModel
	tools     []tool.BaseTool
	agent     compose.Runnable[[]*schema.Message, []*schema.Message]
}

func NewAgent(ctx context.Context, chatModel *openai.ChatModel, tools ...tool.BaseTool) (*Agent, error) {
	//  绑定工具
	toolInfos := make([]*schema.ToolInfo, 0, len(tools))
	for _, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			return nil, fmt.Errorf("获取工具信息失败: %w", err)
		}
		toolInfos = append(toolInfos, info)
	}

	if err := chatModel.BindTools(toolInfos); err != nil {
		return nil, fmt.Errorf("绑定工具失败: %w", err)
	}

	//  创建工具节点
	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{Tools: tools})
	if err != nil {
		return nil, fmt.Errorf("创建工具节点失败: %w", err)
	}

	// 构建处理链
	chain := compose.NewChain[[]*schema.Message, []*schema.Message]()
	chain.
		AppendChatModel(chatModel,
			compose.WithNodeName("chat_model"),
		).
		AppendToolsNode(toolsNode, compose.WithNodeName("tools"))

	// 编译Agent
	agent, err := chain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("编译Agent失败: %w", err)
	}
	return &Agent{
		chatModel: chatModel,
		tools:     tools,
		agent:     agent,
	}, nil
}

func (a *Agent) Invoke(ctx context.Context, in []*schema.Message) ([]*schema.Message, error) {

	return a.agent.Invoke(ctx, in)
}

func (a *Agent) AddTool(ctx context.Context, tool tool.BaseTool) error {
	// 获取工具信息
	info, err := tool.Info(ctx)
	if err != nil {
		return err
	}

	// 绑定新工具
	if err := a.chatModel.BindTools([]*schema.ToolInfo{info}); err != nil {
		return err
	}

	// 更新工具列表
	a.tools = append(a.tools, tool)

	// 需要重新编译Agent
	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{Tools: a.tools})
	if err != nil {
		return err
	}

	chain := compose.NewChain[[]*schema.Message, []*schema.Message]()
	chain.
		AppendChatModel(a.chatModel, compose.WithNodeName("chat_model")).
		AppendToolsNode(toolsNode, compose.WithNodeName("tools"))

	agent, err := chain.Compile(ctx)
	if err != nil {
		return err
	}

	a.agent = agent
	return nil
}

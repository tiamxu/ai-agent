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

func NewChainAgent(ctx context.Context, chatModel *openai.ChatModel, tools ...tool.BaseTool) (*Agent, error) {
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

	// //  创建工具节点
	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{Tools: tools})
	if err != nil {
		return nil, fmt.Errorf("创建工具节点失败: %w", err)
	}
	lambda := compose.ToList[*schema.Message]()

	branchCond := func(ctx context.Context, msgs []*schema.Message) (string, error) {
		fmt.Println("msgs:", msgs)
		if len(msgs) == 0 {
			return "fallback", nil
		}
		lastMsg := msgs[len(msgs)-1]
		fmt.Println("tools:", toolInfos)

		if lastMsg.ToolCalls != nil && len(lastMsg.ToolCalls) > 0 {
			return "tools", nil
		}
		return "fallback", nil
	}
	b1 := compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) ([]*schema.Message, error) {
		fmt.Println("####b1 b1")
		resp, err := toolsNode.Invoke(ctx, input[len(input)-1])
		if err != nil {
			return nil, fmt.Errorf("创建工具b1节点失败: %w", err)
		}
		return resp, nil
	})
	b2 := compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) ([]*schema.Message, error) {
		fmt.Println("####b2 b2")
		resp, err := chatModel.Generate(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("创建工具b2节点失败: %w", err)
		}
		return []*schema.Message{resp}, nil
	})

	// 构建处理链
	chain := compose.NewChain[[]*schema.Message, []*schema.Message]()
	chain.AppendChatModel(chatModel, compose.WithNodeName("chat_model")) //chatModel 返回 *schema.Message
	chain.AppendLambda(lambda)                                           // 将 *schema.Message 转换为 []*schema.Message
	chain.AppendBranch(compose.NewChainBranch(branchCond).AddLambda("tools", b1).AddLambda("fallback", b2))

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

// func (a *Agent) AddTool(ctx context.Context, tool tool.BaseTool) error {
// 	// 获取工具信息
// 	info, err := tool.Info(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	// 绑定新工具
// 	if err := a.chatModel.BindTools([]*schema.ToolInfo{info}); err != nil {
// 		return err
// 	}

// 	// 更新工具列表
// 	a.tools = append(a.tools, tool)

// 	// 需要重新编译Agent
// 	toolsNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{Tools: a.tools})
// 	if err != nil {
// 		return err
// 	}

// 	chain := compose.NewChain[[]*schema.Message, []*schema.Message]()
// 	chain.
// 		AppendChatModel(a.chatModel, compose.WithNodeName("chat_model")).
// 		AppendToolsNode(toolsNode, compose.WithNodeName("tools"))

// 	agent, err := chain.Compile(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	a.agent = agent
// 	return nil
// }

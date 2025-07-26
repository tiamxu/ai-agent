package llm

import (
	"context"

	"github.com/tiamxu/ai-agent/config"

	"github.com/tiamxu/kit/log"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

func NewOpenAIChatModel(ctx context.Context, cfg config.AliyunConfig) model.ToolCallingChatModel {

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: cfg.BaseURL,
		Model:   cfg.LLMModel,
		APIKey:  cfg.APIKey,
	})
	if err != nil {
		log.Fatalf("NewChatModel failed, err=%v", err)
	}
	return chatModel
}

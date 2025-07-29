package llm

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/tiamxu/ai-agent/config"
)

type LLMService struct {
	model model.ToolCallingChatModel
}

func NewOpenAIChatModel(ctx context.Context, cfg config.LLMConfig) model.ToolCallingChatModel {

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
func NewLLMService(model model.ToolCallingChatModel) *LLMService {
	return &LLMService{model: model}
}

func (s *LLMService) Generate(ctx context.Context, in []*schema.Message) (*schema.Message, error) {
	result, err := s.model.Generate(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("生成失败: %w", err)
	}
	return result, nil
}

func (s *LLMService) Stream(ctx context.Context, in []*schema.Message) (*schema.StreamReader[*schema.Message], error) {

	result, err := s.model.Stream(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("流式生成失败: %w", err)
	}
	return result, nil
}

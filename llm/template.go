package llm

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/tiamxu/ai-agent/config"
)

func (s *LLMService) CreateMessagesFromTemplate(cfg *config.Config, question string) ([]*schema.Message, error) {
	if question == "" {
		return nil, errors.New("问题不能为空")
	}

	template := prompt.FromMessages(schema.FString,
		schema.SystemMessage(cfg.MessageTemplates.System.Content),
		schema.UserMessage(cfg.MessageTemplates.User.Template),
	)

	messages, err := template.Format(context.Background(), map[string]any{
		"role":     cfg.MessageTemplates.System.Role,
		"style":    cfg.MessageTemplates.System.Style,
		"question": question,
	})

	if err != nil {
		return nil, fmt.Errorf("格式化模板失败: %w", err)
	}
	return messages, nil
}

package api

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/tiamxu/ai-agent/agent"
	"github.com/tiamxu/ai-agent/config"
	"github.com/tiamxu/ai-agent/llm"
	"github.com/tiamxu/ai-agent/types"
)

type ChatApi struct {
	cfg        *config.Config
	llmService *llm.LLMService
	agent      *agent.Agent
}

func NewChatApi(cfg *config.Config, llmService *llm.LLMService, agent *agent.Agent) *ChatApi {
	return &ChatApi{
		cfg:        cfg,
		llmService: llmService,
		agent:      agent,
	}
}

func (h *ChatApi) Chat(ctx context.Context, c *app.RequestContext) {
	var req types.AskQuestionReq

	if err := c.BindAndValidate(&req); err != nil {
		hlog.Errorf("参数绑定失败: %v", err)
		c.JSON(consts.StatusBadRequest, types.APIResponse{
			Status:  consts.StatusBadRequest,
			Message: "参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 生成消息
	messages, err := h.llmService.CreateMessagesFromTemplate(h.cfg, req.Question)
	if err != nil {
		hlog.Errorf("生成消息失败: %v", err)
		c.JSON(consts.StatusInternalServerError, types.APIResponse{
			Status:  consts.StatusInternalServerError,
			Message: "生成消息失败",
			Error:   err.Error(),
		})
		return
	}

	// 调用Agent处理
	resp, err := h.agent.Invoke(ctx, messages)

	if err != nil {
		hlog.Errorf("Agent调用失败: %v", err)
		c.JSON(consts.StatusInternalServerError, types.APIResponse{
			Status:  consts.StatusInternalServerError,
			Message: "Agent处理失败",
			Error:   err.Error(),
		})
		return
	}

	// 格式化返回结果
	apiResp := types.APIResponse{
		Status:  consts.StatusOK,
		Message: "ok",
		Data:    resp,
	}

	c.JSON(consts.StatusOK, apiResp)

}

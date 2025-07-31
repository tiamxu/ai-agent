package api

import (
	"context"
	"fmt"

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
		c.JSON(consts.StatusBadRequest, RespError(c, err, "参数错误"))
		return
	}
	fmt.Printf("提交参数:%v\n", req)

	messages, err := h.llmService.CreateMessagesFromTemplate(h.cfg, req.Question)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, RespError(c, err, "生成消息模版错误"))
		return
	}
	resp, err := h.agent.Invoke(ctx, messages)

	if err != nil {
		c.JSON(consts.StatusInternalServerError, RespError(c, err, "agent Invoke调用失败"))
		return
	}
	hlog.Infof("返回内容:", resp)
	c.JSON(consts.StatusOK, RespSuccess(c, resp))
}

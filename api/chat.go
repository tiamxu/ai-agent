package api

import (
	"context"
	"io"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/cloudwego/hertz/pkg/protocol/sse"
	"github.com/tiamxu/ai-agent/config"
	"github.com/tiamxu/ai-agent/llm"
	"github.com/tiamxu/ai-agent/types"
)

type ChatApi struct {
	cfg        *config.Config
	llmService *llm.LLMService
}

func NewChatHandler(cfg *config.Config, llmService *llm.LLMService) *ChatApi {
	return &ChatApi{
		cfg:        cfg,
		llmService: llmService,
	}
}

func (h *ChatApi) Health(ctx context.Context, c *app.RequestContext) {
	c.JSON(consts.StatusOK, RespSuccess(c, ""))
}

func (h *ChatApi) Chat(ctx context.Context, c *app.RequestContext) {
	var req types.AskQuestionReq

	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, RespError(c, err, "参数错误"))
		return
	}

	messages, err := h.llmService.CreateMessagesFromTemplate(h.cfg, req.Question)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, RespError(c, err, ""))
		return
	}

	if req.Stream {
		// 获取最后事件ID（用于断线重连）
		lastEventID := sse.GetLastEventID(&c.Request)
		hlog.CtxInfof(ctx, "LastEventID: %s", lastEventID)

		// 创建SSE写入器
		w := sse.NewWriter(c)

		// 获取连接关闭信号
		connClosed := ctx.Done()

		// 启动LLM流式生成
		streamResult, err := h.llmService.Stream(ctx, messages)
		if err != nil {
			c.JSON(consts.StatusInternalServerError, RespError(c, err, "流式错误"))
			return
		}

		// 设置响应头
		c.Response.Header.Set("Content-Type", "text/event-stream")
		c.Response.Header.Set("Cache-Control", "no-cache")
		c.Response.Header.Set("Connection", "keep-alive")

		// 流式传输LLM响应
		for {
			select {
			case <-connClosed:
				hlog.CtxInfof(ctx, "客户端断开连接")
				return
			default:
				msg, err := streamResult.Recv()
				if err != nil {
					if err == io.EOF {
						// 流结束
						w.Close()
						hlog.CtxInfof(ctx, "流式传输完成")
						return
					}
					hlog.CtxErrorf(ctx, "读取流数据错误: %v", err)
					return
				}
				if msg == nil || len(msg.Content) == 0 {
					continue
				}
				// 发送事件
				err = w.WriteEvent("", "message", []byte(strings.TrimSpace(msg.Content)))
				if err != nil {
					hlog.CtxErrorf(ctx, "写入SSE事件错误: %v", err)
					return
				}

			}
		}
	}

	// 非流式处理
	result, err := h.llmService.Generate(ctx, messages)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, RespError(c, err, ""))
		return
	}
	c.JSON(consts.StatusOK, RespSuccess(c, result.Content))
}

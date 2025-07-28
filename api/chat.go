package api

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
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
	hlog.CtxInfof(ctx, "messages: %+v\n\n", messages)

	if err != nil {
		c.JSON(consts.StatusInternalServerError, RespError(c, err, ""))
		return
	}

	// if req.Stream {
	// 	// 创建SSE writer
	// 	w := sse.NewWriter(c)

	// 	// 设置SSE头
	// 	c.Response.Header.Set("Content-Type", "text/event-stream")
	// 	c.Response.Header.Set("Cache-Control", "no-cache")
	// 	c.Response.Header.Set("Connection", "keep-alive")
	// 	c.Response.Header.Set("Transfer-Encoding", "chunked")

	// 	// 从请求中获取上下文
	// 	reqCtx := ctx.Request.Context()

	// 	// 创建可取消的上下文
	// 	reqCtx, cancel := context.WithCancel(reqCtx)
	// 	defer cancel()

	// 	stream, err := h.llmService.Stream(reqCtx, messages)
	// 	if err != nil {
	// 		c.JSON(consts.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	// 		return
	// 	}
	// 	defer stream.Close()

	// 	// 使用通道检测客户端是否断开
	// 	connClosed := reqCtx.Done()

	// 	for {
	// 		select {
	// 		case <-connClosed:
	// 			hlog.CtxInfof(ctx, "客户端断开连接")
	// 			return
	// 		default:
	// 			message, err := stream.Recv()
	// 			if err != nil {
	// 				if err != io.EOF {
	// 					hlog.CtxErrorf(ctx, "流式读取错误: %v", err)
	// 					_ = w.WriteEvent("", "error", []byte(err.Error()))
	// 				}
	// 				_ = w.WriteEvent("", "message", []byte(`{"content":"","finished":true}`))
	// 				return
	// 			}

	// 			if message != nil && len(message.Content) > 0 {
	// 				content := strings.TrimSpace(message.Content)
	// 				hlog.CtxInfof(ctx, "[Stream] Content: %s\n", content)

	// 				// 使用JSON格式发送消息
	// 				eventData := `{"content":"` + content + `","finished":false}`
	// 				if err := w.WriteEvent("", "message", []byte(eventData)); err != nil {
	// 					hlog.CtxErrorf(ctx, "发送SSE事件错误: %v", err)
	// 					return
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	result, err := h.llmService.Generate(ctx, messages)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, RespError(c, err, ""))
		return
	}
	hlog.CtxInfof(ctx, "使用%s 回答: %s", h.cfg.LLMType, result.Content)
	c.JSON(consts.StatusOK, RespSuccess(c, result.Content))
}

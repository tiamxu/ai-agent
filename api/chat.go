package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/hertz/pkg/app"
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

	semanticResp, err := h.parseDNSIntent(ctx, req.Question, messages)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, RespError(c, err, "语义分析失败"))
		return
	}
	if semanticResp.Action == "" {
		// 非流式处理

		result, err := h.llmService.Generate(ctx, messages)
		if err != nil {
			c.JSON(consts.StatusInternalServerError, RespError(c, err, ""))
			return
		}
		c.JSON(consts.StatusOK, RespSuccess(c, result.Content))
	}
	// if req.Stream {
	// 	// 获取最后事件ID（用于断线重连）
	// 	lastEventID := sse.GetLastEventID(&c.Request)
	// 	hlog.CtxInfof(ctx, "LastEventID: %s", lastEventID)

	// 	// 创建SSE写入器
	// 	w := sse.NewWriter(c)

	// 	// 获取连接关闭信号
	// 	connClosed := ctx.Done()

	// 	// 启动LLM流式生成
	// 	streamResult, err := h.llmService.Stream(ctx, messages)
	// 	if err != nil {
	// 		c.JSON(consts.StatusInternalServerError, RespError(c, err, "流式错误"))
	// 		return
	// 	}

	// 	// 设置响应头
	// 	c.Response.Header.Set("Content-Type", "text/event-stream")
	// 	c.Response.Header.Set("Cache-Control", "no-cache")
	// 	c.Response.Header.Set("Connection", "keep-alive")

	// 	// 流式传输LLM响应
	// 	for {
	// 		select {
	// 		case <-connClosed:
	// 			hlog.CtxInfof(ctx, "客户端断开连接")
	// 			return
	// 		default:
	// 			msg, err := streamResult.Recv()
	// 			if err != nil {
	// 				if err == io.EOF {
	// 					// 流结束
	// 					w.Close()
	// 					hlog.CtxInfof(ctx, "流式传输完成")
	// 					return
	// 				}
	// 				hlog.CtxErrorf(ctx, "读取流数据错误: %v", err)
	// 				return
	// 			}
	// 			if msg == nil || len(msg.Content) == 0 {
	// 				continue
	// 			}
	// 			// 发送事件
	// 			err = w.WriteEvent("", "message", []byte(strings.TrimSpace(msg.Content)))
	// 			if err != nil {
	// 				hlog.CtxErrorf(ctx, "写入SSE事件错误: %v", err)
	// 				return
	// 			}

	// 		}
	// 	}
	// }

	result, err := h.executeDNSOperation(ctx, semanticResp)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, RespError(c, err, "DNS操作失败"))
		return
	}

	c.JSON(consts.StatusOK, RespSuccess(c, result))
}

func (h *ChatApi) parseDNSIntent(ctx context.Context, question string, in []*schema.Message) (*types.DNSOperationRequest, error) {
	// 	prompt := fmt.Sprintf(`请分析以下用户请求，提取DNS操作相关信息。用户可能想创建、修改、删除或查询DNS记录。
	// 当前默认域名是：%s

	// 用户请求：%s

	// 请以JSON格式返回分析结果，包含以下可能字段：
	// - action: create/update/delete/enable/disable/query
	// - domain: 域名(如果没有明确指定，使用默认域名)
	// - rr: 主机记录
	// - type: 记录类型(A/CNAME等)
	// - value: 记录值
	// - ttl: TTL时间(可选)
	// - status: enable/disable(仅当操作是启用/禁用时需要)`, "gopron.cn", question)
	resp, err := h.llmService.Generate(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("LLM调用失败: %v", err)
	}
	var dnsReq types.DNSOperationRequest
	if err := json.Unmarshal([]byte(resp.Content), &dnsReq); err != nil {
		return nil, fmt.Errorf("解析AI响应失败: %v", err)
	}
	if dnsReq.Domain == "" {
		dnsReq.Domain = "gopron.cn"
	}
	return &dnsReq, nil
}

func (h *ChatApi) executeDNSOperation(ctx context.Context, req *types.DNSOperationRequest) (string, error) {
	// 这里可以调用前面定义的DNSRecordTool
	// 或者直接调用DNS API

	// 示例实现，实际应该调用你的DNS服务
	switch req.Action {
	case "create":
		return fmt.Sprintf("已创建DNS记录: %s.%s %s记录指向%s",
			req.RR, req.Domain, req.Type, req.Value), nil
	case "update":
		return fmt.Sprintf("已更新DNS记录: %s.%s", req.RR, req.Domain), nil
	case "delete":
		return fmt.Sprintf("已删除DNS记录: %s.%s", req.RR, req.Domain), nil
	default:
		return "", fmt.Errorf("不支持的操作类型: %s", req.Action)
	}
}

package api

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"log"

	"github.com/gin-gonic/gin"
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

func (h *ChatApi) Health(c *gin.Context) {
	c.JSON(http.StatusOK, RespSuccess(c, ""))

}

func (h *ChatApi) Chat(c *gin.Context) {
	var req types.AskQuestionReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, RespError(c, err, "参数错误"))
		return
	}

	messages, err := h.llmService.CreateMessagesFromTemplate(h.cfg, req.Question)
	log.Printf("messages: %+v\n\n", messages)

	if err != nil {
		c.JSON(http.StatusInternalServerError, RespError(c, err, ""))
		return
	}

	if req.Stream {
		// 设置SSE头
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		// 从请求中获取上下文
		ctx := c.Request.Context()

		// 创建可取消的上下文
		ctx, cancel := context.WithCancel(ctx)
		defer cancel() // 确保资源释放

		stream, err := h.llmService.Stream(ctx, messages)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer stream.Close()
		// defer func() {
		// 	// 确保流关闭
		// 	if closeErr := stream.Close(); closeErr != nil {
		// 		log.Printf("关闭流时出错: %v", closeErr)
		// 	}
		// }()

		// 使用通道检测客户端是否断开
		clientGone := c.Writer.CloseNotify()

		c.Stream(func(w io.Writer) bool {
			select {
			case <-clientGone:
				// 客户端断开连接
				cancel() // 取消OpenAI流
				log.Println("客户端断开连接")
				return false
			case <-ctx.Done():
				// 上下文被取消
				log.Printf("上下文被取消: %v", ctx.Err())
				return false
			default:
				message, err := stream.Recv()
				if err != nil {
					if err != io.EOF {
						log.Printf("流式读取错误: %v", err)
						c.SSEvent("error", gin.H{"error": err.Error()})
					}
					return false
				}
				if message != nil && len(message.Content) > 0 {
					c.SSEvent("message", gin.H{
						"content":  message.Content,
						"finished": false,
					})
					return true
				}
				return false
			}
		})

		c.SSEvent("message", gin.H{
			"content":  "",
			"finished": true,
		})
		return
	}

	result, err := h.llmService.Generate(c.Request.Context(), messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, RespError(c, err, ""))
		return
	}
	log.Printf("使用%s 回答: %s", h.cfg.LLMType, result.Content)
	c.JSON(http.StatusOK, RespSuccess(c, result.Content))

}

func (h *ChatApi) Stream(c *gin.Context) {
	// 创建一个不随请求取消的独立上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute) // 适当延长超时
	defer cancel()

	var req struct {
		Question string `json:"question" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 准备消息
	messages, err := h.llmService.CreateMessagesFromTemplate(h.cfg, req.Question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取流式响应
	stream, err := h.llmService.Stream(ctx, messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stream.Close()

	// 设置SSE头部
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // 禁用Nginx缓冲

	// 获取flusher并立即刷新头部
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming unsupported"})
		return
	}
	flusher.Flush()

	// 心跳机制保持连接
	ticker := time.NewTicker(15 * time.Second) // 15秒发送一次心跳
	defer ticker.Stop()

	// 流式传输响应
	c.Stream(func(w io.Writer) bool {
		select {
		case <-ticker.C:
			// 发送心跳保持连接
			c.SSEvent("heartbeat", "ping")
			return true
		default:
			message, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return false
				}
				log.Printf("流式读取错误: %v", err)
				c.SSEvent("error", gin.H{"error": err.Error()})
				return false
			}

			if message != nil && len(message.Content) > 0 {
				c.SSEvent("message", gin.H{
					"content":  message.Content,
					"finished": false,
				})
				return true
			}
			return false
		}
	})

	// 发送结束标记
	c.SSEvent("message", gin.H{
		"content":  "",
		"finished": true,
	})
}

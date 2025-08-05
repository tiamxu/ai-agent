package types

type AskQuestionReq struct {
	Question string `json:"question" binding:"required"`
	Stream   bool   `json:"stream"`
}

type DNSRecord struct {
	Domain   string `json:"domain"`
	RR       string `json:"rr"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	TTL      int    `json:"ttl"`
	RecordID string `json:"record_id,omitempty"`
	Status   string `json:"status,omitempty"`
}

type DNSResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// APIResponse 标准API响应结构
type APIResponse struct {
	Status  int         `json:"status"`          // HTTP状态码
	Message string      `json:"message"`         // 状态描述
	Data    interface{} `json:"data"`            // 返回数据
	Error   string      `json:"error,omitempty"` // 错误信息（可选）
}

// ToolResponse 工具调用的返回结构
type ToolResponse struct {
	Role         string      `json:"role"`                    // "assistant" 或 "tool"
	Content      string      `json:"content"`                 // 返回内容
	ToolCallID   string      `json:"tool_call_id,omitempty"`  // 工具调用ID（可选）
	ToolName     string      `json:"tool_name,omitempty"`     // 工具名称（可选）
	FinishReason string      `json:"finish_reason,omitempty"` // 结束原因（如 "stop"）
	Usage        *TokenUsage `json:"usage,omitempty"`         // Token使用情况
}

// TokenUsage Token统计信息
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

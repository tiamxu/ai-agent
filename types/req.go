package types

type AskQuestionReq struct {
	Question string `json:"question" binding:"required"`
	Stream   bool   `json:"stream"`
}

type DNSOperationRequest struct {
	Action   string `json:"action"`    // create/update/delete/enable/disable
	Domain   string `json:"domain"`    // 域名
	RR       string `json:"rr"`        // 主机记录
	Type     string `json:"type"`      // 记录类型
	Value    string `json:"value"`     // 记录值
	TTL      int    `json:"ttl"`       // TTL
	Status   string `json:"status"`    // 状态
	RecordID string `json:"record_id"` // 记录ID
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

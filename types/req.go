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

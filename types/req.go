package types

type AskQuestionReq struct {
	Question string `json:"question" binding:"required"`
	Stream   bool   `json:"stream"`
}

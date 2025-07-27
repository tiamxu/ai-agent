package api

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
}

func RespSuccess(c *gin.Context, data interface{}, code ...int) *Response {
	status := 200
	if code != nil {
		status = code[0]
	}

	if data == nil {
		data = "操作成功"
	}

	r := &Response{
		Status:  status,
		Data:    data,
		Message: "ok",
	}
	return r
}

func RespError(c *gin.Context, err error, msg string, code ...int) *Response {
	status := 500
	if code != nil {
		status = code[0]
	}
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	r := &Response{
		Status:  status,
		Data:    "",
		Message: msg,
		Error:   errorMsg,
	}
	return r
}

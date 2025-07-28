package api

import (
	"github.com/cloudwego/hertz/pkg/app"
)

type Response struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
}

func RespSuccess(ctx *app.RequestContext, data interface{}, code ...int) *Response {
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

func RespError(ctx *app.RequestContext, err error, msg string, code ...int) *Response {
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

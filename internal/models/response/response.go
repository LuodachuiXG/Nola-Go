package response

import (
	"net/http"
	"nola-go/internal/util"
)

// Response 响应结构体
type Response struct {
	Code   int     `json:"code"`
	ErrMsg *string `json:"errMsg,omitempty"`
	Data   any     `json:"data"`
}

// OK 成功响应体
func OK(data any) Response {
	return Response{
		Code:   http.StatusOK,
		Data:   data,
		ErrMsg: nil,
	}
}

// Fail 失败响应体
func Fail(errMsg string) Response {
	return Response{
		Code:   http.StatusConflict,
		Data:   nil,
		ErrMsg: &errMsg,
	}
}

// ParamMismatch 请求参数不匹配响应体
func ParamMismatch() Response {
	return Fail("请求参数不匹配")
}

// Unauthorized 未授权响应体
func Unauthorized() Response {
	return Response{
		Code:   http.StatusUnauthorized,
		Data:   nil,
		ErrMsg: util.StringPtr("无权访问受保护资源"),
	}
}

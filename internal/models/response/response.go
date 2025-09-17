package response

import (
	"net/http"
	"nola-go/internal/util"

	"github.com/gin-gonic/gin"
)

// Response 响应结构体
type Response struct {
	Code   int     `json:"code"`
	ErrMsg *string `json:"errMsg"`
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

// OkAndResponse 成功并直接返回成功响应体
func OkAndResponse(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, OK(data))
}

// Fail 失败响应体
func Fail(errMsg string) Response {
	return Response{
		Code:   http.StatusConflict,
		Data:   nil,
		ErrMsg: &errMsg,
	}
}

// FailAndResponse 失败并直接返回失败响应体
func FailAndResponse(ctx *gin.Context, errMsg string) {
	ctx.JSON(http.StatusConflict, Fail(errMsg))
}

// ParamMismatch 请求参数不匹配响应
func ParamMismatch(ctx *gin.Context) {
	FailAndResponse(ctx, "请求参数不匹配")
}

// Unauthorized 未授权响应体
func Unauthorized() Response {
	return Response{
		Code:   http.StatusUnauthorized,
		Data:   nil,
		ErrMsg: util.StringPtr("无权访问受保护资源"),
	}
}

// UnauthorizedAndResponse 未授权并直接返回失败响应体
func UnauthorizedAndResponse(ctx *gin.Context) {
	ctx.JSON(http.StatusUnauthorized, Unauthorized())
}

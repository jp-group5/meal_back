package response

import "github.com/gin-gonic/gin"

// APIResponse 统一响应结构。
// code: 业务错误码，0 表示成功；非 0 表示失败。
// message: 响应消息。
// data: 成功时可选返回数据。
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success 返回成功响应。
func Success(c *gin.Context, httpStatus int, data interface{}) {
	c.JSON(httpStatus, APIResponse{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

// Error 返回失败响应。
func Error(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, APIResponse{
		Code:    code,
		Message: message,
	})
}

package model

// Response 通用响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageData 分页数据
type PageData struct {
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
	List  interface{} `json:"list"`
}

// 响应码常量
const (
	CodeSuccess      = 200
	CodeBadRequest   = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
	CodeServerError  = 500
)

// Success 成功响应
func Success(data interface{}) *Response {
	return &Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	}
}

// SuccessMessage 成功响应（带消息）
func SuccessMessage(message string, data interface{}) *Response {
	return &Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	}
}

// Error 错误响应
func Error(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
	}
}

// BadRequest 400 错误
func BadRequest(message string) *Response {
	return Error(CodeBadRequest, message)
}

// Unauthorized 401 错误
func Unauthorized(message string) *Response {
	return Error(CodeUnauthorized, message)
}

// Forbidden 403 错误
func Forbidden(message string) *Response {
	return Error(CodeForbidden, message)
}

// NotFound 404 错误
func NotFound(message string) *Response {
	return Error(CodeNotFound, message)
}

// ServerError 500 错误
func ServerError(message string) *Response {
	return Error(CodeServerError, message)
}

package model

import "fmt"

type TransformRequest struct {
	IsFile          bool   `json:"is_file"`
	URL             string `json:"url"`
	UserAccessToken string `json:"user_access_token"`
}

type TransformResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Content string `json:"content,omitempty"`
	File    []byte `json:"file,omitempty"`
}

// ErrorResponse 定义错误响应结构体
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

// Error 实现 error 接口
func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

type SheetContentResult struct {
	Markdown   string `json:"markdown"`
	SheetTitle string `json:"sheetTitle"`
}

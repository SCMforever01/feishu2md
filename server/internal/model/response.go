package model

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// 通用返回结构体
type Resp struct {
	Code    int         `json:"code"`
	Msg     string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Content string      `json:"content"`
}

// 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Resp{
		Code: 0,
		Msg:  "成功",
		Data: data,
	})
}

// 失败响应（带自定义错误码和提示信息）
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Resp{
		Code: code,
		Msg:  message,
	})
}

// 参数错误快捷方式
func ParamError(c *gin.Context, err error) {
	c.JSON(http.StatusOK, Resp{
		Code: 1000,
		Msg:  "参数错误: " + err.Error(),
	})
}

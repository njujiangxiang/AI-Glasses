package httperr

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type response struct {
	Error APIError `json:"error"`
}

// Respond 将业务错误转换为统一 JSON 错误响应。
func Respond(c *gin.Context, err error) {
	var apiErr APIError
	if !errors.As(err, &apiErr) {
		log.Printf("[HTTP ERROR] %s: %v", c.Request.URL.Path, err)
		apiErr = New(InternalError, "internal error")
	}
	c.JSON(apiErr.HTTPStatus, response{Error: apiErr})
}

// OK 返回统一成功响应，HTTP 状态码为 200。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// Created 返回统一创建成功响应，HTTP 状态码为 201。
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{"data": data})
}

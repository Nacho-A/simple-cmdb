package utils

import "github.com/gin-gonic/gin"

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func JSON(c *gin.Context, httpStatus int, code int, msg string, data interface{}) {
	c.JSON(httpStatus, APIResponse{
		Code:    code,
		Message: msg,
		Data:    data,
	})
}

func OK(c *gin.Context, data interface{}) {
	JSON(c, 200, 200, "success", data)
}

func Fail(c *gin.Context, code int, msg string) {
	httpStatus := 200
	switch code {
	case 401:
		httpStatus = 401
	case 403:
		httpStatus = 403
	case 500:
		httpStatus = 500
	}
	JSON(c, httpStatus, code, msg, gin.H{})
}


package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetPage(c *gin.Context) (page, pageSize, offset, limit int) {
	page = 1
	pageSize = 10
	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := c.Query("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			pageSize = n
		}
	}
	offset = (page - 1) * pageSize
	limit = pageSize
	return
}


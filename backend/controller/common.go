package controller

import (
	"github.com/gin-gonic/gin"

	"cursor-cmdb-backend/utils"
)

func (h *Handler) CloudProviders(c *gin.Context) {
	utils.OK(c, gin.H{
		"items": []string{"阿里云", "腾讯云", "AWS", "华为云", "自建", "其他"},
	})
}


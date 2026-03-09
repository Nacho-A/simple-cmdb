package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"cursor-cmdb-backend/utils"
)

func Recovery(log *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if log != nil {
			log.Error("panic recovered", zap.Any("err", recovered))
		}
		utils.JSON(c, http.StatusInternalServerError, 500, "服务器错误", gin.H{})
		c.Abort()
	})
}


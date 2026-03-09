package controller

import (
	"cursor-cmdb-backend/config"
	"gorm.io/gorm"
	"go.uber.org/zap"
)

type Handler struct {
	Cfg *config.Config
	DB  *gorm.DB
	Log *zap.Logger
}


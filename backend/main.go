package main

import (
	"fmt"
	"path/filepath"

	casbinx "cursor-cmdb-backend/casbin"
	"cursor-cmdb-backend/bootstrap"
	"cursor-cmdb-backend/config"
	"cursor-cmdb-backend/logger"
	"cursor-cmdb-backend/model"
	"cursor-cmdb-backend/router"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load("config")
	if err != nil {
		panic(err)
	}

	log, err := logger.New(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	db, err := model.Init(cfg.MySQL.DSN)
	if err != nil {
		log.Fatal("init db failed", zap.Error(err))
	}
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
	}

	if cfg.App.Bootstrap {
		if err := bootstrap.AutoMigrate(db); err != nil {
			log.Fatal("auto migrate failed", zap.Error(err))
		}
		if err := bootstrap.SeedDefaults(db); err != nil {
			log.Fatal("seed defaults failed", zap.Error(err))
		}
	}

	modelPath := filepath.Join("casbin", "model.conf")
	if _, err := casbinx.Init(modelPath, db); err != nil {
		log.Fatal("init casbin failed", zap.Error(err))
	}
	if cfg.App.Bootstrap {
		if err := bootstrap.SeedCasbinPolicies(); err != nil {
			log.Fatal("seed casbin policies failed", zap.Error(err))
		}
	}

	r := router.New(cfg, db, log)
	log.Info(fmt.Sprintf("server listening on %s", cfg.Server.Addr))
	if err := r.Run(cfg.Server.Addr); err != nil {
		log.Fatal("server run failed", zap.Error(err))
	}
}


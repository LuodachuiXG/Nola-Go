package db

import (
	"nola-go/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectMySQL(cfg *config.Config) (*gorm.DB, error) {
	gCfg := &gorm.Config{}
	if cfg.Env == "dev" {
		gCfg.Logger = logger.Default.LogMode(logger.Info)
	}
	db, err := gorm.Open(mysql.Open(cfg.MySQL.DSN), gCfg)
	if err != nil {
		return nil, err
	}
	return db, nil
}

package db

import (
	"context"
	"math"
	"nola-go/internal/config"
	"nola-go/internal/models"

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

// PagerBuilder 构建分页查询
//   - ctx: 上下文
//   - db: 数据库连接
//   - page: 当前页码
//   - size: 每页大小
//   - queryBuilder: 查询条件构造函数
func PagerBuilder[T any](
	ctx context.Context,
	db *gorm.DB,
	page, size int,
	queryBuilder func(*gorm.DB) *gorm.DB,
) (*models.Pager[T], error) {
	// 计算偏移量
	offset := (page - 1) * size

	// 构建查询条件
	queryDB := queryBuilder(db)

	// 查询总条数
	var totalData int64
	countDB := queryBuilder(db.Session(&gorm.Session{}))
	if err := countDB.WithContext(ctx).Model((*T)(nil)).Count(&totalData).Error; err != nil {
		return nil, err
	}

	// 计算总页数
	totalPages := int64(math.Ceil(float64(totalData) / float64(size)))

	// 查询当前页数据
	var data []T
	if err := queryDB.WithContext(ctx).Limit(size).Offset(offset).Scan(&data).Error; err != nil {
		return nil, err
	}

	return &models.Pager[T]{
		Page:       page,
		Size:       size,
		Data:       data,
		TotalData:  totalData,
		TotalPages: totalPages,
	}, nil
}

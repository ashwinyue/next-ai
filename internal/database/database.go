package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/ashwinyue/next-rag/next-ai/internal/config"
	"github.com/ashwinyue/next-rag/next-ai/internal/model"
)

// DB 数据库封装
type DB struct {
	*gorm.DB
}

// New 创建数据库连接
func New(cfg *config.Config) (*DB, error) {
	logLevel := gormlogger.Silent
	if cfg.App.Debug {
		logLevel = gormlogger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.GetDSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	// 连接池配置
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Database.MaxLifetime) * time.Second)

	// 健康检查
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 自动迁移
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	return &DB{DB: db}, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping 检查数据库连接
func (db *DB) Ping(ctx context.Context) error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

// autoMigrate 自动迁移
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(model.AllModels...)
}

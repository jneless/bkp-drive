package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"

	"bkp-drive/pkg/config"
)

var DB *sql.DB

func InitDB(cfg *config.Config) error {
	// 支持两种连接方式：完整URL或分别提供参数
	var dsn string
	if cfg.DatabaseURL != "" {
		// 使用完整的DATABASE_URL
		dsn = cfg.DatabaseURL
	} else {
		// 使用单独的参数构建连接字符串
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBName,
		)
	}

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("连接Supabase数据库失败: %w", err)
	}

	// 配置连接池
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("Supabase数据库连接测试失败: %w", err)
	}

	return nil
}

func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
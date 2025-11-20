package config

import (
	"os"
)

type Config struct {
	// TOS配置
	TOSEndpoint string
	TOSRegion   string
	AccessKey   string
	SecretKey   string
	BucketName  string

	// Supabase/PostgreSQL配置 (支持两种方式)
	DatabaseURL string // 完整的连接URL (优先使用)
	DBHost      string // 或者使用单独的参数
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string

	// JWT密钥
	JWTSecret string
}

func LoadConfig() *Config {
	return &Config{
		// TOS配置
		AccessKey:   os.Getenv("TOS_ACCESS_KEY"),
		SecretKey:   os.Getenv("TOS_SECRET_KEY"),
		TOSEndpoint: os.Getenv("TOS_ENDPOINT"),
		TOSRegion:   os.Getenv("TOS_REGION"),
		BucketName:  os.Getenv("TOS_BUCKET_NAME"),

		// Supabase/PostgreSQL配置
		DatabaseURL: os.Getenv("DATABASE_URL"),          // 优先使用完整URL
		DBHost:      os.Getenv("SUPABASE_DB_HOST"),      // 或者使用单独参数
		DBPort:      getEnvOrDefault("SUPABASE_DB_PORT", "5432"),
		DBUser:      getEnvOrDefault("SUPABASE_DB_USER", "postgres"),
		DBPassword:  os.Getenv("SUPABASE_DB_PASSWORD"),
		DBName:      getEnvOrDefault("SUPABASE_DB_NAME", "postgres"),

		// JWT密钥
		JWTSecret: os.Getenv("JWT_SECRET"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

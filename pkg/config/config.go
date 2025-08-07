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
	
	// MySQL配置
	MySQLUsername string
	MySQLPassword string
	MySQLHost     string
	MySQLPort     string
	MySQLDatabase string
	
	// JWT密钥
	JWTSecret string
}

func LoadConfig() *Config {
	return &Config{
		// TOS配置
		TOSEndpoint: getEnvOrDefault("TOS_ENDPOINT", "https://tos-cn-beijing.volces.com"),
		TOSRegion:   getEnvOrDefault("TOS_REGION", "cn-beijing"),
		AccessKey:   os.Getenv("TOS_ACCESS_KEY"),
		SecretKey:   os.Getenv("TOS_SECRET_KEY"),
		BucketName:  getEnvOrDefault("TOS_BUCKET_NAME", "bkp-drive-bucket"),
		
		// MySQL配置
		MySQLUsername: getEnvOrDefault("MYSQL_USERNAME", "root"),
		MySQLPassword: getEnvOrDefault("MYSQL_PASSWORD", ""),
		MySQLHost:     getEnvOrDefault("MYSQL_HOST", "localhost"),
		MySQLPort:     getEnvOrDefault("MYSQL_PORT", "3306"),
		MySQLDatabase: getEnvOrDefault("MYSQL_DATABASE", "bkp_drive"),
		
		// JWT密钥
		JWTSecret: getEnvOrDefault("JWT_SECRET", "bkp-drive-jwt-secret-key-2024"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

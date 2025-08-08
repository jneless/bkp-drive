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

		AccessKey: os.Getenv("TOS_ACCESS_KEY"),
		SecretKey: os.Getenv("TOS_SECRET_KEY"),

		// TOSEndpoint: getEnvOrDefault("TOS_ENDPOINT", "https://tos-cn-beijing.volces.com"),
		// TOSRegion:   getEnvOrDefault("TOS_REGION", "cn-beijing"),
		// BucketName:  getEnvOrDefault("TOS_BUCKET_NAME", "bkp-drive-bucket"),
		TOSEndpoint: os.Getenv("TOS_ENDPOINT"),
		TOSRegion:   os.Getenv("TOS_REGION"),
		BucketName:  os.Getenv("TOS_BUCKET_NAME"),

		// mysql conf
		MySQLUsername: os.Getenv("MYSQL_USERNAME"),
		MySQLPassword: os.Getenv("MYSQL_PASSWORD"),
		MySQLHost:     os.Getenv("MYSQL_HOST"),
		MySQLPort:     os.Getenv("MYSQL_PORT"),
		MySQLDatabase: os.Getenv("MYSQL_DATABASE"),

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

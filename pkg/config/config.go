package config

import (
	"os"
)

type Config struct {
	TOSEndpoint   string
	TOSRegion     string
	AccessKey     string
	SecretKey     string
	BucketName    string
}

func LoadConfig() *Config {
	return &Config{
		TOSEndpoint: getEnvOrDefault("TOS_ENDPOINT", "https://tos-cn-beijing.volces.com"),
		TOSRegion:   getEnvOrDefault("TOS_REGION", "cn-beijing"),
		AccessKey:   os.Getenv("TOS_ACCESS_KEY"),
		SecretKey:   os.Getenv("TOS_SECRET_KEY"),
		BucketName:  getEnvOrDefault("TOS_BUCKET_NAME", "bkp-drive-bucket"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
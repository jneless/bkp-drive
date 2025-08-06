package main

import (
	"log"

	"bkp-drive/pkg/config"
	"bkp-drive/pkg/tos"
)

func main() {
	log.Println("启动 bkp-drive 服务...")

	cfg := config.LoadConfig()
	
	log.Printf("配置信息:")
	log.Printf("- TOS端点: %s", cfg.TOSEndpoint)
	log.Printf("- TOS区域: %s", cfg.TOSRegion)
	log.Printf("- 存储桶名称: %s", cfg.BucketName)
	log.Printf("- AccessKey已设置: %t", cfg.AccessKey != "")

	tosClient, err := tos.NewTOSClient(cfg)
	if err != nil {
		log.Fatalf("创建TOS客户端失败: %v", err)
	}

	log.Println("测试TOS连接...")
	if err := tosClient.TestConnection(); err != nil {
		log.Fatalf("TOS连接测试失败: %v", err)
	}

	log.Println("确保存储桶存在...")
	if err := tosClient.EnsureBucketExists(); err != nil {
		log.Fatalf("存储桶操作失败: %v", err)
	}

	log.Println("bkp-drive 初始化完成！")
}
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"

	"bkp-drive/pkg/config"
)

func main() {
	cfg := config.LoadConfig()

	client, err := tos.NewClientV2(cfg.TOSEndpoint, tos.WithRegion(cfg.TOSRegion), tos.WithCredentials(tos.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey)))
	if err != nil {
		log.Fatalf("创建TOS客户端失败: %v", err)
	}

	ctx := context.Background()

	// 我们知道这些方法是有效的，因为它们在client.go中工作
	fmt.Println("测试已知可用的方法...")

	// 测试 ListBuckets - 这个我们知道有效
	buckets, err := client.ListBuckets(ctx, &tos.ListBucketsInput{})
	if err != nil {
		log.Printf("ListBuckets 错误: %v", err)
	} else {
		fmt.Printf("成功获取 %d 个存储桶\n", len(buckets.Buckets))
	}

	// 测试 HeadBucket - 这个我们知道有效
	_, err = client.HeadBucket(ctx, &tos.HeadBucketInput{
		Bucket: cfg.BucketName,
	})
	if err != nil {
		log.Printf("HeadBucket 错误: %v", err)
	} else {
		fmt.Println("HeadBucket 成功")
	}

	// 现在让我们尝试发现其他方法
	// 如果编译器错误告诉我们没有PutObject，那么可能有其他名字的方法

	fmt.Println("SDK客户端已初始化，现在需要发现实际的方法名称...")
	fmt.Printf("客户端类型: %T\n", client)
}
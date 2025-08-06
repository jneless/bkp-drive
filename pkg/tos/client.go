package tos

import (
	"context"
	"fmt"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"

	"bkp-drive/pkg/config"
)

type TOSClient struct {
	client *tos.ClientV2
	config *config.Config
}

func NewTOSClient(cfg *config.Config) (*TOSClient, error) {
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("TOS_ACCESS_KEY 和 TOS_SECRET_KEY 环境变量必须设置")
	}

	client, err := tos.NewClientV2(cfg.TOSEndpoint, tos.WithRegion(cfg.TOSRegion), tos.WithCredentials(tos.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey)))
	if err != nil {
		return nil, fmt.Errorf("创建TOS客户端失败: %w", err)
	}

	return &TOSClient{
		client: client,
		config: cfg,
	}, nil
}

func (tc *TOSClient) TestConnection() error {
	ctx := context.Background()
	
	buckets, err := tc.client.ListBuckets(ctx, &tos.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("测试连接失败: %w", err)
	}

	fmt.Printf("连接成功！找到 %d 个存储桶:\n", len(buckets.Buckets))
	for _, bucket := range buckets.Buckets {
		fmt.Printf("- %s\n", bucket.Name)
	}

	return nil
}

func (tc *TOSClient) EnsureBucketExists() error {
	ctx := context.Background()
	
	_, err := tc.client.HeadBucket(ctx, &tos.HeadBucketInput{
		Bucket: tc.config.BucketName,
	})
	
	if err != nil {
		if tosErr, ok := err.(*tos.TosServerError); ok && tosErr.StatusCode == 404 {
			fmt.Printf("存储桶 %s 不存在，正在创建...\n", tc.config.BucketName)
			
			_, err = tc.client.CreateBucket(ctx, &tos.CreateBucketInput{
				Bucket: tc.config.BucketName,
			})
			if err != nil {
				return fmt.Errorf("创建存储桶失败: %w", err)
			}
			fmt.Printf("存储桶 %s 创建成功！\n", tc.config.BucketName)
		} else {
			return fmt.Errorf("检查存储桶失败: %w", err)
		}
	} else {
		fmt.Printf("存储桶 %s 已存在\n", tc.config.BucketName)
	}
	
	return nil
}
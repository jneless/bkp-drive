package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"

	"bkp-drive/pkg/config"
	tosClient "bkp-drive/pkg/tos"
)

func main() {
	cfg := config.LoadConfig()

	// 创建原生客户端用于对比
	rawClient, err := tos.NewClientV2(cfg.TOSEndpoint, tos.WithRegion(cfg.TOSRegion), tos.WithCredentials(tos.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey)))
	if err != nil {
		log.Fatalf("创建原生TOS客户端失败: %v", err)
	}

	// 创建我们的TOS客户端
	tosClientWrapper, err := tosClient.NewTOSClient(cfg)
	if err != nil {
		log.Fatalf("创建TOS客户端包装器失败: %v", err)
	}

	ctx := context.Background()

	// 测试 PutObjectV2
	fmt.Println("=== 测试 PutObjectV2 API ===")
	testPutObjectV2(rawClient, ctx, cfg.BucketName)

	// 测试 ListObjectsV2
	fmt.Println("\n=== 测试 ListObjectsV2 API ===")
	testListObjectsV2(rawClient, ctx, cfg.BucketName)

	// 测试 GetObjectV2
	fmt.Println("\n=== 测试 GetObjectV2 API ===")
	testGetObjectV2(rawClient, ctx, cfg.BucketName)

	// 测试我们的包装器
	fmt.Println("\n=== 测试包装器方法 ===")
	testWrapper(tosClientWrapper)

	// 测试 DeleteObjectV2
	fmt.Println("\n=== 测试 DeleteObjectV2 API ===")
	testDeleteObjectV2(rawClient, ctx, cfg.BucketName)
}

func testPutObjectV2(client *tos.ClientV2, ctx context.Context, bucket string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PutObjectV2 方法失败: %v\n", r)
		}
	}()

	testContent := "Hello, TOS SDK V2 Test!"
	testKey := "test/hello-v2.txt"

	input := &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:        bucket,
			Key:           testKey,
			ContentLength: int64(len(testContent)),
			ContentType:   "text/plain",
		},
		Content: strings.NewReader(testContent),
	}

	output, err := client.PutObjectV2(ctx, input)
	if err != nil {
		fmt.Printf("PutObjectV2 错误: %v\n", err)
		return
	}

	fmt.Printf("PutObjectV2 成功! ETag: %s\n", output.ETag)
}

func testListObjectsV2(client *tos.ClientV2, ctx context.Context, bucket string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("ListObjectsV2 方法失败: %v\n", r)
		}
	}()

	input := &tos.ListObjectsV2Input{
		Bucket: bucket,
		ListObjectsInput: tos.ListObjectsInput{
			Prefix:  "test/",
			MaxKeys: 10,
		},
	}

	output, err := client.ListObjectsV2(ctx, input)
	if err != nil {
		fmt.Printf("ListObjectsV2 错误: %v\n", err)
		return
	}

	fmt.Printf("ListObjectsV2 成功! 找到 %d 个对象:\n", len(output.Contents))
	for _, obj := range output.Contents {
		fmt.Printf("- %s (大小: %d, 修改时间: %v)\n", obj.Key, obj.Size, obj.LastModified)
	}
}

func testGetObjectV2(client *tos.ClientV2, ctx context.Context, bucket string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("GetObjectV2 方法失败: %v\n", r)
		}
	}()

	testKey := "test/hello-v2.txt"

	input := &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    testKey,
	}

	output, err := client.GetObjectV2(ctx, input)
	if err != nil {
		fmt.Printf("GetObjectV2 错误: %v\n", err)
		return
	}
	defer output.Content.Close()

	// 读取内容
	buf := new(bytes.Buffer)
	buf.ReadFrom(output.Content)
	content := buf.String()

	fmt.Printf("GetObjectV2 成功! 内容: %s\n", content)
	fmt.Printf("ContentType: %s, ContentLength: %d\n", output.ContentType, output.ContentLength)
}

func testDeleteObjectV2(client *tos.ClientV2, ctx context.Context, bucket string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("DeleteObjectV2 方法失败: %v\n", r)
		}
	}()

	testKey := "test/hello-v2.txt"

	input := &tos.DeleteObjectV2Input{
		Bucket: bucket,
		Key:    testKey,
	}

	output, err := client.DeleteObjectV2(ctx, input)
	if err != nil {
		fmt.Printf("DeleteObjectV2 错误: %v\n", err)
		return
	}

	fmt.Printf("DeleteObjectV2 成功! 删除标记: %t\n", output.DeleteMarker)
}

func testWrapper(client *tosClient.TOSClient) {
	// 测试ListObjects包装器
	resp, err := client.ListObjects("test/")
	if err != nil {
		fmt.Printf("ListObjects包装器错误: %v\n", err)
	} else {
		fmt.Printf("ListObjects包装器成功! 找到 %d 个文件和文件夹\n", resp.Total)
	}

	// 测试CreateFolder包装器
	err = client.CreateFolder("test/wrapper-folder/")
	if err != nil {
		fmt.Printf("CreateFolder包装器错误: %v\n", err)
	} else {
		fmt.Println("CreateFolder包装器成功!")
	}
}
package tos

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"

	"bkp-drive/internal/models"
)

// getContentTypeFromKey 根据文件扩展名推测Content-Type
func getContentTypeFromKey(key string) string {
	ext := strings.ToLower(filepath.Ext(key))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".zip":
		return "application/zip"
	case ".mp4":
		return "video/mp4"
	case ".mp3":
		return "audio/mpeg"
	default:
		return "application/octet-stream"
	}
}

// UploadFile 上传文件到 TOS
func (tc *TOSClient) UploadFile(file multipart.File, header *multipart.FileHeader, folder string) (*models.UploadResponse, error) {
	ctx := context.Background()
	
	// 构建对象键
	key := header.Filename
	if folder != "" && folder != "/" {
		// 确保文件夹路径以 / 结尾
		folder = strings.TrimSuffix(folder, "/") + "/"
		key = folder + header.Filename
	}
	
	// 获取文件大小
	file.Seek(0, 0) // 重置文件指针
	fileSize := header.Size
	
	// 检测文件类型
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	
	input := &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:        tc.config.BucketName,
			Key:           key,
			ContentLength: fileSize,
			ContentType:   contentType,
		},
		Content: file,
	}
	
	_, err := tc.client.PutObjectV2(ctx, input)
	if err != nil {
		return &models.UploadResponse{
			Success: false,
			Message: fmt.Sprintf("上传文件失败: %v", err),
		}, nil
	}
	
	return &models.UploadResponse{
		Success: true,
		Message: "文件上传成功",
		Key:     key,
		URL:     fmt.Sprintf("https://%s.%s/%s", tc.config.BucketName, tc.config.TOSEndpoint, key),
	}, nil
}

// ListObjects 列出存储桶中的对象
func (tc *TOSClient) ListObjects(prefix string) (*models.ListResponse, error) {
	ctx := context.Background()
	
	input := &tos.ListObjectsV2Input{
		Bucket: tc.config.BucketName,
		ListObjectsInput: tos.ListObjectsInput{
			Prefix:    prefix,
			MaxKeys:   1000, // 默认最大返回1000个对象
			Delimiter: "/",  // 用于区分文件夹
		},
	}
	
	output, err := tc.client.ListObjectsV2(ctx, input)
	if err != nil {
		return &models.ListResponse{
			Success: false,
			Message: fmt.Sprintf("列出对象失败: %v", err),
		}, nil
	}
	
	var files []models.FileInfo
	var folders []string
	
	// 处理文件
	for _, obj := range output.Contents {
		// 跳过文件夹标记对象（以 / 结尾且大小为0）
		if strings.HasSuffix(obj.Key, "/") && obj.Size == 0 {
			continue
		}
		
		files = append(files, models.FileInfo{
			Key:          obj.Key,
			Name:         filepath.Base(obj.Key),
			Size:         obj.Size,
			LastModified: obj.LastModified,
			ContentType:  getContentTypeFromKey(obj.Key), // 根据文件扩展名推测
			IsFolder:     false,
			ETag:         strings.Trim(obj.ETag, "\""),
		})
	}
	
	// 处理文件夹（公共前缀）
	for _, commonPrefix := range output.CommonPrefixes {
		folderName := strings.TrimSuffix(commonPrefix.Prefix, "/")
		if folderName != "" {
			folders = append(folders, filepath.Base(folderName))
		}
	}
	
	return &models.ListResponse{
		Success: true,
		Message: "列出对象成功",
		Files:   files,
		Folders: folders,
		Total:   len(files) + len(folders),
	}, nil
}

// GetObject 从 TOS 下载对象
func (tc *TOSClient) GetObject(key string) (io.ReadCloser, int64, string, error) {
	ctx := context.Background()
	
	input := &tos.GetObjectV2Input{
		Bucket: tc.config.BucketName,
		Key:    key,
	}
	
	output, err := tc.client.GetObjectV2(ctx, input)
	if err != nil {
		return nil, 0, "", fmt.Errorf("下载对象失败: %w", err)
	}
	
	return output.Content, output.ContentLength, output.ContentType, nil
}

// GetProcessedObject 获取TOS处理后的对象（如缩略图、视频截图等）
func (tc *TOSClient) GetProcessedObject(key string, process string) (io.ReadCloser, int64, string, error) {
	ctx := context.Background()
	
	// 使用TOS SDK的GetObject方法，并通过Process参数指定处理操作
	input := &tos.GetObjectV2Input{
		Bucket:  tc.config.BucketName,
		Key:     key,
		Process: process, // TOS处理参数
	}
	
	output, err := tc.client.GetObjectV2(ctx, input)
	if err != nil {
		return nil, 0, "", fmt.Errorf("获取处理后对象失败: %w", err)
	}
	
	return output.Content, output.ContentLength, output.ContentType, nil
}

// DeleteObject 删除 TOS 中的对象
func (tc *TOSClient) DeleteObject(key string) error {
	ctx := context.Background()
	
	input := &tos.DeleteObjectV2Input{
		Bucket: tc.config.BucketName,
		Key:    key,
	}
	
	_, err := tc.client.DeleteObjectV2(ctx, input)
	if err != nil {
		return fmt.Errorf("删除对象失败: %w", err)
	}
	
	return nil
}

// CreateFolder 创建文件夹（通过创建一个以 / 结尾的空对象）
func (tc *TOSClient) CreateFolder(folderPath string) error {
	ctx := context.Background()
	
	// 确保文件夹路径以 / 结尾
	if !strings.HasSuffix(folderPath, "/") {
		folderPath += "/"
	}
	
	input := &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:        tc.config.BucketName,
			Key:           folderPath,
			ContentLength: 0,
			ContentType:   "application/x-directory",
		},
		Content: strings.NewReader(""),
	}
	
	_, err := tc.client.PutObjectV2(ctx, input)
	if err != nil {
		return fmt.Errorf("创建文件夹失败: %w", err)
	}
	
	return nil
}
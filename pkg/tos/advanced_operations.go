package tos

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"

	"bkp-drive/internal/models"
)

// CopyObject 复制对象（简单实现，先下载再上传）
func (tc *TOSClient) CopyObject(sourceKey, destKey string) error {
	// 由于TOS SDK的CopyObject方法可能不同，我们使用下载-上传的方式
	reader, contentLength, contentType, err := tc.GetObject(sourceKey)
	if err != nil {
		return fmt.Errorf("获取源对象失败: %w", err)
	}
	defer reader.Close()

	ctx := context.Background()
	
	input := &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:        tc.config.BucketName,
			Key:           destKey,
			ContentLength: contentLength,
			ContentType:   contentType,
		},
		Content: reader,
	}

	_, err = tc.client.PutObjectV2(ctx, input)
	if err != nil {
		return fmt.Errorf("复制对象失败: %w", err)
	}

	return nil
}

// MoveObject 移动对象（复制后删除源对象）
func (tc *TOSClient) MoveObject(sourceKey, destKey string) error {
	// 先复制
	if err := tc.CopyObject(sourceKey, destKey); err != nil {
		return err
	}

	// 再删除源对象
	if err := tc.DeleteObject(sourceKey); err != nil {
		// 如果删除失败，尝试清理目标对象
		tc.DeleteObject(destKey)
		return fmt.Errorf("移动对象失败，删除源对象时出错: %w", err)
	}

	return nil
}

// RenameObject 重命名对象
func (tc *TOSClient) RenameObject(oldKey, newKey string) error {
	return tc.MoveObject(oldKey, newKey)
}

// BatchDeleteObjects 批量删除对象
func (tc *TOSClient) BatchDeleteObjects(keys []string) (*models.BatchOperationResponse, error) {
	// 由于不确定批量删除的确切API，我们使用逐个删除的方式
	processed := 0
	failed := 0
	var failedItems []string

	for _, key := range keys {
		if err := tc.DeleteObject(key); err != nil {
			failed++
			failedItems = append(failedItems, key)
		} else {
			processed++
		}
	}

	return &models.BatchOperationResponse{
		Success:     failed == 0,
		Message:     fmt.Sprintf("批量删除完成，成功: %d, 失败: %d", processed, failed),
		Processed:   processed,
		Failed:      failed,
		FailedItems: failedItems,
	}, nil
}

// BatchCopyObjects 批量复制对象
func (tc *TOSClient) BatchCopyObjects(items []string, destination string) (*models.BatchOperationResponse, error) {
	processed := 0
	failed := 0
	var failedItems []string

	destination = strings.TrimSuffix(destination, "/") + "/"

	for _, sourceKey := range items {
		fileName := getFileName(sourceKey)
		destKey := destination + fileName

		if err := tc.CopyObject(sourceKey, destKey); err != nil {
			failed++
			failedItems = append(failedItems, sourceKey)
		} else {
			processed++
		}
	}

	return &models.BatchOperationResponse{
		Success:     failed == 0,
		Message:     fmt.Sprintf("批量复制完成，成功: %d, 失败: %d", processed, failed),
		Processed:   processed,
		Failed:      failed,
		FailedItems: failedItems,
	}, nil
}

// BatchMoveObjects 批量移动对象
func (tc *TOSClient) BatchMoveObjects(items []string, destination string) (*models.BatchOperationResponse, error) {
	processed := 0
	failed := 0
	var failedItems []string

	destination = strings.TrimSuffix(destination, "/") + "/"

	for _, sourceKey := range items {
		fileName := getFileName(sourceKey)
		destKey := destination + fileName

		if err := tc.MoveObject(sourceKey, destKey); err != nil {
			failed++
			failedItems = append(failedItems, sourceKey)
		} else {
			processed++
		}
	}

	return &models.BatchOperationResponse{
		Success:     failed == 0,
		Message:     fmt.Sprintf("批量移动完成，成功: %d, 失败: %d", processed, failed),
		Processed:   processed,
		Failed:      failed,
		FailedItems: failedItems,
	}, nil
}

// SearchObjects 搜索对象
func (tc *TOSClient) SearchObjects(req *models.SearchRequest) (*models.SearchResponse, error) {
	ctx := context.Background()

	// 设置默认限制
	limit := req.Limit
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	input := &tos.ListObjectsV2Input{
		Bucket: tc.config.BucketName,
		ListObjectsInput: tos.ListObjectsInput{
			Prefix:  req.Folder,
			MaxKeys: limit * 2, // 获取更多结果用于过滤
		},
	}

	output, err := tc.client.ListObjectsV2(ctx, input)
	if err != nil {
		return &models.SearchResponse{
			Success: false,
			Message: fmt.Sprintf("搜索失败: %v", err),
		}, nil
	}

	var results []models.ExtendedFileInfo
	
	for _, obj := range output.Contents {
		// 跳过文件夹标记
		if strings.HasSuffix(obj.Key, "/") && obj.Size == 0 {
			continue
		}

		// 应用搜索过滤条件
		if !matchesSearchCriteria(obj, req) {
			continue
		}

		extInfo := models.ExtendedFileInfo{
			FileInfo: models.FileInfo{
				Key:          obj.Key,
				Name:         getFileName(obj.Key),
				Size:         obj.Size,
				LastModified: obj.LastModified,
				ContentType:  getContentTypeFromKey(obj.Key),
				IsFolder:     false,
				ETag:         strings.Trim(obj.ETag, "\""),
			},
			Path: obj.Key,
		}

		results = append(results, extInfo)

		// 限制结果数量
		if len(results) >= limit {
			break
		}
	}

	return &models.SearchResponse{
		Success: true,
		Message: "搜索完成",
		Results: results,
		Total:   len(results),
		Query:   req.Query,
	}, nil
}

// GetStorageStats 获取存储统计信息
func (tc *TOSClient) GetStorageStats() (*models.StorageStats, error) {
	ctx := context.Background()

	input := &tos.ListObjectsV2Input{
		Bucket: tc.config.BucketName,
		ListObjectsInput: tos.ListObjectsInput{
			MaxKeys: 10000, // 获取大量对象用于统计
		},
	}

	output, err := tc.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("获取存储统计失败: %w", err)
	}

	stats := &models.StorageStats{
		FileTypeStats: make(map[string]int64),
		RecentUsage:   []models.DayUsage{},
	}

	var totalSize int64
	var fileCount int64
	var folderCount int64

	// 统计文件和文件夹
	for _, obj := range output.Contents {
		if strings.HasSuffix(obj.Key, "/") && obj.Size == 0 {
			folderCount++
			continue
		}

		fileCount++
		totalSize += obj.Size

		// 按文件类型统计
		contentType := getContentTypeFromKey(obj.Key)
		stats.FileTypeStats[contentType]++
	}

	stats.UsedSpace = totalSize
	stats.FileCount = fileCount
	stats.FolderCount = folderCount
	// 假设总空间为100GB（实际应该从配置或TOS服务获取）
	stats.TotalSpace = 100 * 1024 * 1024 * 1024
	stats.FreeSpace = stats.TotalSpace - stats.UsedSpace

	return stats, nil
}

// 辅助函数
func getFileName(key string) string {
	parts := strings.Split(key, "/")
	return parts[len(parts)-1]
}

func matchesSearchCriteria(obj tos.ListedObjectV2, req *models.SearchRequest) bool {
	// 文件名匹配
	if req.Query != "" && !strings.Contains(strings.ToLower(obj.Key), strings.ToLower(req.Query)) {
		return false
	}

	// 文件大小过滤
	if req.MinSize > 0 && obj.Size < req.MinSize {
		return false
	}
	if req.MaxSize > 0 && obj.Size > req.MaxSize {
		return false
	}

	// 时间范围过滤
	if req.StartDate != "" {
		if startTime, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			if obj.LastModified.Before(startTime) {
				return false
			}
		}
	}
	if req.EndDate != "" {
		if endTime, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			if obj.LastModified.After(endTime.Add(24 * time.Hour)) {
				return false
			}
		}
	}

	// 文件类型过滤
	if len(req.FileTypes) > 0 {
		contentType := getContentTypeFromKey(obj.Key)
		found := false
		for _, ft := range req.FileTypes {
			if strings.Contains(contentType, ft) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
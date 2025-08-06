package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"bkp-drive/internal/models"
	"bkp-drive/pkg/tos"
)

type AdvancedHandler struct {
	tosClient *tos.TOSClient
}

func NewAdvancedHandler(tosClient *tos.TOSClient) *AdvancedHandler {
	return &AdvancedHandler{
		tosClient: tosClient,
	}
}

// BatchDelete 批量删除文件
func (h *AdvancedHandler) BatchDelete(c *gin.Context) {
	var req models.BatchOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "请求参数错误: " + err.Error(),
		})
		return
	}

	result, err := h.tosClient.BatchDeleteObjects(req.Items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if result.Success {
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusPartialContent, result)
	}
}

// BatchMove 批量移动文件
func (h *AdvancedHandler) BatchMove(c *gin.Context) {
	var req models.BatchMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "请求参数错误: " + err.Error(),
		})
		return
	}

	result, err := h.tosClient.BatchMoveObjects(req.Items, req.Destination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if result.Success {
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusPartialContent, result)
	}
}

// BatchCopy 批量复制文件
func (h *AdvancedHandler) BatchCopy(c *gin.Context) {
	var req models.BatchCopyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "请求参数错误: " + err.Error(),
		})
		return
	}

	result, err := h.tosClient.BatchCopyObjects(req.Items, req.Destination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if result.Success {
		c.JSON(http.StatusOK, result)
	} else {
		c.JSON(http.StatusPartialContent, result)
	}
}

// MoveFile 移动文件或文件夹
func (h *AdvancedHandler) MoveFile(c *gin.Context) {
	var req models.MoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "请求参数错误: " + err.Error(),
		})
		return
	}

	if err := h.tosClient.MoveObject(req.Source, req.Destination); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "文件移动成功",
		"source":  req.Source,
		"destination": req.Destination,
	})
}

// CopyFile 复制文件或文件夹
func (h *AdvancedHandler) CopyFile(c *gin.Context) {
	var req models.CopyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "请求参数错误: " + err.Error(),
		})
		return
	}

	if err := h.tosClient.CopyObject(req.Source, req.Destination); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "文件复制成功",
		"source":  req.Source,
		"destination": req.Destination,
	})
}

// RenameFile 重命名文件或文件夹
func (h *AdvancedHandler) RenameFile(c *gin.Context) {
	var req models.RenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "请求参数错误: " + err.Error(),
		})
		return
	}

	if err := h.tosClient.RenameObject(req.OldKey, req.NewKey); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "文件重命名成功",
		"oldKey":  req.OldKey,
		"newKey":  req.NewKey,
	})
}

// SearchFiles 搜索文件
func (h *AdvancedHandler) SearchFiles(c *gin.Context) {
	// 从URL参数构建搜索请求
	req := &models.SearchRequest{
		Query:     c.Query("q"),
		Folder:    c.Query("folder"),
	}

	// 处理文件类型过滤
	if types := c.Query("types"); types != "" {
		req.FileTypes = strings.Split(types, ",")
	}

	// 处理大小过滤
	if minSize := c.Query("minSize"); minSize != "" {
		if size, err := strconv.ParseInt(minSize, 10, 64); err == nil {
			req.MinSize = size
		}
	}
	if maxSize := c.Query("maxSize"); maxSize != "" {
		if size, err := strconv.ParseInt(maxSize, 10, 64); err == nil {
			req.MaxSize = size
		}
	}

	// 处理时间范围
	req.StartDate = c.Query("startDate")
	req.EndDate = c.Query("endDate")

	// 处理限制
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			req.Limit = l
		}
	}

	result, err := h.tosClient.SearchObjects(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetStorageStats 获取存储空间统计
func (h *AdvancedHandler) GetStorageStats(c *gin.Context) {
	stats, err := h.tosClient.GetStorageStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取存储统计成功",
		"stats":   stats,
	})
}

// GetRecentFiles 获取最近访问的文件
func (h *AdvancedHandler) GetRecentFiles(c *gin.Context) {
	limit := 20 // 默认20个
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	// 使用搜索功能，按修改时间排序
	req := &models.SearchRequest{
		Limit: limit,
	}

	result, err := h.tosClient.SearchObjects(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// 按修改时间排序（最新的在前面）
	files := result.Results
	for i, j := 0, len(files)-1; i < j; i, j = i+1, j-1 {
		if files[i].LastModified.Before(files[j].LastModified) {
			files[i], files[j] = files[j], files[i]
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取最近文件成功",
		"files":   files[:min(len(files), limit)],
		"total":   len(files),
	})
}

// FilterFiles 按条件过滤文件
func (h *AdvancedHandler) FilterFiles(c *gin.Context) {
	fileType := c.Query("type")     // image, video, document, etc.
	sizeRange := c.Query("size")    // small, medium, large
	timeRange := c.Query("time")    // today, week, month, year
	folder := c.Query("folder")

	req := &models.SearchRequest{
		Folder: folder,
		Limit:  100,
	}

	// 根据文件类型过滤
	switch fileType {
	case "image":
		req.FileTypes = []string{"image"}
	case "video":
		req.FileTypes = []string{"video"}
	case "document":
		req.FileTypes = []string{"application/pdf", "text", "application/msword", "application/vnd.openxmlformats"}
	case "audio":
		req.FileTypes = []string{"audio"}
	}

	// 根据大小范围过滤
	switch sizeRange {
	case "small":
		req.MaxSize = 10 * 1024 * 1024 // 10MB
	case "medium":
		req.MinSize = 10 * 1024 * 1024   // 10MB
		req.MaxSize = 100 * 1024 * 1024  // 100MB
	case "large":
		req.MinSize = 100 * 1024 * 1024 // 100MB
	}

	// 根据时间范围过滤
	now := gin.H{}
	switch timeRange {
	case "today":
		req.StartDate = now["today"].(string)
	case "week":
		req.StartDate = now["week"].(string)
	case "month":
		req.StartDate = now["month"].(string)
	case "year":
		req.StartDate = now["year"].(string)
	}

	result, err := h.tosClient.SearchObjects(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
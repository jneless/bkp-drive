package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"bkp-drive/internal/models"
	"bkp-drive/pkg/tos"
)

// ShareHandler 分享功能处理器
type ShareHandler struct {
	tosClient *tos.TOSClient
	shares    map[string]*models.ShareInfo // 内存存储分享信息（生产环境应使用数据库）
	mu        sync.RWMutex
}

func NewShareHandler(tosClient *tos.TOSClient) *ShareHandler {
	return &ShareHandler{
		tosClient: tosClient,
		shares:    make(map[string]*models.ShareInfo),
	}
}

// CreateShare 创建分享链接
func (h *ShareHandler) CreateShare(c *gin.Context) {
	var req models.ShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证文件是否存在
	_, _, _, err := h.tosClient.GetObject(req.FileKey)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "文件不存在",
		})
		return
	}

	// 生成分享ID
	shareId, err := generateShareId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "生成分享ID失败",
		})
		return
	}

	// 创建分享信息
	shareInfo := &models.ShareInfo{
		ShareId:       shareId,
		FileKey:       req.FileKey,
		FileName:      getFileName(req.FileKey),
		ShareUrl:      fmt.Sprintf("/api/v1/share/%s", shareId),
		ExpiresAt:     req.ExpiresAt,
		Password:      req.Password,
		AllowDownload: req.AllowDownload,
		AccessCount:   0,
		CreatedAt:     time.Now(),
	}

	// 存储分享信息
	h.mu.Lock()
	h.shares[shareId] = shareInfo
	h.mu.Unlock()

	c.JSON(http.StatusOK, models.ShareResponse{
		Success:   true,
		Message:   "分享创建成功",
		ShareInfo: *shareInfo,
	})
}

// AccessShare 访问分享内容
func (h *ShareHandler) AccessShare(c *gin.Context) {
	shareId := c.Param("shareId")
	password := c.Query("password")

	h.mu.RLock()
	shareInfo, exists := h.shares[shareId]
	h.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "分享不存在或已过期",
		})
		return
	}

	// 检查是否过期
	if time.Now().After(shareInfo.ExpiresAt) {
		h.mu.Lock()
		delete(h.shares, shareId)
		h.mu.Unlock()

		c.JSON(http.StatusGone, models.ErrorResponse{
			Success: false,
			Error:   "分享已过期",
		})
		return
	}

	// 检查密码
	if shareInfo.Password != "" && shareInfo.Password != password {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "密码错误",
		})
		return
	}

	// 增加访问计数
	h.mu.Lock()
	shareInfo.AccessCount++
	h.mu.Unlock()

	// 返回文件信息
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"message":  "访问成功",
		"fileInfo": gin.H{
			"name":          shareInfo.FileName,
			"size":          shareInfo.FileSize,
			"allowDownload": shareInfo.AllowDownload,
			"downloadUrl":   fmt.Sprintf("/api/v1/share/%s/download", shareId),
		},
		"shareInfo": gin.H{
			"accessCount": shareInfo.AccessCount,
			"createdAt":   shareInfo.CreatedAt,
			"expiresAt":   shareInfo.ExpiresAt,
		},
	})
}

// DownloadSharedFile 下载分享的文件
func (h *ShareHandler) DownloadSharedFile(c *gin.Context) {
	shareId := c.Param("shareId")
	password := c.Query("password")

	h.mu.RLock()
	shareInfo, exists := h.shares[shareId]
	h.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "分享不存在或已过期",
		})
		return
	}

	// 检查是否过期
	if time.Now().After(shareInfo.ExpiresAt) {
		c.JSON(http.StatusGone, models.ErrorResponse{
			Success: false,
			Error:   "分享已过期",
		})
		return
	}

	// 检查密码
	if shareInfo.Password != "" && shareInfo.Password != password {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "密码错误",
		})
		return
	}

	// 检查是否允许下载
	if !shareInfo.AllowDownload {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Success: false,
			Error:   "该分享不允许下载",
		})
		return
	}

	// 下载文件
	reader, contentLength, contentType, err := h.tosClient.GetObject(shareInfo.FileKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "下载文件失败: " + err.Error(),
		})
		return
	}
	defer reader.Close()

	c.Header("Content-Disposition", "attachment; filename="+shareInfo.FileName)
	c.Header("Content-Type", contentType)
	c.Header("Content-Length", fmt.Sprintf("%d", contentLength))

	_, err = io.Copy(c.Writer, reader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "下载文件失败: " + err.Error(),
		})
		return
	}
}

// DeleteShare 删除分享链接
func (h *ShareHandler) DeleteShare(c *gin.Context) {
	shareId := c.Param("shareId")

	h.mu.Lock()
	_, exists := h.shares[shareId]
	if exists {
		delete(h.shares, shareId)
	}
	h.mu.Unlock()

	if !exists {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   "分享不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "分享删除成功",
	})
}

// ListShares 列出用户的分享
func (h *ShareHandler) ListShares(c *gin.Context) {
	h.mu.RLock()
	var shares []models.ShareInfo
	for _, share := range h.shares {
		// 清理过期分享
		if time.Now().After(share.ExpiresAt) {
			continue
		}
		shares = append(shares, *share)
	}
	h.mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取分享列表成功",
		"shares":  shares,
		"total":   len(shares),
	})
}

// 辅助函数
func generateShareId() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func getFileName(key string) string {
	parts := strings.Split(key, "/")
	return parts[len(parts)-1]
}
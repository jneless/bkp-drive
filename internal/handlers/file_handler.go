package handlers

import (
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"bkp-drive/internal/models"
	"bkp-drive/pkg/tos"
)

type FileHandler struct {
	tosClient *tos.TOSClient
}

func NewFileHandler(tosClient *tos.TOSClient) *FileHandler {
	return &FileHandler{
		tosClient: tosClient,
	}
}

func (h *FileHandler) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "无法获取上传文件: " + err.Error(),
		})
		return
	}
	defer file.Close()

	folder := c.DefaultPostForm("folder", "")
	
	result, err := h.tosClient.UploadFile(file, header, folder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if !result.Success {
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *FileHandler) ListFiles(c *gin.Context) {
	prefix := c.DefaultQuery("prefix", "")
	
	result, err := h.tosClient.ListObjects(prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if !result.Success {
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *FileHandler) DownloadFile(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "文件键不能为空",
		})
		return
	}

	key = strings.TrimPrefix(key, "/")
	
	reader, contentLength, contentType, err := h.tosClient.GetObject(key)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	defer reader.Close()

	filename := filepath.Base(key)
	
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", contentType)
	c.Header("Content-Length", strconv.FormatInt(contentLength, 10))

	_, err = io.Copy(c.Writer, reader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "文件下载失败: " + err.Error(),
		})
		return
	}
}

func (h *FileHandler) DeleteFile(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "文件键不能为空",
		})
		return
	}

	key = strings.TrimPrefix(key, "/")
	
	err := h.tosClient.DeleteObject(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.DeleteResponse{
		Success: true,
		Message: "文件删除成功",
	})
}

func (h *FileHandler) CreateFolder(c *gin.Context) {
	var request struct {
		FolderPath string `json:"folderPath" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "请求参数错误: " + err.Error(),
		})
		return
	}

	err := h.tosClient.CreateFolder(request.FolderPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "文件夹创建成功",
		"folder":  request.FolderPath,
	})
}
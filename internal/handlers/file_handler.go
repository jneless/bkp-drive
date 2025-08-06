package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// UploadFile 上传文件到TOS
// @Summary      上传文件
// @Description  上传单个文件到指定文件夹
// @Tags         文件操作
// @Accept       multipart/form-data
// @Produce      json
// @Param        file     formData  file    true  "要上传的文件"
// @Param        folder   formData  string  false "目标文件夹路径"
// @Success      200      {object}  models.UploadResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /upload [post]
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

// ListFiles 列出文件和文件夹
// @Summary      列出文件
// @Description  列出指定前缀下的文件和文件夹
// @Tags         文件操作
// @Accept       json
// @Produce      json
// @Param        prefix   query     string  false  "文件夹前缀路径"
// @Success      200      {object}  models.ListResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /files [get]
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

// DownloadFile 下载文件或获取处理后的内容
// @Summary      下载文件
// @Description  下载原文件或获取TOS处理后的内容（如缩略图、视频截图等）
// @Tags         文件操作
// @Accept       json
// @Produce      octet-stream
// @Param        key           path      string  true   "文件路径（URL编码）"
// @Param        x-tos-process query     string  false  "TOS处理参数，如image/resize,w_128或video/snapshot,t_0,w_128,h_128,f_jpg"
// @Success      200           {file}    binary  "文件内容或处理后的内容"
// @Failure      400           {object}  models.ErrorResponse
// @Failure      404           {object}  models.ErrorResponse
// @Failure      500           {object}  models.ErrorResponse
// @Router       /download/{key} [get]
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
	
	// 检查是否有TOS处理参数（如图片处理、视频截图等）
	tosProcess := c.Query("x-tos-process")
	
	// 添加详细调试日志
	fmt.Printf("DownloadFile - Key: %s, TOS Process: %s\n", key, tosProcess)
	
	if tosProcess != "" {
		// 如果有TOS处理参数，使用SDK的GetProcessedObject方法（包含正确签名）
		reader, contentLength, contentType, err := h.tosClient.GetProcessedObject(key, tosProcess)
		if err != nil {
			fmt.Printf("GetProcessedObject 失败: %v\n", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Success: false,
				Error:   "处理内容获取失败: " + err.Error(),
			})
			return
		}
		defer reader.Close()
		
		fmt.Printf("SDK处理请求成功 - Key: %s, Process: %s\n", key, tosProcess)
		
		// 设置响应头
		c.Header("Content-Type", contentType)
		c.Header("Content-Length", strconv.FormatInt(contentLength, 10))
		
		// 流式传输内容
		_, err = io.Copy(c.Writer, reader)
		if err != nil {
			fmt.Printf("内容传输失败: %v\n", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Success: false,
				Error:   "内容传输失败: " + err.Error(),
			})
			return
		}
		
		fmt.Printf("SDK处理内容传输成功\n")
		return
	}
	
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

// DeleteFile 删除文件或文件夹
// @Summary      删除文件
// @Description  删除指定的文件或空文件夹
// @Tags         文件操作
// @Accept       json
// @Produce      json
// @Param        key   path      string  true  "要删除的文件路径（URL编码）"
// @Success      200   {object}  models.DeleteResponse
// @Failure      400   {object}  models.ErrorResponse
// @Failure      500   {object}  models.ErrorResponse
// @Router       /files/{key} [delete]
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

// CreateFolder 创建文件夹
// @Summary      创建文件夹
// @Description  在TOS中创建一个新的文件夹
// @Tags         文件操作
// @Accept       json
// @Produce      json
// @Param        request   body      object{folderPath=string}  true  "文件夹创建请求"
// @Success      200       {object}  object{success=bool,message=string,folder=string}
// @Failure      400       {object}  models.ErrorResponse
// @Failure      500       {object}  models.ErrorResponse
// @Router       /folders [post]
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

// serveProcessedContent 获取TOS处理后的内容并返回给客户端
func (h *FileHandler) serveProcessedContent(c *gin.Context, key string, tosProcess string) error {
	// URL编码文件路径，特别是处理包含斜杠的路径
	encodedKey := url.PathEscape(key)
	
	// 构建完整的TOS处理URL
	tosURL := fmt.Sprintf("https://%s.%s/%s?x-tos-process=%s", 
		h.tosClient.GetBucketName(), 
		h.tosClient.GetEndpoint(), 
		encodedKey, 
		tosProcess)
	
	// 添加详细调试日志
	fmt.Printf("获取TOS处理内容:\n")
	fmt.Printf("  原始Key: %s\n", key)
	fmt.Printf("  编码后Key: %s\n", encodedKey)
	fmt.Printf("  处理参数: %s\n", tosProcess)
	fmt.Printf("  完整URL: %s\n", tosURL)
	
	// 创建HTTP客户端请求TOS处理后的内容
	resp, err := http.Get(tosURL)
	if err != nil {
		fmt.Printf("  HTTP请求失败: %v\n", err)
		return fmt.Errorf("请求TOS处理内容失败: %w", err)
	}
	defer resp.Body.Close()
	
	fmt.Printf("  HTTP响应状态: %d %s\n", resp.StatusCode, resp.Status)
	fmt.Printf("  Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("  Content-Length: %s\n", resp.Header.Get("Content-Length"))
	
	// 检查响应状态码
	if resp.StatusCode != 200 {
		fmt.Printf("  TOS处理请求失败，状态码: %d\n", resp.StatusCode)
		return fmt.Errorf("TOS处理请求失败，状态码: %d", resp.StatusCode)
	}
	
	// 设置适当的头部
	c.Header("Content-Type", resp.Header.Get("Content-Type"))
	
	// 将TOS的响应内容流式传输给客户端
	bytesWritten, err := io.Copy(c.Writer, resp.Body)
	fmt.Printf("  传输字节数: %d\n", bytesWritten)
	if err != nil {
		fmt.Printf("  内容传输失败: %v\n", err)
	} else {
		fmt.Printf("  内容传输成功\n")
	}
	
	return err
}

// redirectToTOSProcessedURL 重定向到TOS处理过的URL（备用方法）
func (h *FileHandler) redirectToTOSProcessedURL(c *gin.Context, key string, tosProcess string) {
	// 构建完整的TOS URL
	tosURL := fmt.Sprintf("https://%s.%s/%s?x-tos-process=%s", 
		h.tosClient.GetBucketName(), 
		h.tosClient.GetEndpoint(), 
		key, 
		tosProcess)
	
	// 添加调试日志
	fmt.Printf("重定向到TOS处理URL: %s\n", tosURL)
	
	// 302重定向到TOS处理URL
	c.Redirect(http.StatusFound, tosURL)
}
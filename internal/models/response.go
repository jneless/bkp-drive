package models

import "time"

// 基础文件信息
type FileInfo struct {
	Key          string            `json:"key"`
	Name         string            `json:"name"`
	Size         int64             `json:"size"`
	LastModified time.Time         `json:"lastModified"`
	ContentType  string            `json:"contentType"`
	IsFolder     bool              `json:"isFolder"`
	ETag         string            `json:"etag"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// 扩展文件信息
type ExtendedFileInfo struct {
	FileInfo
	Path        string `json:"path"`
	Thumbnail   string `json:"thumbnail,omitempty"`
	ShareCount  int    `json:"shareCount"`
	VersionCount int   `json:"versionCount"`
}

// 批量操作请求
type BatchOperationRequest struct {
	Items []string `json:"items" binding:"required"` // 文件或文件夹的keys
}

type BatchMoveRequest struct {
	Items       []string `json:"items" binding:"required"`
	Destination string   `json:"destination" binding:"required"`
}

type BatchCopyRequest struct {
	Items       []string `json:"items" binding:"required"`
	Destination string   `json:"destination" binding:"required"`
}

// 文件操作请求
type MoveRequest struct {
	Source      string `json:"source" binding:"required"`
	Destination string `json:"destination" binding:"required"`
}

type CopyRequest struct {
	Source      string `json:"source" binding:"required"`
	Destination string `json:"destination" binding:"required"`
}

type RenameRequest struct {
	OldKey string `json:"oldKey" binding:"required"`
	NewKey string `json:"newKey" binding:"required"`
}

// 搜索请求
type SearchRequest struct {
	Query      string   `json:"query"`
	FileTypes  []string `json:"fileTypes,omitempty"`
	MinSize    int64    `json:"minSize,omitempty"`
	MaxSize    int64    `json:"maxSize,omitempty"`
	StartDate  string   `json:"startDate,omitempty"`
	EndDate    string   `json:"endDate,omitempty"`
	Folder     string   `json:"folder,omitempty"`
	Limit      int      `json:"limit,omitempty"`
}

// 分享相关
type ShareRequest struct {
	FileKey    string    `json:"fileKey" binding:"required"`
	ExpiresAt  time.Time `json:"expiresAt"`
	Password   string    `json:"password,omitempty"`
	AllowDownload bool   `json:"allowDownload"`
}

type ShareInfo struct {
	ShareId       string    `json:"shareId"`
	FileKey       string    `json:"fileKey"`
	FileName      string    `json:"fileName"`
	FileSize      int64     `json:"fileSize"`
	ShareUrl      string    `json:"shareUrl"`
	ExpiresAt     time.Time `json:"expiresAt"`
	Password      string    `json:"password,omitempty"`
	AllowDownload bool      `json:"allowDownload"`
	AccessCount   int       `json:"accessCount"`
	CreatedAt     time.Time `json:"createdAt"`
}

// 存储统计
type StorageStats struct {
	TotalSpace    int64             `json:"totalSpace"`
	UsedSpace     int64             `json:"usedSpace"`
	FreeSpace     int64             `json:"freeSpace"`
	FileCount     int64             `json:"fileCount"`
	FolderCount   int64             `json:"folderCount"`
	FileTypeStats map[string]int64  `json:"fileTypeStats"`
	RecentUsage   []DayUsage        `json:"recentUsage"`
}

type DayUsage struct {
	Date      string `json:"date"`
	FileCount int64  `json:"fileCount"`
	SizeAdded int64  `json:"sizeAdded"`
}

// 回收站相关
type TrashItem struct {
	FileInfo
	DeletedAt    time.Time `json:"deletedAt"`
	OriginalPath string    `json:"originalPath"`
	DeletedBy    string    `json:"deletedBy,omitempty"`
}

type RestoreRequest struct {
	Items []string `json:"items" binding:"required"`
}

// 版本管理
type FileVersion struct {
	VersionId    string    `json:"versionId"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
	ETag         string    `json:"etag"`
	IsLatest     bool      `json:"isLatest"`
	Comment      string    `json:"comment,omitempty"`
}

// 压缩相关
type CompressRequest struct {
	Items      []string `json:"items" binding:"required"`
	OutputName string   `json:"outputName" binding:"required"`
	Format     string   `json:"format"` // zip, tar, tar.gz
}

type ExtractRequest struct {
	ArchiveKey string `json:"archiveKey" binding:"required"`
	OutputPath string `json:"outputPath"`
}

// 通用响应结构
type UploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Key     string `json:"key,omitempty"`
	URL     string `json:"url,omitempty"`
}

type ListResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message"`
	Files   []FileInfo `json:"files,omitempty"`
	Folders []string   `json:"folders,omitempty"`
	Total   int        `json:"total"`
}

type SearchResponse struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
	Results []ExtendedFileInfo `json:"results"`
	Total   int                `json:"total"`
	Query   string             `json:"query"`
}

type BatchOperationResponse struct {
	Success     bool     `json:"success"`
	Message     string   `json:"message"`
	Processed   int      `json:"processed"`
	Failed      int      `json:"failed"`
	FailedItems []string `json:"failedItems,omitempty"`
}

type DeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ShareResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	ShareInfo ShareInfo `json:"shareInfo,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
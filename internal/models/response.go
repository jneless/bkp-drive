package models

import "time"

type FileInfo struct {
	Key          string    `json:"key"`
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
	ContentType  string    `json:"contentType"`
	IsFolder     bool      `json:"isFolder"`
	ETag         string    `json:"etag"`
}

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

type DeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
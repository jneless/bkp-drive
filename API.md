# bkp-drive API 文档

基于火山引擎TOS的云网盘后端服务

## 快速开始

### 1. 环境配置

复制并编辑环境变量：
```bash
cp .env.example .env
```

设置必需的环境变量：
```bash
export TOS_ACCESS_KEY="your_access_key_here"
export TOS_SECRET_KEY="your_secret_key_here"

# 可选配置
export TOS_ENDPOINT="https://tos-cn-beijing.volces.com"
export TOS_REGION="cn-beijing"
export TOS_BUCKET_NAME="bkp-drive-bucket"
```

### 2. 启动服务

启动HTTP服务器：
```bash
go run cmd/server/main.go
```

或者构建后运行：
```bash
go build -o bkp-drive cmd/server/main.go
./bkp-drive
```

服务默认运行在端口 8080

### 3. 健康检查

```bash
curl http://localhost:8080/health
```

## API 接口

### 基础信息
- **Base URL**: `http://localhost:8080/api/v1`
- **Content-Type**: `application/json` (除文件上传外)

### 1. 文件上传
```http
POST /api/v1/upload
Content-Type: multipart/form-data
```

**参数:**
- `file`: 文件 (required)
- `folder`: 目标文件夹路径 (optional)

**示例:**
```bash
curl -X POST \
  http://localhost:8080/api/v1/upload \
  -F "file=@/path/to/your/file.jpg" \
  -F "folder=images"
```

**响应:**
```json
{
  "success": true,
  "message": "文件上传成功",
  "key": "images/file.jpg",
  "url": "https://bucket.endpoint.com/images/file.jpg"
}
```

### 2. 文件列表
```http
GET /api/v1/files?prefix=folder_path
```

**参数:**
- `prefix`: 文件夹路径前缀 (optional)

**示例:**
```bash
curl http://localhost:8080/api/v1/files
curl http://localhost:8080/api/v1/files?prefix=images
```

**响应:**
```json
{
  "success": true,
  "message": "获取文件列表成功",
  "files": [
    {
      "key": "images/photo1.jpg",
      "name": "photo1.jpg",
      "size": 1024000,
      "lastModified": "2023-01-01T12:00:00Z",
      "contentType": "image/jpeg",
      "isFolder": false,
      "etag": "abc123"
    }
  ],
  "folders": ["documents", "videos"],
  "total": 10
}
```

### 3. 文件下载
```http
GET /api/v1/download/{file_key}
```

**示例:**
```bash
curl -o downloaded_file.jpg http://localhost:8080/api/v1/download/images/photo1.jpg
```

### 4. 删除文件
```http
DELETE /api/v1/files/{file_key}
```

**示例:**
```bash
curl -X DELETE http://localhost:8080/api/v1/files/images/photo1.jpg
```

**响应:**
```json
{
  "success": true,
  "message": "文件删除成功"
}
```

### 5. 创建文件夹
```http
POST /api/v1/folders
Content-Type: application/json
```

**请求体:**
```json
{
  "folderPath": "new-folder/sub-folder"
}
```

**示例:**
```bash
curl -X POST \
  http://localhost:8080/api/v1/folders \
  -H "Content-Type: application/json" \
  -d '{"folderPath": "documents/2024"}'
```

**响应:**
```json
{
  "success": true,
  "message": "文件夹创建成功",
  "folder": "documents/2024"
}
```

## 错误响应

所有错误响应格式：
```json
{
  "success": false,
  "error": "错误信息描述"
}
```

常见状态码：
- `200` - 成功
- `400` - 请求参数错误
- `404` - 文件不存在
- `500` - 服务器内部错误

## CORS 支持

服务器已启用 CORS，支持所有来源的跨域请求，适合前端应用直接调用。

## 项目结构

```
bkp-drive/
├── cmd/
│   ├── bkp-drive/     # 原始连接测试程序
│   └── server/        # HTTP API 服务器
├── internal/
│   ├── handlers/      # HTTP 处理器
│   └── models/        # 数据模型
├── pkg/
│   ├── config/        # 配置管理
│   └── tos/          # TOS 客户端封装
├── .env.example       # 环境变量模板
└── CLAUDE.md         # 项目说明
```

## 开发计划

当前版本为基础框架，后续将实现：
- ✅ TOS SDK 集成
- ✅ 基础 HTTP API
- 🔄 文件操作完整实现
- 📋 用户认证
- 📋 多用户支持
- 📋 文件搜索
- 📋 批量操作
- 📋 缩略图生成
- 📋 回收站功能

## 技术栈

- **后端框架**: Gin (Go)
- **对象存储**: 火山引擎 TOS
- **部署**: 支持 Docker / 裸机部署
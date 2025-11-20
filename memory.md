# bkp-drive 项目记忆文档

> 本文档记录了对 bkp-drive (不靠谱网盘) 项目的深度分析和理解，用于后续开发工作的快速恢复和参考。

## 项目概述

**bkp-drive** 是一个基于火山引擎TOS的云存储项目，采用Go后端 + Web前端架构。

- **技术栈**: Go 1.23.4 + Gin + MySQL + JWT + 火山引擎TOS
- **部署**: 单端口18666服务，后端提供静态文件服务
- **存储**: 火山引擎TOS对象存储作为文件存储后端

## 项目结构分析

### 核心目录结构
```
bkp-drive/
├── cmd/server/main.go           # HTTP服务器入口
├── internal/
│   ├── handlers/                # API处理器层
│   │   ├── auth_handler.go      # 用户认证
│   │   ├── file_handler.go      # 文件操作
│   │   ├── advanced_handler.go  # 高级功能
│   │   └── share_handler.go     # 分享功能
│   ├── middleware/auth.go       # JWT认证中间件
│   ├── models/                  # 数据模型
│   └── services/user_service.go # 用户服务
├── pkg/
│   ├── config/config.go         # 配置管理
│   ├── database/mysql.go        # 数据库连接
│   └── tos/                     # TOS客户端封装
└── public/                      # Web前端
```

## 核心功能模块

### 1. 用户认证系统

**位置**: `internal/handlers/auth_handler.go`, `internal/services/user_service.go`

**核心逻辑**:
- **用户注册**: 生成`bkp-`前缀8位随机字符用户ID
- **密码安全**: 使用bcrypt加密存储
- **JWT认证**: 24小时过期，包含用户ID和用户名
- **中间件保护**: Bearer token验证，提取用户上下文

**关键函数**:
- `RegisterUser()` - 用户注册，生成唯一用户ID
- `LoginUser()` - 用户登录，生成JWT令牌
- `AuthMiddleware()` - JWT认证中间件

### 2. 文件操作核心

**位置**: `internal/handlers/file_handler.go`, `pkg/tos/operations.go`

**核心功能**:
- **上传**: 多部分表单上传到TOS，支持文件夹分类
- **列表**: 支持前缀查询，区分文件和文件夹
- **下载**: 原文件下载 + TOS处理功能(缩略图、视频截图)
- **删除/创建**: 基于TOS SDK的CRUD操作

**TOS集成特色**:
- 支持`x-tos-process`参数进行图片resize和视频截图
- 使用`GetProcessedObject()`方法获取处理后内容
- 自动内容类型检测和文件夹管理

### 3. TOS云存储集成

**位置**: `pkg/tos/client.go`, `pkg/tos/operations.go`

**核心功能**:
- **认证配置**: AK/SK凭据管理，支持多区域
- **存储桶管理**: 自动创建和验证存储桶存在性
- **多媒体处理**: 支持图片resize、视频snapshot等高级功能
- **错误处理**: 完善的TOS错误处理和重试机制

**关键方法**:
- `NewTOSClient()` - 创建TOS客户端
- `GetProcessedObject()` - 获取处理后的媒体内容
- `EnsureBucketExists()` - 确保存储桶存在

### 4. 高级功能模块

**位置**: `internal/handlers/advanced_handler.go`, `pkg/tos/advanced_operations.go`

**批量操作**:
- `BatchDelete()` - 批量删除文件
- `BatchMove()` - 批量移动文件  
- `BatchCopy()` - 批量复制文件

**搜索和过滤**:
- 支持按文件类型、大小、时间范围搜索
- 获取最近文件和存储统计信息

### 5. 分享系统

**位置**: `internal/handlers/share_handler.go`

**功能特点**:
- 生成唯一分享ID和链接
- 支持密码保护和过期时间设置
- 访问计数和权限控制
- 内存存储分享信息(生产环境需改为数据库)

## 数据库设计

**位置**: `scripts/init_db.sql`, `pkg/database/mysql.go`

```sql
-- 用户表
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(12) UNIQUE NOT NULL,  -- bkp-开头的唯一标识符
    username VARCHAR(30) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,       -- bcrypt加密
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

## 环境配置

**必需环境变量**:
```bash
# TOS存储配置
export TOS_ENDPOINT="https://tos-cn-beijing.volces.com"
export TOS_REGION="cn-beijing"
export TOS_ACCESS_KEY="your-access-key"
export TOS_SECRET_KEY="your-secret-key"
export TOS_BUCKET_NAME="bkp-drive-bucket"

# MySQL数据库配置
export MYSQL_USERNAME="your-username"
export MYSQL_PASSWORD="your-password"
export MYSQL_HOST="localhost"
export MYSQL_PORT="3306"
export MYSQL_DATABASE="bkp_drive"

# JWT密钥
export JWT_SECRET="your-jwt-secret"
```

## 启动流程

1. **环境准备**: 配置上述环境变量
2. **数据库初始化**: `mysql -u root -p < scripts/init_db.sql`
3. **安装依赖**: `go mod tidy`
4. **启动服务**: `go run cmd/server/main.go`

**访问地址**:
- 主页: http://localhost:18666/
- 网盘: http://localhost:18666/pan.html
- 登录: http://localhost:18666/login.html
- 注册: http://localhost:18666/register.html
- API文档: http://localhost:18666/swagger/index.html

## 前端架构

**位置**: `public/`

**技术栈**: 纯HTML/CSS/JavaScript

**主要页面**:
- `index.html` - Apple风格首页
- `pan.html` - 文件管理界面
- `login.html/register.html` - 用户认证页面
- `swagger.html` - API文档界面

## API接口总览

### 认证接口
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `GET /api/v1/auth/profile` - 获取用户信息

### 文件操作接口
- `POST /api/v1/upload` - 上传文件
- `GET /api/v1/files` - 列出文件
- `GET /api/v1/download/*key` - 下载文件
- `DELETE /api/v1/files/*key` - 删除文件
- `POST /api/v1/folders` - 创建文件夹

### 批量操作接口
- `POST /api/v1/batch/delete` - 批量删除
- `POST /api/v1/batch/move` - 批量移动
- `POST /api/v1/batch/copy` - 批量复制

### 分享接口
- `POST /api/v1/share/create` - 创建分享
- `GET /api/v1/share/:id` - 访问分享
- `DELETE /api/v1/share/:id` - 删除分享

## 关键技术要点

### 1. 用户隔离
- 每个用户拥有唯一的`bkp-`前缀用户ID
- 文件操作通过JWT中间件进行用户身份验证
- 未来可扩展为用户级别的存储隔离

### 2. TOS多媒体处理
- 支持图片实时缩略图: `x-tos-process=image/resize,w_32`
- 支持视频截图: `x-tos-process=video/snapshot,t_0,w_32,h_32,f_jpg`
- 使用SDK的签名机制确保安全访问

### 3. 错误处理模式
- 统一的错误响应结构 (`models.ErrorResponse`)
- TOS操作的完善错误处理和回滚机制
- HTTP状态码的合理使用

### 4. 扩展性考虑
- 模块化的处理器设计
- 配置驱动的环境管理
- 为批量操作预留的响应结构

## 待优化点

1. **分享系统**: 当前使用内存存储，生产环境需改为数据库存储
2. **用户隔离**: 可增强为基于用户ID的文件路径隔离
3. **缓存机制**: 可添加Redis缓存提升性能
4. **日志系统**: 可增加结构化日志记录
5. **监控告警**: 可集成健康检查和监控指标

## 文件关键位置索引

- 服务器入口: `cmd/server/main.go:40-237`
- 认证逻辑: `internal/services/user_service.go:74-169`
- 文件上传: `internal/handlers/file_handler.go:40-68`
- TOS客户端: `pkg/tos/client.go:18-32`
- 下载处理: `internal/handlers/file_handler.go:113-190`
- 批量操作: `pkg/tos/advanced_operations.go:65-88`

---

*最后更新时间: 2025-08-09*
*当前项目版本: 2.0.0 (Vercel Ready)*

## Vercel 部署改造

**项目已成功适配Vercel平台部署**

### 改造要点

1. **项目结构重组**:
   - 前端文件移至 `public/` 目录
   - API函数放置在 `api/` 目录
   - 使用单文件多路由模式 (`api/index.go`)

2. **API架构变更**:
   - 统一的Handler函数处理所有API路由
   - 基于URL路径的内部路由分发
   - 支持CORS跨域请求

3. **核心API端点**:
   - `GET /api/health` - 健康检查
   - `POST /api/register` - 用户注册
   - `POST /api/login` - 用户登录
   - `GET /api/files` - 文件列表 (需JWT认证)
   - `POST /api/upload` - 文件上传 (需JWT认证)

4. **前端配置适配**:
   - API基础URL动态配置 (`localhost:3000/api` vs `/api`)
   - 支持本地开发和生产部署

5. **环境变量配置**:
   - 创建 `.env.local` 用于本地开发
   - `.env.example` 提供配置模板

### 本地测试验证

✅ **API功能验证**:
- 健康检查端点正常
- 用户注册功能正常 (数据库连接成功)
- 用户登录和JWT生成正常
- 受保护端点的认证机制正常

✅ **前端访问验证**:
- 静态文件服务正常
- 各页面可正常访问

### 部署准备

项目现已准备好部署到Vercel平台:

1. **环境变量设置** - 需在Vercel Dashboard中配置所有必要的环境变量
2. **数据库连接** - 确保MySQL数据库可从Vercel访问
3. **TOS服务集成** - 后续需集成完整的文件操作功能

---

*最后更新时间: 2025-08-09*
*当前项目版本: 2.0.0 (Vercel Ready)*
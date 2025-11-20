# bkp-drive (不靠谱网盘)

> 基于火山引擎TOS的个人网盘 🚀 从不靠谱到靠谱的路上

![](src/img/logo_v1.gif)

## 📋 项目介绍

**bkp-drive** 是一个基于火山引擎对象存储(TOS)服务构建的网盘服务。项目基于 TOS Go SDK 开发.

## ✨ 功能特性

### 🔐 用户认证
- **用户注册**: 支持用户账号注册，自动生成唯一用户ID（bkp-前缀）
- **用户登录**: JWT令牌认证，支持记住登录状态
- **权限控制**: 所有文件操作需要登录验证

### 📁 文件管理
- **文件上传**: 支持单文件和文件夹批量上传
- **文件下载**: 安全的认证下载功能
- **文件删除**: 支持单文件和批量删除
- **文件夹操作**: 创建文件夹、文件夹导航
- **文件预览**: 图片在线预览（缩略图功能开发中）

### 🎨 界面特色
- **Apple风格首页**: 现代化的用户界面设计
- **响应式设计**: 支持桌面和移动设备
- **网格/列表视图**: 多种文件显示模式
- **面包屑导航**: 便捷的文件路径导航

## 🚀 快速开始

### 环境变量配置

```bash
# TOS 存储配置
export TOS_ENDPOINT="your-tos-endpoint"
export TOS_REGION="your-region"
export TOS_ACCESS_KEY="your-access-key"
export TOS_SECRET_KEY="your-secret-key"
export TOS_BUCKET_NAME="your-bucket-name"

# Supabase PostgreSQL 数据库配置 (使用Session Pooler - IPv4兼容)
export DATABASE_URL="postgresql://postgres.[PROJECT_REF]:[YOUR_PASSWORD]@aws-1-[region].pooler.supabase.com:5432/postgres"

# JWT密钥
export JWT_SECRET="jwt_secret_key"

```

### 数据库初始化

**Supabase方式** (推荐):
1. 在Supabase Dashboard创建新项目
2. 复制Session Pooler连接字符串（IPv4兼容）
3. 在Supabase SQL Editor执行初始化脚本:
```bash
cat scripts/init_supabase.sql
# 或者直接在Supabase SQL Editor中执行脚本内容
```

**本地测试方式**:
```bash
# 使用psql连接Supabase并执行初始化脚本
psql "postgresql://postgres.[PROJECT_REF]:[YOUR_PASSWORD]@aws-1-[region].pooler.supabase.com:5432/postgres" < scripts/init_supabase.sql
```

### 启动服务

#### 本地开发 (Go服务器)
```bash
# 安装依赖
go mod tidy

# 启动服务器
go run cmd/server/main.go
```

#### 本地开发 (Vercel Dev)
```bash
# 安装Vercel CLI
npm i -g vercel

# 启动Vercel开发服务器
vercel dev
```

#### Vercel部署
```bash
# 部署到Vercel
vercel --prod

# 记得在Vercel Dashboard配置环境变量:
# - DATABASE_URL
# - TOS_ENDPOINT, TOS_REGION, TOS_ACCESS_KEY, TOS_SECRET_KEY, TOS_BUCKET_NAME
# - JWT_SECRET
```

### 访问服务

**本地Go服务器** (端口18666):
- **首页**: http://localhost:18666/ (Apple风格首页)
- **网盘功能**: http://localhost:18666/pan.html (文件管理界面)
- **用户登录**: http://localhost:18666/login.html
- **用户注册**: http://localhost:18666/register.html
- **API文档**: http://localhost:18666/swagger/index.html
- **健康检查**: http://localhost:18666/health

**Vercel Dev服务器** (端口自动分配，通常3000-3002):
- **首页**: http://localhost:3002/
- **网盘功能**: http://localhost:3002/pan.html
- **用户登录**: http://localhost:3002/login.html
- **用户注册**: http://localhost:3002/register.html
- **API端点**: http://localhost:3002/api/*

**Vercel生产环境**:
- 部署后访问: https://your-project.vercel.app

## 📖 API 文档

### Swagger 文档
访问 http://localhost:18666/swagger/index.html 查看完整的API文档

### 核心模块说明

#### 🌐 HTTP服务层 (`cmd/server/`, `internal/handlers/`)
- **main.go**: 服务器启动和路由配置
- **file_handler.go**: 基础文件操作（上传、下载、删除、列表）
- **advanced_handler.go**: 高级功能（批量操作、搜索、统计）  
- **share_handler.go**: 文件分享和权限管理

#### 🗃️ TOS存储层 (`pkg/tos/`)
- **client.go**: TOS客户端连接和认证
- **operations.go**: 基础存储操作（GetObject、PutObject等）
- **advanced_operations.go**: 批量操作和搜索功能

#### 📊 配置管理 (`pkg/config/`)
- 环境变量管理和TOS连接配置
- 服务器端口和CORS设置

#### 🎨 前端界面 (`frontend/`)
- 响应式Web界面和Electron桌面应用
- 文件上传、预览、批量操作交互
- 图片和视频缩略图显示

#### 📚 API文档 (`docs/`)
- Swagger自动生成的API文档
- 支持在线测试和接口说明

## 🔧 技术栈

- **后端**: Go 1.23.4, Gin Web框架
- **存储**: 火山引擎TOS对象存储
- **数据库**: Supabase PostgreSQL (Session Pooler)
- **部署**: Vercel Serverless Functions + 本地Go服务器
- **认证**: JWT令牌 (24小时有效期)
- **密码加密**: bcrypt
- **前端**: HTML5, CSS3, Vanilla JavaScript
- **文档**: Swagger/OpenAPI 3.0
- **依赖管理**: Go Modules

## 📄 参考资料

- [火山引擎对象存储TOS API文档](https://www.volcengine.com/docs/6349/74837)
- [TOS Go SDK文档](https://github.com/volcengine/ve-tos-golang-sdk)
- [Supabase PostgreSQL文档](https://supabase.com/docs/guides/database)
- [Vercel部署文档](https://vercel.com/docs)
- [Gin Web框架文档](https://gin-gonic.com/zh-cn/docs/)
- [Swagger/OpenAPI文档](https://swagger.io/docs/)

## 📚 项目文档

- [ARCHITECTURE.md](./ARCHITECTURE.md) - 项目架构和API调用关系详解
- [MIGRATION_TO_SUPABASE.md](./MIGRATION_TO_SUPABASE.md) - MySQL到Supabase迁移指南
- [DATABASE_TEST_REPORT.md](./DATABASE_TEST_REPORT.md) - 数据库测试报告
- [API.md](./API.md) - 基础API文档
- [API_EXTENDED.md](./API_EXTENDED.md) - 扩展API文档

## license
Apache-2.0

## Thanks
* 感谢 Jinpu Hu 对本项目的前端架构建议

* 感谢 Weibin Ma 对 ai 相关技术的讲解

* 感谢 claude code 和 instcopilot 提供 ai 相关能力
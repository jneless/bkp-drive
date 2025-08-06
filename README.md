# bkp-drive (不靠谱网盘)

> 基于火山引擎TOS的云存储后端服务 🚀
> 从不靠谱到靠谱的路上

![](src/img/logo_v1.gif)

## 📋 项目介绍

**bkp-drive** 是一个基于火山引擎对象存储(TOS)服务构建的云存储后端，提供完整的网盘功能API。项目使用Go语言开发，具备企业级的功能特性和扩展性。

### 🎯 核心特性

- ✅ **完整的文件操作** - 上传、下载、删除、文件夹管理
- ✅ **批量操作支持** - 批量删除、移动、复制文件
- ✅ **高级文件操作** - 文件移动、复制、重命名
- ✅ **智能搜索功能** - 支持多维度文件搜索和过滤
- ✅ **安全文件分享** - 密码保护、过期时间、访问统计
- ✅ **存储统计分析** - 详细的存储使用情况和文件类型统计
- ⚠️ **RESTful API** - 15+ 个标准化API接口

### 📊 项目规模

- **代码行数**: 2,010 行
- **Go文件数**: 12 个
- **API接口数**: 15+
- **功能模块数**: 6 个主要模块

## 🚀 快速开始

### 环境要求

- Go 1.23.4+
- 火山引擎 TOS 服务账号
- macOS (目前仅在 macOS 上测试)

### 环境变量配置

```bash
export TOS_ENDPOINT="your-tos-endpoint"
export TOS_REGION="your-region" 
export TOS_ACCESS_KEY="your-access-key"
export TOS_SECRET_KEY="your-secret-key"
export TOS_BUCKET_NAME="your-bucket-name"
```

### 启动服务

```bash
# 安装依赖
go mod tidy

# 启动HTTP服务器
go run cmd/server/main.go

# 访问健康检查
curl http://localhost:8082/health
```

## 🏗️ 架构设计

### 项目结构
```
bkp-drive/
├── cmd/                    # 可执行程序
│   └── server/main.go     # HTTP服务器主程序
├── internal/              # 内部业务逻辑
│   ├── handlers/          # HTTP处理器
│   └── models/           # 数据模型
├── pkg/                   # 可复用包
│   ├── config/           # 配置管理
│   └── tos/              # TOS客户端封装
└── 文档/                  # API文档和功能说明
```

### 核心模块

#### 🗂️ 文件操作模块 (`pkg/tos/`)
- `client.go` - TOS客户端连接管理
- `operations.go` - 基础文件操作 (上传/下载/删除/列表)
- `advanced_operations.go` - 高级操作 (批量/搜索/统计)

#### 🌐 HTTP服务模块 (`internal/handlers/`)
- `file_handler.go` - 基础文件操作API
- `advanced_handler.go` - 批量操作和搜索API
- `share_handler.go` - 文件分享和权限管理

#### ⚙️ 配置模块 (`pkg/config/`)
- 环境变量管理
- TOS连接配置
- 服务器配置

## 📚 API 文档

### 🔥 核心 API (基础功能)

| 功能 | 方法 | 路径 | 状态 |
|------|------|------|------|
| 文件上传 | POST | `/api/v1/upload` | ✅ |
| 文件下载 | GET | `/api/v1/download/*key` | ✅ |
| 文件列表 | GET | `/api/v1/files` | ✅ |
| 删除文件 | DELETE | `/api/v1/files/*key` | ✅ |
| 创建文件夹 | POST | `/api/v1/folders` | ✅ |

### 🚀 扩展 API (高级功能)

| 功能类别 | API数量 | 实现状态 |
|----------|---------|----------|
| **批量操作** | 3个 | ✅ 完全实现 |
| **文件操作** | 3个 | ✅ 完全实现 |
| **搜索功能** | 3个 | ⚠️ 功能有限 |
| **分享功能** | 4个 | ✅ 完全实现 |
| **存储统计** | 1个 | ✅ 完全实现 |

详细的API使用说明请参考：
- [基础API文档](API.md)
- [扩展功能文档](API_EXTENDED.md)

## ⚡ 功能特性

### ✅ 已实现功能

#### 🗂️ 文件管理
- **基础操作**: 上传、下载、删除、文件夹创建
- **高级操作**: 文件移动、复制、重命名
- **批量操作**: 支持批量删除、移动、复制多个文件

#### 🔍 智能搜索
- 文件名搜索和过滤
- 按文件类型、大小、时间过滤
- 最近文件查看

#### 🔗 安全分享
- 创建带密码保护的分享链接
- 设置分享过期时间
- 分享访问统计
- 下载权限控制

#### 📊 存储统计
- 存储空间使用分析
- 文件类型分布统计
- 文件和文件夹计数

### ⚠️ 功能限制

1. **搜索功能限制**
   - 仅支持文件名搜索，无法搜索文件内容
   - TOS API限制最多返回1000个对象
   - 大量文件时性能较差

2. **分享功能限制**
   - 分享信息存储在内存中，服务重启后丢失
   - 生产环境需要数据库持久化存储

3. **性能考虑**
   - 文件复制/移动使用下载-上传方式
   - 批量操作逐个执行，大量操作时较慢

### ❌ 未实现功能

- 文件版本管理和历史记录
- 回收站功能
- 文件压缩和解压
- 缩略图生成
- 全文搜索
- 用户认证和权限管理
- 多用户支持

## 🔄 开发历程

### Phase 1: 基础框架 ✅
- [x] Go模块初始化和TOS SDK集成
- [x] 环境变量配置管理
- [x] HTTP服务器和CORS配置
- [x] 基础文件操作API

### Phase 2: 核心功能 ✅
- [x] 完整的文件CRUD操作
- [x] 流式文件传输
- [x] 错误处理和响应格式统一
- [x] TOS客户端封装

### Phase 3: 高级功能 ✅
- [x] 批量文件操作
- [x] 文件搜索和过滤
- [x] 文件分享系统
- [x] 存储统计分析
- [x] 15+个扩展API接口

### Phase 4: 文档和测试 ✅
- [x] 完整的API文档
- [x] 功能实现状态标注
- [x] 使用示例和测试流程
- [x] 技术限制说明


## 🔧 开发指南

### 添加新功能
1. 在 `pkg/tos/` 中添加TOS操作
2. 在 `internal/handlers/` 中添加HTTP处理器
3. 在 `internal/models/` 中定义数据模型
4. 在 `cmd/server/main.go` 中注册路由
5. 更新API文档

### 测试方法
```bash
# 运行服务
go run cmd/server/main.go

# 执行API测试 
curl http://localhost:8082/health

# 查看详细API文档
cat API_EXTENDED.md
```

## 📖 参考资料

- [火山引擎对象存储TOS API文档](https://www.volcengine.com/docs/6349/74837)
- [TOS Go SDK文档](https://github.com/volcengine/ve-tos-golang-sdk)
- [Gin Web框架文档](https://gin-gonic.com/zh-cn/docs/)

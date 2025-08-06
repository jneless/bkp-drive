# bkp-drive (不靠谱网盘)

> 基于火山引擎TOS的云存储后端服务 🚀
> 从不靠谱到靠谱的路上

![](src/img/logo_v1.gif)

## 📋 项目介绍

**bkp-drive** 是一个基于火山引擎对象存储(TOS)服务构建的网盘服务。项目基于 TOS Go SDK开发。


## 🚀 快速开始


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
```

### 访问服务

- **API服务**: http://localhost:8082
- **健康检查**: http://localhost:8082/health
- **API文档**: http://localhost:8082/swagger/index.html
- **前端界面**: 参见 [frontend/README.md](frontend/README.md)

## 📖 API 文档

### Swagger 文档
访问 http://localhost:8082/swagger/index.html 查看完整的API文档

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
- **前端**: HTML5, CSS3, Vanilla JavaScript
- **桌面**: Electron
- **文档**: Swagger/OpenAPI 3.0
- **依赖管理**: Go Modules

## 📄 参考资料

- [火山引擎对象存储TOS API文档](https://www.volcengine.com/docs/6349/74837)
- [TOS Go SDK文档](https://github.com/volcengine/ve-tos-golang-sdk)
- [Gin Web框架文档](https://gin-gonic.com/zh-cn/docs/)
- [Swagger/OpenAPI文档](https://swagger.io/docs/)

## license
Apache-2.0

## Thanks
* 感谢 Jinpu Hu 对本项目的前端架构建议

* 感谢 Weibin Ma 对 ai 相关技术的讲解

* 感谢 claude code 和 instcopilot 提供 ai 相关能力
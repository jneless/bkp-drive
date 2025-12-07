# bkp-drive (不靠谱网盘)

> 基于火山引擎TOS的个人网盘 🚀 从不靠谱到靠谱的路上

![](src/img/logo_v1.gif)

## 🎉 最新更新 (2025-12-07)

### ⭐ AI 文件理解功能上线
- 集成火山引擎 ARK 平台多模态理解能力
- 支持图片、视频、PDF 文档的智能分析
- "不靠谱助手"侧栏实时对话
- SSE 流式响应，实时显示分析结果

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
- **文件删除**: 支持单文件和批量删除（含文件夹递归删除）
- **文件夹操作**: 创建文件夹、文件夹导航
- **文件预览**:
  - 图片在线预览（支持缩略图）
  - 视频首帧预览（基于TOS视频处理）
  - PDF 文档预览占位（可通过 AI 助手理解内容）

### 🤖 AI 文件理解 (新增)
- **多模态理解**: 基于火山引擎 ARK 平台的 doubao-seed-1-6-251015 模型
- **图片分析**: 自动识别图片内容、场景、文字等
- **视频分析**: 理解视频内容和关键帧
- **PDF 文档理解**: 提取和总结文档内容
- **智能对话**: "不靠谱助手"侧栏，支持多轮对话
- **流式输出**: 实时显示 AI 分析结果

### 🎨 界面特色
- **Apple风格首页**: 现代化的用户界面设计
- **响应式设计**: 支持桌面和移动设备
- **网格/列表视图**: 多种文件显示模式
- **面包屑导航**: 便捷的文件路径导航

## 🚀 快速开始

### 环境变量配置

```bash
# TOS 存储配置
export TOS_ENDPOINT="https://tos-cn-beijing.volces.com"
export TOS_REGION="cn-beijing"
export TOS_ACCESS_KEY="your-access-key"
export TOS_SECRET_KEY="your-secret-key"
export TOS_BUCKET_NAME="your-bucket-name"

# Supabase PostgreSQL 数据库配置 (使用Session Pooler - IPv4兼容)
export DATABASE_URL="postgresql://postgres.[PROJECT_REF]:[YOUR_PASSWORD]@aws-1-[region].pooler.supabase.com:5432/postgres"

# JWT密钥
export JWT_SECRET="your-jwt-secret-key"

# ARK AI 平台配置 (新增 - 用于文件内容理解)
export ARK_API_KEY="your-ark-api-key"
# 获取ARK API Key: https://console.volcengine.com/ark/region:ark+cn-beijing/apikey
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
# - DATABASE_URL (PostgreSQL连接字符串)
# - TOS_ENDPOINT, TOS_REGION, TOS_ACCESS_KEY, TOS_SECRET_KEY, TOS_BUCKET_NAME (TOS存储)
# - JWT_SECRET (JWT认证密钥)
# - ARK_API_KEY (ARK AI平台密钥) ⭐ 新增
```

## 📖 API 文档

### 核心API端点

#### 认证相关
- `POST /api/register` - 用户注册（需要邀请码: "bkp"）
- `POST /api/login` - 用户登录
- `GET /api/health` - 健康检查

#### 文件操作
- `GET /api/files` - 列出文件（支持prefix参数）
- `POST /api/upload` - 上传文件
- `POST /api/folders` - 创建文件夹
- `DELETE /api/files/{path}` - 删除文件/文件夹
- `POST /api/batch/delete` - 批量删除
- `GET /api/download/{path}` - 下载文件（支持TOS处理参数）

#### AI 功能 (新增)
- `POST /api/ark/upload` - 上传文件到ARK平台进行预处理
- `POST /api/ark/chat` - 与AI助手对话（支持SSE流式响应）

### Swagger 文档
访问 http://localhost:18666/swagger/index.html 查看完整的API文档

### 核心模块说明

#### 🌐 Vercel Serverless API (`api/index.go`)
- **统一入口**: 处理所有 `/api/*` 路由
- **认证模块**: 用户注册、登录、JWT验证
- **文件操作**: 上传、下载、删除、列表、文件夹管理
- **ARK集成**: 文件上传到AI平台、流式对话响应
- **TOS集成**: 对象存储操作、图片视频处理

#### 🎨 前端界面 (`public/`)
- **pan.html**: 主文件管理界面
- **script.js**: 文件操作、AI对话、预览功能
- **style.css**: 响应式样式、模态框、聊天界面
- **index.html**: Apple风格首页
- **login.html / register.html**: 认证页面

#### 📚 参考文档 (`reference/ark/`)
- ARK多模态理解示例代码
- Responses API流式响应示例

## 💡 特色亮点

### AI 文件理解实现细节
1. **文件类型智能识别**
   - 图片文件使用 `ContentItem_Image` 类型
   - 视频文件使用 `ContentItem_Video` 类型
   - PDF文档使用 `ContentItem_File` 类型
   - 自动保留文件扩展名以确保ARK平台正确识别

2. **流式响应体验**
   - 使用 Server-Sent Events (SSE) 实现实时输出
   - 最小加载动画显示时间（500ms）优化用户体验
   - 支持增量式显示 AI 分析结果

3. **智能对话上下文**
   - 首次打开文件自动发送 "这个文件是什么内容"
   - 支持多轮对话，保持上下文连贯性
   - 文件内容仅在首次对话时上传，后续对话复用

4. **视频处理优化**
   - 使用 TOS 的 `video/snapshot` 功能提取首帧
   - 参数: `t_0,f_jpg,w_0,h_0,m_fast` 快速生成预览
   - 点击视频预览图显示提示信息

### 架构优势
- **无服务器部署**: Vercel Serverless Functions，按需计费
- **全球CDN加速**: Vercel边缘网络分发
- **安全认证**: JWT + bcrypt 双重保障
- **弹性存储**: TOS对象存储，无限扩展
- **智能AI**: 火山引擎ARK平台，多模态理解

## 🔧 技术栈

- **后端**: Go 1.23.4, Gin Web框架 (本地) / Vercel Serverless Functions (生产)
- **存储**: 火山引擎TOS对象存储
- **AI平台**: 火山引擎ARK平台 (doubao-seed-1-6-251015 多模态模型) ⭐ 新增
- **数据库**: Supabase PostgreSQL (Session Pooler)
- **部署**: Vercel Serverless Functions + 本地Go服务器
- **认证**: JWT令牌 (24小时有效期)
- **密码加密**: bcrypt
- **前端**: HTML5, CSS3, Vanilla JavaScript
- **实时通信**: Server-Sent Events (SSE) - 用于 AI 流式响应
- **文档**: Swagger/OpenAPI 3.0
- **依赖管理**: Go Modules

## 📄 参考资料

- [火山引擎对象存储TOS API文档](https://www.volcengine.com/docs/6349/74837)
- [TOS Go SDK文档](https://github.com/volcengine/ve-tos-golang-sdk)
- [火山引擎ARK平台文档](https://www.volcengine.com/docs/82379/1099320) ⭐ 新增
- [ARK多模态模型API](https://www.volcengine.com/docs/82379/1298454) ⭐ 新增
- [Supabase PostgreSQL文档](https://supabase.com/docs/guides/database)
- [Vercel部署文档](https://vercel.com/docs)
- [Gin Web框架文档](https://gin-gonic.com/zh-cn/docs/)
- [Swagger/OpenAPI文档](https://swagger.io/docs/)

## ❓ 常见问题

### Q: AI 文件理解功能需要额外配置吗？
A: 是的，需要在 Vercel 环境变量中配置 `ARK_API_KEY`。可以从 [ARK控制台](https://console.volcengine.com/ark/region:ark+cn-beijing/apikey) 获取。

### Q: 支持哪些文件类型的 AI 理解？
A: 目前支持：
- 图片格式：JPG, PNG, GIF, WebP, BMP
- 视频格式：MP4, AVI, MOV, WMV, FLV, WebM, MKV 等
- 文档格式：PDF

### Q: 视频预览为什么只显示首帧？
A: 基于成本和性能考虑，目前使用 TOS 的 `video/snapshot` 功能提取首帧进行预览。完整视频可通过下载功能获取。

### Q: AI 分析结果会保存吗？
A: 当前版本不保存对话历史。每次打开文件会重新分析。文件在 ARK 平台的缓存有效期内可以快速响应。

### Q: 本地开发如何测试 AI 功能？
A: 在本地 `.env` 文件或环境变量中配置 `ARK_API_KEY`，然后运行 `vercel dev`。

## 🛠️ 故障排查

### 文件上传到 ARK 失败
检查：
1. `ARK_API_KEY` 是否正确配置
2. ARK 控制台是否有可用余额
3. 文件大小是否超过限制（建议 < 100MB）

### AI 对话返回错误
检查：
1. 浏览器控制台的错误信息
2. Vercel 函数日志（`vercel logs`）
3. 文件类型是否受支持

## license
Apache-2.0

## Thanks
* 感谢 Jinpu Hu 对本项目的前端架构建议

* 感谢 Weibin Ma 对 AI 相关技术的讲解

* 感谢 Claude Code 提供 AI 编程能力支持
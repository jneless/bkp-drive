# bkp-drive 网盘扩展功能 API 文档

## 🎯 功能实现状态概览

### ✅ 已完全实现的功能

- ✅ **批量操作** - 批量删除、移动、复制文件
- ✅ **高级文件操作** - 移动、复制、重命名单个文件
- ✅ **文件分享** - 创建带密码和过期时间的分享链接
- ✅ **存储统计** - 详细的存储空间和文件类型统计

### ⚠️ 部分实现/有限制的功能

- ⚠️ **基础搜索** - 仅支持文件名搜索，性能有限
- ⚠️ **最近文件** - 基于修改时间排序，非真实访问历史

### ❌ 未实现的功能

- ❌ **文件版本管理** - 文件历史版本控制
- ❌ **回收站功能** - 删除文件的临时存储和恢复
- ❌ **文件压缩/解压** - ZIP、TAR等格式支持
- ❌ **缩略图生成** - 图片、视频缩略图
- ❌ **全文搜索** - 文档内容搜索
- ❌ **文件同步** - 多端文件同步
- ❌ **访问日志** - 详细的用户操作日志

## 🔧 环境准备

```bash
# 启动服务
go run cmd/server/main.go

# 服务地址
http://localhost:8082

# 健康检查
curl http://localhost:8082/health
```

## 📁 批量操作 API

### 1. 批量删除文件

```bash
curl -X POST http://localhost:8081/api/v1/batch/delete \
  -H "Content-Type: application/json" \
  -d '{
    "items": ["test/file1.txt", "test/file2.txt", "images/photo.jpg"]
  }'
```

### 2. 批量移动文件

```bash
curl -X POST http://localhost:8081/api/v1/batch/move \
  -H "Content-Type: application/json" \
  -d '{
    "items": ["test/file1.txt", "test/file2.txt"],
    "destination": "backup/"
  }'
```

### 3. 批量复制文件

```bash
curl -X POST http://localhost:8081/api/v1/batch/copy \
  -H "Content-Type: application/json" \
  -d '{
    "items": ["documents/report.pdf", "documents/data.xlsx"],
    "destination": "archive/"
  }'
```

## 🔄 高级文件操作 API

### 1. 移动文件

```bash
curl -X PUT http://localhost:8081/api/v1/files/move \
  -H "Content-Type: application/json" \
  -d '{
    "source": "temp/document.pdf",
    "destination": "documents/document.pdf"
  }'
```

### 2. 复制文件

```bash
curl -X PUT http://localhost:8081/api/v1/files/copy \
  -H "Content-Type: application/json" \
  -d '{
    "source": "documents/template.docx",
    "destination": "projects/new-template.docx"
  }'
```

### 3. 重命名文件

```bash
curl -X PUT http://localhost:8081/api/v1/files/rename \
  -H "Content-Type: application/json" \
  -d '{
    "oldKey": "documents/old-name.txt",
    "newKey": "documents/new-name.txt"
  }'
```

## 🔍 搜索功能 API

⚠️ **重要提示：当前搜索功能有重要限制**

### 功能限制说明

1. **仅支持文件名搜索**
   - ❌ 无法搜索文件内容
   - ❌ 不支持全文检索
   - ⚠️ 只能匹配文件路径中的关键词

2. **性能限制**
   - ⚠️ 每次搜索需遍历所有文件
   - ⚠️ 文件数量多时响应缓慢
   - ⚠️ TOS API 限制最多返回1000个对象

3. **搜索算法简单**
   - ⚠️ 仅支持简单字符串包含匹配
   - ❌ 不支持正则表达式
   - ❌ 不支持模糊搜索或相似度匹配

### 1. 基础搜索（有限制）

```bash
# 搜索包含"report"的文件路径
curl "http://localhost:8082/api/v1/search?q=report"

# 在特定文件夹中搜索
curl "http://localhost:8082/api/v1/search?q=photo&folder=images/"
```

### 2. 高级搜索（部分支持）

```bash
# 按文件类型搜索
curl "http://localhost:8082/api/v1/search?types=image,video&limit=20"

# 按大小范围搜索 (单位: bytes)
curl "http://localhost:8082/api/v1/search?minSize=1048576&maxSize=10485760"

# 按时间范围搜索
curl "http://localhost:8082/api/v1/search?startDate=2024-01-01&endDate=2024-12-31"
```

### 3. 获取最近文件（功能有限）

⚠️ **注意：此功能基于文件修改时间，而非真实访问历史**

```bash
# 获取最近20个文件
curl "http://localhost:8082/api/v1/files/recent?limit=20"
```

### 4. 文件过滤

```bash
# 按类型过滤
curl "http://localhost:8082/api/v1/files/filter?type=image"
curl "http://localhost:8082/api/v1/files/filter?type=document"

# 按大小过滤
curl "http://localhost:8082/api/v1/files/filter?size=large"  # >100MB
curl "http://localhost:8082/api/v1/files/filter?size=small" # <10MB
```

## 🔗 文件分享 API

### 1. 创建分享链接

```bash
curl -X POST http://localhost:8082/api/v1/share/create \
  -H "Content-Type: application/json" \
  -d '{
    "fileKey": "documents/presentation.pdf",
    "expiresAt": "2025-12-31T23:59:59Z",
    "password": "abc123",
    "allowDownload": true
  }'
```

### 2. 访问分享内容

```bash
# 无密码分享
curl "http://localhost:8082/api/v1/share/YOUR_SHARE_ID"

# 有密码分享
curl "http://localhost:8082/api/v1/share/YOUR_SHARE_ID?password=abc123"
```

### 3. 下载分享文件

```bash
# 下载分享的文件
curl -o downloaded.pdf "http://localhost:8082/api/v1/share/YOUR_SHARE_ID/download?password=abc123"
```

### 4. 管理分享

```bash
# 查看所有分享
curl "http://localhost:8082/api/v1/share/"

# 删除分享
curl -X DELETE "http://localhost:8082/api/v1/share/YOUR_SHARE_ID"
```

## 📊 存储统计 API

### 获取存储统计信息

```bash
curl "http://localhost:8082/api/v1/stats/storage"
```

**响应示例:**

```json
{
  "success": true,
  "message": "获取存储统计成功",
  "stats": {
    "totalSpace": 107374182400,
    "usedSpace": 1073741824,
    "freeSpace": 106300440576,
    "fileCount": 156,
    "folderCount": 12,
    "fileTypeStats": {
      "image/jpeg": 45,
      "application/pdf": 23,
      "text/plain": 88
    },
    "recentUsage": []
  }
}
```

## 🧪 完整测试流程

### 1. 准备测试文件

```bash
# 创建测试文件
echo "测试文档内容" > test-doc.txt
echo "另一个测试文件" > test-doc2.txt

# 上传文件
curl -X POST http://localhost:8081/api/v1/upload \
  -F "file=@test-doc.txt" \
  -F "folder=test"

curl -X POST http://localhost:8081/api/v1/upload \
  -F "file=@test-doc2.txt" \
  -F "folder=test"
```

### 2. 测试批量操作

```bash
# 批量复制到备份文件夹
curl -X POST http://localhost:8081/api/v1/batch/copy \
  -H "Content-Type: application/json" \
  -d '{
    "items": ["test/test-doc.txt", "test/test-doc2.txt"],
    "destination": "backup/"
  }'
```

### 3. 测试搜索功能

```bash
# 搜索测试文件
curl "http://localhost:8081/api/v1/search?q=test-doc"
```

### 4. 测试分享功能

```bash
# 创建分享
curl -X POST http://localhost:8081/api/v1/share/create \
  -H "Content-Type: application/json" \
  -d '{
    "fileKey": "test/test-doc.txt",
    "expiresAt": "2024-12-31T23:59:59Z",
    "allowDownload": true
  }'

# 记录返回的 shareId，然后访问
curl "http://localhost:8081/api/v1/share/YOUR_SHARE_ID"
```

### 5. 测试统计功能

```bash
# 查看存储统计
curl "http://localhost:8081/api/v1/stats/storage"
```

## 🎯 响应格式说明

所有 API 返回统一的响应格式：

**成功响应:**
```json
{
  "success": true,
  "message": "操作成功",
  "data": { ... }
}
```

**错误响应:**
```json
{
  "success": false,
  "error": "错误描述"
}
```

**批量操作响应:**
```json
{
  "success": true,
  "message": "批量操作完成",
  "processed": 2,
  "failed": 0,
  "failedItems": []
}
```

## 💡 实现细节和限制

### 批量操作实现方式
- **实现方法**: 逐个调用单文件操作API
- **性能影响**: 文件数量多时会有延迟
- **改进方向**: 未来可使用TOS批量API（如果支持）

### 文件复制/移动实现方式
- **实现方法**: 下载源文件 → 重新上传到目标位置 → 删除源文件（移动）
- **性能影响**: 大文件操作会消耗较多带宽和时间
- **改进方向**: 使用TOS服务端复制API（如果支持）

### 搜索功能技术限制

#### 当前实现原理
```go
1. 调用 ListObjectsV2 获取文件列表（最多1000个）
2. 在内存中逐个过滤文件
3. 应用搜索条件：文件名、大小、时间、类型
4. 返回匹配结果
```

#### 具体限制
1. **规模限制**
   - ⚠️ TOS API 单次最多返回1000个对象
   - ⚠️ 超过1000个文件需要分页处理（未实现）
   - ⚠️ 大型网盘（>1万文件）性能会显著下降

2. **搜索能力限制**
   - ❌ 仅支持文件路径关键词匹配
   - ❌ 无法搜索PDF、Word等文档内容
   - ❌ 不支持拼音搜索、近似搜索
   - ❌ 无搜索结果排序和相关性评分

3. **实时性限制**
   - ⚠️ 每次搜索都需要实时查询TOS
   - ❌ 没有搜索结果缓存
   - ❌ 没有增量索引更新

#### 搜索功能改进建议
```go
// 推荐的企业级搜索方案
1. 使用 Elasticsearch 建立文件索引
2. 定期同步文件元数据到搜索引擎
3. 支持全文检索和复杂查询
4. 添加Redis缓存热门搜索结果
```

### 分享功能存储方式
- **当前实现**: 内存存储分享信息
- **重要限制**: 服务重启后分享链接失效
- **生产环境建议**: 使用数据库（MySQL/PostgreSQL）持久化存储

### 存储统计实现方式
- **实现方法**: 实时遍历所有文件计算统计
- **性能影响**: 文件多时响应较慢
- **改进方向**: 定期计算并缓存统计结果

## 🔄 未来功能规划

### Phase 2: 高级功能
- [ ] **文件版本管理**: 支持文件历史版本和回滚
- [ ] **回收站**: 软删除和恢复机制
- [ ] **文件压缩**: ZIP/TAR格式在线压缩解压
- [ ] **缩略图**: 图片和视频缩略图生成

### Phase 3: 企业级功能
- [ ] **全文搜索**: Elasticsearch集成
- [ ] **用户权限**: 多用户和权限管理
- [ ] **文件同步**: 客户端同步支持
- [ ] **API限流**: 防止API滥用

### Phase 4: 性能优化
- [ ] **CDN集成**: 加速文件访问
- [ ] **缓存优化**: Redis缓存热点数据
- [ ] **异步处理**: 大文件操作异步化
- [ ] **监控告警**: 完整的运维监控

## 🔥 高级功能特性（已实现）

- **智能搜索**: 支持多种过滤条件组合（功能有限）
- **安全分享**: 支持密码保护和过期时间
- **批量处理**: 高效的批量文件操作
- **实时统计**: 详细的存储使用情况
- **扩展性强**: 易于添加新的功能模块

## ⚠️ 生产环境注意事项

1. **性能考虑**
   - 文件数量 < 1000时，功能正常
   - 文件数量 > 10000时，搜索等功能会明显变慢

2. **数据持久化**
   - 分享链接存储在内存中，重启服务会丢失
   - 建议生产环境使用数据库存储

3. **安全考虑**
   - 所有API都没有用户认证
   - 分享链接可被任意访问（有密码保护时除外）

现在 bkp-drive 是一个功能基础但实用的网盘后端服务！🎉
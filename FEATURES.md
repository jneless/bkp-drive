# bkp-drive 网盘功能扩展计划

## 核心网盘功能 API 设计

### 1. 批量操作
- `POST /api/v1/batch/delete` - 批量删除文件
- `POST /api/v1/batch/move` - 批量移动文件
- `POST /api/v1/batch/copy` - 批量复制文件
- `POST /api/v1/batch/download` - 批量下载文件

### 2. 文件/文件夹操作
- `PUT /api/v1/files/move` - 移动文件或文件夹
- `PUT /api/v1/files/copy` - 复制文件或文件夹
- `PUT /api/v1/files/rename` - 重命名文件或文件夹
- `GET /api/v1/folders/{path}` - 获取文件夹详细信息

### 3. 搜索和过滤
- `GET /api/v1/search` - 搜索文件
- `GET /api/v1/files/filter` - 按类型/大小/时间过滤文件
- `GET /api/v1/files/recent` - 最近使用的文件

### 4. 文件分享和预览
- `POST /api/v1/share/create` - 创建分享链接
- `GET /api/v1/share/{shareId}` - 访问分享内容
- `DELETE /api/v1/share/{shareId}` - 删除分享链接
- `GET /api/v1/preview/{fileKey}` - 文件预览

### 5. 存储统计
- `GET /api/v1/stats/storage` - 存储空间统计
- `GET /api/v1/stats/usage` - 使用情况统计
- `GET /api/v1/stats/files` - 文件类型统计

### 6. 回收站
- `GET /api/v1/trash` - 查看回收站
- `POST /api/v1/trash/restore` - 恢复文件
- `DELETE /api/v1/trash/empty` - 清空回收站

### 7. 文件版本管理
- `GET /api/v1/files/{fileKey}/versions` - 获取文件版本历史
- `POST /api/v1/files/{fileKey}/versions/restore` - 恢复到指定版本

### 8. 压缩和解压
- `POST /api/v1/compress` - 压缩文件/文件夹
- `POST /api/v1/extract` - 解压文件

### 9. 高级功能
- `GET /api/v1/files/{fileKey}/metadata` - 获取文件元数据
- `PUT /api/v1/files/{fileKey}/metadata` - 更新文件元数据
- `POST /api/v1/files/{fileKey}/thumbnail` - 生成缩略图
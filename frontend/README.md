# 不靠谱网盘 - 前端界面

基于Electron的桌面客户端和Web界面，用于访问不靠谱网盘云存储服务。

## 功能特性

- 📁 文件夹浏览和导航
- 📄 文件预览和下载
- 📤 文件上传
- 🗂️ 新建文件夹
- ✅ 多选操作
- 🗑️ 批量删除
- 🍞 面包屑导航
- 📱 响应式设计

## 项目结构

```
frontend/
├── index.html      # 主页面
├── style.css       # 样式文件
├── script.js       # 主要JavaScript逻辑
├── main.js         # Electron主进程
├── package.json    # 项目配置
└── README.md       # 使用说明
```

## 快速开始

### 1. 启动后端服务

确保后端服务已经启动：
```bash
cd /Users/bytedance/Documents/bkp-drive
go run cmd/server/main.go
```

后端服务将在 http://localhost:8082 上运行

### 2. 运行Web版本

使用Python简单服务器：
```bash
cd frontend
python3 -m http.server 3000
```

然后在浏览器中访问：http://localhost:3000

### 3. 运行Electron桌面版本

首先安装Electron依赖（当网络稳定时）：
```bash
cd frontend
npm install electron --save-dev
```

然后启动应用：
```bash
npm start
# 或者
npx electron .
```

## API配置

前端默认连接到 http://localhost:8082 的后端API。如果需要修改，请编辑 `script.js` 中的 `API_BASE_URL` 常量。

## 使用说明

### 基本操作

1. **浏览文件**: 打开界面后自动显示根目录的文件和文件夹
2. **导航**: 点击文件夹进入，使用面包屑导航返回上级目录
3. **下载文件**: 点击文件名或下载按钮即可下载
4. **上传文件**: 点击"上传文件"按钮选择文件上传
5. **新建文件夹**: 点击"新建文件夹"按钮创建新目录

### 多选操作

1. **选择文件**: 勾选文件前的复选框
2. **全选**: 点击"全选"按钮选择当前页面所有项目
3. **取消选择**: 点击"取消选择"按钮清除所有选择
4. **批量删除**: 选择文件后点击"删除选中"按钮

### 快捷键（Electron版本）

- `F5` - 刷新页面
- `Ctrl+Shift+I` (Windows/Linux) 或 `Alt+Command+I` (macOS) - 开发者工具
- `Ctrl+Q` (Windows/Linux) 或 `Command+Q` (macOS) - 退出应用

## 技术栈

- **前端**: HTML5, CSS3, Vanilla JavaScript
- **桌面应用**: Electron
- **后端通信**: REST API (JSON)
- **样式**: 现代扁平化设计

## 浏览器兼容性

- Chrome 60+
- Firefox 55+
- Safari 12+
- Edge 79+

## 开发说明

### 文件结构说明

- `index.html` - 主页面结构
- `style.css` - 界面样式，采用现代设计
- `script.js` - 核心JavaScript逻辑
  - API通信
  - 文件操作
  - UI交互
  - 状态管理
- `main.js` - Electron主进程配置

### 自定义配置

可以在 `script.js` 顶部修改以下配置：

```javascript
const API_BASE_URL = 'http://localhost:8082/api/v1';  // 后端API地址
```

## 部署建议

### Web版本部署
1. 将frontend目录中的文件部署到Web服务器
2. 配置正确的API地址
3. 确保CORS配置正确

### Electron桌面版本打包
```bash
npm install electron-builder --save-dev
npm run build
```

## 故障排除

1. **无法连接后端**: 检查后端服务是否启动，确认API地址配置
2. **CORS错误**: 确保后端已启用CORS支持
3. **上传失败**: 检查文件大小限制和网络连接
4. **Electron无法启动**: 检查Node.js版本和Electron依赖安装

## 未来计划

- [ ] 图片预览功能
- [ ] 拖拽上传
- [ ] 文件搜索
- [ ] 右键菜单
- [ ] 键盘快捷键
- [ ] 主题切换
- [ ] 离线缓存
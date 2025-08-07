// API配置
const API_BASE_URL = 'http://localhost:18666/api/v1';

// 认证相关
let authToken = null;
let currentUser = null;

// 全局状态
let currentPath = '';
let selectedFiles = new Set();
let allFiles = [];
let uploadCancelToken = null;
let currentViewMode = 'list'; // 'list' 或 'grid'

// DOM元素
const fileList = document.getElementById('file-list');
const breadcrumb = document.getElementById('breadcrumb');
const selectionCount = document.getElementById('selection-count');
const deleteBtn = document.getElementById('delete-selected-btn');
const selectAllBtn = document.getElementById('select-all-btn');
const clearSelectionBtn = document.getElementById('clear-selection-btn');
const refreshBtn = document.getElementById('refresh-btn');
const uploadBtn = document.getElementById('upload-btn');
const uploadFolderBtn = document.getElementById('upload-folder-btn');
const fileInput = document.getElementById('file-input');
const folderInput = document.getElementById('folder-input');
const newFolderBtn = document.getElementById('new-folder-btn');
const folderModal = document.getElementById('folder-modal');
const notificationBanner = document.getElementById('notification-banner');
const notificationClose = document.getElementById('notification-close');
const uploadProgressModal = document.getElementById('upload-progress-modal');
const progressFill = document.getElementById('progress-fill');
const progressText = document.getElementById('progress-text');
const progressPercentage = document.getElementById('progress-percentage');
const uploadDetails = document.getElementById('upload-details');
const cancelUploadBtn = document.getElementById('cancel-upload-btn');
const listViewBtn = document.getElementById('list-view-btn');
const gridViewBtn = document.getElementById('grid-view-btn');
const imagePreviewModal = document.getElementById('image-preview-modal');
const previewImage = document.getElementById('preview-image');
const imageTitle = document.getElementById('image-title');
const imagePath = document.getElementById('image-path');
const imageLoading = document.getElementById('image-loading');
const imageError = document.getElementById('image-error');
const closeImageModal = document.getElementById('close-image-modal');
const downloadImageBtn = document.getElementById('download-image');
const retryImageBtn = document.getElementById('retry-image');
const userInfo = document.getElementById('user-info');
const usernameDisplay = document.getElementById('username-display');
const logoutBtn = document.getElementById('logout-btn');

// =============== 认证相关函数 ===============

// 检查登录状态
function checkAuth() {
    authToken = localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token');
    const userInfoStr = localStorage.getItem('user_info') || sessionStorage.getItem('user_info');
    const expiry = localStorage.getItem('auth_expiry');
    
    if (authToken && userInfoStr) {
        // 检查是否过期
        if (expiry && Date.now() > parseInt(expiry)) {
            // 已过期，清除数据
            clearAuth();
            redirectToLogin();
            return false;
        }
        
        try {
            currentUser = JSON.parse(userInfoStr);
            showUserInfo();
            return true;
        } catch (e) {
            console.error('解析用户信息失败:', e);
            clearAuth();
            redirectToLogin();
            return false;
        }
    } else {
        redirectToLogin();
        return false;
    }
}

// 清除认证信息
function clearAuth() {
    authToken = null;
    currentUser = null;
    localStorage.removeItem('auth_token');
    localStorage.removeItem('user_info');
    localStorage.removeItem('auth_expiry');
    sessionStorage.removeItem('auth_token');
    sessionStorage.removeItem('user_info');
}

// 跳转到登录页面
function redirectToLogin() {
    // 显示未登录的界面状态
    showLoginPrompt();
    return false;
}

// 显示登录提示界面
function showLoginPrompt() {
    const fileListEl = document.getElementById('file-list');
    if (fileListEl) {
        fileListEl.innerHTML = `
            <div style="text-align: center; padding: 60px 20px; color: #666;">
                <div style="font-size: 48px; margin-bottom: 20px;">🔒</div>
                <h2 style="margin-bottom: 10px; color: #333;">需要登录</h2>
                <p style="margin-bottom: 30px;">请先登录您的账号以访问网盘功能</p>
                <div style="gap: 15px; display: flex; justify-content: center; flex-wrap: wrap;">
                    <a href="login.html" style="display: inline-block; padding: 12px 24px; background: #007AFF; color: white; text-decoration: none; border-radius: 8px; font-weight: 500;">立即登录</a>
                    <a href="register.html" style="display: inline-block; padding: 12px 24px; background: #f8f9fa; color: #333; text-decoration: none; border-radius: 8px; font-weight: 500; border: 1px solid #e9ecef;">注册新账号</a>
                </div>
            </div>
        `;
    }
    
    // 隐藏用户信息
    if (userInfo) {
        userInfo.style.display = 'none';
    }
}

// 显示用户信息
function showUserInfo() {
    if (currentUser && usernameDisplay) {
        usernameDisplay.textContent = currentUser.username;
        userInfo.style.display = 'flex';
    }
}

// 登出功能
function logout() {
    clearAuth();
    showAlert('已登出', 'success');
    setTimeout(() => {
        window.location.href = 'login.html';
    }, 1000);
}

// 获取认证头
function getAuthHeaders() {
    if (!authToken) {
        throw new Error('未登录');
    }
    return {
        'Authorization': `Bearer ${authToken}`,
        'Content-Type': 'application/json'
    };
}

// 显示提示消息
function showAlert(message, type) {
    // 创建提示元素
    const alert = document.createElement('div');
    alert.className = `auth-alert alert-${type}`;
    alert.textContent = message;
    alert.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 12px 20px;
        border-radius: 5px;
        color: white;
        font-weight: 500;
        z-index: 10000;
        opacity: 0;
        transition: opacity 0.3s ease;
        ${type === 'success' ? 'background-color: #28a745;' : 'background-color: #dc3545;'}
    `;
    
    document.body.appendChild(alert);
    
    // 显示动画
    setTimeout(() => alert.style.opacity = '1', 100);
    
    // 自动移除
    setTimeout(() => {
        alert.style.opacity = '0';
        setTimeout(() => {
            if (alert.parentNode) {
                alert.parentNode.removeChild(alert);
            }
        }, 300);
    }, 3000);
}

// =============== 原有函数 ===============

// 初始化
document.addEventListener('DOMContentLoaded', function() {
    // 先设置基础的事件监听器，确保页面功能正常
    setupEventListeners();
    initNotificationBanner();
    
    // 然后检查登录状态
    if (!checkAuth()) {
        // 未登录时不加载文件，但页面基础功能仍然可用
        return;
    }
    
    // 已登录才加载文件
    loadFiles();
});

// 设置事件监听器
function setupEventListeners() {
    // 工具栏按钮
    selectAllBtn.addEventListener('click', selectAllFiles);
    clearSelectionBtn.addEventListener('click', clearSelection);
    deleteBtn.addEventListener('click', deleteSelectedFiles);
    refreshBtn.addEventListener('click', () => loadFiles());
    
    // 上传文件
    uploadBtn.addEventListener('click', () => fileInput.click());
    uploadFolderBtn.addEventListener('click', () => folderInput.click());
    fileInput.addEventListener('change', handleFileUpload);
    folderInput.addEventListener('change', handleFolderUpload);
    
    // 新建文件夹
    newFolderBtn.addEventListener('click', showNewFolderModal);
    document.getElementById('create-folder-btn').addEventListener('click', createFolder);
    document.getElementById('cancel-folder-btn').addEventListener('click', hideNewFolderModal);
    
    // 模态框
    folderModal.addEventListener('click', (e) => {
        if (e.target === folderModal) {
            hideNewFolderModal();
        }
    });
    
    // 通知栏关闭按钮
    if (notificationClose) {
        notificationClose.addEventListener('click', closeNotificationBanner);
    }
    
    // 登出按钮
    if (logoutBtn) {
        logoutBtn.addEventListener('click', logout);
    }
    
    // 上传取消按钮
    if (cancelUploadBtn) {
        cancelUploadBtn.addEventListener('click', cancelUpload);
    }
    
    // 上传进度模态框点击外部关闭
    if (uploadProgressModal) {
        uploadProgressModal.addEventListener('click', (e) => {
            if (e.target === uploadProgressModal) {
                // 上传中不允许点击外部关闭
            }
        });
    }
    
    // 视图模式切换
    if (listViewBtn) {
        listViewBtn.addEventListener('click', () => switchViewMode('list'));
    }
    if (gridViewBtn) {
        gridViewBtn.addEventListener('click', () => switchViewMode('grid'));
    }
    
    // 图片预览模态框
    if (closeImageModal) {
        closeImageModal.addEventListener('click', closeImagePreview);
    }
    if (imagePreviewModal) {
        imagePreviewModal.addEventListener('click', (e) => {
            if (e.target === imagePreviewModal) {
                closeImagePreview();
            }
        });
    }
    if (downloadImageBtn) {
        downloadImageBtn.addEventListener('click', downloadCurrentImage);
    }
    if (retryImageBtn) {
        retryImageBtn.addEventListener('click', retryImageLoad);
    }
}

// 加载文件列表
async function loadFiles(path = currentPath) {
    try {
        showLoading();
        // 确保path是字符串
        if (typeof path !== 'string') {
            path = currentPath;
        }
        currentPath = path;
        
        const response = await fetch(`${API_BASE_URL}/files?prefix=${encodeURIComponent(path)}`, {
            headers: getAuthHeaders()
        });
        const result = await response.json();
        
        console.log('API Response:', result); // 调试日志
        
        if (result.success) {
            // 处理API返回的数据结构
            allFiles = [];
            
            // 处理文件夹
            if (result.folders && Array.isArray(result.folders)) {
                result.folders.forEach(folderName => {
                    // 确保文件夹名称格式一致
                    const cleanName = folderName.replace(/\/$/, ''); // 移除末尾斜杠用于显示
                    // 构建完整的文件夹路径
                    const folderKey = (path || '') + cleanName + '/';
                    
                    console.log('Processing folder:', { folderName, cleanName, folderKey, currentPath: path });
                    
                    allFiles.push({
                        name: cleanName,
                        key: folderKey,
                        isFolder: true,
                        size: 0,
                        lastModified: null
                    });
                });
            }
            
            // 处理文件
            if (result.files && Array.isArray(result.files)) {
                result.files.forEach(file => {
                    allFiles.push({
                        name: file.name || file.key,
                        key: file.key,
                        isFolder: false,
                        size: file.size || 0,
                        lastModified: file.lastModified || file.lastModified
                    });
                });
            }
            
            renderFiles(allFiles);
            updateBreadcrumb();
        } else {
            showError('加载文件失败: ' + result.error);
        }
    } catch (error) {
        showError('网络错误: ' + error.message);
    }
}

// 渲染文件列表
function renderFiles(files) {
    // 设置视图模式
    fileList.className = `file-list ${currentViewMode}-view`;
    
    if (files.length === 0) {
        fileList.innerHTML = `
            <div class="empty-state">
                <div class="empty-state-icon">📂</div>
                <p>此文件夹为空</p>
            </div>
        `;
        return;
    }

    const fileItems = files.map(file => createFileItem(file)).join('');
    fileList.innerHTML = fileItems;
    
    // 添加事件监听器
    setupFileItemListeners();
    updateSelectionUI();
}

// 创建文件项HTML
function createFileItem(file) {
    const isFolder = file.isFolder;
    const icon = getFileIcon(file, isFolder);
    const size = isFolder ? '' : formatFileSize(file.size);
    const lastModified = (!isFolder && file.lastModified) ? new Date(file.lastModified).toLocaleDateString('zh-CN') : '';
    
    // 构建meta信息，文件夹不显示大小和时间
    let metaInfo = '';
    if (!isFolder) {
        const metaParts = [];
        if (size) metaParts.push(`大小: ${size}`);
        if (lastModified) metaParts.push(`修改时间: ${lastModified}`);
        metaInfo = metaParts.join('</span><span>');
        if (metaInfo) {
            metaInfo = `<span>${metaInfo}</span>`;
        }
    }
    
    // 检查文件类型以显示适当的下载按钮
    const filename = file.name;
    const ext = filename.toLowerCase().split('.').pop();
    const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp'];
    const isImage = !isFolder && imageExts.includes(ext);
    
    return `
        <div class="file-item" data-path="${file.key}" data-is-folder="${isFolder}">
            <input type="checkbox" class="file-checkbox" data-path="${file.key}">
            ${icon}
            <div class="file-info">
                <div class="file-name">${file.name}</div>
                <div class="file-meta">
                    ${metaInfo}
                </div>
            </div>
            <div class="file-actions">
                ${!isFolder ? `<button class="action-btn download-btn" data-path="${file.key}" title="下载文件">下载</button>` : ''}
                <button class="action-btn delete-btn" data-path="${file.key}" data-is-folder="${isFolder}">删除</button>
            </div>
        </div>
    `;
}

// 设置文件项事件监听器
function setupFileItemListeners() {
    // 文件项点击
    document.querySelectorAll('.file-item').forEach(item => {
        item.addEventListener('click', (e) => {
            if (e.target.type === 'checkbox' || e.target.classList.contains('action-btn')) {
                return;
            }
            
            const path = item.dataset.path;
            const isFolder = item.dataset.isFolder === 'true';
            
            if (isFolder) {
                navigateToFolder(path);
            } else {
                // 检查是否为图片文件
                const filename = item.dataset.path.split('/').pop();
                const ext = filename.toLowerCase().split('.').pop();
                const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp'];
                
                if (imageExts.includes(ext)) {
                    // 图片文件：打开预览
                    openImagePreview(item.dataset.path, filename);
                } else {
                    // 其他文件（包括视频）：不执行任何操作（不提供预览和下载）
                    const videoExts = ['mp4', 'avi', 'mov', 'wmv', 'flv', 'webm', 'mkv', '3gp', 'f4v', 'rmvb'];
                    if (videoExts.includes(ext)) {
                        console.log(`点击了视频文件: ${filename}，当前不支持预览功能`);
                    } else {
                        console.log(`点击了文件: ${filename}，当前不支持预览此类型文件`);
                    }
                }
            }
        });
    });
    
    // 复选框
    document.querySelectorAll('.file-checkbox').forEach(checkbox => {
        checkbox.addEventListener('change', handleCheckboxChange);
    });
    
    // 下载按钮
    document.querySelectorAll('.download-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            downloadFile(btn.dataset.path);
        });
    });
    
    // 删除按钮
    document.querySelectorAll('.delete-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            const filePath = btn.dataset.path;
            const isFolder = btn.dataset.isFolder === 'true';
            deleteFile(filePath, isFolder);
        });
    });
    
    // 异步加载缩略图
    loadThumbnails();
}

// 异步加载所有缩略图
async function loadThumbnails() {
    // 如果用户未登录，跳过缩略图加载
    if (!authToken) {
        return;
    }
    
    // 加载图片缩略图
    const imageContainers = document.querySelectorAll('.file-image-container');
    imageContainers.forEach(async container => {
        const fileKey = container.dataset.fileKey;
        const size = parseInt(container.dataset.size);
        const imgElement = container.querySelector('.file-image-preview');
        const iconElement = container.querySelector('.file-icon');
        
        try {
            const blobUrl = await loadAuthenticatedImage(fileKey, size);
            if (blobUrl) {
                imgElement.src = blobUrl;
                imgElement.style.display = 'block';
                iconElement.style.display = 'none';
                iconElement.classList.remove('loading-thumbnail');
            }
        } catch (error) {
            console.warn(`加载图片缩略图失败: ${fileKey}`, error);
            iconElement.classList.remove('loading-thumbnail');
        }
    });
    
    // 加载视频缩略图
    const videoContainers = document.querySelectorAll('.video-thumbnail-container');
    videoContainers.forEach(async container => {
        const fileKey = container.dataset.fileKey;
        const size = parseInt(container.dataset.size);
        const imgElement = container.querySelector('.file-video-thumbnail');
        const fallbackElement = container.querySelector('.video-fallback-icon');
        const playOverlay = container.querySelector('.video-play-overlay');
        
        try {
            const blobUrl = await loadAuthenticatedVideoThumbnail(fileKey, size);
            if (blobUrl) {
                imgElement.src = blobUrl;
                imgElement.style.display = 'block';
                playOverlay.style.display = 'block';
                fallbackElement.style.display = 'none';
            }
        } catch (error) {
            console.warn(`加载视频缩略图失败: ${fileKey}`, error);
            fallbackElement.style.display = 'block';
        }
    });
}

// 导航到文件夹
function navigateToFolder(folderKey) {
    console.log('Navigating to folder:', folderKey); // 调试日志
    
    // 确保文件夹路径以斜杠结尾
    let folderPath = folderKey;
    if (!folderPath.endsWith('/')) {
        folderPath += '/';
    }
    
    loadFiles(folderPath);
}

// 下载文件
async function downloadFile(filePath) {
    try {
        // 显示下载提示
        showAlert('正在准备下载...', 'success');
        
        const response = await fetch(`${API_BASE_URL}/download/${filePath}`, {
            method: 'GET',
            headers: getAuthHeaders()
        });
        
        if (!response.ok) {
            throw new Error(`下载失败: ${response.status} ${response.statusText}`);
        }
        
        // 获取文件blob
        const blob = await response.blob();
        
        // 从filePath提取文件名
        const fileName = filePath.split('/').pop() || 'download';
        
        // 创建下载链接
        const downloadUrl = window.URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = downloadUrl;
        link.download = fileName;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        
        // 清理blob URL
        window.URL.revokeObjectURL(downloadUrl);
        
        showAlert('下载开始', 'success');
    } catch (error) {
        console.error('下载文件失败:', error);
        showAlert('下载失败: ' + error.message, 'error');
    }
}

// 删除文件或文件夹
async function deleteFile(filePath, isFolder = false) {
    const itemType = isFolder ? '文件夹' : '文件';
    let confirmMessage = `确定要删除这个${itemType}吗？`;
    
    if (isFolder) {
        confirmMessage = `确定要删除文件夹"${filePath.replace(/\/$/, '')}"吗？\n\n⚠️ 警告：这将删除文件夹及其所有内容！`;
    }
    
    if (!confirm(confirmMessage)) {
        return;
    }
    
    console.log('Deleting:', filePath, 'isFolder:', isFolder);
    
    try {
        if (isFolder) {
            // 递归删除文件夹及其内容
            await deleteFolderRecursively(filePath);
        } else {
            // 删除单个文件
            await deleteSingleItem(filePath);
        }
        
        showSuccess(`${itemType}删除成功`);
        loadFiles();
    } catch (error) {
        showError(`删除${itemType}失败: ${error.message}`);
    }
}

// 递归删除文件夹
async function deleteFolderRecursively(folderPath) {
    console.log('开始递归删除文件夹:', folderPath);
    
    // 获取文件夹内的所有内容
    const allItems = await getAllItemsInFolder(folderPath);
    
    if (allItems.length > 0) {
        console.log('文件夹内容:', allItems);
        
        // 使用批量删除API删除所有内容
        const response = await fetch(`${API_BASE_URL}/batch/delete`, {
            method: 'POST',
            headers: getAuthHeaders(),
            body: JSON.stringify({
                items: allItems
            })
        });
        
        const result = await response.json();
        if (!result.success) {
            throw new Error(result.error || '批量删除内容失败');
        }
        
        console.log('批量删除结果:', result);
    }
    
    // 最后删除文件夹本身
    await deleteSingleItem(folderPath);
}

// 获取文件夹内的所有项目（递归）
async function getAllItemsInFolder(folderPath) {
    const allItems = [];
    
    async function scanFolder(path) {
        try {
            const response = await fetch(`${API_BASE_URL}/files?prefix=${encodeURIComponent(path)}`);
            const result = await response.json();
            
            if (!result.success) {
                console.warn('无法扫描文件夹:', path);
                return;
            }
            
            // 添加文件
            if (result.files && Array.isArray(result.files)) {
                result.files.forEach(file => {
                    allItems.push(file.key);
                });
            }
            
            // 递归处理子文件夹
            if (result.folders && Array.isArray(result.folders)) {
                for (const folderName of result.folders) {
                    const subFolderPath = path + folderName + (folderName.endsWith('/') ? '' : '/');
                    await scanFolder(subFolderPath);
                    // 将子文件夹也加入删除列表
                    allItems.push(subFolderPath);
                }
            }
        } catch (error) {
            console.error('扫描文件夹出错:', path, error);
        }
    }
    
    await scanFolder(folderPath);
    
    // 确保父文件夹最后被删除（排序确保深层文件夹先被删除）
    return allItems.sort((a, b) => b.split('/').length - a.split('/').length);
}

// 删除单个项目
async function deleteSingleItem(itemPath) {
    const encodedPath = itemPath.split('/').map(part => encodeURIComponent(part)).join('/');
    const response = await fetch(`${API_BASE_URL}/files/${encodedPath}`, {
        method: 'DELETE',
        headers: getAuthHeaders()
    });
    
    const result = await response.json();
    if (!result.success) {
        throw new Error(result.error || '删除失败');
    }
}

// 复选框变化处理
function handleCheckboxChange(e) {
    const path = e.target.dataset.path;
    if (e.target.checked) {
        selectedFiles.add(path);
    } else {
        selectedFiles.delete(path);
    }
    updateSelectionUI();
}

// 全选
function selectAllFiles() {
    selectedFiles.clear();
    document.querySelectorAll('.file-checkbox').forEach(checkbox => {
        checkbox.checked = true;
        selectedFiles.add(checkbox.dataset.path);
    });
    updateSelectionUI();
}

// 清除选择
function clearSelection() {
    selectedFiles.clear();
    document.querySelectorAll('.file-checkbox').forEach(checkbox => {
        checkbox.checked = false;
    });
    updateSelectionUI();
}

// 删除选中文件
async function deleteSelectedFiles() {
    if (selectedFiles.size === 0) return;
    
    // 检查选中的项目中是否有文件夹
    const selectedItems = Array.from(selectedFiles);
    const folderCount = selectedItems.filter(path => path.endsWith('/')).length;
    const fileCount = selectedItems.length - folderCount;
    
    let confirmMessage = `确定要删除选中的 ${selectedItems.length} 个项目吗？\n\n`;
    if (folderCount > 0) {
        confirmMessage += `⚠️ 包含 ${folderCount} 个文件夹和 ${fileCount} 个文件\n`;
        confirmMessage += `文件夹将被递归删除（包括其所有内容）！`;
    }
    
    if (!confirm(confirmMessage)) {
        return;
    }
    
    try {
        // 收集所有需要删除的项目（包括文件夹内容）
        const allItemsToDelete = new Set();
        
        for (const itemPath of selectedItems) {
            if (itemPath.endsWith('/')) {
                // 这是一个文件夹，需要递归获取其内容
                console.log('处理文件夹:', itemPath);
                const folderContents = await getAllItemsInFolder(itemPath);
                folderContents.forEach(item => allItemsToDelete.add(item));
                allItemsToDelete.add(itemPath);
            } else {
                // 这是一个文件
                allItemsToDelete.add(itemPath);
            }
        }
        
        const finalItemsList = Array.from(allItemsToDelete)
            .sort((a, b) => b.split('/').length - a.split('/').length); // 深层项目优先删除
        
        console.log('最终删除列表:', finalItemsList);
        
        if (finalItemsList.length > 0) {
            const response = await fetch(`${API_BASE_URL}/batch/delete`, {
                method: 'POST',
                headers: getAuthHeaders(),
                body: JSON.stringify({
                    items: finalItemsList
                })
            });
            
            const result = await response.json();
            if (result.success) {
                showSuccess(`批量删除成功，共删除 ${finalItemsList.length} 个项目`);
            } else {
                showError('批量删除失败: ' + result.error);
            }
        }
        
        selectedFiles.clear();
        loadFiles();
    } catch (error) {
        showError('批量删除失败: ' + error.message);
    }
}

// 更新选择UI
function updateSelectionUI() {
    selectionCount.textContent = `已选择: ${selectedFiles.size} 项`;
    deleteBtn.disabled = selectedFiles.size === 0;
    
    // 更新文件项选中状态
    document.querySelectorAll('.file-item').forEach(item => {
        if (selectedFiles.has(item.dataset.path)) {
            item.classList.add('selected');
        } else {
            item.classList.remove('selected');
        }
    });
}

// 更新面包屑导航
function updateBreadcrumb() {
    // 确保currentPath是字符串
    const pathStr = typeof currentPath === 'string' ? currentPath : '';
    const parts = pathStr ? pathStr.split('/').filter(p => p) : [];
    let breadcrumbHTML = '<span class="breadcrumb-item" data-path="">根目录</span>';
    
    let buildPath = '';
    parts.forEach(part => {
        buildPath += part + '/';
        breadcrumbHTML += `<span class="breadcrumb-item" data-path="${buildPath}">${part}</span>`;
    });
    
    breadcrumb.innerHTML = breadcrumbHTML;
    
    // 添加面包屑点击事件
    document.querySelectorAll('.breadcrumb-item').forEach(item => {
        item.addEventListener('click', () => {
            loadFiles(item.dataset.path);
        });
    });
}

// 文件上传
async function handleFileUpload(e) {
    const files = Array.from(e.target.files);
    if (files.length === 0) return;
    
    for (const file of files) {
        try {
            const formData = new FormData();
            formData.append('file', file);
            formData.append('folder', currentPath);
            
            const response = await fetch(`${API_BASE_URL}/upload`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${authToken}`
                },
                body: formData
            });
            
            const result = await response.json();
            if (!result.success) {
                showError(`上传 ${file.name} 失败: ${result.error}`);
            }
        } catch (error) {
            showError(`上传 ${file.name} 失败: ${error.message}`);
        }
    }
    
    showSuccess(`成功上传 ${files.length} 个文件`);
    fileInput.value = '';
    loadFiles();
}

// 新建文件夹相关
function showNewFolderModal() {
    folderModal.style.display = 'block';
    document.getElementById('folder-name').value = '';
    document.getElementById('folder-name').focus();
}

function hideNewFolderModal() {
    folderModal.style.display = 'none';
}

async function createFolder() {
    const folderName = document.getElementById('folder-name').value.trim();
    if (!folderName) {
        alert('请输入文件夹名称');
        return;
    }
    
    // 确保currentPath是字符串
    const pathStr = typeof currentPath === 'string' ? currentPath : '';
    const folderPath = pathStr ? `${pathStr}${folderName}/` : `${folderName}/`;
    
    console.log('Creating folder:', folderPath); // 调试日志
    
    try {
        const response = await fetch(`${API_BASE_URL}/folders`, {
            method: 'POST',
            headers: getAuthHeaders(),
            body: JSON.stringify({
                folderPath: folderPath
            })
        });
        
        const result = await response.json();
        if (result.success) {
            showSuccess('文件夹创建成功');
            hideNewFolderModal();
            loadFiles();
        } else {
            showError('创建文件夹失败: ' + result.error);
        }
    } catch (error) {
        showError('创建文件夹失败: ' + error.message);
    }
}

// 工具函数
function getFileIcon(file, isFolder) {
    if (isFolder) {
        return '<div class="file-icon">📁</div>';
    }
    
    const filename = file.name;
    const ext = filename.toLowerCase().split('.').pop();
    
    // 检查是否为图片文件
    const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp'];
    if (imageExts.includes(ext)) {
        const size = currentViewMode === 'list' ? 32 : 128;
        return `<div class="file-image-container" data-file-key="${file.key}" data-size="${size}">
                    <div class="file-icon loading-thumbnail">${getFileTypeIcon(ext)}</div>
                    <img class="file-image-preview" style="display:none;" alt="${filename}">
                </div>`;
    }
    
    // 检查是否为视频文件
    const videoExts = ['mp4', 'avi', 'mov', 'wmv', 'flv', 'webm', 'mkv', '3gp', 'f4v', 'rmvb'];
    if (videoExts.includes(ext)) {
        const size = currentViewMode === 'list' ? 32 : 128;
        return `<div class="video-thumbnail-container" data-file-key="${file.key}" data-size="${size}">
                    <div class="video-fallback-icon">${getFileTypeIcon(ext)}</div>
                    <img class="file-video-thumbnail" style="display:none;" alt="${filename}">
                    <div class="video-play-overlay" style="display:none;">▶</div>
                </div>`;
    }
    
    return `<div class="file-icon">${getFileTypeIcon(ext)}</div>`;
}

function getFileTypeIcon(ext) {
    const iconMap = {
        'txt': '📄', 'doc': '📄', 'docx': '📄', 'pdf': '📄',
        'jpg': '🖼️', 'jpeg': '🖼️', 'png': '🖼️', 'gif': '🖼️',
        'mp4': '🎬', 'avi': '🎬', 'mov': '🎬',
        'mp3': '🎵', 'wav': '🎵', 'flac': '🎵',
        'zip': '📦', 'rar': '📦', '7z': '📦',
        'js': '📜', 'html': '📜', 'css': '📜', 'json': '📜'
    };
    return iconMap[ext] || '📄';
}

function getImagePreviewUrl(fileKey, size = 128) {
    // 火山引擎TOS图片处理URL格式，只使用宽度参数
    // 参数格式：?x-tos-process=image/resize,w_128
    const baseUrl = `${API_BASE_URL}/download/${fileKey}`;
    return `${baseUrl}?x-tos-process=image/resize,w_${size}`;
}

// 异步加载认证图片并返回blob URL
async function loadAuthenticatedImage(fileKey, size = 128) {
    try {
        const response = await fetch(`${API_BASE_URL}/download/${fileKey}?x-tos-process=image/resize,w_${size}`, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${authToken}`
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}`);
        }
        
        const blob = await response.blob();
        return URL.createObjectURL(blob);
    } catch (error) {
        console.warn('加载认证图片失败:', error);
        return null;
    }
}

// 异步加载认证视频缩略图并返回blob URL
async function loadAuthenticatedVideoThumbnail(fileKey, size = 128) {
    try {
        const response = await fetch(`${API_BASE_URL}/download/${fileKey}?x-tos-process=video/snapshot,t_0,w_${size},h_${size},f_jpg`, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${authToken}`
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}`);
        }
        
        const blob = await response.blob();
        return URL.createObjectURL(blob);
    } catch (error) {
        console.warn('加载认证视频缩略图失败:', error);
        return null;
    }
}

function getVideoThumbnailUrl(fileKey, size = 128) {
    // 火山引擎TOS视频缩略图URL格式
    // 参数说明：t_0 表示首帧截图，w_128,h_128 表示缩略图尺寸，f_jpg 表示jpg格式
    const baseUrl = `${API_BASE_URL}/download/${fileKey}`;
    const thumbnailUrl = `${baseUrl}?x-tos-process=video/snapshot,t_0,w_${size},h_${size},f_jpg`;
    
    return thumbnailUrl;
}

// 视图模式切换
function switchViewMode(mode) {
    currentViewMode = mode;
    
    // 更新按钮状态
    if (listViewBtn && gridViewBtn) {
        listViewBtn.classList.toggle('active', mode === 'list');
        gridViewBtn.classList.toggle('active', mode === 'grid');
    }
    
    // 重新渲染文件列表
    renderFiles(allFiles);
}

function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function showLoading() {
    fileList.innerHTML = '<div class="loading">加载中...</div>';
}

function showError(message) {
    console.error(message);
    alert(message);
}

function showSuccess(message) {
    console.log(message);
    // 这里可以添加更优雅的成功提示
}

// 通知栏相关功能
function initNotificationBanner() {
    // 每次页面加载时显示通知栏
    if (notificationBanner) {
        notificationBanner.style.display = 'block';
        document.body.classList.remove('notification-hidden');
    }
}

function closeNotificationBanner() {
    if (notificationBanner) {
        // 添加关闭动画
        notificationBanner.classList.add('closing');
        
        // 等待动画完成后隐藏元素并调整body padding
        setTimeout(() => {
            notificationBanner.style.display = 'none';
            document.body.classList.add('notification-hidden');
        }, 300);
    }
}

// 文件夹上传功能
async function handleFolderUpload(e) {
    const files = Array.from(e.target.files);
    if (files.length === 0) return;

    console.log('选中的文件夹文件:', files);
    
    // 显示上传进度
    showUploadProgress();
    
    uploadCancelToken = { cancelled: false };
    
    try {
        await uploadFolderFiles(files);
        hideUploadProgress();
        showSuccess(`成功上传文件夹，共 ${files.length} 个文件`);
        folderInput.value = '';
        loadFiles();
    } catch (error) {
        hideUploadProgress();
        if (!uploadCancelToken.cancelled) {
            showError('文件夹上传失败: ' + error.message);
        }
        folderInput.value = '';
    }
}

async function uploadFolderFiles(files) {
    const totalFiles = files.length;
    let uploadedCount = 0;
    const folderStructure = new Map();
    
    // 分析文件夹结构
    files.forEach(file => {
        const relativePath = file.webkitRelativePath || file.name;
        const pathParts = relativePath.split('/');
        
        if (pathParts.length > 1) {
            // 这是文件夹中的文件
            const folders = pathParts.slice(0, -1);
            let currentPath = '';
            
            folders.forEach(folder => {
                currentPath = currentPath ? `${currentPath}/${folder}` : folder;
                if (!folderStructure.has(currentPath)) {
                    folderStructure.set(currentPath, []);
                }
            });
            
            // 将文件添加到对应文件夹
            const fileFolder = pathParts.slice(0, -1).join('/');
            if (!folderStructure.has(fileFolder)) {
                folderStructure.set(fileFolder, []);
            }
            folderStructure.get(fileFolder).push(file);
        }
    });
    
    console.log('文件夹结构分析:', folderStructure);
    
    // 先创建所有需要的文件夹
    updateUploadProgress(0, totalFiles, '正在创建文件夹结构...');
    
    const sortedFolders = Array.from(folderStructure.keys()).sort();
    for (const folderPath of sortedFolders) {
        if (uploadCancelToken.cancelled) throw new Error('上传已取消');
        
        await createFolderIfNeeded(folderPath);
    }
    
    // 然后上传所有文件
    for (const [, folderFiles] of folderStructure) {
        if (uploadCancelToken.cancelled) throw new Error('上传已取消');
        
        for (const file of folderFiles) {
            if (uploadCancelToken.cancelled) throw new Error('上传已取消');
            
            const relativePath = file.webkitRelativePath || file.name;
            const targetPath = currentPath ? `${currentPath}${relativePath}` : relativePath;
            
            updateUploadProgress(uploadedCount, totalFiles, `正在上传: ${relativePath}`);
            
            await uploadSingleFile(file, targetPath);
            uploadedCount++;
            
            updateUploadProgress(uploadedCount, totalFiles, `已上传: ${relativePath}`);
        }
    }
}

async function createFolderIfNeeded(folderPath) {
    const fullFolderPath = currentPath ? `${currentPath}${folderPath}/` : `${folderPath}/`;
    
    try {
        const response = await fetch(`${API_BASE_URL}/folders`, {
            method: 'POST',
            headers: getAuthHeaders(),
            body: JSON.stringify({
                folderPath: fullFolderPath
            })
        });
        
        const result = await response.json();
        console.log(`创建文件夹 ${fullFolderPath}:`, result.success ? '成功' : '失败');
    } catch (error) {
        console.warn(`创建文件夹 ${fullFolderPath} 失败:`, error);
        // 文件夹可能已存在，继续执行
    }
}

async function uploadSingleFile(file, targetPath) {
    const formData = new FormData();
    formData.append('file', file);
    
    // 提取文件夹路径
    const pathParts = targetPath.split('/');
    pathParts.pop(); // 移除文件名
    const folderPath = pathParts.join('/');
    
    if (folderPath) {
        formData.append('folder', folderPath + '/');
    } else {
        formData.append('folder', currentPath);
    }
    
    const response = await fetch(`${API_BASE_URL}/upload`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${authToken}`
        },
        body: formData
    });
    
    const result = await response.json();
    if (!result.success) {
        throw new Error(`上传文件 ${targetPath} 失败: ${result.error}`);
    }
    
    return result;
}

// 上传进度相关函数
function showUploadProgress() {
    if (uploadProgressModal) {
        uploadProgressModal.style.display = 'block';
        updateUploadProgress(0, 100, '准备上传...');
    }
}

function hideUploadProgress() {
    if (uploadProgressModal) {
        uploadProgressModal.style.display = 'none';
    }
}

function updateUploadProgress(current, total, message) {
    const percentage = total > 0 ? Math.round((current / total) * 100) : 0;
    
    if (progressFill) {
        progressFill.style.width = `${percentage}%`;
    }
    
    if (progressPercentage) {
        progressPercentage.textContent = `${percentage}%`;
    }
    
    if (progressText) {
        progressText.textContent = `${current}/${total} 个文件`;
    }
    
    if (uploadDetails) {
        uploadDetails.textContent = message;
    }
}

function cancelUpload() {
    if (uploadCancelToken) {
        uploadCancelToken.cancelled = true;
        hideUploadProgress();
        showSuccess('上传已取消');
    }
}

// 图片预览相关功能
let currentImagePath = '';
let currentImageBlobUrl = null; // 用于管理当前预览图片的blob URL

function openImagePreview(imageFilePath, filename) {
    currentImagePath = imageFilePath;
    
    if (imageTitle) {
        imageTitle.textContent = filename;
    }
    
    // 显示完整的网盘路径
    if (imagePath) {
        imagePath.textContent = `/${imageFilePath}`;
    }
    
    // 显示模态框和加载状态
    if (imagePreviewModal) {
        imagePreviewModal.style.display = 'block';
    }
    
    showImageLoading();
    loadPreviewImage(imageFilePath);
}

function showImageLoading() {
    if (previewImage) previewImage.style.display = 'none';
    if (imageError) imageError.style.display = 'none';
    if (imageLoading) imageLoading.style.display = 'block';
}

function showImageError() {
    if (previewImage) previewImage.style.display = 'none';
    if (imageLoading) imageLoading.style.display = 'none';
    if (imageError) imageError.style.display = 'block';
}

function showImage() {
    if (imageLoading) imageLoading.style.display = 'none';
    if (imageError) imageError.style.display = 'none';
    if (previewImage) previewImage.style.display = 'block';
}

function loadPreviewImage(imagePath) {
    if (!previewImage) return;
    
    // 清理之前的blob URL
    if (currentImageBlobUrl) {
        URL.revokeObjectURL(currentImageBlobUrl);
        currentImageBlobUrl = null;
    }
    
    showImageLoading();
    
    // 使用认证方式加载原图
    loadAuthenticatedOriginalImage(imagePath)
        .then(blobUrl => {
            if (blobUrl) {
                currentImageBlobUrl = blobUrl; // 保存blob URL用于后续清理
                previewImage.src = blobUrl;
                showImage();
            } else {
                console.error('图片加载失败: blob URL为空');
                showImageError();
            }
        })
        .catch(error => {
            console.error('图片加载失败:', error);
            showImageError();
        });
}

// 加载认证的原图（不使用缩略图处理）
async function loadAuthenticatedOriginalImage(fileKey) {
    try {
        const response = await fetch(`${API_BASE_URL}/download/${fileKey}`, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${authToken}`
            }
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        const blob = await response.blob();
        return URL.createObjectURL(blob);
    } catch (error) {
        console.warn('加载认证原图失败:', error);
        return null;
    }
}

function closeImagePreview() {
    if (imagePreviewModal) {
        imagePreviewModal.style.display = 'none';
    }
    
    // 清理blob URL以释放内存
    if (currentImageBlobUrl) {
        URL.revokeObjectURL(currentImageBlobUrl);
        currentImageBlobUrl = null;
    }
    
    currentImagePath = '';
    if (previewImage) {
        previewImage.src = '';
    }
}

function downloadCurrentImage() {
    if (!currentImagePath) return;
    
    // 使用修复后的下载函数
    downloadFile(currentImagePath);
}

function retryImageLoad() {
    if (currentImagePath) {
        showImageLoading();
        loadPreviewImage(currentImagePath);
    }
}
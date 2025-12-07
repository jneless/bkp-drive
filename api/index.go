package handler

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"net/url"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	_ "github.com/lib/pq"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model/file"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model/responses"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
)

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// 设置CORS头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 根据URL路径路由到不同的处理函数
	path := strings.TrimPrefix(r.URL.Path, "/api")
	
	switch {
	case path == "/health":
		handleHealth(w, r)
	case path == "/register":
		handleRegister(w, r)
	case path == "/login":
		handleLogin(w, r)
	case path == "/files" && r.Method == "GET":
		handleFiles(w, r)
	case strings.HasPrefix(path, "/files/") && r.Method == "DELETE":
		handleDeleteFile(w, r)
	case path == "/batch/delete":
		handleBatchDelete(w, r)
	case path == "/upload":
		handleUpload(w, r)
	case path == "/folders":
		handleFolders(w, r)
	case strings.HasPrefix(path, "/download/"):
		handleDownload(w, r)
	case path == "/ark/upload":
		handleArkUpload(w, r)
	case path == "/ark/chat":
		handleArkChat(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "API endpoint not found",
		})
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"message": "bkp-drive API is running on Vercel",
		"version": "2.0.0",
	})
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Method not allowed",
		})
		return
	}

	var req struct {
		Username   string `json:"username"`
		Password   string `json:"password"`
		InviteCode string `json:"invite_code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证邀请码
	if req.InviteCode != "bkp" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "邀请码错误",
		})
		return
	}

	// 验证输入
	if req.Username == "" || len(req.Username) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "用户名至少3个字符",
		})
		return
	}

	if req.Password == "" || len(req.Password) < 6 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "密码至少6个字符",
		})
		return
	}

	// 连接数据库
	db, err := getDBConnection()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "数据库连接失败: " + err.Error(),
		})
		return
	}
	defer db.Close()

	// 检查用户名是否已存在
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", req.Username).Scan(&count)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "检查用户名失败: " + err.Error(),
		})
		return
	}
	if count > 0 {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "用户名已存在",
		})
		return
	}

	// 生成用户ID
	userID, err := generateUniqueUserID(db)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "生成用户ID失败: " + err.Error(),
		})
		return
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "密码加密失败: " + err.Error(),
		})
		return
	}

	// 插入数据库并返回自增ID
	var insertedID int
	err = db.QueryRow(
		"INSERT INTO users (user_id, username, password, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id",
		userID, req.Username, string(hashedPassword),
	).Scan(&insertedID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "创建用户失败: " + err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "注册成功",
		"user": map[string]interface{}{
			"id":         insertedID,
			"user_id":    userID,
			"username":   req.Username,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		},
	})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Method not allowed",
		})
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	// 连接数据库
	db, err := getDBConnection()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "数据库连接失败: " + err.Error(),
		})
		return
	}
	defer db.Close()

	var user struct {
		ID        int       `json:"id"`
		UserID    string    `json:"user_id"`
		Username  string    `json:"username"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	var hashedPassword string

	err = db.QueryRow(
		"SELECT id, user_id, username, password, created_at, updated_at FROM users WHERE username = $1",
		req.Username,
	).Scan(&user.ID, &user.UserID, &user.Username, &hashedPassword, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "用户名或密码错误",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "查询用户失败: " + err.Error(),
		})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "用户名或密码错误",
		})
		return
	}

	// 生成JWT token
	token, err := generateJWTToken(user.UserID, user.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "生成令牌失败: " + err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "登录成功",
		"token":   token,
		"user":    user,
	})
}

func handleFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Method not allowed",
		})
		return
	}

	// 验证JWT token
	if !validateJWTToken(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "未授权访问",
		})
		return
	}

	// 检查TOS环境变量
	bucketName := os.Getenv("TOS_BUCKET_NAME")
	if bucketName == "" {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "TOS_BUCKET_NAME环境变量未设置",
		})
		return
	}

	// 获取TOS客户端
	tosClient, err := getTOSClient()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "TOS服务初始化失败: " + err.Error(),
		})
		return
	}

	prefix := r.URL.Query().Get("prefix")

	// 调用TOS列出对象
	ctx := context.Background()
	input := &tos.ListObjectsV2Input{
		Bucket: bucketName,
		ListObjectsInput: tos.ListObjectsInput{
			Prefix:    prefix,
			MaxKeys:   1000,
			Delimiter: "/",
		},
	}

	output, err := tosClient.ListObjectsV2(ctx, input)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "列出对象失败: " + err.Error(),
		})
		return
	}

	var files []map[string]interface{}
	var folders []string

	// 处理文件
	for _, obj := range output.Contents {
		// 跳过文件夹标记对象（以 / 结尾且大小为0）
		if strings.HasSuffix(obj.Key, "/") && obj.Size == 0 {
			continue
		}

		files = append(files, map[string]interface{}{
			"key":          obj.Key,
			"name":         filepath.Base(obj.Key),
			"size":         obj.Size,
			"lastModified": obj.LastModified,
			"contentType":  getContentTypeFromKey(obj.Key),
			"isFolder":     false,
			"etag":         strings.Trim(obj.ETag, "\""),
		})
	}

	// 处理文件夹（公共前缀）
	for _, commonPrefix := range output.CommonPrefixes {
		folderName := strings.TrimSuffix(commonPrefix.Prefix, "/")
		if folderName != "" {
			folders = append(folders, filepath.Base(folderName))
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "列出对象成功",
		"files":   files,
		"folders": folders,
		"total":   len(files) + len(folders),
	})
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Method not allowed",
		})
		return
	}

	// 验证JWT token
	if !validateJWTToken(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "未授权访问",
		})
		return
	}

	// 解析multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "解析上传文件失败: " + err.Error(),
		})
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "获取上传文件失败: " + err.Error(),
		})
		return
	}
	defer file.Close()

	folder := r.FormValue("folder")
	if folder == "" {
		folder = ""
	} else if !strings.HasSuffix(folder, "/") {
		folder += "/"
	}

	// 构造文件路径
	fileName := fileHeader.Filename
	filePath := folder + fileName

	// 获取TOS客户端
	tosClient, err := getTOSClient()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "TOS服务初始化失败: " + err.Error(),
		})
		return
	}

	bucketName := os.Getenv("TOS_BUCKET_NAME")
	if bucketName == "" {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "TOS_BUCKET_NAME环境变量未设置",
		})
		return
	}

	// 推测Content-Type
	contentType := getContentTypeFromKey(fileName)

	// 上传文件到TOS
	ctx := context.Background()
	input := &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:      bucketName,
			Key:         filePath,
			ContentType: contentType,
		},
		Content: file,
	}

	output, err := tosClient.PutObjectV2(ctx, input)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "文件上传失败: " + err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "文件上传成功",
		"key":     filePath,
		"etag":    strings.Trim(output.ETag, "\""),
		"size":    fileHeader.Size,
	})
}

func handleFolders(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Method not allowed",
		})
		return
	}

	// 验证JWT token
	if !validateJWTToken(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "未授权访问",
		})
		return
	}

	var req struct {
		FolderPath string `json:"folderPath"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	if req.FolderPath == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "文件夹路径不能为空",
		})
		return
	}

	// 确保文件夹路径以 / 结尾
	if !strings.HasSuffix(req.FolderPath, "/") {
		req.FolderPath += "/"
	}

	// 获取TOS客户端
	tosClient, err := getTOSClient()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "TOS服务初始化失败: " + err.Error(),
		})
		return
	}

	bucketName := os.Getenv("TOS_BUCKET_NAME")
	if bucketName == "" {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "TOS_BUCKET_NAME环境变量未设置",
		})
		return
	}

	// 创建文件夹（通过创建一个以 / 结尾的空对象）
	ctx := context.Background()
	input := &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucketName,
			Key:    req.FolderPath,
		},
		Content: strings.NewReader(""),
	}

	_, err = tosClient.PutObjectV2(ctx, input)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "创建文件夹失败: " + err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "文件夹创建成功",
		"path":    req.FolderPath,
	})
}

func handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Method not allowed",
		})
		return
	}

	// 验证JWT token
	if !validateJWTToken(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "未授权访问",
		})
		return
	}

	// 提取文件路径 (去掉 /files/ 前缀)
	path := strings.TrimPrefix(r.URL.Path, "/api/files/")
	if path == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "文件路径不能为空",
		})
		return
	}

	// URL解码文件路径
	decodedPath, err := url.QueryUnescape(path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "文件路径格式错误: " + err.Error(),
		})
		return
	}

	// 获取TOS客户端
	tosClient, err := getTOSClient()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "TOS服务初始化失败: " + err.Error(),
		})
		return
	}

	bucketName := os.Getenv("TOS_BUCKET_NAME")
	if bucketName == "" {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "TOS_BUCKET_NAME环境变量未设置",
		})
		return
	}

	// 删除文件/文件夹
	ctx := context.Background()
	input := &tos.DeleteObjectV2Input{
		Bucket: bucketName,
		Key:    decodedPath,
	}

	_, err = tosClient.DeleteObjectV2(ctx, input)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "删除失败: " + err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "删除成功",
		"path":    decodedPath,
	})
}

func handleBatchDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Method not allowed",
		})
		return
	}

	// 验证JWT token
	if !validateJWTToken(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "未授权访问",
		})
		return
	}

	var req struct {
		Items []string `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	if len(req.Items) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "删除项目列表不能为空",
		})
		return
	}

	// 获取TOS客户端
	tosClient, err := getTOSClient()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "TOS服务初始化失败: " + err.Error(),
		})
		return
	}

	bucketName := os.Getenv("TOS_BUCKET_NAME")
	if bucketName == "" {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "TOS_BUCKET_NAME环境变量未设置",
		})
		return
	}

	// 构造批量删除请求（逐个删除）
	ctx := context.Background()
	
	var successCount, failCount int
	var errors []string

	for _, item := range req.Items {
		input := &tos.DeleteObjectV2Input{
			Bucket: bucketName,
			Key:    item,
		}

		_, err = tosClient.DeleteObjectV2(ctx, input)
		if err != nil {
			failCount++
			errors = append(errors, fmt.Sprintf("%s: %s", item, err.Error()))
		} else {
			successCount++
		}
	}

	if failCount > 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("部分删除失败，成功: %d，失败: %d", successCount, failCount),
			"errors":  errors,
			"deleted": successCount,
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("批量删除成功，共删除 %d 个项目", successCount),
		"deleted": successCount,
	})
}

// 辅助函数
func getDBConnection() (*sql.DB, error) {
	// 优先使用完整的DATABASE_URL
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		db, err := sql.Open("postgres", databaseURL)
		if err != nil {
			return nil, err
		}
		if err = db.Ping(); err != nil {
			return nil, err
		}
		return db, nil
	}

	// 否则使用单独的参数构建连接字符串
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		getEnvOrDefault("SUPABASE_DB_HOST", ""),
		getEnvOrDefault("SUPABASE_DB_PORT", "5432"),
		getEnvOrDefault("SUPABASE_DB_USER", "postgres"),
		os.Getenv("SUPABASE_DB_PASSWORD"),
		getEnvOrDefault("SUPABASE_DB_NAME", "postgres"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func generateUserID() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyz"
	const length = 8

	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}

	return "bkp-" + string(b), nil
}

func generateUniqueUserID(db *sql.DB) (string, error) {
	for i := 0; i < 10; i++ {
		userID, err := generateUserID()
		if err != nil {
			return "", err
		}

		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM users WHERE user_id = $1", userID).Scan(&count)
		if err != nil {
			return "", err
		}

		if count == 0 {
			return userID, nil
		}
	}
	return "", fmt.Errorf("生成唯一用户ID失败")
}

func generateJWTToken(userID, username string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func validateJWTToken(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return false
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	return err == nil && token.Valid
}

// getTOSClient 获取TOS客户端
func getTOSClient() (*tos.ClientV2, error) {
	endpoint := getEnvOrDefault("TOS_ENDPOINT", "https://tos-cn-beijing.volces.com")
	region := getEnvOrDefault("TOS_REGION", "cn-beijing")
	accessKey := os.Getenv("TOS_ACCESS_KEY")
	secretKey := os.Getenv("TOS_SECRET_KEY")

	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("TOS_ACCESS_KEY 和 TOS_SECRET_KEY 环境变量必须设置")
	}

	tosClient, err := tos.NewClientV2(
		endpoint,
		tos.WithRegion(region),
		tos.WithCredentials(tos.NewStaticCredentials(accessKey, secretKey)),
	)
	if err != nil {
		return nil, fmt.Errorf("创建TOS客户端失败: %w", err)
	}

	return tosClient, nil
}

// getContentTypeFromKey 根据文件扩展名推测Content-Type
func getContentTypeFromKey(key string) string {
	ext := strings.ToLower(filepath.Ext(key))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".zip":
		return "application/zip"
	case ".mp4":
		return "video/mp4"
	case ".mp3":
		return "audio/mpeg"
	default:
		return "application/octet-stream"
	}
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// handleDownload 处理文件下载请求
func handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Method not allowed",
		})
		return
	}

	// 验证JWT token
	if !validateJWTToken(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "未授权访问",
		})
		return
	}

	// 提取文件路径 (去掉 /download/ 前缀)
	path := strings.TrimPrefix(r.URL.Path, "/api/download/")
	if path == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "文件路径不能为空",
		})
		return
	}

	// URL解码文件路径
	decodedPath, err := url.QueryUnescape(path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "文件路径格式错误: " + err.Error(),
		})
		return
	}

	// 获取TOS客户端
	tosClient, err := getTOSClient()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "TOS服务初始化失败: " + err.Error(),
		})
		return
	}

	// 检查是否有TOS处理参数（图片缩略图或视频截图）
	processParam := r.URL.Query().Get("x-tos-process")
	bucketName := os.Getenv("TOS_BUCKET_NAME")
	
	ctx := context.Background()

	if processParam != "" {
		// 使用TOS处理功能（缩略图/视频截图）
		processInput := &tos.GetObjectV2Input{
			Bucket:  bucketName,
			Key:     decodedPath,
			Process: processParam,
		}
		
		output, err := tosClient.GetObjectV2(ctx, processInput)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "文件下载失败: " + err.Error(),
			})
			return
		}
		defer output.Content.Close()

		// 设置响应头
		if output.ContentType != "" {
			w.Header().Set("Content-Type", output.ContentType)
		} else {
			// 根据处理类型推测内容类型
			if strings.Contains(processParam, "video/snapshot") {
				w.Header().Set("Content-Type", "image/jpeg")
			} else if strings.Contains(processParam, "image/resize") {
				w.Header().Set("Content-Type", getContentTypeFromKey(decodedPath))
			}
		}

		// 设置文件下载头
		fileName := filepath.Base(decodedPath)
		if strings.Contains(processParam, "video/snapshot") {
			// 视频截图保存为jpg格式
			fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName)) + "_thumb.jpg"
		} else if strings.Contains(processParam, "image/resize") {
			// 图片缩略图保持原格式
			fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName)) + "_thumb" + filepath.Ext(fileName)
		}

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

		// 复制文件内容
		_, err = io.Copy(w, output.Content)
		if err != nil {
			return // 无法再发送错误响应，因为已经开始写入响应体
		}
	} else {
		// 普通文件下载
		input := &tos.GetObjectV2Input{
			Bucket: bucketName,
			Key:    decodedPath,
		}
		
		output, err := tosClient.GetObjectV2(ctx, input)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "文件下载失败: " + err.Error(),
			})
			return
		}
		defer output.Content.Close()

		// 设置响应头
		if output.ContentType != "" {
			w.Header().Set("Content-Type", output.ContentType)
		} else {
			w.Header().Set("Content-Type", getContentTypeFromKey(decodedPath))
		}
		
		if output.ContentLength > 0 {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", output.ContentLength))
		}

		// 设置文件下载头
		fileName := filepath.Base(decodedPath)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

		// 复制文件内容
		_, err = io.Copy(w, output.Content)
		if err != nil {
			return // 无法再发送错误响应，因为已经开始写入响应体
		}
	}
}

// handleArkUpload 处理文件上传到 ark 平台
func handleArkUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Method not allowed",
		})
		return
	}

	// 验证JWT token
	if !validateJWTToken(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "未授权访问",
		})
		return
	}

	// 获取文件路径参数
	filePath := r.URL.Query().Get("file_path")
	if filePath == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "文件路径不能为空",
		})
		return
	}

	fmt.Printf("[ARK Upload] Uploading file: %s\n", filePath)

	// 从 TOS 下载文件
	tosClient, err := getTOSClient()
	if err != nil {
		fmt.Printf("[ARK Upload] TOS client error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "TOS服务初始化失败: " + err.Error(),
		})
		return
	}

	bucketName := os.Getenv("TOS_BUCKET_NAME")
	ctx := context.Background()

	getObjectInput := &tos.GetObjectV2Input{
		Bucket: bucketName,
		Key:    filePath,
	}

	fmt.Printf("[ARK Upload] Downloading from TOS bucket: %s, key: %s\n", bucketName, filePath)

	output, err := tosClient.GetObjectV2(ctx, getObjectInput)
	if err != nil {
		fmt.Printf("[ARK Upload] TOS download error: %v\n", err)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "文件下载失败: " + err.Error(),
		})
		return
	}
	defer output.Content.Close()

	fmt.Printf("[ARK Upload] File downloaded from TOS, size: %d bytes\n", output.ContentLength)

	// 创建临时文件，保留原始文件扩展名
	tmpFile, err := os.CreateTemp("", "ark-upload-*"+filepath.Ext(filePath))
	if err != nil {
		fmt.Printf("[ARK Upload] Failed to create temp file: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "创建临时文件失败: " + err.Error(),
		})
		return
	}
	tmpFilePath := tmpFile.Name()
	defer os.Remove(tmpFilePath) // 确保清理临时文件

	// 将 TOS 内容写入临时文件
	_, err = io.Copy(tmpFile, output.Content)
	if err != nil {
		tmpFile.Close()
		fmt.Printf("[ARK Upload] Failed to write temp file: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "写入临时文件失败: " + err.Error(),
		})
		return
	}
	tmpFile.Close()

	// 重新打开临时文件用于上传
	uploadFile, err := os.Open(tmpFilePath)
	if err != nil {
		fmt.Printf("[ARK Upload] Failed to open temp file: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "打开临时文件失败: " + err.Error(),
		})
		return
	}
	defer uploadFile.Close()

	fmt.Printf("[ARK Upload] Temp file created: %s\n", tmpFilePath)

	// 上传到 ark 平台
	arkApiKey := os.Getenv("ARK_API_KEY")
	if arkApiKey == "" {
		fmt.Printf("[ARK Upload] ARK_API_KEY not set\n")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "ARK_API_KEY环境变量未设置",
		})
		return
	}

	arkClient := arkruntime.NewClientWithApiKey(arkApiKey)

	fmt.Printf("[ARK Upload] Uploading to ARK platform...\n")

	fileInfo, err := arkClient.UploadFile(ctx, &file.UploadFileRequest{
		File:    uploadFile,
		Purpose: file.PurposeUserData,
	})
	if err != nil {
		fmt.Printf("[ARK Upload] ARK upload error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "上传到ark失败: " + err.Error(),
		})
		return
	}

	fmt.Printf("[ARK Upload] File uploaded to ARK, ID: %s, Status: %s\n", fileInfo.ID, fileInfo.Status)

	// 等待文件处理完成
	maxRetries := 30 // 最多等待60秒
	retryCount := 0
	for fileInfo.Status == file.StatusProcessing {
		if retryCount >= maxRetries {
			fmt.Printf("[ARK Upload] Timeout waiting for file processing\n")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "文件处理超时",
			})
			return
		}

		time.Sleep(2 * time.Second)
		retryCount++
		fileInfo, err = arkClient.RetrieveFile(ctx, fileInfo.ID)
		if err != nil {
			fmt.Printf("[ARK Upload] Error retrieving file status: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "获取文件状态失败: " + err.Error(),
			})
			return
		}
		fmt.Printf("[ARK Upload] File status check %d: %s\n", retryCount, fileInfo.Status)
	}

	fmt.Printf("[ARK Upload] File processing completed, Status: %s\n", fileInfo.Status)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "文件上传成功",
		"file_id": fileInfo.ID,
		"status":  fileInfo.Status,
	})
}

// handleArkChat 处理流式对话
func handleArkChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Method not allowed",
		})
		return
	}

	// 验证JWT token
	if !validateJWTToken(r) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "未授权访问",
		})
		return
	}

	// 解析请求
	var req struct {
		FileID   string   `json:"file_id"`
		FileType string   `json:"file_type"`
		Messages []string `json:"messages"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "请求参数错误: " + err.Error(),
		})
		return
	}

	fmt.Printf("[ARK Chat] Received request - FileID: %s, Messages: %d\n", req.FileID, len(req.Messages))

	arkApiKey := os.Getenv("ARK_API_KEY")
	if arkApiKey == "" {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "ARK_API_KEY环境变量未设置",
		})
		return
	}

	arkClient := arkruntime.NewClientWithApiKey(arkApiKey)
	ctx := context.Background()

	// 检查消息数组
	if len(req.Messages) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "消息不能为空",
		})
		return
	}

	// 构建消息
	var inputMessages []*responses.InputItem

	// 如果有文件ID，添加首个带文件的消息
	if req.FileID != "" && len(req.Messages) > 0 {
		// 根据文件类型创建对应的 ContentItem
		var fileContentItem *responses.ContentItem

		switch req.FileType {
		case "image":
			fileContentItem = &responses.ContentItem{
				Union: &responses.ContentItem_Image{
					Image: &responses.ContentItemImage{
						Type:   responses.ContentItemType_input_image,
						FileId: volcengine.String(req.FileID),
					},
				},
			}
		case "video":
			fileContentItem = &responses.ContentItem{
				Union: &responses.ContentItem_Video{
					Video: &responses.ContentItemVideo{
						Type:   responses.ContentItemType_input_video,
						FileId: volcengine.String(req.FileID),
					},
				},
			}
		case "pdf":
			fileContentItem = &responses.ContentItem{
				Union: &responses.ContentItem_File{
					File: &responses.ContentItemFile{
						Type:   responses.ContentItemType_input_file,
						FileId: volcengine.String(req.FileID),
					},
				},
			}
		default:
			// 默认使用 File 类型
			fileContentItem = &responses.ContentItem{
				Union: &responses.ContentItem_File{
					File: &responses.ContentItemFile{
						Type:   responses.ContentItemType_input_file,
						FileId: volcengine.String(req.FileID),
					},
				},
			}
		}

		inputMessage := &responses.ItemInputMessage{
			Role: responses.MessageRole_user,
			Content: []*responses.ContentItem{
				fileContentItem,
				{
					Union: &responses.ContentItem_Text{
						Text: &responses.ContentItemText{
							Type: responses.ContentItemType_input_text,
							Text: req.Messages[0],
						},
					},
				},
			},
		}
		inputMessages = append(inputMessages, &responses.InputItem{
			Union: &responses.InputItem_InputMessage{
				InputMessage: inputMessage,
			},
		})

		// 添加后续消息（不包含文件）
		for i := 1; i < len(req.Messages); i++ {
			inputMessage := &responses.ItemInputMessage{
				Role: responses.MessageRole_user,
				Content: []*responses.ContentItem{
					{
						Union: &responses.ContentItem_Text{
							Text: &responses.ContentItemText{
								Type: responses.ContentItemType_input_text,
								Text: req.Messages[i],
							},
						},
					},
				},
			}
			inputMessages = append(inputMessages, &responses.InputItem{
				Union: &responses.InputItem_InputMessage{
					InputMessage: inputMessage,
				},
			})
		}
	} else {
		// 仅文本消息
		for _, msg := range req.Messages {
			inputMessage := &responses.ItemInputMessage{
				Role: responses.MessageRole_user,
				Content: []*responses.ContentItem{
					{
						Union: &responses.ContentItem_Text{
							Text: &responses.ContentItemText{
								Type: responses.ContentItemType_input_text,
								Text: msg,
							},
						},
					},
				},
			}
			inputMessages = append(inputMessages, &responses.InputItem{
				Union: &responses.InputItem_InputMessage{
					InputMessage: inputMessage,
				},
			})
		}
	}

	createResponsesReq := &responses.ResponsesRequest{
		Model: "doubao-seed-1-6-251015",
		Input: &responses.ResponsesInput{
			Union: &responses.ResponsesInput_ListValue{
				ListValue: &responses.InputItemList{ListValue: inputMessages},
			},
		},
		Thinking: &responses.ResponsesThinking{Type: responses.ThinkingType_enabled.Enum()},
	}

	fmt.Printf("[ARK Chat] Creating stream request...\n")

	// 设置SSE响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// 重新设置 CORS 头（确保不被覆盖）
	w.Header().Set("Access-Control-Allow-Origin", "*")

	resp, err := arkClient.CreateResponsesStream(ctx, createResponsesReq)
	if err != nil {
		fmt.Printf("[ARK Chat] Stream creation error: %v\n", err)
		// 发送错误事件
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		return
	}

	fmt.Printf("[ARK Chat] Stream created successfully\n")

	// 尝试获取 Flusher,但不强制要求(Vercel 环境可能有不同实现)
	flusher, _ := w.(http.Flusher)

	for {
		event, err := resp.Recv()
		if err == io.EOF {
			fmt.Printf("[ARK Chat] Stream ended (EOF)\n")
			fmt.Fprintf(w, "event: done\ndata: \n\n")
			if flusher != nil {
				flusher.Flush()
			}
			break
		}
		if err != nil {
			fmt.Printf("[ARK Chat] Stream error: %v\n", err)
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
			if flusher != nil {
				flusher.Flush()
			}
			break
		}

		eventType := event.GetEventType()
		// fmt.Printf("[ARK Chat] Received event: %s\n", eventType) // 可选：详细日志

		// 处理不同类型的事件
		switch eventType {
		case responses.EventType_response_reasoning_summary_text_delta.String():
			// 推理文本增量
			delta := event.GetReasoningText().GetDelta()
			data, _ := json.Marshal(map[string]string{"type": "reasoning", "delta": delta})
			fmt.Fprintf(w, "data: %s\n\n", data)
			if flusher != nil {
				flusher.Flush()
			}

		case responses.EventType_response_output_text_delta.String():
			// 输出文本增量
			delta := event.GetText().GetDelta()
			data, _ := json.Marshal(map[string]string{"type": "output", "delta": delta})
			fmt.Fprintf(w, "data: %s\n\n", data)
			if flusher != nil {
				flusher.Flush()
			}

		case responses.EventType_response_output_text_done.String():
			// 输出完成
			text := event.GetTextDone().GetText()
			data, _ := json.Marshal(map[string]string{"type": "complete", "text": text})
			fmt.Fprintf(w, "data: %s\n\n", data)
			if flusher != nil {
				flusher.Flush()
			}
		}
	}

	fmt.Printf("[ARK Chat] Request completed\n")
}
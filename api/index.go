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
	_ "github.com/go-sql-driver/mysql"
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
	case path == "/files":
		handleFiles(w, r)
	case path == "/upload":
		handleUpload(w, r)
	case strings.HasPrefix(path, "/download/"):
		handleDownload(w, r)
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
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", req.Username).Scan(&count)
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

	// 插入数据库
	result, err := db.Exec(
		"INSERT INTO users (user_id, username, password, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())",
		userID, req.Username, string(hashedPassword),
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "创建用户失败: " + err.Error(),
		})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "获取用户ID失败: " + err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "注册成功",
		"user": map[string]interface{}{
			"id":         int(id),
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
		"SELECT id, user_id, username, password, created_at, updated_at FROM users WHERE username = ?",
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

	// 暂时返回上传成功的占位响应
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "文件上传成功",
		"key":     "placeholder-file-key",
		"url":     "https://example.com/placeholder-file",
	})
}

// 辅助函数
func getDBConnection() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("MYSQL_USERNAME"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DATABASE"),
	)

	db, err := sql.Open("mysql", dsn)
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
		err = db.QueryRow("SELECT COUNT(*) FROM users WHERE user_id = ?", userID).Scan(&count)
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
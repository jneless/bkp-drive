package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "bkp-drive/docs" // 导入生成的swagger文档
	"bkp-drive/internal/handlers"
	"bkp-drive/internal/middleware"
	"bkp-drive/internal/services"
	"bkp-drive/pkg/config"
	"bkp-drive/pkg/database"
	"bkp-drive/pkg/tos"
)

// @title           bkp-drive API
// @version         2.0.0
// @description     基于火山引擎TOS的云存储后端服务 - 不靠谱网盘API文档
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://github.com/jneless/bkp-drive
// @contact.email  support@bkp-drive.com

// @license.name  MIT
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:18666
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/

func main() {
	log.Println("启动 bkp-drive HTTP 服务...")

	cfg := config.LoadConfig()

	// 初始化数据库连接
	if err := database.InitDB(cfg); err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}
	defer database.CloseDB()

	// 初始化JWT
	middleware.InitJWT(cfg.JWTSecret)

	tosClient, err := tos.NewTOSClient(cfg)
	if err != nil {
		log.Fatalf("创建TOS客户端失败: %v", err)
	}

	if err := tosClient.EnsureBucketExists(); err != nil {
		log.Fatalf("存储桶操作失败: %v", err)
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 静态文件服务
	r.Static("/static", "./public")
	r.StaticFile("/", "./public/index.html")
	r.StaticFile("/index.html", "./public/index.html")
	r.StaticFile("/pan.html", "./public/pan.html")
	r.StaticFile("/login.html", "./public/login.html")
	r.StaticFile("/register.html", "./public/register.html")
	r.StaticFile("/swagger.html", "./public/swagger.html")

	// 创建处理器
	fileHandler := handlers.NewFileHandler(tosClient)
	advancedHandler := handlers.NewAdvancedHandler(tosClient)
	shareHandler := handlers.NewShareHandler(tosClient)

	// 用户服务和认证处理器
	userService := services.NewUserService(cfg.JWTSecret)
	authHandler := handlers.NewAuthHandler(userService)

	api := r.Group("/api/v1")
	{
		// 认证相关API (不需要登录)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/profile", middleware.AuthMiddleware(), authHandler.GetProfile)
		}

		// 文件操作API (需要登录)
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// 基础文件操作
			protected.POST("/upload", fileHandler.UploadFile)
			protected.GET("/files", fileHandler.ListFiles)
			protected.GET("/download/*key", fileHandler.DownloadFile)
			protected.DELETE("/files/*key", fileHandler.DeleteFile)
			protected.POST("/folders", fileHandler.CreateFolder)

			// 高级文件操作
			protected.PUT("/files/move", advancedHandler.MoveFile)
			protected.PUT("/files/copy", advancedHandler.CopyFile)
			protected.PUT("/files/rename", advancedHandler.RenameFile)

			// 批量操作
			batch := protected.Group("/batch")
			{
				batch.POST("/delete", advancedHandler.BatchDelete)
				batch.POST("/move", advancedHandler.BatchMove)
				batch.POST("/copy", advancedHandler.BatchCopy)
			}

			// 搜索和过滤
			protected.GET("/search", advancedHandler.SearchFiles)
			protected.GET("/files/recent", advancedHandler.GetRecentFiles)
			protected.GET("/files/filter", advancedHandler.FilterFiles)

			// 存储统计
			protected.GET("/stats/storage", advancedHandler.GetStorageStats)

			// 分享功能
			share := protected.Group("/share")
			{
				share.POST("/create", shareHandler.CreateShare)
				share.GET("/:shareId", shareHandler.AccessShare)
				share.GET("/:shareId/download", shareHandler.DownloadSharedFile)
				share.DELETE("/:shareId", shareHandler.DeleteShare)
				share.GET("/", shareHandler.ListShares)
			}
		}
	}

	// Swagger 文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API健康检查和示例 (移动到/api路径)
	r.GET("/api", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":      "ok",
			"service":     "bkp-drive (不靠谱网盘)",
			"version":     "2.0.0",
			"description": "基于火山引擎TOS的云存储后端服务",
			"health": gin.H{
				"database": "ok",
				"storage":  "ok",
				"uptime":   "running",
			},
			"features": []string{
				"用户认证",
				"基础文件操作",
				"批量操作",
				"文件搜索",
				"文件分享",
				"存储统计",
			},
			"api_examples": gin.H{
				"register":      "curl -X POST http://localhost:18666/api/v1/auth/register -H \"Content-Type: application/json\" -d '{\"username\":\"test\",\"password\":\"password\"}'",
				"login":         "curl -X POST http://localhost:18666/api/v1/auth/login -H \"Content-Type: application/json\" -d '{\"username\":\"test\",\"password\":\"password\"}'",
				"upload_file":   "curl -X POST http://localhost:18666/api/v1/upload -H \"Authorization: Bearer YOUR_TOKEN\" -F \"file=@test.txt\" -F \"folder=documents\"",
				"list_files":    "curl -H \"Authorization: Bearer YOUR_TOKEN\" http://localhost:18666/api/v1/files",
				"download_file": "curl -H \"Authorization: Bearer YOUR_TOKEN\" -o downloaded.txt \"http://localhost:18666/api/v1/download/documents/test.txt\"",
				"delete_file":   "curl -X DELETE -H \"Authorization: Bearer YOUR_TOKEN\" \"http://localhost:18666/api/v1/files/documents/test.txt\"",
				"create_folder": "curl -X POST -H \"Authorization: Bearer YOUR_TOKEN\" http://localhost:18666/api/v1/folders -H \"Content-Type: application/json\" -d '{\"name\":\"new-folder\"}'",
				"batch_delete":  "curl -X POST -H \"Authorization: Bearer YOUR_TOKEN\" http://localhost:18666/api/v1/batch/delete -H \"Content-Type: application/json\" -d '{\"items\":[\"file1.txt\",\"file2.txt\"]}'",
				"search_files":  "curl -H \"Authorization: Bearer YOUR_TOKEN\" \"http://localhost:18666/api/v1/search?q=document&limit=10\"",
				"create_share":  "curl -X POST -H \"Authorization: Bearer YOUR_TOKEN\" http://localhost:18666/api/v1/share/create -H \"Content-Type: application/json\" -d '{\"fileKey\":\"documents/report.pdf\",\"password\":\"123456\",\"allowDownload\":true}'",
				"storage_stats": "curl -H \"Authorization: Bearer YOUR_TOKEN\" http://localhost:18666/api/v1/stats/storage",
			},
			"documentation": gin.H{
				"api_docs": "查看 API.md 和 API_EXTENDED.md 了解完整API文档",
				"github":   "https://github.com/jneless/bkp-drive",
			},
		})
	})

	// 保持原有的 /health 端点用于简单健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "bkp-drive",
			"version": "2.0.0",
		})
	})

	port := ":18666"
	log.Printf("HTTP服务器启动在端口%s", port)
	log.Printf("主页访问: http://localhost%s/ (Apple风格首页)", port)
	log.Printf("网盘功能: http://localhost%s/pan.html (文件管理)", port)
	log.Printf("用户登录: http://localhost%s/login.html (用户登录)", port)
	log.Printf("用户注册: http://localhost%s/register.html (用户注册)", port)
	log.Printf("API示例: http://localhost%s/api (包含API示例)", port)
	log.Printf("健康检查: http://localhost%s/health", port)
	log.Printf("API文档: http://localhost%s/swagger/index.html", port)
	log.Printf("认证API:")
	log.Printf("  POST   /api/v1/auth/register     - 用户注册")
	log.Printf("  POST   /api/v1/auth/login        - 用户登录")
	log.Printf("  POST   /api/v1/auth/logout       - 用户登出")
	log.Printf("  GET    /api/v1/auth/profile      - 获取用户信息")
	log.Printf("扩展功能API (需要认证):")
	log.Printf("  批量操作:")
	log.Printf("    POST   /api/v1/batch/delete    - 批量删除")
	log.Printf("    POST   /api/v1/batch/move      - 批量移动")
	log.Printf("    POST   /api/v1/batch/copy      - 批量复制")
	log.Printf("  高级操作:")
	log.Printf("    PUT    /api/v1/files/move      - 移动文件")
	log.Printf("    PUT    /api/v1/files/copy      - 复制文件")
	log.Printf("    PUT    /api/v1/files/rename    - 重命名文件")
	log.Printf("  搜索功能:")
	log.Printf("    GET    /api/v1/search          - 搜索文件")
	log.Printf("    GET    /api/v1/files/recent    - 最近文件")
	log.Printf("    GET    /api/v1/files/filter    - 过滤文件")
	log.Printf("  分享功能:")
	log.Printf("    POST   /api/v1/share/create    - 创建分享")
	log.Printf("    GET    /api/v1/share/:id       - 访问分享")
	log.Printf("    DELETE /api/v1/share/:id       - 删除分享")
	log.Printf("  统计功能:")
	log.Printf("    GET    /api/v1/stats/storage   - 存储统计")

	if err := r.Run(port); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

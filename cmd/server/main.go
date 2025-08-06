package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"bkp-drive/internal/handlers"
	"bkp-drive/pkg/config"
	"bkp-drive/pkg/tos"
)

func main() {
	log.Println("启动 bkp-drive HTTP 服务...")

	cfg := config.LoadConfig()

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

	// 创建处理器
	fileHandler := handlers.NewFileHandler(tosClient)
	advancedHandler := handlers.NewAdvancedHandler(tosClient)
	shareHandler := handlers.NewShareHandler(tosClient)

	api := r.Group("/api/v1")
	{
		// 基础文件操作
		api.POST("/upload", fileHandler.UploadFile)
		api.GET("/files", fileHandler.ListFiles)
		api.GET("/download/*key", fileHandler.DownloadFile)
		api.DELETE("/files/*key", fileHandler.DeleteFile)
		api.POST("/folders", fileHandler.CreateFolder)

		// 高级文件操作
		api.PUT("/files/move", advancedHandler.MoveFile)
		api.PUT("/files/copy", advancedHandler.CopyFile)
		api.PUT("/files/rename", advancedHandler.RenameFile)

		// 批量操作
		batch := api.Group("/batch")
		{
			batch.POST("/delete", advancedHandler.BatchDelete)
			batch.POST("/move", advancedHandler.BatchMove)
			batch.POST("/copy", advancedHandler.BatchCopy)
		}

		// 搜索和过滤
		api.GET("/search", advancedHandler.SearchFiles)
		api.GET("/files/recent", advancedHandler.GetRecentFiles)
		api.GET("/files/filter", advancedHandler.FilterFiles)

		// 存储统计
		api.GET("/stats/storage", advancedHandler.GetStorageStats)

		// 分享功能
		share := api.Group("/share")
		{
			share.POST("/create", shareHandler.CreateShare)
			share.GET("/:shareId", shareHandler.AccessShare)
			share.GET("/:shareId/download", shareHandler.DownloadSharedFile)
			share.DELETE("/:shareId", shareHandler.DeleteShare)
			share.GET("/", shareHandler.ListShares)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "bkp-drive",
			"version": "2.0.0",
			"features": []string{
				"基础文件操作",
				"批量操作",
				"文件搜索",
				"文件分享",
				"存储统计",
			},
		})
	})

	port := ":8082"
	log.Printf("HTTP服务器启动在端口%s", port)
	log.Printf("健康检查: http://localhost%s/health", port)
	log.Printf("扩展功能API:")
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
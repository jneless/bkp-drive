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

	fileHandler := handlers.NewFileHandler(tosClient)

	api := r.Group("/api/v1")
	{
		api.POST("/upload", fileHandler.UploadFile)
		api.GET("/files", fileHandler.ListFiles)
		api.GET("/download/*key", fileHandler.DownloadFile)
		api.DELETE("/files/*key", fileHandler.DeleteFile)
		api.POST("/folders", fileHandler.CreateFolder)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "bkp-drive",
			"version": "1.0.0",
		})
	})

	port := ":8081"
	log.Printf("HTTP服务器启动在端口%s", port)
	log.Printf("健康检查: http://localhost%s/health", port)
	log.Printf("API文档:")
	log.Printf("  POST   /api/v1/upload     - 上传文件")
	log.Printf("  GET    /api/v1/files      - 列出文件")
	log.Printf("  GET    /api/v1/download/* - 下载文件")
	log.Printf("  DELETE /api/v1/files/*    - 删除文件")
	log.Printf("  POST   /api/v1/folders    - 创建文件夹")

	if err := r.Run(port); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}
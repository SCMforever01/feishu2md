package server

import (
	"feishu2md/server/internal/config"
	"feishu2md/server/internal/handler"
	"feishu2md/server/internal/logger"
	"feishu2md/server/internal/middlewares"
	"feishu2md/server/internal/repository/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"log"
)

type Server struct {
	Router *gin.Engine
	cfg    *config.Config
}

func NewServer(cfg *config.Config) *Server {
	router := gin.Default()
	return &Server{
		Router: router,
		cfg:    cfg,
	}
}

func (s *Server) Start() {
	s.Router.Run(":" + s.cfg.Port)
}

func StartServer(cfg *config.Config) {
	// 初始化 Gin 引擎
	router := gin.Default()

	// 设置上传文件大小限制（例如 64 MB）
	router.MaxMultipartMemory = 64 << 20 // 64 MB
	// 启用 CORS 中间件，允许 http://localhost:3000 来源的请求
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},                            // 允许的来源
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},                     // 允许的请求方法
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "token"}, // 允许的请求头
		AllowCredentials: true,                                                         // 允许携带凭证
	}))
	// 初始化本地存储服务
	newStorage, err := storage.NewLocalStorage(cfg.Storage.LocalDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	// 注册路由
	handler.RegisterRoutes(router, newStorage)
	protected := router.Group("/v1")
	protected.Use(middlewares.JWTMiddleware()) // 应用 JWT 验证中间件

	// 添加受保护的路由
	protected.GET("/", func(c *gin.Context) {
		// 从上下文中获取 userID
		userID, _ := c.Get("userID")
		c.JSON(200, gin.H{
			"user_id": userID,
			"message": "This is a protected route",
		})
	})
	// 启动服务器
	logger.L.Info("Server is running on port " + cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		logger.L.Fatal("Failed to start server", zap.Error(err))
	}
}

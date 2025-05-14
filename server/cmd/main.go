package main

import (
	"context"
	"feishu2md/server/internal/config"
	"feishu2md/server/internal/logger"
	"feishu2md/server/internal/server"
	"fmt"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 打印配置验证
	fmt.Printf("Server Port: %s\n", cfg.Port)
	fmt.Printf("Feishu AppID: %s\n", cfg.Feishu.AppID)

	// 初始化日志
	ctx := context.Background()
	err := logger.Init(ctx, cfg.LogConfig) // 传递指针
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	// 启动服务器
	server.StartServer(cfg)
}

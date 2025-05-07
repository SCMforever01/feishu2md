package app

import (
	"feishu2md/server/internal/config"
	"feishu2md/server/internal/handler"
	"feishu2md/server/internal/server"
)

type App struct {
	cfg    *config.Config
	server *server.Server
}

func NewApp(cfg *config.Config) *App {
	return &App{
		cfg:    cfg,
		server: server.NewServer(cfg),
	}
}

func (a *App) Run() {
	// 注册路由
	handler.RegisterRoutes(a.server.Router)

	// 启动服务器
	a.server.Start()
}

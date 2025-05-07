package storage

import (
	"context"
	"feishu2md/server/pkg/conf"
	"fmt"
)

// StorageType 存储类型枚举
type StorageType string

const (
	StorageTypeLocal StorageType = "local" // 本地存储
	StorageTypeFDS   StorageType = "fds"   // 远程FDS存储
)

//// Config 存储配置
//type Config struct {
//	Type     StorageType // 存储类型
//	LocalDir string      // 本地存储目录（仅当Type=local时生效）
//	Endpoint string      // FDS服务地址
//	Bucket   string      // 存储桶名称
//}

// ObjectStorage 对象存储接口
type ObjectStorage interface {
	Upload(ctx context.Context, filename string, content []byte) (string, error)
	Type() StorageType
}

// InitStorageClient 初始化存储客户端（工厂方法）
func InitStorageClient(cfg conf.StorageConfig) (ObjectStorage, error) {
	switch cfg.Type {
	case string(StorageTypeLocal):
		return NewLocalStorage(cfg.LocalDir)
	case string(StorageTypeFDS):
		return NewFDSStorage(cfg.Endpoint, cfg.Bucket)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Type)
	}
}

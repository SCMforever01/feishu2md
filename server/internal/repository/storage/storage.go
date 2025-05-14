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

package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"
)

type LocalStorage struct {
	BaseDir     string // 存储根目录
	httpBaseURL string // 访问基础URL（示例）
}

// NewLocalStorage 创建本地存储实例
func NewLocalStorage(baseDir string) (*LocalStorage, error) {
	fmt.Println("创建本地存储实例！！！！")
	// 自动创建存储目录
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorage{
		BaseDir:     baseDir,
		httpBaseURL: "http://localhost:8080/storage", // 假设本地服务映射
	}, nil
}

func (s *LocalStorage) Upload(ctx context.Context, filename string, content []byte) (string, error) {
	// 生成带时间戳的文件名防止冲突
	ext := path.Ext(filename)
	base := filename[0 : len(filename)-len(ext)]
	newFilename := fmt.Sprintf("%s_%d%s",
		base,
		time.Now().UnixNano(),
		ext,
	)

	// 构建完整存储路径
	fullPath := filepath.Join(s.BaseDir, newFilename)
	fmt.Println("Attempting to write file to:", fullPath) // 输出日志检查路径
	// 写入文件
	if err := ioutil.WriteFile(fullPath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// 返回访问URL
	return fmt.Sprintf("%s/%s", s.httpBaseURL, newFilename), nil
}

func (s *LocalStorage) Type() StorageType {
	return StorageTypeLocal
}

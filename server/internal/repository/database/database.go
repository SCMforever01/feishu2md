package database

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// InitializeDB 连接数据库并返回数据库实例
func InitializeDB(dsn string) (*sql.DB, error) {
	// 数据源名称（连接字符串）
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// 测试数据库连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return db, nil
}

package services

import (
	"database/sql"
	"errors"
	"feishu2md/server/internal/model"
	"fmt"
	"time"
)

// TransformService 提供与转换相关的操作
type TransformService struct {
	DB *sql.DB
}

// NewTransformService 创建 TransformService 实例
func NewTransformService(db *sql.DB) *TransformService {
	return &TransformService{DB: db}
}

// CreateTransform 创建一条新的 Transform 记录
func (s *TransformService) CreateTransform(userID int, url string, result string) (*model.Transform, error) {
	// 创建新记录
	transform := &model.Transform{
		UserID:    userID,
		Url:       url,
		Result:    result, // 使用 Result 字段
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 插入数据
	query := `INSERT INTO transform (user_id, url, result, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	resultSet, err := s.DB.Exec(query, transform.UserID, transform.Url, transform.Result, transform.CreatedAt, transform.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("插入 Transform 记录失败: %v", err)
	}

	// 获取插入后的 ID
	id, err := resultSet.LastInsertId() // 使用 LastInsertId 方法
	if err != nil {
		return nil, fmt.Errorf("获取插入后的 ID 失败: %v", err)
	}

	// 设置 ID 并返回 Transform 记录
	transform.ID = int(id)
	return transform, nil
}

// GetTransform 获取一条 Transform 记录（通过 ID）
func (s *TransformService) GetTransform(id int) (*model.Transform, error) {
	var transform model.Transform

	// 根据 ID 查找记录
	query := `SELECT id, user_id, url, result, created_at, updated_at FROM transform WHERE id = ?`
	row := s.DB.QueryRow(query, id)

	// 填充数据
	err := row.Scan(&transform.ID, &transform.UserID, &transform.Url, &transform.Result, &transform.CreatedAt, &transform.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Transform 记录未找到")
		}
		return nil, fmt.Errorf("查询 Transform 记录失败: %v", err)
	}

	return &transform, nil
}

// UpdateTransform 更新一条 Transform 记录
func (s *TransformService) UpdateTransform(id int, url, result string) (*model.Transform, error) {
	var transform model.Transform

	// 查找记录
	query := `SELECT id, user_id, url, result, created_at, updated_at FROM transform WHERE id = ?`
	row := s.DB.QueryRow(query, id)

	// 填充数据
	err := row.Scan(&transform.ID, &transform.UserID, &transform.Url, &transform.Result, &transform.CreatedAt, &transform.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Transform 记录未找到")
		}
		return nil, fmt.Errorf("查询 Transform 记录失败: %v", err)
	}

	// 更新记录
	transform.Url = url
	transform.Result = result
	transform.UpdatedAt = time.Now()

	// 执行更新
	updateQuery := `UPDATE transform SET url = ?, result = ?, updated_at = ? WHERE id = ?`
	_, err = s.DB.Exec(updateQuery, transform.Url, transform.Result, transform.UpdatedAt, transform.ID)
	if err != nil {
		return nil, fmt.Errorf("更新 Transform 记录失败: %v", err)
	}

	return &transform, nil
}

// DeleteTransform 删除一条 Transform 记录
func (s *TransformService) DeleteTransform(id int) error {
	// 查找记录
	query := `SELECT id FROM transform WHERE id = ?`
	row := s.DB.QueryRow(query, id)

	var transformID int
	err := row.Scan(&transformID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("Transform 记录未找到")
		}
		return fmt.Errorf("查询 Transform 记录失败: %v", err)
	}

	// 删除记录
	deleteQuery := `DELETE FROM transform WHERE id = ?`
	_, err = s.DB.Exec(deleteQuery, transformID)
	if err != nil {
		return fmt.Errorf("删除 Transform 记录失败: %v", err)
	}

	return nil
}

// GetHistory 获取某个用户的所有历史记录
func (s *TransformService) GetHistory(userID int) ([]model.Transform, error) {
	var transforms []model.Transform

	// 根据 userID 查找历史记录
	query := `SELECT id, user_id, url, result, created_at, updated_at FROM transform WHERE user_id = ? ORDER BY created_at DESC`
	rows, err := s.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("查询 Transform 记录失败: %v", err)
	}
	defer rows.Close()

	// 遍历查询结果
	for rows.Next() {
		var transform model.Transform
		if err := rows.Scan(&transform.ID, &transform.UserID, &transform.Url, &transform.Result, &transform.CreatedAt, &transform.UpdatedAt); err != nil {
			return nil, fmt.Errorf("扫描 Transform 记录失败: %v", err)
		}
		transforms = append(transforms, transform)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("读取 Transform 记录失败: %v", err)
	}

	return transforms, nil
}

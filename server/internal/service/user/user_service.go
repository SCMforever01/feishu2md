package user

import (
	"database/sql"
	"errors"
	"feishu2md/server/internal/model"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// UserService 提供与用户相关的操作
type UserService struct {
	DB *sql.DB
}

// NewUserService 创建 UserService 实例
func NewUserService(db *sql.DB) *UserService {
	return &UserService{DB: db}
}

// RegisterUser 处理用户注册
func (s *UserService) RegisterUser(phone, password string) (*model.User, error) {
	// 1. 检查手机号是否已存在
	var count int
	query := `SELECT COUNT(*) FROM user WHERE phone = ?`
	err := s.DB.QueryRow(query, phone).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("检查手机号是否已存在失败: %v", err)
	}
	// 如果已存在相同手机号，返回错误
	if count > 0 {
		return nil, fmt.Errorf("手机号已被注册")
	}

	// 2. 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %v", err)
	}

	// 3. 插入用户数据
	query = `INSERT INTO user (phone, password) VALUES (?, ?)`
	result, err := s.DB.Exec(query, phone, string(hashedPassword))
	if err != nil {
		return nil, fmt.Errorf("插入用户数据失败: %v", err)
	}

	// 4. 获取插入后的用户ID
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("获取插入后ID失败: %v", err)
	}

	// 5. 返回用户信息
	user := &model.User{
		ID:       int(id),
		Phone:    phone,
		Password: string(hashedPassword),
	}
	return user, nil
}

// LoginUser 处理用户登录
func (s *UserService) LoginUser(phone, password string) (*model.User, error) {
	var user model.User

	// 根据手机号查找用户
	query := `SELECT id, phone, password FROM user WHERE phone = ?`
	row := s.DB.QueryRow(query, phone)

	// 填充用户数据
	err := row.Scan(&user.ID, &user.Phone, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}

	// 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("密码错误")
	}

	// 密码验证成功，返回用户
	return &user, nil
}

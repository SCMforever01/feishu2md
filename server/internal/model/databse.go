package model

import "time"

// User 代表数据库中的用户表
type User struct {
	ID        int       `json:"id"`
	Phone     string    `json:"phone"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type Transform struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Url       string    `json:"url"`
	Result    string    `json:"result"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

package utils

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"time"
)

// JWT 密钥，签名时需要使用
var jwtKey = []byte("your_secret_key")

// Claims 用来封装 JWT 中的数据
type Claims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

// GenerateJWTToken 生成 JWT Token
func GenerateJWTToken(userID int) (string, error) {
	// 创建声明
	claims := &Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}

	// 创建 JWT Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名生成 token
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT 验证 JWT Token
func ValidateJWT(tokenString string) (*Claims, error) {
	// 解析 JWT Token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 检查 token 的签名方法是否为 HS256
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})

	// 如果解析出错，返回错误
	if err != nil {
		return nil, err
	}

	// 断言并返回解析后的 claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("invalid token")
	}
}

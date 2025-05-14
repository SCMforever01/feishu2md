package model

type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Phone       string `json:"phone" binding:"required"`
	CaptchaID   string `json:"captcha_id" binding:"required"`
	CaptchaCode string `json:"captcha_code" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

package model

// TokenRequest 接收前端请求结构
type TokenRequest struct {
	Code         string `json:"code" binding:"required"`
	RedirectURI  string `json:"redirect_uri"`
	CodeVerifier string `json:"code_verifier"`
	Scope        string `json:"scope"`
}

// FeishuTokenResponse 飞书API响应结构
type FeishuTokenResponse struct {
	Code             int    `json:"code"`
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	Scope            string `json:"scope"`
	Err              string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type FeishuTokenReq struct {
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id" binding:"required"`
	ClientSecret string `json:"client_secret" binding:"required"`
	Code         string `json:"code" binding:"required"`
	RedirectURI  string `json:"redirect_uri"`
	CodeVerifier string `json:"code_verifier"`
	Scope        string `json:"scope"`
}

// 错误定义
var (
	InvalidRequest = Resp{Code: 400, Msg: "Invalid request"}
	AuthFailed     = Resp{Code: 401, Msg: "Authentication failed"}
	InternalError  = Resp{Code: 500, Msg: "Internal server error"}
	FeishuAPIError = Resp{Code: 502, Msg: "Feishu API error"}
)

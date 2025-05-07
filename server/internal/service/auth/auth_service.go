package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"feishu2md/server/internal/config"
	"feishu2md/server/internal/model"
	"fmt"
	"net/http"
	"time"
)

type Service struct {
	client *http.Client
}

func NewAuthService() *Service {
	return &Service{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}
func (s *Service) GetAccessToken(req *model.TokenRequest) (*model.FeishuTokenResponse, error) {
	cfg := config.LoadConfig()
	clientID := cfg.Feishu.AppID
	clientSecret := cfg.Feishu.AppSecret

	if clientID == "" || clientSecret == "" {
		return nil, errors.New("missing feishu credentials")
	}

	// 构造飞书请求体
	payload := map[string]interface{}{
		"grant_type":    "authorization_code",
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          req.Code,
	}

	// 可选参数处理
	if req.RedirectURI != "" {
		payload["redirect_uri"] = req.RedirectURI
	}
	if req.CodeVerifier != "" {
		payload["code_verifier"] = req.CodeVerifier
	}
	if req.Scope != "" {
		payload["scope"] = req.Scope
	}
	jsonData, _ := json.Marshal(payload)

	// 调用飞书API
	resp, err := s.client.Post(
		"https://open.feishu.cn/open-apis/authen/v2/oauth/token",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析响应
	var result model.FeishuTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// 处理飞书错误码
	if result.Code != 0 {
		return nil, fmt.Errorf("feishu error: %d - %s - %s", result.Code, result.Err, result.ErrorDescription)
	}

	return &result, nil
}

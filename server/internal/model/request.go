package model

// Req 定义request的结构体
type Req struct {
	Id                string `json:"id"`
	Url               string `json:"url"`
	Collection        string `json:"collection"`
	AccessKey         string `json:"access_key"`
	UserAccessToken   string `json:"user_access_token"`
	WithImageDownload bool   `json:"with_image_download"`
	IsFile            bool   `json:"is_file"`
}

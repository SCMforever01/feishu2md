package conf

import "time"

type LogConfig struct {
	Path          string `yaml:"path"` // 严格匹配YAML键
	Level         string `yaml:"level"`
	Format        string `yaml:"format"`
	MaxSize       int    `yaml:"max_size"` // 注意下划线
	MaxBackups    int    `yaml:"max_backups"`
	MaxAge        int    `yaml:"max_age"`
	Compress      bool   `yaml:"compress"`
	ConsoleEnable bool   `yaml:"console_enable"` // 必须带下划线
}

type FeishuConfig struct {
	AppID     string `yaml:"app_id"`     //  严格匹配
	AppSecret string `yaml:"app_secret"` //  严格匹配
}

type StorageConfig struct {
	Type     string `yaml:"type"`
	LocalDir string `yaml:"local_dir"`
	Endpoint string `yaml:"endpoint"`
	Bucket   string `yaml:"bucket"`
}

type ImgConfig struct {
	DownloadRate int           `yaml:"download_rate"`
	MaxRetries   int           `yaml:"max_retries"`
	MaxWaitTime  time.Duration `yaml:"max_wait_time"`
}
type CaptchaConfig struct {
	CaptchaType   string        `yaml:"captcha_type"`
	RandomCaptcha bool          `yaml:"random_captcha"`
	ExpireTime    time.Duration `yaml:"expire_time"`
	MaxStore      int           `yaml:"max_store"`
}

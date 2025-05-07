package config

import (
	"feishu2md/server/pkg/conf"
	"fmt"
	"github.com/go-viper/mapstructure/v2"
)
import "github.com/spf13/viper"

type Config struct {
	Port      string              `yaml:"port"`
	LogConfig *conf.LogConfig     `yaml:"log"`    // 改为指针类型
	Feishu    *conf.FeishuConfig  `yaml:"feishu"` // 改为指针类型
	Storage   *conf.StorageConfig `yaml:"storage"`
	ImgConfig *conf.ImgConfig     `yaml:"image"`
	CptConfig *conf.CaptchaConfig `yaml:"captcha"`
}

func LoadConfig() *Config {
	v := viper.New()

	// 配置文件设置
	v.SetConfigName("base")              // 配置文件名称（不带扩展名）
	v.SetConfigType("yaml")              // 配置文件格式
	v.AddConfigPath("./internal/config") // 查找当前目录

	// 读取配置
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("❌ Config file not found in paths:")
		} else {
			fmt.Println("❌ Config file error:", err)
		}
		panic(fmt.Errorf("failed to read config: %w", err))
	} else {
		fmt.Println("✅ Using config file:", v.ConfigFileUsed())
	}

	// 调试输出
	fmt.Println("All Settings:", v.AllSettings())

	// 初始化嵌套字段
	cfg := &Config{
		LogConfig: &conf.LogConfig{},    // 初始化 LogConfig
		Feishu:    &conf.FeishuConfig{}, // 初始化 Feishu
	}

	// 解析配置时添加解码器选项
	if err := v.Unmarshal(cfg, func(decoderConfig *mapstructure.DecoderConfig) {
		decoderConfig.TagName = "yaml"        // 强制使用 yaml 标签
		decoderConfig.WeaklyTypedInput = true // 允许弱类型转换
	}); err != nil {
		panic(err)
	}

	// 验证必要字段
	if cfg.Feishu.AppID == "" {
		panic("Missing feishu.app_id in config")
	}
	if cfg.Feishu.AppSecret == "" {
		panic("Missing feishu.app_secret in config")
	}

	return cfg
}

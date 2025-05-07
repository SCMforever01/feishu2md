package captcha

import (
	"feishu2md/server/internal/config"
	"github.com/mojocn/base64Captcha"
	"image/color"
	"log"
	"math/rand"
	"time"
)

var result base64Captcha.Store

type CaptService struct {
}

func NewCaptService() *CaptService {
	return &CaptService{}
}
func init() {
	cfg := config.LoadConfig()
	expire := cfg.CptConfig.ExpireTime
	if expire == 0 {
		expire = 180 // 默认3分钟
	}
	maxStore := cfg.CptConfig.MaxStore
	if maxStore == 0 {
		maxStore = 20240
	}

	result = base64Captcha.NewMemoryStore(maxStore, expire*time.Second)
}
func (s *CaptService) CreateCode() (string, string, error) {
	var driver base64Captcha.Driver
	cfg := config.LoadConfig()
	var driverType string
	if cfg.CptConfig.RandomCaptcha {
		// 随机模式
		driverList := []struct {
			Type   string
			Driver base64Captcha.Driver
		}{
			{"string", s.stringConfig()},
			{"math", s.mathConfig()},
			{"digit", s.digitConfig()},
		}
		rand.Seed(time.Now().UnixNano())
		selected := driverList[rand.Intn(len(driverList))]
		driver = selected.Driver
		driverType = selected.Type
	} else {
		// 固定模式
		driverType = cfg.CptConfig.CaptchaType
		switch driverType {
		case "string":
			driver = s.stringConfig()
		case "math":
			driver = s.mathConfig()
		case "digit":
			driver = s.digitConfig()
		default:
			panic("生成验证码的类型没有配置，请在yaml文件中配置后再启动项目")
		}
	}

	c := base64Captcha.NewCaptcha(driver, result)
	id, b64s, err := c.Generate()
	if err != nil {
		log.Printf("[Captcha] 生成失败: %v\n", err)
	} else {
		log.Printf("[Captcha] 生成成功，类型: %s，ID: %s\n", driverType, id)
	}
	return id, b64s, err
}

func (s *CaptService) VerifyCaptcha(id, verifyValue string) bool {
	return result.Verify(id, verifyValue, true)
}

func (s *CaptService) GetCodeAnswer(id string) string {
	return result.Get(id, false)
}

// mathConfig 生成图形化算术验证码配置
func (s *CaptService) mathConfig() *base64Captcha.DriverMath {
	return &base64Captcha.DriverMath{
		Height:          50,
		Width:           100,
		NoiseCount:      0,
		ShowLineOptions: base64Captcha.OptionShowHollowLine,
		BgColor: &color.RGBA{
			R: 40, G: 30, B: 89, A: 29,
		},
		Fonts: []string{"./server/internal/service/captcha/Inkfree.ttf"},
	}
}

// digitConfig 生成图形化数字验证码配置
func (s *CaptService) digitConfig() *base64Captcha.DriverDigit {
	return &base64Captcha.DriverDigit{
		Height:   50,
		Width:    100,
		Length:   5,
		MaxSkew:  0.45,
		DotCount: 80,
	}
}

// stringConfig 生成图形化字符串验证码配置
func (s *CaptService) stringConfig() *base64Captcha.DriverString {
	return &base64Captcha.DriverString{
		Height:          100,
		Width:           50,
		NoiseCount:      0,
		ShowLineOptions: base64Captcha.OptionShowHollowLine | base64Captcha.OptionShowSlimeLine,
		Length:          5,
		Source:          "123456789qwertyuiopasdfghjklzxcvb",
		BgColor: &color.RGBA{
			R: 40, G: 30, B: 89, A: 29,
		},
		Fonts: []string{"./server/internal/service/captcha/Inkfree.ttf"},
	}
}

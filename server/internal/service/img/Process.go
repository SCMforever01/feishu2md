package img

import (
	"bytes"
	"context"
	"feishu2md/server/internal/feishu"
	"feishu2md/server/internal/logger"
	"feishu2md/server/internal/model"
	"feishu2md/server/internal/repository/cache"
	"feishu2md/server/internal/repository/storage"
	"feishu2md/server/pkg/conf"
	"feishu2md/server/pkg/utils"
	"fmt"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
	"time"
)

type Processor struct {
	cache     cache.RedisCache      // 缓存接口
	storage   storage.ObjectStorage // 对象存储
	client    feishu.Client         //飞书客户端
	imgConfig conf.ImgConfig
}

func NewProcessor(
	cache cache.RedisCache,
	storage storage.ObjectStorage,
	client feishu.Client,
	cfg conf.ImgConfig,
) *Processor {
	return &Processor{
		cache:     cache,
		storage:   storage,
		client:    client,
		imgConfig: cfg,
	}
}
func (p *Processor) ProcessImages(ctx context.Context, markdown string, tokens []string, req model.Req) (string, error) {
	var (
		wg           sync.WaitGroup
		successCount int64
		limiter      = make(chan struct{}, p.imgConfig.DownloadRate)
		mtx          sync.Mutex
		result       = []byte(markdown)
	)

	startTime := time.Now()
	total := len(tokens)

	for _, token := range tokens {
		wg.Add(1)
		limiter <- struct{}{}

		go func(t string) {
			defer func() {
				<-limiter
				wg.Done()
			}()

			if url := p.processSingleImage(ctx, token, p.imgConfig.MaxRetries, p.imgConfig.MaxWaitTime, req); url != "" {
				atomic.AddInt64(&successCount, 1)
				mtx.Lock()
				result = bytes.Replace(result, []byte(t), []byte(url), 1)
				mtx.Unlock()
			}
		}(token)
	}

	wg.Wait()

	logger.L.Info("图片处理完成",
		zap.Int("total", total),
		zap.Int64("success", successCount),
		zap.Duration("duration", time.Since(startTime)),
	)

	return string(result), nil
}

func (p *Processor) processSingleImage(ctx context.Context, token string, maxRetries int, maxWaitTime time.Duration, req model.Req) string {
	// 1. 检查缓存
	if url, _ := p.cache.GetURL(ctx, token); url != "" {
		return url
	}

	//启用全局令牌桶进行限流，保证全局请求不超过5QPS
	if !p.cache.RetryDownloadRateLimit(ctx, token, maxRetries, maxWaitTime) {
		logger.L.Warn("下载被限流", zap.String("token", token))
		return "" // 超过重试次数或等待超时，跳过该图片
	}
	// 3. 下载并上传
	downloadImageRaw := func() (string, []byte, error) {
		return p.client.DownloadImageRaw(ctx, token, fmt.Sprintf("%s/%s", "images", req.Collection), req.UserAccessToken)
	}
	filename, content, err := utils.ExponentialBackoff(ctx, 0, maxRetries, maxWaitTime, downloadImageRaw)
	if err != nil {
		logger.L.Error("图片下载失败",
			zap.String("token", token),
			zap.Error(err),
		)
		return ""
	}

	url, err := p.storage.Upload(ctx, filename, content)
	if err != nil {
		logger.L.Error("图片上传失败",
			zap.String("token", token),
			zap.Error(err),
		)
		return ""
	}

	// 4. 更新缓存
	if err := p.cache.SetURL(ctx, token, url); err != nil {
		logger.L.Warn("缓存更新失败",
			zap.String("token", token),
			zap.Error(err),
		)
	}

	return url
}

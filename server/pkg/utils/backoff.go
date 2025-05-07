package utils

import (
	"context"
	"feishu2md/server/internal/logger"
	"fmt"
	"go.uber.org/zap"
	"regexp"
	"strconv"
	"time"
)

// 实现一个指数退避算法调用一个函数
//const (
//	maxRetries   = 5
//	maxWaitTime  = 20 * time.Second
//	downloadRate = 5 // 每秒最多下载 5 张图片
//)

var (
	retryableCodes = map[int]bool{
		500: true,
	}
	nonRetryableCodes = map[int]bool{
		403: true,
		404: true,
	}
)

func ExponentialBackoff(ctx context.Context, retry, maxRetries int, maxWaitTime time.Duration, fn func() (string, []byte, error)) (string, []byte, error) {
	if retry >= maxRetries {
		return "", nil, fmt.Errorf("failed after retrying %d times", retry)
	}

	filename, content, err := fn()
	if err == nil {
		return filename, content, nil
	}

	// 提取错误码
	re := regexp.MustCompile(`fail: (\d+)`)
	matches := re.FindStringSubmatch(err.Error())

	if len(matches) <= 1 {
		return "", nil, fmt.Errorf(err.Error())
	}

	code, _ := strconv.Atoi(matches[1])

	if nonRetryableCodes[code] {
		return "", nil, fmt.Errorf("non-retryable error (code: %d): %s", code, err.Error())
	}

	if retryableCodes[code] {
		logger.L.Error("retrying function due to retryable error",
			zap.Int("code: ", code),
			zap.Int("retry:", retry),
			zap.Error(err),
		)
		waitTime := time.Duration(1<<uint(retry)) * time.Second
		if waitTime > maxWaitTime {
			waitTime = maxWaitTime
		}

		select {
		case <-time.After(waitTime):
			return ExponentialBackoff(ctx, retry+1, maxRetries, maxWaitTime, fn)
		case <-ctx.Done():
			return "", nil, ctx.Err()
		}
	}

	return "", nil, fmt.Errorf("unknown error (code: %d): %s", code, err.Error())
}

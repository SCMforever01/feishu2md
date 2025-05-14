package cache

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"
)
import "github.com/redis/go-redis/v9"

type RedisConfig struct {
	Cluster    string
	Host       string
	Port       string
	Password   string
	MaxRetries int
}

type RedisCache struct {
	client *redis.Client
	config RedisConfig
}

// InitRedisClient 根据集群环境初始化客户端
func InitRedisClient(cluster string) *RedisCache {
	//baseStaging := "cWhaUG1POWhKMERlRFhkaEYwX2NDWlZsX3Y3d3k0Tjk="
	//baseProduction := "MEk3NGpUR1k0WF9xY1VtOVVnZnl6OVVBQ09iT3pKQV8="
	baseStaging := "local"
	baseProduction := "test"
	decodePassword := func(encoded string) string {
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return ""
		}
		return string(decoded)
	}

	var config RedisConfig
	switch cluster {
	case "staging":
		config = RedisConfig{
			Host:     "ares.test.common.cache.srv",
			Port:     "22127",
			Password: decodePassword(baseStaging),
			Cluster:  cluster,
		}
	case "production":
		config = RedisConfig{
			Host:     "ares.ai.ai.cache.srv",
			Port:     "5123",
			Password: decodePassword(baseProduction),
			Cluster:  cluster,
		}
	default:
		config = RedisConfig{
			Host:    "127.0.0.1",
			Port:    "6379",
			Cluster: "local",
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password: config.Password,
		DB:       0,
	})

	// 连接健康检查
	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(fmt.Sprintf("Redis connection failed: %v", err))
	} else {
		log.Println("Connected to Redis successfully")
	}

	return &RedisCache{
		client: client,
		config: config,
	}
}

func (c *RedisCache) GetURL(ctx context.Context, imgToken string) (string, error) {
	// 从Redis获取预签名URL
	return c.client.Get(ctx, imgToken).Result()
}

func (c *RedisCache) SetURL(ctx context.Context, imgToken string, url string) error {
	// 设置Redis缓存
	return c.client.Set(ctx, imgToken, url, time.Hour*24*30).Err()
}

// TokenBucketRateLimit 令牌桶限流器
func (c *RedisCache) TokenBucketRateLimit(ctx context.Context, key string, limit int, refillRate int) (bool, error) {
	luaScript := `
        local tokens = tonumber(redis.call("GET", KEYS[1]) or "0")
        local timestamp = tonumber(redis.call("GET", KEYS[2]) or "0")
        local now = tonumber(ARGV[1])
        local refill_rate = tonumber(ARGV[2])
        local capacity = tonumber(ARGV[3])

        -- 计算需要补充的令牌
        local new_tokens = math.min(capacity, tokens + math.floor((now - timestamp) * refill_rate))
        if new_tokens < 1 then
            return 0
        else
            redis.call("SET", KEYS[1], new_tokens - 1)
            redis.call("SET", KEYS[2], now)
            return 1
        end
    `
	now := time.Now().Unix()
	log.Println("程序运行到 TokenBucketRateLimit 函数中")
	result, err := c.client.Eval(ctx, luaScript, []string{key + ":tokens", key + ":timestamp"}, now, refillRate, limit).Result()
	log.Println("程序运行到 TokenBucketRateLimit 函数后")
	log.Printf("result:%v", result)
	if err != nil {
		return false, err
	}
	return result == int64(1), nil
}

// RetryDownloadRateLimit 全局限流和重试
func (c *RedisCache) RetryDownloadRateLimit(ctx context.Context, imgToken string, maxRetries int, maxWaitTime time.Duration) bool {
	retryCount := 0
	totalWaitTime := time.Duration(0)

	for retryCount < maxRetries {
		log.Println("程序运行到 RetryDownloadRateLimit 函数中")

		allowed, err := c.TokenBucketRateLimit(ctx, "global_download_rate", 5, 5)
		if err != nil {
			log.Printf("Error checking rate limit for image %s: %v", imgToken, err)
			return false
		}
		log.Println("程序运行到 RetryDownloadRateLimit 函数后")
		if allowed {
			return true // 限流成功，返回 true
		}

		// 限流失败，等待一段时间再试
		log.Printf("Rate limit exceeded for image %s, retrying...", imgToken)
		retryCount++
		waitTime := time.Second * time.Duration(retryCount) // 每次重试等待增加
		totalWaitTime += waitTime
		if totalWaitTime > maxWaitTime {
			log.Printf("Total wait time exceeded %v, skipping image %s", maxWaitTime, imgToken)
			return false // 等待时间超过最大值，返回 false
		}

		time.Sleep(waitTime)
	}

	log.Printf("Rate limit exceeded for image %s after %d retries, skipping", imgToken, maxRetries)
	return false // 超过最大重试次数，返回 false
}

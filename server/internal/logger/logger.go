package logger

import (
	"context"
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"net/http"
	"os"
	"strings"

	"feishu2md/server/pkg/conf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	L           *zap.Logger
	AtomicLevel = zap.NewAtomicLevel()
)

// Init 初始化日志模块
func Init(ctx context.Context, cfg *conf.LogConfig) (err error) {
	err = initLogger(cfg)
	if err != nil {
		return
	}

	L = zap.L()

	// 在上下文结束时同步日志
	go func() {
		<-ctx.Done()
		err = L.Sync()
		if err != nil {
			fmt.Println("Failed to sync logger:", err)
		}
	}()

	return nil
}

// WithRequest 为日志添加请求上下文
func WithRequest(r *http.Request) *zap.Logger {
	return L.With(
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("client_ip", r.RemoteAddr),
	)
}

// WithError 为日志添加错误上下文
func WithError(err error) zap.Field {
	return zap.NamedError("error_detail", err)
}

// getEncoder 获取日志编码器
func getEncoder(format string) zapcore.Encoder {
	encodeConfig := zap.NewProductionEncoderConfig()
	encodeConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encodeConfig.TimeKey = "time"
	encodeConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encodeConfig.EncodeCaller = zapcore.ShortCallerEncoder

	if strings.ToUpper(format) == "JSON" {
		return zapcore.NewJSONEncoder(encodeConfig)
	} else {
		encodeConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		return zapcore.NewConsoleEncoder(encodeConfig)
	}
}

// Info 记录结构化信息日志
func Info(msg string, fields ...zap.Field) {
	L.Info(msg, fields...)
}

// Infof 格式化日志（性能略低，适合调试）
func Infof(template string, args ...interface{}) {
	L.Sugar().Infof(template, args...)
}

// Error 记录错误日志
func Error(msg string, err error, fields ...zap.Field) {
	fields = append(fields, WithError(err))
	L.Error(msg, fields...)
}

// getLogWriter 获取日志写入器
func getLogWriter(cfg *conf.LogConfig) zapcore.Core {
	var cores []zapcore.Core

	// 文件日志
	if cfg.Path != "" {
		logRotate := &lumberjack.Logger{
			Filename:   cfg.Path,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		fileEncoder := getEncoder(cfg.Format)
		cores = append(cores, zapcore.NewCore(fileEncoder, zapcore.AddSync(logRotate), AtomicLevel))
	}

	// 控制台日志
	if cfg.ConsoleEnable {
		consoleEncoder := getEncoder("console")
		cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), AtomicLevel))
	}

	return zapcore.NewTee(cores...)
}

// // initLogger 初始化日志
//
//	func initLogger(cfg *conf.LogConfig) (err error) {
//		level, err := zap.ParseAtomicLevel(cfg.Level)
//		if err != nil {
//			return err
//		}
//		AtomicLevel.SetLevel(level.Level())
//
//		core := getLogWriter(cfg)
//		logger := zap.New(core, zap.AddCaller())
//		zap.ReplaceGlobals(logger)
//
//		return nil
//	}
//
// 在 initLogger 函数中添加更丰富的初始化配置
func initLogger(cfg *conf.LogConfig) error {
	// 设置原子级别
	level, err := zap.ParseAtomicLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("failed to parse log level: %w", err)
	}
	AtomicLevel.SetLevel(level.Level())

	// 创建核心
	core := getLogWriter(cfg)

	// 创建带调用方信息的 logger
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1), // 跳过封装层
		zap.AddStacktrace(zap.ErrorLevel),
	)

	// 替换全局 logger
	zap.ReplaceGlobals(logger)

	// 设置自定义字段
	L = logger.With(
		zap.String("app", "feishu2md"),
		zap.String("version", "1.0.0"),
	)

	return nil
}

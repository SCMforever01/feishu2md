package handler

import (
	"context"
	"encoding/json"
	"feishu2md/server/internal/config"
	"feishu2md/server/internal/feishu"
	"feishu2md/server/internal/logger"
	"feishu2md/server/internal/model"
	"feishu2md/server/internal/repository/cache"
	"feishu2md/server/internal/repository/database"
	"feishu2md/server/internal/repository/storage"
	"feishu2md/server/internal/service/auth"
	"feishu2md/server/internal/service/captcha"
	"feishu2md/server/internal/service/img"
	services "feishu2md/server/internal/service/transform"
	user2 "feishu2md/server/internal/service/user"
	"feishu2md/server/pkg/metrics"
	"feishu2md/server/pkg/utils"
	"fmt"
	"github.com/Wsine/feishu2md/core"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func RegisterRoutes(router *gin.Engine) {
	router.GET("/health", healthCheck)
	router.POST("/v1/upload", uploadFile)                  // 新增上传接口
	router.POST("/v1/transform", transformV1)              // 文件解析接口
	router.POST("/v1/feishu/access_token", getAccessToken) //获取accessToken接口
	router.GET("/api/captcha/get", getCaptcha)
	router.POST("/api/captcha/refresh", refreshCaptcha)
	router.POST("/api/register", register)
	router.POST("/api/login", login)
	router.GET("/v1/getHistory", getHistory)
}

func getHistory(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, model.Resp{
			Code:    http.StatusUnauthorized,
			Msg:     "Unauthorized",
			Content: "userID not found in token",
		})
		return
	}
	userID := userIDVal.(int) // 根据你的 claims 结构类型转换

	db, err := database.InitializeDB(DSN) // 更新为你的数据库连接信息
	if err != nil {
		logger.L.Fatal("Failed to initialize database")
		return
	}
	// 创建 TransformService 实例
	service := services.NewTransformService(db) // db 是你已连接的数据库实例

	// 获取历史解析数据
	transforms, err := service.GetHistory(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Resp{
			Code:    http.StatusInternalServerError,
			Msg:     "Internal Server Error",
			Data:    nil,
			Content: err.Error(),
		})
		return
	}

	// 返回多条历史记录
	c.JSON(http.StatusOK, model.Resp{
		Code:    http.StatusOK,
		Msg:     "Success",
		Data:    transforms,
		Content: "History data fetched successfully",
	})
}

const DSN = "root:SCM15172008724mysql@tcp(127.0.0.1:3306)/feishu?charset=utf8mb4&parseTime=True&loc=Local"

func login(c *gin.Context) {
	var req model.LoginRequest
	// 绑定请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		model.Error(c, 1003, "请求参数错误")
		return
	}
	db, err := database.InitializeDB(DSN) // 更新为你的数据库连接信息
	if err != nil {
		log.Fatal("Failed to initialize database: ", err)
		return
	}
	userService := user2.NewUserService(db)
	user, err := userService.LoginUser(req.Phone, req.Password)
	if err != nil {
		model.Error(c, 1002, "登录失败")
		return
	}
	// 生成JWT Token
	token, err := utils.GenerateJWTToken(user.ID)
	if err != nil {
		model.Error(c, 1001, "生成token失败")
		return
	}

	model.Success(c, gin.H{
		"token": token,
		"user":  user,
	})
}
func register(c *gin.Context) {
	var req model.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误", "content": err.Error()})
		return
	}
	service := captcha.NewCaptService()
	// 验证图形验证码
	if !service.VerifyCaptcha(req.CaptchaID, req.CaptchaCode) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "图形验证码错误"})
		return
	}
	db, err := database.InitializeDB(DSN) // 更新为你的数据库连接信息
	if err != nil {
		logger.L.Fatal("Failed to initialize database")
		model.Error(c, 2000, "数据库连接失败")
		return
	}
	userService := user2.NewUserService(db)
	// 注册时校验图形验证码
	user, err := userService.RegisterUser(req.Phone, req.Password)
	if err != nil {
		model.Error(c, 1002, "注册失败："+err.Error())
		return
	}
	model.Success(c, gin.H{"user": user})
}

func refreshCaptcha(c *gin.Context) {
	type Request struct {
		OldID string `json:"old_id" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "参数错误",
			"error":   err.Error(),
		})
		return
	}

	service := captcha.NewCaptService()

	// 先把老的验证码删除掉
	service.GetCodeAnswer(req.OldID) // 调用一次 Get，会把验证码拿出来（并清除）
	// (base64Captcha 的 Get方法带clear，第二个参数设为false，所以要读一下)

	// 生成新的验证码
	newID, b64s, err := service.CreateCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "生成新验证码失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":     newID,
		"base64": b64s,
	})
}

func getCaptcha(c *gin.Context) {
	service := captcha.NewCaptService()
	id, b64s, err := service.CreateCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "生成验证码失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":     id,
		"base64": b64s,
	})
}

func getAccessToken(c *gin.Context) {
	log := logger.WithRequest(c.Request)
	// 验证请求方法
	if c.Request.Method != http.MethodPost {
		log.Error("Invalid request method")
		c.JSON(http.StatusMethodNotAllowed, model.MethodNotAllowed)
		return
	}
	var req model.TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.InvalidRequest)
		return
	}
	authService := auth.NewAuthService()
	resp, err := authService.GetAccessToken(&req)
	if err != nil {
		log.Error("Failed to get accessToken", logger.WithError(err))
		c.JSON(http.StatusBadRequest, model.AuthFailed)
		return
	}
	accessToken := resp.AccessToken
	fmt.Println("accessToken:" + accessToken)
	c.JSON(http.StatusOK, model.Resp{
		Code:    0,
		Msg:     "Success",
		Data:    resp,
		Content: accessToken,
	})
}

func transformV1(c *gin.Context) {
	start := time.Now()
	log := logger.WithRequest(c.Request)

	// 初始化监控指标
	metrics.TotalRequests.WithLabelValues(c.Request.Method, c.FullPath()).Inc()
	defer func() {
		metrics.RequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(time.Since(start).Seconds())
	}()

	// 验证请求方法
	if c.Request.Method != http.MethodPost {
		log.Error("Invalid request method")
		c.JSON(http.StatusMethodNotAllowed, model.MethodNotAllowed)
		return
	}

	// 解析请求体
	var req model.Req
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("Failed to decode request body", logger.WithError(err))
		model.Error(c, 1002, "解析请求体错误")
		return
	}
	fmt.Println(req)
	// 验证请求字段
	if err := validateRequestFields(&req); err != nil {
		log.Error("Request validation failed", logger.WithError(err))
		model.Error(c, 2001, "Request validation failed")
		return
	}

	// 选择处理流程
	var handler func(*gin.Context, *model.Req) (string, string, error)
	if req.IsFile {
		//处理pdf、docx转置的飞书文档
		handler = handleFileTransform
	} else {
		handler = handleDocTransform
	}
	markdown, tittle, err := handler(c, &req)
	// 执行处理逻辑
	if err != nil {
		log.Error("Processing failed", zap.Error(err))
		// 处理自定义错误类型
		if resp, ok := err.(*model.ErrorResponse); ok {
			model.Error(c, resp.Code, resp.Message)
		} else {
			model.Error(c, 1007, "执行解析失败")
		}
		return
	}
	db, err := database.InitializeDB(DSN) // 更新为你的数据库连接信息
	histtroyService := services.NewTransformService(db)
	userID, err := strconv.Atoi(req.Id)
	if err != nil {
		log.Error("Invalid user ID", zap.Error(err))
		model.Error(c, 1001, "用户ID格式不正确")
		return
	}
	if req.WithImageDownload {
		domain, _, _, _ := parseDocumentURL(req.Url)
		cfg := config.LoadConfig()
		redis := cache.InitRedisClient("local")
		storageClient, err := storage.InitStorageClient(*cfg.Storage)
		if err != nil {
			log.Error("IntiStorage client failed", zap.Error(err))
		}
		client := feishu.NewClient(cfg.Feishu.AppID, cfg.Feishu.AppSecret, domain)
		processor := img.NewProcessor(*redis, storageClient, *client, *cfg.ImgConfig)
		newConfig := core.NewConfig(cfg.Feishu.AppID, cfg.Feishu.AppSecret)
		parser := core.NewParser(newConfig.Output)
		imgTokens := parser.ImgTokens
		// 图片处理方法
		resultMarkdown, err := processor.ProcessImages(c.Request.Context(), markdown, imgTokens, req)
		if err != nil {
			log.Error("Image processing failed", zap.Error(err))
			model.Error(c, 1003, "图片处理失败")
			return
		}
		histtroyService.CreateTransform(userID, req.Url, resultMarkdown)
		// 返回处理后的结果
		model.Success(c, gin.H{
			"markdown":   resultMarkdown,
			"sheetTitle": tittle,
		})

	} else {
		histtroyService.CreateTransform(userID, req.Url, markdown)
		model.Success(c, gin.H{
			"markdown":   markdown,
			"sheetTitle": tittle,
		}) // 返回未处理图片的 markdown
	}
	// 记录QPS
	metrics.QPS.WithLabelValues(c.Request.Method, c.FullPath()).Inc()
	return
}

func handleDocTransform(c *gin.Context, req *model.Req) (string, string, error) {
	start := time.Now()

	// 记录指标
	defer func() {
		metrics.RequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(time.Since(start).Seconds())
	}()

	// 处理文档转换
	markdown, tittle, err := handleURLArgument(c, req)
	if err != nil {
		// 特定错误处理
		if strings.Contains(err.Error(), "Invalid access token") ||
			strings.Contains(err.Error(), "Invalid URL format") {
			metrics.ErrorRequests.WithLabelValues(c.Request.Method, c.FullPath()).Inc()
			return "", "", &model.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Unsupported document type",
			}
		}

		metrics.ErrorRequests.WithLabelValues(c.Request.Method, c.FullPath()).Inc()
		return "", "", &model.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Document transformation failed",
			Detail:  err.Error(),
		}
	}

	// 记录成功指标
	metrics.SuccessRequests.WithLabelValues(
		c.Request.Method,
		c.FullPath(),
		http.StatusText(http.StatusOK),
	).Inc()
	return markdown, tittle, nil
}

// DocHandler 文档处理接口
type DocHandler interface {
	Process(ctx context.Context, token, domain, userAccessToken string, downLoadImg bool, req model.Req) ([]byte, error)
}

// 文档处理器注册表
var docHandlers = map[string]DocHandler{
	"doc":     &DocHandlerImpl{},
	"docs":    &DocHandlerImpl{},
	"docx":    &DocHandlerImpl{},
	"wiki":    &WikiHandler{},
	"sheet":   &SheetHandler{},
	"sheets":  &SheetHandler{},
	"base":    &BitableHandler{},
	"bitable": &BitableHandler{},
}

func handleURLArgument(c *gin.Context, req *model.Req) (string, string, error) {
	start := time.Now()
	log := logger.WithRequest(c.Request)
	ctx := c.Request.Context()

	defer func() {
		log.Info("Document processing completed",
			zap.String("url", req.Url),
			zap.Duration("duration", time.Since(start)),
		)
	}()

	// 1. 解析URL获取文档类型和token
	domain, docType, token, err := parseDocumentURL(req.Url)
	if err != nil {
		return "", "", &model.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid document URL",
			Detail:  err.Error(),
		}
	}

	// 2. 获取文档处理器
	handler, ok := docHandlers[docType]
	if !ok {
		return "", "", &model.ErrorResponse{
			Code:    http.StatusUnprocessableEntity,
			Message: "Unsupported document type",
			Detail:  fmt.Sprintf("Doctype '%s' not supported", docType),
		}
	}

	// 3. 处理文档内容
	content, err := handler.Process(ctx, token, domain, req.UserAccessToken, req.WithImageDownload, *req)
	if err != nil {
		return "", "", wrapProcessingError(err, docType)
	}
	var result struct {
		Markdown   string `json:"markdown"`
		SheetTitle string `json:"sheetTitle"`
	}
	if err := json.Unmarshal(content, &result); err != nil {
		return "", "", fmt.Errorf("failed to parse sheet response: %w", err)
	}

	return result.Markdown, result.SheetTitle, nil
}

// parseDocumentURL 解析飞书文档URL
func parseDocumentURL(url string) (domain, docType, token string, err error) {
	// 匹配飞书文档URL格式
	// 示例：https://xxx.feishu.cn/docx/ABC123 或 https://xxx.larksuite.com/docs/ABC123
	pattern := `^https://[a-zA-Z0-9-]+.(feishu.cn|larksuite.com|f.mioffice.cn)/(doc|docs|docx|wiki|sheets|base|sheet|bitable)/([a-zA-Z0-9]+)`
	reg := regexp.MustCompile(pattern)
	matches := reg.FindStringSubmatch(url)

	if matches == nil || len(matches) != 4 {
		return "", "", "", fmt.Errorf("Invalid feishu/larksuite URL format\n")
	}

	// matches[2] 是文档类型，matches[3] 是文档token
	return matches[1], normalizeDocType(matches[2]), matches[3], nil
}

// normalizeDocType 标准化文档类型
func normalizeDocType(rawType string) string {
	switch rawType {
	case "docs":
		return "doc" // 统一处理docs和doc类型
	default:
		return rawType
	}
}

// wrapProcessingError 包装处理错误
func wrapProcessingError(err error, docType string) error {
	switch {
	case strings.Contains(err.Error(), "permission denied"):
		return &model.ErrorResponse{
			Code:    http.StatusForbidden,
			Message: "Access denied",
			Detail:  fmt.Sprintf("No permission to access %s document", docType),
		}
	case strings.Contains(err.Error(), "not found"):
		return &model.ErrorResponse{
			Code:    http.StatusNotFound,
			Message: "Document not found",
			Detail:  fmt.Sprintf("%s document not exist or deleted", docType),
		}
	default:
		return &model.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Document processing failed",
			Detail:  fmt.Sprintf("Failed to process %s document: %v", docType, err),
		}
	}
}

// ========== 具体文档处理器实现 ==========

// DocHandlerImpl 旧文档处理器
type DocHandlerImpl struct{}

func (h *DocHandlerImpl) Process(ctx context.Context, token, domain, userAccessToken string, downLoadImg bool, req model.Req) ([]byte, error) {
	cfg := config.LoadConfig()
	client := feishu.NewClient(cfg.Feishu.AppID, cfg.Feishu.AppSecret, domain)

	content, err := client.GetDocumentContent(ctx, token, userAccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get document content: %w", err)
	}

	return []byte(content), nil
}

// WikiHandler 处理知识库文档
type WikiHandler struct{}

func (h *WikiHandler) Process(ctx context.Context, token, domain, userAccessToken string, downLoadImg bool, req model.Req) ([]byte, error) {
	cfg := config.LoadConfig()
	client := feishu.NewClient(cfg.Feishu.AppID, cfg.Feishu.AppSecret, domain)

	// 获取知识库节点信息
	node, err := client.GetWikiNodeInfo(ctx, token, userAccessToken)
	fmt.Printf("node:%v", node.ObjType)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Type:%v", node.ObjType)
	handler, exists := docHandlers[node.ObjType]
	if !exists {
		return nil, fmt.Errorf("暂不支持处理 %s 类型文档", node.ObjType)
	}

	// 转发到实际文档处理器
	return handler.Process(ctx, node.ObjToken, domain, userAccessToken, downLoadImg, req)
}

// SheetHandler 表格处理器
type SheetHandler struct{}

func (s *SheetHandler) Process(ctx context.Context, token, domain, userAccessToken string, downLoadImg bool, req model.Req) ([]byte, error) {
	cfg := config.LoadConfig()
	client := feishu.NewClient(cfg.Feishu.AppID, cfg.Feishu.AppSecret, domain)
	fmt.Printf("token:%s,userAccessToken:%s,Url:%s", token, userAccessToken, req.Url)
	result, err := client.GetSheetsContent(ctx, token, userAccessToken, req.Url)
	if err != nil {
		return nil, fmt.Errorf("failed to get document content: %w", err)
	}
	fmt.Printf("result:%v", result)

	resp := map[string]string{
		"markdown":   result.Markdown,
		"sheetTitle": result.SheetTitle,
	}
	//生成 .md 文件并写入内容
	err = writeContentToMDFile(result.Markdown)
	if err != nil {
		log.Printf("failed to write .md file: %v", err)
		// 非致命错误，继续执行
	}
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sheet result: %w", err)
	}

	return jsonBytes, nil
}

func writeContentToMDFile(content string) error {
	fmt.Println("Start writeContentToMDFile")
	// 从 content 中提取标题
	re := regexp.MustCompile(`(?m)^# (.+)$`)
	matches := re.FindStringSubmatch(content)
	title := "Untitled"
	if len(matches) >= 2 {
		title = matches[1]
	}
	// 移除非法文件名字符
	title = sanitizeFileName(title)
	fileName := fmt.Sprintf("%s.md", title)

	// 获取当前目录路径
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get current directory: %v\n", err)
		return err
	}
	fmt.Println("Current directory:", currentDir)

	// 确保 testmd 文件夹存在
	testDir := filepath.Join(currentDir, "testmd")
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		fmt.Println("Creating directory:", testDir)
		err = os.Mkdir(testDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", testDir, err)
		}
	} else {
		fmt.Println("Directory already exists:", testDir)
	}

	// 检查 content 是否为空
	if len(content) == 0 {
		return fmt.Errorf("content is empty, cannot write to file")
	}
	fmt.Println("Content length:", len(content))

	// 写入文件
	filePath := filepath.Join(testDir, fileName)
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write content to file %s: %w", filePath, err)
	}

	fmt.Printf("File successfully written: %s\n", filePath)
	return nil
}
func sanitizeFileName(fileName string) string {
	// 移除文件名中的非法字符
	illegalChars := regexp.MustCompile(`[\/:*?"<>|]`)
	return illegalChars.ReplaceAllString(fileName, "")
}

// BitableHandler 多为表格处理器
type BitableHandler struct {
}

func (s *BitableHandler) Process(ctx context.Context, token, domain, userAccessToken string, downLoadImg bool, req model.Req) ([]byte, error) {
	cfg := config.LoadConfig()
	client := feishu.NewClient(cfg.Feishu.AppID, cfg.Feishu.AppSecret, domain)

	sheet, err := client.GetBitablesContent(ctx, token, userAccessToken, req.Url)
	if err != nil {
		return nil, fmt.Errorf("failed to get document content: %w", err)
	}

	return []byte(sheet), nil
}

// ========== 辅助函数 ==========

// validateRequestFields 验证请求字段
func validateRequestFields(req *model.Req) error {
	if !req.IsFile {
		if req.Id == "" || req.Url == "" {
			return &model.ErrorResponse{
				Code:    http.StatusBadRequest,
				Message: "Missing required fields",
			}
		}

		//if _, ok := accessKeySet[req.AccessKey]; !ok {
		//	return &model.ErrorResponse{
		//		Code:    http.StatusUnauthorized,
		//		Message: "Invalid access key",
		//	}
		//}
	}
	return nil
}

// 定义一个全局的set，用来存储access_key，并初始化
var accessKeySet = map[string]struct{}{
	"7r9dWfGq2L": {},
	"Y6tE5z4P1X": {},
	"3yC7sB2eRv": {},
	"G5D4hJ6j8S": {},
	"1mV9uN0wFq": {},
	"T4xY3bW2dV": {},
	"K7lO6aP3eI": {},
	"Z8mN7pB9cX": {},
	"D2fA4sG6rH": {},
	"Q9wE8rT6yU": {},
}

func handleFileTransform(c *gin.Context, req *model.Req) (string, string, error) {
	start := time.Now()

	// 记录指标
	defer func() {
		metrics.RequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(time.Since(start).Seconds())
	}()

	reg := regexp.MustCompile("^https://[a-zA-Z0-9-]+.(feishu.cn|larksuite.com|f.mioffice.cn)/(doc|docs|docx|wiki|file)/([a-zA-Z0-9]+)")
	matchResult := reg.FindStringSubmatch(req.Url)

	if matchResult == nil || len(matchResult) != 4 {
		return "", "", &model.ErrorResponse{
			Code:    http.StatusBadRequest,
			Message: "Invalid feishu/larksuite URL format",
		}
	}

	cfg := config.LoadConfig()
	// 初始化飞书客户端
	domain := matchResult[1]
	docType := matchResult[2]
	docToken := matchResult[3]
	fmt.Printf("Debug: Extracted domain: %s, docType: %s, docToken: %s\n", domain, docType, docToken)

	client := feishu.NewClient(cfg.Feishu.AppID, cfg.Feishu.AppSecret, domain)

	logger.Info("handle info: ",
		zap.String("domain", domain),
		zap.String("docType", docType),
		zap.String("doc token", docToken),
	)

	// 下载文件
	data, err := client.DownloadFile(docToken, req.UserAccessToken)
	if err != nil {
		metrics.ErrorRequests.WithLabelValues(c.Request.Method, c.FullPath()).Inc()
		return "", "", &model.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "Failed to download file",
			Detail:  err.Error(),
		}
	}
	// 处理文件内容
	contentBytes, err := io.ReadAll(data.File)
	if err != nil {
		return "", "", &model.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to read file content",
			Detail:  err.Error(),
		}
	}

	// 构建响应头
	baseName := filepath.Base(data.Filename)
	fileNameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	//fileType := strings.TrimPrefix(filepath.Ext(fileName), ".")

	//c.Header("Content-Disposition", "attachment; filename="+baseName+"."+fileType)
	//c.Data(http.StatusOK, "application/octet-stream", contentBytes)

	// 记录成功指标
	metrics.SuccessRequests.WithLabelValues(
		c.Request.Method,
		c.FullPath(),
		http.StatusText(http.StatusOK),
	).Inc()
	return string(contentBytes), fileNameWithoutExt, nil
}

func healthCheck(c *gin.Context) {
	logger.Info("Health check endpoint called")
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// uploadFile 处理文件上传和解析
func uploadFile(c *gin.Context) {
	// 1. 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		logger.L.Error("Failed to read uploaded file", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请提供有效的文件",
		})
		return
	}

	// 2. 创建保存文件的目录（如果不存在）
	uploadDir := "uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		if err := os.Mkdir(uploadDir, 0755); err != nil {
			logger.L.Error("Failed to create upload directory", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "服务器内部错误",
			})
			return
		}
	}

	// 3. 保存文件到服务器
	filePath := filepath.Join(uploadDir, file.Filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		logger.L.Error("Failed to save file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "文件保存失败",
		})
		return
	}

	// 4. 解析文件
	content, err := parseFile(filePath)
	if err != nil {
		logger.L.Error("Failed to parse file", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "文件解析失败",
		})
		return
	}

	// 5. 返回解析结果
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("文件 %s 上传成功", file.Filename),
		"content": content,
	})
}

// parseFile 解析文件内容（示例：读取文本文件）
func parseFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

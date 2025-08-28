package middleware

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// I18nConfig 国际化配置
type I18nConfig struct {
	DefaultLanguage    string   // 默认语言
	SupportedLanguages []string // 支持的语言列表
	LanguageHeader     string   // 语言头名称
	LanguageParam      string   // 语言参数名
	LanguageCookie     string   // 语言Cookie名
	TranslationPath    string   // 翻译文件路径
	FallbackToDefault  bool     // 是否回退到默认语言
}

// DefaultI18nConfig 默认国际化配置
func DefaultI18nConfig() *I18nConfig {
	return &I18nConfig{
		DefaultLanguage:    "zh-CN",
		SupportedLanguages: []string{"zh-CN", "en-US", "ja-JP"},
		LanguageHeader:     "Accept-Language",
		LanguageParam:      "lang",
		LanguageCookie:     "lang",
		TranslationPath:    "locales",
		FallbackToDefault:  true,
	}
}

// Translation 翻译数据结构
type Translation map[string]interface{}

// I18nManager 国际化管理器
type I18nManager struct {
	config       *I18nConfig
	translations map[string]Translation
	mutex        sync.RWMutex
}

// NewI18nManager 创建国际化管理器
func NewI18nManager(config *I18nConfig) *I18nManager {
	if config == nil {
		config = DefaultI18nConfig()
	}

	manager := &I18nManager{
		config:       config,
		translations: make(map[string]Translation),
	}

	// 加载翻译文件
	if err := manager.LoadTranslations(); err != nil {
		// 如果加载失败，至少初始化默认语言
		manager.translations[config.DefaultLanguage] = make(Translation)
	}

	return manager
}

// LoadTranslations 加载翻译文件
func (i *I18nManager) LoadTranslations() error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	for _, lang := range i.config.SupportedLanguages {
		filePath := filepath.Join(i.config.TranslationPath, fmt.Sprintf("%s.yaml", lang))

		// 检查文件是否存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// 文件不存在，创建空的翻译映射
			i.translations[lang] = make(Translation)
			continue
		}

		// 验证文件路径安全性
		if !isValidTranslationPath(filePath, i.config.TranslationPath) {
			return fmt.Errorf("无效的翻译文件路径: %s", filePath)
		}

		// 读取文件内容
		// #nosec G304 - filePath is validated above
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("读取翻译文件 %s 失败: %w", filePath, err)
		}

		// 解析YAML
		var translation Translation
		if err := yaml.Unmarshal(data, &translation); err != nil {
			return fmt.Errorf("解析翻译文件 %s 失败: %w", filePath, err)
		}

		i.translations[lang] = translation
	}

	return nil
}

// GetTranslation 获取翻译
func (i *I18nManager) GetTranslation(lang, key string, args ...interface{}) string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	// 查找指定语言的翻译
	if translation, exists := i.translations[lang]; exists {
		if value := i.getNestedValue(translation, key); value != "" {
			return i.formatMessage(value, args...)
		}
	}

	// 回退到默认语言
	if i.config.FallbackToDefault && lang != i.config.DefaultLanguage {
		if translation, exists := i.translations[i.config.DefaultLanguage]; exists {
			if value := i.getNestedValue(translation, key); value != "" {
				return i.formatMessage(value, args...)
			}
		}
	}

	// 返回键值本身（开发时有用）
	return key
}

// getNestedValue 获取嵌套值
func (i *I18nManager) getNestedValue(translation Translation, key string) string {
	keys := strings.Split(key, ".")
	current := translation

	for i, k := range keys {
		if i == len(keys)-1 {
			// 最后一个键，应该是字符串值
			if value, ok := current[k].(string); ok {
				return value
			}
			return ""
		}

		// 中间键，应该是map
		if next, ok := current[k].(map[string]interface{}); ok {
			current = next
		} else {
			return ""
		}
	}

	return ""
}

// formatMessage 格式化消息
func (i *I18nManager) formatMessage(template string, args ...interface{}) string {
	if len(args) == 0 {
		return template
	}

	return fmt.Sprintf(template, args...)
}

// 全局国际化管理器实例
var globalI18nManager *I18nManager
var once sync.Once

// InitI18n 初始化国际化
func InitI18n(config *I18nConfig) error {
	var err error
	once.Do(func() {
		globalI18nManager = NewI18nManager(config)
	})
	return err
}

// I18nMiddleware 创建国际化中间件
func I18nMiddleware(config ...*I18nConfig) gin.HandlerFunc {
	var cfg *I18nConfig
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	} else {
		cfg = DefaultI18nConfig()
	}

	// 确保全局管理器已初始化
	if globalI18nManager == nil {
		if err := InitI18n(cfg); err != nil {
			panic(fmt.Sprintf("初始化国际化失败: %v", err))
		}
	}

	return gin.HandlerFunc(func(c *gin.Context) {
		lang := extractLanguage(c, cfg)

		// 验证语言是否支持
		if !isLanguageSupported(lang, cfg.SupportedLanguages) {
			lang = cfg.DefaultLanguage
		}

		// 设置语言信息到上下文
		c.Set("language", lang)
		c.Set("i18n_config", cfg)
		c.Set("i18n_manager", globalI18nManager)

		// 添加语言信息到响应头
		c.Header("Content-Language", lang)

		c.Next()
	})
}

// extractLanguage 提取语言设置
func extractLanguage(c *gin.Context, cfg *I18nConfig) string {
	var lang string

	// 1. 优先从查询参数获取
	lang = c.Query(cfg.LanguageParam)

	// 2. 从Cookie获取
	if lang == "" {
		if cookie, err := c.Cookie(cfg.LanguageCookie); err == nil {
			lang = cookie
		}
	}

	// 3. 从Header获取
	if lang == "" {
		acceptLang := c.GetHeader(cfg.LanguageHeader)
		lang = parseAcceptLanguage(acceptLang)
	}

	// 4. 使用默认语言
	if lang == "" {
		lang = cfg.DefaultLanguage
	}

	return normalizeLanguage(lang)
}

// parseAcceptLanguage 解析Accept-Language头
func parseAcceptLanguage(acceptLang string) string {
	if acceptLang == "" {
		return ""
	}

	// 简单解析，取第一个语言标识
	parts := strings.Split(acceptLang, ",")
	if len(parts) > 0 {
		lang := strings.TrimSpace(parts[0])
		// 移除权重信息
		if idx := strings.Index(lang, ";"); idx != -1 {
			lang = lang[:idx]
		}
		return lang
	}

	return ""
}

// normalizeLanguage 标准化语言格式
func normalizeLanguage(lang string) string {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		return ""
	}

	// 转换常见格式
	switch strings.ToLower(lang) {
	case "zh", "zh-cn", "chinese":
		return "zh-CN"
	case "en", "en-us", "english":
		return "en-US"
	case "ja", "ja-jp", "japanese":
		return "ja-JP"
	default:
		return lang
	}
}

// isLanguageSupported 检查语言是否支持
func isLanguageSupported(lang string, supportedLanguages []string) bool {
	for _, supported := range supportedLanguages {
		if lang == supported {
			return true
		}
	}
	return false
}

// GetLanguage 获取当前请求的语言
func GetLanguage(c *gin.Context) string {
	if lang, exists := c.Get("language"); exists {
		if l, ok := lang.(string); ok {
			return l
		}
	}
	return "zh-CN" // 默认语言
}

// T 翻译函数（简写）
func T(c *gin.Context, key string, args ...interface{}) string {
	return Translate(c, key, args...)
}

// Translate 翻译函数
func Translate(c *gin.Context, key string, args ...interface{}) string {
	lang := GetLanguage(c)

	if manager, exists := c.Get("i18n_manager"); exists {
		if i18nManager, ok := manager.(*I18nManager); ok {
			return i18nManager.GetTranslation(lang, key, args...)
		}
	}

	// 如果没有管理器，返回键值
	return key
}

// LanguageInfoResponse 语言信息响应
type LanguageInfoResponse struct {
	CurrentLanguage    string   `json:"current_language"`
	SupportedLanguages []string `json:"supported_languages"`
	DefaultLanguage    string   `json:"default_language"`
}

// LanguageInfoHandler 语言信息处理器
func LanguageInfoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := GetI18nConfig(c)
		currentLang := GetLanguage(c)

		response := LanguageInfoResponse{
			CurrentLanguage:    currentLang,
			SupportedLanguages: cfg.SupportedLanguages,
			DefaultLanguage:    cfg.DefaultLanguage,
		}

		c.JSON(http.StatusOK, gin.H{
			"code":      200,
			"message":   Translate(c, "common.success"),
			"data":      response,
			"timestamp": c.GetInt64("timestamp"),
		})
	}
}

// isValidTranslationPath 验证翻译文件路径是否安全
func isValidTranslationPath(filePath, basePath string) bool {
	// 获取绝对路径
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}

	absBasePath, err := filepath.Abs(basePath)
	if err != nil {
		return false
	}

	// 检查文件是否在基础路径内
	relPath, err := filepath.Rel(absBasePath, absFilePath)
	if err != nil {
		return false
	}

	// 禁止路径穿越
	if strings.HasPrefix(relPath, "..") || strings.Contains(relPath, string(filepath.Separator)+".."+string(filepath.Separator)) {
		return false
	}

	// 只允许YAML文件
	if !strings.HasSuffix(strings.ToLower(filePath), ".yaml") && !strings.HasSuffix(strings.ToLower(filePath), ".yml") {
		return false
	}

	return true
}

// GetI18nConfig 获取国际化配置
func GetI18nConfig(c *gin.Context) *I18nConfig {
	if config, exists := c.Get("i18n_config"); exists {
		if cfg, ok := config.(*I18nConfig); ok {
			return cfg
		}
	}
	return DefaultI18nConfig()
}

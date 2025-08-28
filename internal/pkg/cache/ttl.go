package cache

import (
	"cloudpan/internal/pkg/config"
	pkgErrors "cloudpan/internal/pkg/errors"
	"time"
)

// 引用统一的错误定义
var (
	ErrCacheNotFound   = pkgErrors.ErrCacheNotFound
	ErrCacheExpired    = pkgErrors.ErrCacheExpired
	ErrInvalidCacheKey = pkgErrors.ErrInvalidCacheKey
	ErrCacheServerDown = pkgErrors.ErrCacheServerDown
	ErrInvalidTTL      = pkgErrors.ErrInvalidTTL
)

// TTLManager TTL管理器，管理不同类型缓存的TTL策略
type TTLManager struct {
	ttlMap map[string]time.Duration
}

// NewTTLManager 创建TTL管理器
func NewTTLManager() *TTLManager {
	tm := &TTLManager{
		ttlMap: make(map[string]time.Duration),
	}
	tm.initTTLMap()
	return tm
}

// initTTLMap 初始化TTL映射表
func (tm *TTLManager) initTTLMap() {
	tm.ttlMap = map[string]time.Duration{
		"user_session":     2 * time.Hour,    // 用户会话2小时
		"user_permissions": 1 * time.Hour,    // 用户权限1小时
		"file_preview":     30 * time.Minute, // 文件预览30分钟
		"file_share":       1 * time.Hour,    // 文件分享1小时
		"file_upload":      24 * time.Hour,   // 文件上传状态24小时
		"team_info":        30 * time.Minute, // 团队信息30分钟
		"team_members":     15 * time.Minute, // 团队成员15分钟
		"verify_attempt":   15 * time.Minute, // 验证尝试15分钟
		"verify_block":     1 * time.Hour,    // 验证封锁1小时
		"rate_limit":       1 * time.Minute,  // 限流1分钟
		"user_rate_limit":  5 * time.Minute,  // 用户限流5分钟
		"api_rate_limit":   1 * time.Minute,  // API限流1分钟
		"lock":             10 * time.Minute, // 分布式锁10分钟
		"search_result":    15 * time.Minute, // 搜索结果15分钟
		"search_history":   24 * time.Hour,   // 搜索历史24小时
		"stats_user":       10 * time.Minute, // 用户统计10分钟
		"stats_file":       5 * time.Minute,  // 文件统计5分钟
		"stats_system":     1 * time.Minute,  // 系统统计1分钟
		"message":          1 * time.Hour,    // 消息缓存1小时
		"conversation":     30 * time.Minute, // 会话缓存30分钟
		"online_users":     5 * time.Minute,  // 在线用户5分钟
	}
}

// GetTTL 根据缓存类型获取对应的TTL
func (tm *TTLManager) GetTTL(cacheType string) time.Duration {
	cfg := config.AppConfig.Cache

	// 优先检查映射表中的固定TTL
	if ttl, exists := tm.ttlMap[cacheType]; exists {
		return ttl
	}

	// 处理需要从配置读取的特殊类型
	return tm.getConfigBasedTTL(cacheType, cfg)
}

// getConfigBasedTTL 获取基于配置的TTL
func (tm *TTLManager) getConfigBasedTTL(cacheType string, cfg config.CacheConfig) time.Duration {
	switch cacheType {
	case "user_profile", "user_info":
		return cfg.UserInfoTTL
	case "file_info":
		return cfg.FileInfoTTL
	case "verify_code":
		return cfg.VerificationCodeTTL
	default:
		return cfg.DefaultTTL
	}
}

// GetShortTTL 获取短期缓存TTL (5分钟)
func (tm *TTLManager) GetShortTTL() time.Duration {
	return 5 * time.Minute
}

// GetMediumTTL 获取中期缓存TTL (30分钟)
func (tm *TTLManager) GetMediumTTL() time.Duration {
	return 30 * time.Minute
}

// GetLongTTL 获取长期缓存TTL (2小时)
func (tm *TTLManager) GetLongTTL() time.Duration {
	return 2 * time.Hour
}

// GetPersistentTTL 获取持久缓存TTL (24小时)
func (tm *TTLManager) GetPersistentTTL() time.Duration {
	return 24 * time.Hour
}

// ValidateTTL 验证TTL值是否合法
func (tm *TTLManager) ValidateTTL(ttl time.Duration) error {
	if ttl < 0 {
		return ErrInvalidTTL
	}
	if ttl > 7*24*time.Hour { // 最大7天
		return ErrInvalidTTL
	}
	return nil
}

// CacheWrapper 缓存包装器，提供带TTL的缓存操作
type CacheWrapper struct {
	manager    *CacheManager
	ttlManager *TTLManager
}

// NewCacheWrapper 创建缓存包装器
func NewCacheWrapper() *CacheWrapper {
	return &CacheWrapper{
		manager:    NewCacheManager(),
		ttlManager: NewTTLManager(),
	}
}

// SetByType 根据缓存类型设置缓存
func (cw *CacheWrapper) SetByType(key string, value interface{}, cacheType string) error {
	ttl := cw.ttlManager.GetTTL(cacheType)
	return cw.manager.SetWithTTL(key, value, ttl)
}

// SetUserSession 设置用户会话缓存
func (cw *CacheWrapper) SetUserSession(token string, value interface{}) error {
	key := Keys.UserSession(token)
	return cw.SetByType(key, value, "user_session")
}

// GetUserSession 获取用户会话缓存
func (cw *CacheWrapper) GetUserSession(token string, dest interface{}) error {
	key := Keys.UserSession(token)
	return cw.manager.Get(key, dest)
}

// SetUserPermissions 设置用户权限缓存
func (cw *CacheWrapper) SetUserPermissions(userID string, permissions []string) error {
	key := Keys.UserPermissions(userID)
	return cw.SetByType(key, permissions, "user_permissions")
}

// GetUserPermissions 获取用户权限缓存
func (cw *CacheWrapper) GetUserPermissions(userID string) ([]string, error) {
	key := Keys.UserPermissions(userID)
	var permissions []string
	err := cw.manager.Get(key, &permissions)
	return permissions, err
}

// SetFileInfo 设置文件信息缓存
func (cw *CacheWrapper) SetFileInfo(fileID string, fileInfo interface{}) error {
	key := Keys.FileInfo(fileID)
	return cw.SetByType(key, fileInfo, "file_info")
}

// GetFileInfo 获取文件信息缓存
func (cw *CacheWrapper) GetFileInfo(fileID string, dest interface{}) error {
	key := Keys.FileInfo(fileID)
	return cw.manager.Get(key, dest)
}

// SetVerificationCode 设置验证码
func (cw *CacheWrapper) SetVerificationCode(codeType, target, code string) error {
	key := Keys.VerifyCode(codeType, target)
	return cw.SetByType(key, code, "verify_code")
}

// GetVerificationCode 获取验证码
func (cw *CacheWrapper) GetVerificationCode(codeType, target string) (string, error) {
	key := Keys.VerifyCode(codeType, target)
	var code string
	err := cw.manager.Get(key, &code)
	return code, err
}

// IncrementRateLimit 增加限流计数
func (cw *CacheWrapper) IncrementRateLimit(ip, endpoint string) (int64, error) {
	key := Keys.RateLimit(ip, endpoint)
	count, err := cw.manager.Increment(key)
	if err != nil {
		return count, err
	}

	// 设置过期时间
	ttl := cw.ttlManager.GetTTL("rate_limit")
	if err := cw.manager.Expire(key, ttl); err != nil {
		// 即使设置TTL失败，也返回计数结果，但记录错误
		// 这里可以添加日志记录
	}
	return count, nil
}

// SetOnlineUser 设置用户在线状态
func (cw *CacheWrapper) SetOnlineUser(userID string) error {
	key := Keys.UserOnline(userID)
	return cw.SetByType(key, time.Now().Unix(), "online_users")
}

// IsUserOnline 检查用户是否在线
func (cw *CacheWrapper) IsUserOnline(userID string) bool {
	key := Keys.UserOnline(userID)
	exists, _ := cw.manager.Exists(key)
	return exists > 0
}

// ClearUserCache 清理用户相关缓存
func (cw *CacheWrapper) ClearUserCache(userID string) error {
	keys := []string{
		Keys.UserProfile(userID),
		Keys.UserPermissions(userID),
		Keys.UserQuota(userID),
		Keys.UserStats(userID),
		Keys.UserOnline(userID),
	}
	return cw.manager.Delete(keys...)
}

// ClearFileCache 清理文件相关缓存
func (cw *CacheWrapper) ClearFileCache(fileID string) error {
	keys := []string{
		Keys.FileInfo(fileID),
		Keys.FilePreview(fileID),
		Keys.FileDownload(fileID),
		Keys.FileStats(fileID),
	}
	return cw.manager.Delete(keys...)
}

// 全局TTL管理器和缓存包装器实例
var (
	TTLMgr *TTLManager
	Cache  *CacheWrapper
)

// InitGlobalCache 初始化全局实例（在Redis初始化后调用）
func InitGlobalCache() {
	TTLMgr = NewTTLManager()
	Cache = NewCacheWrapper()
}

// GetTTLManager 获取全局TTL管理器
func GetTTLManager() *TTLManager {
	if TTLMgr == nil {
		TTLMgr = NewTTLManager()
	}
	return TTLMgr
}

// GetCacheWrapper 获取全局缓存包装器
func GetCacheWrapper() *CacheWrapper {
	if Cache == nil {
		// 只有在Redis初始化后才创建缓存包装器
		if RedisClient != nil {
			Cache = NewCacheWrapper()
		}
	}
	return Cache
}

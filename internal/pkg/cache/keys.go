package cache

import "fmt"

// 缓存键命名规范常量
const (
	// 用户相关
	KeyUserSession     = "session:%s"     // session:token
	KeyUserPermissions = "permissions:%s" // permissions:user_id
	KeyUserProfile     = "profile:%s"     // profile:user_id
	KeyUserOnline      = "online:%s"      // online:user_id
	KeyUserQuota       = "quota:%s"       // quota:user_id

	// 文件相关
	KeyFileInfo     = "file:%s"     // file:file_id
	KeyFileShare    = "share:%s"    // share:token
	KeyFileUpload   = "upload:%s"   // upload:upload_id
	KeyFileChunk    = "chunk:%s:%d" // chunk:upload_id:chunk_num
	KeyFilePreview  = "preview:%s"  // preview:file_id
	KeyFileDownload = "download:%s" // download:file_id

	// 团队相关
	KeyTeamInfo        = "team:%s"          // team:team_id
	KeyTeamMembers     = "team:members:%s"  // team:members:team_id
	KeyTeamFiles       = "team:files:%s"    // team:files:team_id
	KeyTeamPermissions = "team:perms:%s:%s" // team:perms:team_id:user_id

	// 验证码相关
	KeyVerifyCode    = "code:%s:%s"    // code:type:target
	KeyVerifyAttempt = "attempt:%s:%s" // attempt:type:target
	KeyVerifyBlock   = "block:%s:%s"   // block:type:target

	// 限流相关
	KeyRateLimit     = "rate:%s:%s"      // rate:ip:endpoint
	KeyUserRateLimit = "user_rate:%s:%s" // user_rate:user_id:action
	KeyAPIRateLimit  = "api_rate:%s:%s"  // api_rate:api_key:endpoint

	// 锁相关
	KeyFileLock   = "lock:file:%s"   // lock:file:file_id
	KeyUserLock   = "lock:user:%s"   // lock:user:user_id
	KeyTeamLock   = "lock:team:%s"   // lock:team:team_id
	KeyUploadLock = "lock:upload:%s" // lock:upload:upload_id

	// 队列相关
	KeyTaskQueue   = "queue:task"   // 任务队列
	KeyEmailQueue  = "queue:email"  // 邮件队列
	KeyNotifyQueue = "queue:notify" // 通知队列
	KeyFileQueue   = "queue:file"   // 文件处理队列

	// 消息相关
	KeyConversation = "msg:conv:%s"    // msg:conv:conversation_id
	KeyMessage      = "msg:%s"         // msg:message_id
	KeyMessageRead  = "msg:read:%s:%s" // msg:read:conversation_id:user_id
	KeyUserMessages = "msg:user:%s"    // msg:user:user_id

	// 统计相关
	KeyUserStats   = "stats:user:%s" // stats:user:user_id
	KeyFileStats   = "stats:file:%s" // stats:file:file_id
	KeyTeamStats   = "stats:team:%s" // stats:team:team_id
	KeySystemStats = "stats:system"  // 系统统计

	// 搜索相关
	KeySearchIndex   = "search:index:%s"   // search:index:type
	KeySearchResult  = "search:result:%s"  // search:result:query_hash
	KeySearchHistory = "search:history:%s" // search:history:user_id
)

// KeyBuilder 缓存键构建器
type KeyBuilder struct{}

// NewKeyBuilder 创建键构建器
func NewKeyBuilder() *KeyBuilder {
	return &KeyBuilder{}
}

// build 通用键构建方法，减少重复代码
func (kb *KeyBuilder) build(template string, args ...interface{}) string {
	return fmt.Sprintf(template, args...)
}

// UserSession 生成用户会话缓存键
func (kb *KeyBuilder) UserSession(token string) string {
	return kb.build(KeyUserSession, token)
}

// UserPermissions 生成用户权限缓存键
func (kb *KeyBuilder) UserPermissions(userID string) string {
	return kb.build(KeyUserPermissions, userID)
}

// UserProfile 生成用户个人资料缓存键
func (kb *KeyBuilder) UserProfile(userID string) string {
	return kb.build(KeyUserProfile, userID)
}

// UserOnline 生成用户在线状态缓存键
func (kb *KeyBuilder) UserOnline(userID string) string {
	return kb.build(KeyUserOnline, userID)
}

// UserQuota 生成用户配额缓存键
func (kb *KeyBuilder) UserQuota(userID string) string {
	return kb.build(KeyUserQuota, userID)
}

// FileInfo 生成文件信息缓存键
func (kb *KeyBuilder) FileInfo(fileID string) string {
	return kb.build(KeyFileInfo, fileID)
}

// FileShare 生成文件分享缓存键
func (kb *KeyBuilder) FileShare(token string) string {
	return kb.build(KeyFileShare, token)
}

// FileUpload 生成文件上传缓存键
func (kb *KeyBuilder) FileUpload(uploadID string) string {
	return kb.build(KeyFileUpload, uploadID)
}

// FileChunk 生成文件分片缓存键
func (kb *KeyBuilder) FileChunk(uploadID string, chunkNum int) string {
	return kb.build(KeyFileChunk, uploadID, chunkNum)
}

// FilePreview 生成文件预览缓存键
func (kb *KeyBuilder) FilePreview(fileID string) string {
	return kb.build(KeyFilePreview, fileID)
}

// FileDownload 生成文件下载缓存键
func (kb *KeyBuilder) FileDownload(fileID string) string {
	return kb.build(KeyFileDownload, fileID)
}

// 团队相关键构建方法
// TeamInfo 生成团队信息缓存键
func (kb *KeyBuilder) TeamInfo(teamID string) string {
	return kb.build(KeyTeamInfo, teamID)
}

// TeamMembers 生成团队成员缓存键
func (kb *KeyBuilder) TeamMembers(teamID string) string {
	return kb.build(KeyTeamMembers, teamID)
}

// TeamFiles 生成团队文件缓存键
func (kb *KeyBuilder) TeamFiles(teamID string) string {
	return kb.build(KeyTeamFiles, teamID)
}

// TeamPermissions 生成团队权限缓存键
func (kb *KeyBuilder) TeamPermissions(teamID, userID string) string {
	return kb.build(KeyTeamPermissions, teamID, userID)
}

// 验证码相关键构建方法
// VerifyCode 生成验证码缓存键
func (kb *KeyBuilder) VerifyCode(codeType, target string) string {
	return kb.build(KeyVerifyCode, codeType, target)
}

// VerifyAttempt 生成验证码尝试次数缓存键
func (kb *KeyBuilder) VerifyAttempt(codeType, target string) string {
	return kb.build(KeyVerifyAttempt, codeType, target)
}

// VerifyBlock 生成验证码封锁缓存键
func (kb *KeyBuilder) VerifyBlock(codeType, target string) string {
	return kb.build(KeyVerifyBlock, codeType, target)
}

// 限流相关键构建方法
// RateLimit 生成限流缓存键
func (kb *KeyBuilder) RateLimit(ip, endpoint string) string {
	return kb.build(KeyRateLimit, ip, endpoint)
}

// UserRateLimit 生成用户限流缓存键
func (kb *KeyBuilder) UserRateLimit(userID, action string) string {
	return kb.build(KeyUserRateLimit, userID, action)
}

// APIRateLimit 生成API限流缓存键
func (kb *KeyBuilder) APIRateLimit(apiKey, endpoint string) string {
	return kb.build(KeyAPIRateLimit, apiKey, endpoint)
}

// 锁相关键构建方法
// FileLock 生成文件锁缓存键
func (kb *KeyBuilder) FileLock(fileID string) string {
	return kb.build(KeyFileLock, fileID)
}

// UserLock 生成用户锁缓存键
func (kb *KeyBuilder) UserLock(userID string) string {
	return kb.build(KeyUserLock, userID)
}

// TeamLock 生成团队锁缓存键
func (kb *KeyBuilder) TeamLock(teamID string) string {
	return kb.build(KeyTeamLock, teamID)
}

// UploadLock 生成上传锁缓存键
func (kb *KeyBuilder) UploadLock(uploadID string) string {
	return kb.build(KeyUploadLock, uploadID)
}

// 消息相关键构建方法
// Conversation 生成会话缓存键
func (kb *KeyBuilder) Conversation(conversationID string) string {
	return kb.build(KeyConversation, conversationID)
}

// Message 生成消息缓存键
func (kb *KeyBuilder) Message(messageID string) string {
	return kb.build(KeyMessage, messageID)
}

// MessageRead 生成消息已读状态缓存键
func (kb *KeyBuilder) MessageRead(conversationID, userID string) string {
	return kb.build(KeyMessageRead, conversationID, userID)
}

// UserMessages 生成用户消息缓存键
func (kb *KeyBuilder) UserMessages(userID string) string {
	return kb.build(KeyUserMessages, userID)
}

// 统计相关键构建方法
// UserStats 生成用户统计缓存键
func (kb *KeyBuilder) UserStats(userID string) string {
	return kb.build(KeyUserStats, userID)
}

// FileStats 生成文件统计缓存键
func (kb *KeyBuilder) FileStats(fileID string) string {
	return kb.build(KeyFileStats, fileID)
}

// TeamStats 生成团队统计缓存键
func (kb *KeyBuilder) TeamStats(teamID string) string {
	return kb.build(KeyTeamStats, teamID)
}

// SystemStats 生成系统统计缓存键
func (kb *KeyBuilder) SystemStats() string {
	return KeySystemStats
}

// 搜索相关键构建方法
// SearchIndex 生成搜索索引缓存键
func (kb *KeyBuilder) SearchIndex(indexType string) string {
	return kb.build(KeySearchIndex, indexType)
}

// SearchResult 生成搜索结果缓存键
func (kb *KeyBuilder) SearchResult(queryHash string) string {
	return kb.build(KeySearchResult, queryHash)
}

// SearchHistory 生成搜索历史缓存键
func (kb *KeyBuilder) SearchHistory(userID string) string {
	return kb.build(KeySearchHistory, userID)
}

// Keys 全局键构建器实例
var Keys = NewKeyBuilder()

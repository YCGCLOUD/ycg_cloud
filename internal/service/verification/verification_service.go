package verification

import (
	"context"
	"time"

	"cloudpan/internal/repository/models"
)

// VerificationService 验证码服务接口
//
// 提供完整的验证码管理功能，包括：
// 1. 验证码生成：支持多种类型的验证码生成
// 2. 验证码验证：验证码格式验证和内容验证
// 3. 验证码管理：过期清理、使用状态管理
// 4. 安全防护：频率限制、尝试次数限制
//
// 使用示例：
//
//	service := NewVerificationService(db, emailService, logger)
//	code, err := service.GenerateEmailCode(ctx, email, "password_reset", userID, request.RemoteAddr)
//	isValid, err := service.VerifyEmailCode(ctx, email, "password_reset", inputCode)
type VerificationService interface {
	// 验证码生成
	GenerateEmailCode(ctx context.Context, email, codeType string, userID *uint, ipAddress string) (*models.VerificationCode, error)
	GeneratePhoneCode(ctx context.Context, phone, codeType string, userID *uint, ipAddress string) (*models.VerificationCode, error)

	// 验证码验证
	VerifyEmailCode(ctx context.Context, email, codeType, code string) (*models.VerificationCode, error)
	VerifyPhoneCode(ctx context.Context, phone, codeType, code string) (*models.VerificationCode, error)

	// 验证码管理
	GetActiveCode(ctx context.Context, target, codeType string) (*models.VerificationCode, error)
	InvalidateCode(ctx context.Context, codeID uint) error
	CleanupExpiredCodes(ctx context.Context) error

	// 安全检查
	CheckRateLimit(ctx context.Context, target, codeType string, ipAddress string) error
	GetAttemptCount(ctx context.Context, target, codeType string, timeWindow time.Duration) (int, error)

	// 验证码状态
	IsCodeValid(ctx context.Context, codeID uint) (bool, error)
	MarkCodeAsUsed(ctx context.Context, codeID uint) error

	// 密码重置专用方法
	GeneratePasswordResetCode(ctx context.Context, email string, userID uint, ipAddress string) (*models.VerificationCode, error)
	VerifyPasswordResetCode(ctx context.Context, email, code string) (*models.VerificationCode, error)
	CompletePasswordReset(ctx context.Context, codeID uint) error

	// 邮箱验证专用方法
	GenerateEmailVerificationCode(ctx context.Context, email string, userID uint, ipAddress string) (*models.VerificationCode, error)
	VerifyEmailVerificationCode(ctx context.Context, email, code string) (*models.VerificationCode, error)

	// 批量操作
	CleanupUserCodes(ctx context.Context, userID uint, codeType string) error
	GetUserActiveCodes(ctx context.Context, userID uint) ([]*models.VerificationCode, error)
}

// CodeGenerationRequest 验证码生成请求
type CodeGenerationRequest struct {
	Target    string  `json:"target"`     // 目标（邮箱/手机号）
	Type      string  `json:"type"`       // 验证码类型
	UserID    *uint   `json:"user_id"`    // 用户ID（可选）
	IPAddress string  `json:"ip_address"` // 请求IP
	UserAgent *string `json:"user_agent"` // 用户代理（可选）
	ExpiresIn int     `json:"expires_in"` // 过期时间（分钟，可选）
}

// CodeVerificationRequest 验证码验证请求
type CodeVerificationRequest struct {
	Target string `json:"target"` // 目标（邮箱/手机号）
	Type   string `json:"type"`   // 验证码类型
	Code   string `json:"code"`   // 验证码
}

// CodeGenerationResponse 验证码生成响应
type CodeGenerationResponse struct {
	CodeID    uint      `json:"code_id"`    // 验证码ID
	Target    string    `json:"target"`     // 目标
	Type      string    `json:"type"`       // 类型
	ExpiresAt time.Time `json:"expires_at"` // 过期时间
	Success   bool      `json:"success"`    // 是否成功
	Message   string    `json:"message"`    // 消息
}

// CodeVerificationResponse 验证码验证响应
type CodeVerificationResponse struct {
	CodeID       uint   `json:"code_id"`       // 验证码ID
	IsValid      bool   `json:"is_valid"`      // 是否有效
	Message      string `json:"message"`       // 消息
	AttemptCount int    `json:"attempt_count"` // 尝试次数
	MaxAttempts  int    `json:"max_attempts"`  // 最大尝试次数
}

// RateLimitInfo 频率限制信息
type RateLimitInfo struct {
	Limit     int           `json:"limit"`      // 限制次数
	Window    time.Duration `json:"window"`     // 时间窗口
	Current   int           `json:"current"`    // 当前次数
	ResetTime time.Time     `json:"reset_time"` // 重置时间
	IsBlocked bool          `json:"is_blocked"` // 是否被阻塞
}

// CodeStatistics 验证码统计信息
type CodeStatistics struct {
	TotalGenerated int     `json:"total_generated"` // 总生成数
	TotalVerified  int     `json:"total_verified"`  // 总验证数
	TotalExpired   int     `json:"total_expired"`   // 总过期数
	TotalUsed      int     `json:"total_used"`      // 总使用数
	SuccessRate    float64 `json:"success_rate"`    // 成功率
}

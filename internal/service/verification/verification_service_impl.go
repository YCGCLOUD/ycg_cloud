package verification

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"cloudpan/internal/pkg/email"
	"cloudpan/internal/pkg/errors"
	"cloudpan/internal/pkg/utils"
	"cloudpan/internal/repository/models"
)

// verificationService 验证码服务实现
type verificationService struct {
	db           *gorm.DB
	emailService email.EmailService
	logger       *zap.Logger
	codeManager  utils.EmailCodeManager
	validator    utils.Validator
}

// NewVerificationService 创建验证码服务实例
func NewVerificationService(db *gorm.DB, emailService email.EmailService, logger *zap.Logger) VerificationService {
	return &verificationService{
		db:           db,
		emailService: emailService,
		logger:       logger,
		codeManager:  utils.NewEmailCodeManager(),
		validator:    utils.NewValidator(),
	}
}

// GenerateEmailCode 生成邮箱验证码
func (s *verificationService) GenerateEmailCode(ctx context.Context, email, codeType string, userID *uint, ipAddress string) (*models.VerificationCode, error) {
	// 验证输入参数
	if err := s.validateCodeGenerationParams(email, codeType); err != nil {
		return nil, err
	}

	// 检查频率限制
	if err := s.CheckRateLimit(ctx, email, codeType, ipAddress); err != nil {
		return nil, err
	}

	// 生成验证码和盐值
	code, salt, err := s.generateCodeAndSalt(codeType)
	if err != nil {
		return nil, err
	}

	// 失效旧验证码
	if err := s.invalidateOldCodes(ctx, email, codeType); err != nil {
		s.logger.Warn("Failed to invalidate old codes", zap.Error(err))
	}

	// 创建和保存验证码记录
	verificationCode, err := s.createAndSaveCode(ctx, email, codeType, code, salt, ipAddress, userID)
	if err != nil {
		return nil, err
	}

	// 发送邮件
	if err := s.sendVerificationEmail(ctx, email, code, codeType); err != nil {
		s.logger.Error("Failed to send verification email",
			zap.String("email", email),
			zap.String("type", codeType),
			zap.Error(err))
		// 不返回错误，验证码已生成成功
	}

	s.logger.Info("Verification code generated successfully",
		zap.String("target", email),
		zap.String("type", codeType),
		zap.String("ip", ipAddress),
		zap.Uint("code_id", verificationCode.ID))

	return verificationCode, nil
}

// validateCodeGenerationParams 验证验证码生成参数
func (s *verificationService) validateCodeGenerationParams(email, codeType string) error {
	// 验证邮箱格式
	if err := s.validator.ValidateEmail(email); err != nil {
		return errors.NewValidationError("email", err.Error())
	}

	// 验证验证码类型
	if err := s.codeManager.ValidateCodeType(codeType); err != nil {
		return errors.NewValidationError("code_type", err.Error())
	}
	return nil
}

// generateCodeAndSalt 生成验证码和盐值
func (s *verificationService) generateCodeAndSalt(codeType string) (string, string, error) {
	// 生成验证码
	code, err := s.codeManager.GenerateVerificationCode(codeType)
	if err != nil {
		s.logger.Error("Failed to generate verification code", zap.Error(err))
		return "", "", errors.NewInternalError("验证码生成失败")
	}

	// 生成盐值
	salt, err := s.codeManager.GenerateSalt()
	if err != nil {
		s.logger.Error("Failed to generate salt", zap.Error(err))
		return "", "", errors.NewInternalError("验证码生成失败")
	}
	return code, salt, nil
}

// createAndSaveCode 创建和保存验证码记录
func (s *verificationService) createAndSaveCode(ctx context.Context, email, codeType, code, salt, ipAddress string, userID *uint) (*models.VerificationCode, error) {
	// 哈希验证码
	codeHash := s.codeManager.HashVerificationCode(code, salt)

	// 设置过期时间
	expiresAt := s.calculateExpirationTime(codeType)

	// 创建验证码记录
	verificationCode := &models.VerificationCode{
		Target:    email,
		Type:      codeType,
		CodeHash:  codeHash,
		Salt:      salt,
		ExpiresAt: expiresAt,
		IPAddress: ipAddress,
		UserID:    userID,
	}

	if err := s.db.WithContext(ctx).Create(verificationCode).Error; err != nil {
		s.logger.Error("Failed to save verification code", zap.Error(err))
		return nil, errors.NewInternalError("验证码保存失败")
	}
	return verificationCode, nil
}

// calculateExpirationTime 计算过期时间
func (s *verificationService) calculateExpirationTime(codeType string) time.Time {
	switch codeType {
	case models.VerificationTypeResetPassword:
		return time.Now().Add(30 * time.Minute) // 密码重置30分钟
	case models.VerificationTypeLogin:
		return time.Now().Add(5 * time.Minute) // 登录验证5分钟
	default:
		return time.Now().Add(15 * time.Minute) // 默认15分钟
	}
}

// GeneratePasswordResetCode 生成密码重置验证码
func (s *verificationService) GeneratePasswordResetCode(ctx context.Context, email string, userID uint, ipAddress string) (*models.VerificationCode, error) {
	return s.GenerateEmailCode(ctx, email, models.VerificationTypeResetPassword, &userID, ipAddress)
}

// VerifyEmailCode 验证邮箱验证码
func (s *verificationService) VerifyEmailCode(ctx context.Context, email, codeType, code string) (*models.VerificationCode, error) {
	// 验证输入参数
	if err := s.validator.ValidateEmail(email); err != nil {
		return nil, errors.NewValidationError("email", err.Error())
	}

	if err := s.codeManager.ValidateCodeFormat(code); err != nil {
		return nil, errors.NewValidationError("code", err.Error())
	}

	if err := s.codeManager.ValidateCodeType(codeType); err != nil {
		return nil, errors.NewValidationError("code_type", err.Error())
	}

	// 查找有效的验证码
	var verificationCode models.VerificationCode
	err := s.db.WithContext(ctx).Where(
		"target = ? AND type = ? AND is_used = false AND expires_at > ?",
		email, codeType, time.Now(),
	).Order("created_at DESC").First(&verificationCode).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewValidationError("code", "验证码不存在或已过期")
		}
		s.logger.Error("Failed to query verification code", zap.Error(err))
		return nil, errors.NewInternalError("验证码查询失败")
	}

	// 检查尝试次数
	if verificationCode.AttemptCount >= verificationCode.MaxAttempts {
		return nil, errors.NewValidationError("code", "验证码尝试次数过多，请重新获取")
	}

	// 增加尝试次数
	verificationCode.AttemptCount++
	s.db.WithContext(ctx).Model(&verificationCode).Update("attempt_count", verificationCode.AttemptCount)

	// 验证验证码
	isValid := s.codeManager.HashVerificationCode(code, verificationCode.Salt) == verificationCode.CodeHash
	if !isValid {
		s.logger.Warn("Invalid verification code attempt",
			zap.String("target", email),
			zap.String("type", codeType),
			zap.Int("attempt", verificationCode.AttemptCount))
		return nil, errors.NewValidationError("code", "验证码错误")
	}

	s.logger.Info("Verification code verified successfully",
		zap.String("target", email),
		zap.String("type", codeType),
		zap.Uint("code_id", verificationCode.ID))

	return &verificationCode, nil
}

// VerifyPasswordResetCode 验证密码重置验证码
func (s *verificationService) VerifyPasswordResetCode(ctx context.Context, email, code string) (*models.VerificationCode, error) {
	return s.VerifyEmailCode(ctx, email, models.VerificationTypeResetPassword, code)
}

// CompletePasswordReset 完成密码重置
func (s *verificationService) CompletePasswordReset(ctx context.Context, codeID uint) error {
	return s.MarkCodeAsUsed(ctx, codeID)
}

// MarkCodeAsUsed 标记验证码为已使用
func (s *verificationService) MarkCodeAsUsed(ctx context.Context, codeID uint) error {
	now := time.Now()
	result := s.db.WithContext(ctx).Model(&models.VerificationCode{}).
		Where("id = ? AND is_used = false", codeID).
		Updates(map[string]interface{}{
			"is_used": true,
			"used_at": now,
		})

	if result.Error != nil {
		s.logger.Error("Failed to mark code as used",
			zap.Uint("code_id", codeID),
			zap.Error(result.Error))
		return errors.NewInternalError("验证码状态更新失败")
	}

	if result.RowsAffected == 0 {
		return errors.NewValidationError("code", "验证码不存在或已使用")
	}

	return nil
}

// CheckRateLimit 检查频率限制
func (s *verificationService) CheckRateLimit(ctx context.Context, target, codeType string, ipAddress string) error {
	// 检查同一邮箱的频率限制（5分钟内最多3次）
	count := int64(0)
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)

	err := s.db.WithContext(ctx).Model(&models.VerificationCode{}).
		Where("target = ? AND type = ? AND created_at > ?", target, codeType, fiveMinutesAgo).
		Count(&count).Error

	if err != nil {
		s.logger.Error("Failed to check rate limit", zap.Error(err))
		return errors.NewInternalError("频率检查失败")
	}

	if count >= 3 {
		return errors.NewValidationError("rate_limit", "获取验证码过于频繁，请5分钟后再试")
	}

	// 检查同一IP的频率限制（1小时内最多10次）
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	err = s.db.WithContext(ctx).Model(&models.VerificationCode{}).
		Where("ip_address = ? AND created_at > ?", ipAddress, oneHourAgo).
		Count(&count).Error

	if err != nil {
		s.logger.Error("Failed to check IP rate limit", zap.Error(err))
		return errors.NewInternalError("频率检查失败")
	}

	if count >= 10 {
		return errors.NewValidationError("rate_limit", "该IP获取验证码过于频繁，请稍后再试")
	}

	return nil
}

// GetActiveCode 获取活跃的验证码
func (s *verificationService) GetActiveCode(ctx context.Context, target, codeType string) (*models.VerificationCode, error) {
	var verificationCode models.VerificationCode
	err := s.db.WithContext(ctx).Where(
		"target = ? AND type = ? AND is_used = false AND expires_at > ?",
		target, codeType, time.Now(),
	).Order("created_at DESC").First(&verificationCode).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &verificationCode, nil
}

// CleanupExpiredCodes 清理过期验证码
func (s *verificationService) CleanupExpiredCodes(ctx context.Context) error {
	result := s.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&models.VerificationCode{})
	if result.Error != nil {
		s.logger.Error("Failed to cleanup expired codes", zap.Error(result.Error))
		return result.Error
	}

	s.logger.Info("Cleaned up expired verification codes", zap.Int64("count", result.RowsAffected))
	return nil
}

// sendVerificationEmail 发送验证邮件
func (s *verificationService) sendVerificationEmail(ctx context.Context, email, code, codeType string) error {
	switch codeType {
	case models.VerificationTypeResetPassword:
		return s.emailService.SendPasswordReset(ctx, email, code)
	case models.VerificationTypeRegister:
		return s.emailService.SendVerificationCode(ctx, email, code)
	default:
		return s.emailService.SendVerificationCode(ctx, email, code)
	}
}

// invalidateOldCodes 使旧验证码失效
func (s *verificationService) invalidateOldCodes(ctx context.Context, target, codeType string) error {
	return s.db.WithContext(ctx).Model(&models.VerificationCode{}).
		Where("target = ? AND type = ? AND is_used = false", target, codeType).
		Update("is_used", true).Error
}

// 实现其他接口方法的简化版本

func (s *verificationService) GeneratePhoneCode(ctx context.Context, phone, codeType string, userID *uint, ipAddress string) (*models.VerificationCode, error) {
	return nil, errors.NewValidationError("phone", "手机验证码功能尚未实现")
}

func (s *verificationService) VerifyPhoneCode(ctx context.Context, phone, codeType, code string) (*models.VerificationCode, error) {
	return nil, errors.NewValidationError("phone", "手机验证码功能尚未实现")
}

func (s *verificationService) InvalidateCode(ctx context.Context, codeID uint) error {
	return s.MarkCodeAsUsed(ctx, codeID)
}

func (s *verificationService) GetAttemptCount(ctx context.Context, target, codeType string, timeWindow time.Duration) (int, error) {
	var count int64
	since := time.Now().Add(-timeWindow)
	err := s.db.WithContext(ctx).Model(&models.VerificationCode{}).
		Where("target = ? AND type = ? AND created_at > ?", target, codeType, since).
		Count(&count).Error
	return int(count), err
}

func (s *verificationService) IsCodeValid(ctx context.Context, codeID uint) (bool, error) {
	var verificationCode models.VerificationCode
	err := s.db.WithContext(ctx).First(&verificationCode, codeID).Error
	if err != nil {
		return false, err
	}
	return verificationCode.IsValid(), nil
}

func (s *verificationService) GenerateEmailVerificationCode(ctx context.Context, email string, userID uint, ipAddress string) (*models.VerificationCode, error) {
	return s.GenerateEmailCode(ctx, email, models.VerificationTypeRegister, &userID, ipAddress)
}

func (s *verificationService) VerifyEmailVerificationCode(ctx context.Context, email, code string) (*models.VerificationCode, error) {
	return s.VerifyEmailCode(ctx, email, models.VerificationTypeRegister, code)
}

func (s *verificationService) CleanupUserCodes(ctx context.Context, userID uint, codeType string) error {
	return s.db.WithContext(ctx).Where("user_id = ? AND type = ?", userID, codeType).Delete(&models.VerificationCode{}).Error
}

func (s *verificationService) GetUserActiveCodes(ctx context.Context, userID uint) ([]*models.VerificationCode, error) {
	var codes []*models.VerificationCode
	err := s.db.WithContext(ctx).Where(
		"user_id = ? AND is_used = false AND expires_at > ?",
		userID, time.Now(),
	).Find(&codes).Error
	return codes, err
}

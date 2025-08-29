package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"cloudpan/internal/pkg/errors"
	"cloudpan/internal/pkg/utils"
	"cloudpan/internal/service/user"
	"cloudpan/internal/service/verification"
)

// ForgotPasswordRequest 忘记密码请求结构体
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required" example:"user@example.com"`
}

// ForgotPasswordResponse 忘记密码响应结构体
type ForgotPasswordResponse struct {
	Message   string    `json:"message" example:"密码重置邮件已发送"`
	Email     string    `json:"email" example:"user@example.com"`
	ExpiresAt time.Time `json:"expires_at" example:"2024-01-01T01:30:00Z"`
	Success   bool      `json:"success" example:"true"`
}

// ResetPasswordRequest 重置密码请求结构体
type ResetPasswordRequest struct {
	Email            string `json:"email" binding:"required,email" example:"user@example.com"`
	VerificationCode string `json:"verification_code" binding:"required" example:"123456"`
	NewPassword      string `json:"new_password" binding:"required" example:"NewPassword123!"`
	ConfirmPassword  string `json:"confirm_password" binding:"required" example:"NewPassword123!"`
}

// ResetPasswordResponse 重置密码响应结构体
type ResetPasswordResponse struct {
	Message string `json:"message" example:"密码重置成功"`
	Success bool   `json:"success" example:"true"`
}

// ChangePasswordRequest 修改密码请求结构体
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required" example:"OldPassword123!"`
	NewPassword     string `json:"new_password" binding:"required" example:"NewPassword123!"`
	ConfirmPassword string `json:"confirm_password" binding:"required" example:"NewPassword123!"`
}

// ChangePasswordResponse 修改密码响应结构体
type ChangePasswordResponse struct {
	Message string `json:"message" example:"密码修改成功"`
	Success bool   `json:"success" example:"true"`
}

// PasswordStrengthRequest 密码强度检查请求结构体
type PasswordStrengthRequest struct {
	Password string `json:"password" binding:"required" example:"TestPassword123!"`
}

// PasswordStrengthResponse 密码强度检查响应结构体
type PasswordStrengthResponse struct {
	Strength     int      `json:"strength" example:"3"`
	StrengthText string   `json:"strength_text" example:"强"`
	Score        int      `json:"score" example:"85"`
	Suggestions  []string `json:"suggestions,omitempty"`
	IsValid      bool     `json:"is_valid" example:"true"`
}

// PasswordManagerHandler 密码管理处理器
type PasswordManagerHandler struct {
	userService         user.UserService
	verificationService verification.VerificationService
	logger              *zap.Logger
	validator           utils.ParameterValidator
	passwordHasher      utils.PasswordHasher
}

// NewPasswordManagerHandler 创建新的密码管理处理器
func NewPasswordManagerHandler(
	userService user.UserService,
	verificationService verification.VerificationService,
	logger *zap.Logger,
) *PasswordManagerHandler {
	return &PasswordManagerHandler{
		userService:         userService,
		verificationService: verificationService,
		logger:              logger,
		validator:           utils.NewParameterValidator(),
		passwordHasher:      utils.NewDefaultPasswordHasher(),
	}
}

// ForgotPassword 忘记密码
//
// @Summary 忘记密码
// @Description 发送密码重置邮件到用户邮箱
// @Tags 密码管理
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "忘记密码请求"
// @Success 200 {object} utils.Response{data=ForgotPasswordResponse} "请求成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 404 {object} utils.Response "用户不存在"
// @Failure 429 {object} utils.Response "请求频率限制"
// @Failure 500 {object} utils.Response "内部服务器错误"
// @Router /api/v1/password/forgot [post]
func (h *PasswordManagerHandler) ForgotPassword(c *gin.Context) {
	ctx := c.Request.Context()

	// 解析请求参数
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid forgot password request", zap.Error(err), zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeBadRequest, "请求参数格式错误")
		return
	}

	// 验证邮箱格式
	if err := utils.ValidatePasswordResetRequest(req.Email); err != nil {
		h.logger.Warn("Invalid email for password reset",
			zap.String("email", req.Email),
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeValidationError, err.Error())
		return
	}

	// 查找用户
	user, err := h.userService.GetUserByEmail(ctx, req.Email)
	if err != nil {
		h.logger.Warn("User not found for password reset",
			zap.String("email", req.Email),
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		// 为了安全，不透露用户是否存在
		utils.SuccessWithMessage(c, "如果邮箱存在，密码重置邮件已发送", ForgotPasswordResponse{
			Message:   "如果邮箱存在，密码重置邮件已发送",
			Email:     req.Email,
			ExpiresAt: time.Now().Add(30 * time.Minute),
			Success:   true,
		})
		return
	}

	// 检查用户状态
	if user.Status != "active" {
		h.logger.Warn("Inactive user attempted password reset",
			zap.String("email", req.Email),
			zap.String("status", user.Status),
			zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeUnauthorized, "账户状态异常，无法重置密码")
		return
	}

	// 生成密码重置验证码
	verificationCode, err := h.verificationService.GeneratePasswordResetCode(
		ctx,
		req.Email,
		user.ID,
		c.ClientIP(),
	)
	if err != nil {
		h.logger.Error("Failed to generate password reset code",
			zap.String("email", req.Email),
			zap.Uint("user_id", user.ID),
			zap.Error(err),
			zap.String("ip", c.ClientIP()))

		// 检查是否是频率限制错误
		if validationErr, ok := err.(*errors.ValidationError); ok && validationErr.Field == "rate_limit" {
			utils.ErrorWithMessage(c, utils.CodeTooManyRequests, validationErr.Message)
			return
		}

		utils.InternalErrorWithMessage(c, "密码重置邮件发送失败，请稍后重试")
		return
	}

	h.logger.Info("Password reset code generated successfully",
		zap.String("email", req.Email),
		zap.Uint("user_id", user.ID),
		zap.Uint("code_id", verificationCode.ID),
		zap.String("ip", c.ClientIP()))

	response := ForgotPasswordResponse{
		Message:   "密码重置邮件已发送到您的邮箱",
		Email:     req.Email,
		ExpiresAt: verificationCode.ExpiresAt,
		Success:   true,
	}

	utils.SuccessWithMessage(c, "密码重置邮件已发送", response)
}

// ResetPassword 重置密码
//
// @Summary 重置密码
// @Description 使用验证码重置用户密码
// @Tags 密码管理
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "重置密码请求"
// @Success 200 {object} utils.Response{data=ResetPasswordResponse} "重置成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 401 {object} utils.Response "验证码错误"
// @Failure 404 {object} utils.Response "用户不存在"
// @Failure 500 {object} utils.Response "内部服务器错误"
// @Router /api/v1/password/reset [post]
func (h *PasswordManagerHandler) ResetPassword(c *gin.Context) {
	ctx := c.Request.Context()

	// 解析请求参数
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid reset password request", zap.Error(err), zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeBadRequest, "请求参数格式错误")
		return
	}

	// 验证密码重置参数
	if err := utils.ValidatePasswordResetConfirm(req.Email, req.VerificationCode, req.NewPassword, req.ConfirmPassword); err != nil {
		h.logger.Warn("Invalid password reset parameters",
			zap.String("email", req.Email),
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeValidationError, err.Error())
		return
	}

	// 验证验证码
	verificationCode, err := h.verificationService.VerifyPasswordResetCode(ctx, req.Email, req.VerificationCode)
	if err != nil {
		h.logger.Warn("Invalid verification code for password reset",
			zap.String("email", req.Email),
			zap.String("code", req.VerificationCode),
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeUnauthorized, err.Error())
		return
	}

	// 查找用户
	user, err := h.userService.GetUserByEmail(ctx, req.Email)
	if err != nil {
		h.logger.Error("User not found during password reset",
			zap.String("email", req.Email),
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeNotFound, "用户不存在")
		return
	}

	// 检查验证码是否属于该用户
	if verificationCode.UserID == nil || *verificationCode.UserID != user.ID {
		h.logger.Warn("Verification code user mismatch",
			zap.String("email", req.Email),
			zap.Uint("user_id", user.ID),
			zap.Uint("code_user_id", *verificationCode.UserID),
			zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeUnauthorized, "验证码无效")
		return
	}

	// 哈希新密码
	hashedPassword, err := h.passwordHasher.HashPassword(req.NewPassword)
	if err != nil {
		h.logger.Error("Failed to hash new password",
			zap.String("email", req.Email),
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		utils.InternalErrorWithMessage(c, "密码加密失败")
		return
	}

	// 更新用户密码
	if err := h.userService.UpdatePassword(ctx, user.ID, hashedPassword); err != nil {
		h.logger.Error("Failed to update user password",
			zap.String("email", req.Email),
			zap.Uint("user_id", user.ID),
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		utils.InternalErrorWithMessage(c, "密码更新失败")
		return
	}

	// 标记验证码为已使用
	if err := h.verificationService.CompletePasswordReset(ctx, verificationCode.ID); err != nil {
		h.logger.Error("Failed to mark verification code as used",
			zap.Uint("code_id", verificationCode.ID),
			zap.Error(err))
		// 不影响密码重置成功
	}

	h.logger.Info("Password reset completed successfully",
		zap.String("email", req.Email),
		zap.Uint("user_id", user.ID),
		zap.Uint("code_id", verificationCode.ID),
		zap.String("ip", c.ClientIP()))

	response := ResetPasswordResponse{
		Message: "密码重置成功",
		Success: true,
	}

	utils.SuccessWithMessage(c, "密码重置成功", response)
}

// ChangePassword 修改密码
//
// @Summary 修改密码
// @Description 用户修改自己的密码
// @Tags 密码管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} utils.Response{data=ChangePasswordResponse} "修改成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 401 {object} utils.Response "当前密码错误"
// @Failure 500 {object} utils.Response "内部服务器错误"
// @Router /api/v1/password/change [post]
func (h *PasswordManagerHandler) ChangePassword(c *gin.Context) {
	ctx := c.Request.Context()

	// 获取当前用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		utils.ErrorWithMessage(c, utils.CodeUnauthorized, "用户身份验证失败")
		return
	}

	currentUserID, ok := userID.(uint)
	if !ok {
		h.logger.Error("Invalid user ID type in context")
		utils.ErrorWithMessage(c, utils.CodeUnauthorized, "用户身份验证失败")
		return
	}

	// 解析请求参数
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid change password request", zap.Error(err), zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeBadRequest, "请求参数格式错误")
		return
	}

	// 验证密码修改参数
	if err := h.validator.ValidatePasswordChangeParams(req.CurrentPassword, req.NewPassword, req.ConfirmPassword); err != nil {
		h.logger.Warn("Invalid password change parameters",
			zap.Uint("user_id", currentUserID),
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeValidationError, err.Error())
		return
	}

	// 获取用户信息
	user, err := h.userService.GetUserByID(ctx, currentUserID)
	if err != nil {
		h.logger.Error("User not found during password change",
			zap.Uint("user_id", currentUserID),
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeNotFound, "用户不存在")
		return
	}

	// 验证当前密码
	if !h.passwordHasher.VerifyPassword(user.PasswordHash, req.CurrentPassword) {
		h.logger.Warn("Current password verification failed",
			zap.Uint("user_id", currentUserID),
			zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeUnauthorized, "当前密码错误")
		return
	}

	// 哈希新密码
	hashedPassword, err := h.passwordHasher.HashPassword(req.NewPassword)
	if err != nil {
		h.logger.Error("Failed to hash new password",
			zap.Uint("user_id", currentUserID),
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		utils.InternalErrorWithMessage(c, "密码加密失败")
		return
	}

	// 更新密码
	if err := h.userService.UpdatePassword(ctx, currentUserID, hashedPassword); err != nil {
		h.logger.Error("Failed to update password",
			zap.Uint("user_id", currentUserID),
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		utils.InternalErrorWithMessage(c, "密码更新失败")
		return
	}

	h.logger.Info("Password changed successfully",
		zap.Uint("user_id", currentUserID),
		zap.String("ip", c.ClientIP()))

	response := ChangePasswordResponse{
		Message: "密码修改成功",
		Success: true,
	}

	utils.SuccessWithMessage(c, "密码修改成功", response)
}

// CheckPasswordStrength 检查密码强度
//
// @Summary 检查密码强度
// @Description 检查密码强度并返回安全建议
// @Tags 密码管理
// @Accept json
// @Produce json
// @Param request body PasswordStrengthRequest true "密码强度检查请求"
// @Success 200 {object} utils.Response{data=PasswordStrengthResponse} "检查成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Router /api/v1/password/strength [post]
func (h *PasswordManagerHandler) CheckPasswordStrength(c *gin.Context) {
	// 解析请求参数
	var req PasswordStrengthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid password strength request", zap.Error(err), zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeBadRequest, "请求参数格式错误")
		return
	}

	// 检查密码强度
	strength, err := h.passwordHasher.ValidatePasswordStrength(req.Password)
	if err != nil {
		// 密码不符合要求，但仍返回强度信息
		response := PasswordStrengthResponse{
			Strength:     utils.PasswordWeek,
			StrengthText: h.getStrengthText(utils.PasswordWeek),
			Score:        h.calculatePasswordScore(req.Password),
			Suggestions:  h.getPasswordSuggestions(req.Password, err),
			IsValid:      false,
		}
		utils.SuccessWithMessage(c, "密码强度检查完成", response)
		return
	}

	// 生成密码建议
	suggestions := h.getPasswordSuggestions(req.Password, nil)

	response := PasswordStrengthResponse{
		Strength:     strength,
		StrengthText: h.getStrengthText(strength),
		Score:        h.calculatePasswordScore(req.Password),
		Suggestions:  suggestions,
		IsValid:      true,
	}

	utils.SuccessWithMessage(c, "密码强度检查完成", response)
}

// getStrengthText 获取强度文本描述
func (h *PasswordManagerHandler) getStrengthText(strength int) string {
	switch strength {
	case utils.PasswordWeek:
		return "弱"
	case utils.PasswordMedium:
		return "中等"
	case utils.PasswordStrong:
		return "强"
	default:
		return "未知"
	}
}

// calculatePasswordScore 计算密码评分（0-100）
func (h *PasswordManagerHandler) calculatePasswordScore(password string) int {
	score := 0

	// 长度评分（最多30分）
	score += h.calculateLengthScore(password)

	// 字符类型评分
	score += h.calculateCharTypeScore(password)

	// 复杂度奖励
	score += h.calculateComplexityBonus(password)

	// 确保不超过100分
	if score > 100 {
		score = 100
	}

	return score
}

// calculateLengthScore 计算长度得分
func (h *PasswordManagerHandler) calculateLengthScore(password string) int {
	score := 0
	length := len(password)

	if length >= 8 {
		score += 10
	}
	if length >= 12 {
		score += 10
	}
	if length >= 16 {
		score += 10
	}
	return score
}

// calculateCharTypeScore 计算字符类型得分
func (h *PasswordManagerHandler) calculateCharTypeScore(password string) int {
	score := 0
	hasUpper, hasLower, hasDigit, hasSpecial := h.analyzePasswordChars(password)

	if hasUpper {
		score += 15
	}
	if hasLower {
		score += 15
	}
	if hasDigit {
		score += 15
	}
	if hasSpecial {
		score += 15
	}
	return score
}

// calculateComplexityBonus 计算复杂度奖励
func (h *PasswordManagerHandler) calculateComplexityBonus(password string) int {
	hasUpper, hasLower, hasDigit, hasSpecial := h.analyzePasswordChars(password)

	complexityCount := 0
	if hasUpper {
		complexityCount++
	}
	if hasLower {
		complexityCount++
	}
	if hasDigit {
		complexityCount++
	}
	if hasSpecial {
		complexityCount++
	}

	if complexityCount >= 3 {
		return 10
	}
	return 0
}

// analyzePasswordChars 分析密码字符类型
func (h *PasswordManagerHandler) analyzePasswordChars(password string) (hasUpper, hasLower, hasDigit, hasSpecial bool) {
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}
	return
}

// getPasswordSuggestions 获取密码建议
func (h *PasswordManagerHandler) getPasswordSuggestions(password string, validationErr error) []string {
	var suggestions []string

	// 添加长度建议
	suggestions = append(suggestions, h.getLengthSuggestions(password)...)

	// 添加字符类型建议
	suggestions = append(suggestions, h.getCharacterTypeSuggestions(password)...)

	// 添加验证错误信息
	if validationErr != nil {
		suggestions = append(suggestions, validationErr.Error())
	}

	// 如果没有建议，添加默认建议
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "密码强度良好，建议定期更换")
	}

	return suggestions
}

// getLengthSuggestions 获取长度建议
func (h *PasswordManagerHandler) getLengthSuggestions(password string) []string {
	var suggestions []string

	if len(password) < 8 {
		suggestions = append(suggestions, "密码长度至少8位")
	}

	if len(password) < 12 {
		suggestions = append(suggestions, "建议密码长度12位以上，增强安全性")
	}

	return suggestions
}

// getCharacterTypeSuggestions 获取字符类型建议
func (h *PasswordManagerHandler) getCharacterTypeSuggestions(password string) []string {
	var suggestions []string
	hasUpper, hasLower, hasDigit, hasSpecial := h.analyzePasswordChars(password)

	if !hasUpper {
		suggestions = append(suggestions, "添加大写字母")
	}
	if !hasLower {
		suggestions = append(suggestions, "添加小写字母")
	}
	if !hasDigit {
		suggestions = append(suggestions, "添加数字")
	}
	if !hasSpecial {
		suggestions = append(suggestions, "添加特殊字符（如!@#$%^&*）")
	}

	return suggestions
}

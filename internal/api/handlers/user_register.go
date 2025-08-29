package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"cloudpan/internal/pkg/email"
	"cloudpan/internal/pkg/utils"
	"cloudpan/internal/repository/models"
	"cloudpan/internal/service/user"
)

// CacheInterface 缓存接口，用于支持Mock测试
type CacheInterface interface {
	SetWithTTL(key string, value interface{}, ttl time.Duration) error
	Get(key string, dest interface{}) error
	Delete(keys ...string) error
}

// RegisterRequest 用户注册请求结构体
type RegisterRequest struct {
	Email            string `json:"email" binding:"required,email" validate:"required,email"`                    // 邮箱地址
	Username         string `json:"username" binding:"required,min=3,max=50" validate:"required,min=3,max=50"`   // 用户名
	Password         string `json:"password" binding:"required,min=8,max=128" validate:"required,min=8,max=128"` // 密码
	ConfirmPassword  string `json:"confirm_password" binding:"required" validate:"required"`                     // 确认密码
	VerificationCode string `json:"verification_code" binding:"required,len=6" validate:"required,len=6"`        // 邮箱验证码
	DisplayName      string `json:"display_name,omitempty" validate:"omitempty,min=1,max=100"`                   // 显示名称（可选）
	AcceptTerms      bool   `json:"accept_terms" binding:"required" validate:"required"`                         // 接受服务条款
}

// RegisterResponse 用户注册响应结构体
type RegisterResponse struct {
	UserID      uint   `json:"user_id"`      // 用户ID
	UUID        string `json:"uuid"`         // 用户UUID
	Email       string `json:"email"`        // 邮箱地址
	Username    string `json:"username"`     // 用户名
	DisplayName string `json:"display_name"` // 显示名称
	Status      string `json:"status"`       // 用户状态
	CreatedAt   string `json:"created_at"`   // 创建时间
	Message     string `json:"message"`      // 响应消息
}

// SendVerificationCodeRequest 发送验证码请求结构体
type SendVerificationCodeRequest struct {
	Email string `json:"email" binding:"required,email" validate:"required,email"`                  // 邮箱地址
	Type  string `json:"type" binding:"required" validate:"required,oneof=register password_reset"` // 验证码类型
}

// SendVerificationCodeResponse 发送验证码响应结构体
type SendVerificationCodeResponse struct {
	Email     string `json:"email"`      // 邮箱地址
	ExpiresIn int    `json:"expires_in"` // 过期时间(秒)
	Message   string `json:"message"`    // 响应消息
}

// UserRegisterHandler 用户注册处理器
type UserRegisterHandler struct {
	userService  user.UserService
	emailService email.EmailService
	cacheManager CacheInterface
}

// NewUserRegisterHandler 创建用户注册处理器
func NewUserRegisterHandler(userService user.UserService, emailService email.EmailService, cacheManager CacheInterface) *UserRegisterHandler {
	return &UserRegisterHandler{
		userService:  userService,
		emailService: emailService,
		cacheManager: cacheManager,
	}
}

// createUserFromRequest 从请求创建用户对象
func (h *UserRegisterHandler) createUserFromRequest(req *RegisterRequest) (*models.User, error) {
	// 密码加密
	hashedPassword, err := h.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %s", err.Error())
	}

	// 创建用户
	user := &models.User{
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		Username:     strings.TrimSpace(req.Username),
		PasswordHash: hashedPassword,
		Status:       "active",
		StorageQuota: 10737418240, // 10GB 默认配额
		StorageUsed:  0,
	}

	if req.DisplayName != "" {
		displayName := strings.TrimSpace(req.DisplayName)
		user.DisplayName = &displayName
	}

	return user, nil
}

// sendWelcomeEmailAsync 异步发送欢迎邮件
func (h *UserRegisterHandler) sendWelcomeEmailAsync(email, username string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := h.emailService.SendWelcomeEmail(ctx, email, username); err != nil {
			// 记录邮件发送失败，但不影响注册成功
			// 可以在这里添加日志记录
			_ = err // 明确忽略错误
		}
	}()
}

// buildRegisterResponse 构建注册响应
func (h *UserRegisterHandler) buildRegisterResponse(user *models.User) RegisterResponse {
	return RegisterResponse{
		UserID:   user.ID,
		UUID:     user.UUID,
		Email:    user.Email,
		Username: user.Username,
		DisplayName: func() string {
			if user.DisplayName != nil {
				return *user.DisplayName
			}
			return ""
		}(),
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		Message:   "注册成功，欢迎使用云盘服务",
	}
}

// Register 用户注册接口
// @Summary 用户注册
// @Description 新用户通过邮箱验证注册账号
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册请求参数"
// @Success 201 {object} utils.APIResponse{data=RegisterResponse} "注册成功"
// @Failure 400 {object} utils.APIResponse{} "请求参数错误"
// @Failure 409 {object} utils.APIResponse{} "用户已存在"
// @Failure 422 {object} utils.APIResponse{} "验证失败"
// @Failure 500 {object} utils.APIResponse{} "内部服务器错误"
// @Router /api/v1/auth/register [post]
func (h *UserRegisterHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithMessage(c, utils.CodeBadRequest, "参数格式错误: "+err.Error())
		return
	}

	// 验证请求参数
	if err := h.validateRegisterRequest(&req); err != nil {
		utils.ErrorWithMessage(c, utils.CodeValidationError, "参数验证失败: "+err.Error())
		return
	}

	// 验证邮箱验证码
	if err := h.verifyEmailCode(c.Request.Context(), req.Email, req.VerificationCode, "register"); err != nil {
		utils.ErrorWithMessage(c, utils.CodeUnauthorized, "邮箱验证码错误或已过期: "+err.Error())
		return
	}

	// 检查用户是否已存在
	exists, err := h.userService.CheckUserExists(c.Request.Context(), req.Email, req.Username)
	if err != nil {
		utils.ErrorWithMessage(c, utils.CodeInternalError, "检查用户存在性失败: "+err.Error())
		return
	}
	if exists {
		utils.ErrorWithMessage(c, utils.CodeDuplicateData, "用户已存在: 邮箱或用户名已被注册")
		return
	}

	// 创建用户对象
	user, err := h.createUserFromRequest(&req)
	if err != nil {
		utils.ErrorWithMessage(c, utils.CodeInternalError, err.Error())
		return
	}

	// 保存用户
	if err := h.userService.CreateUser(c.Request.Context(), user); err != nil {
		utils.ErrorWithMessage(c, utils.CodeInternalError, "创建用户失败: "+err.Error())
		return
	}

	// 清除验证码
	h.clearEmailCode(c.Request.Context(), req.Email, "register")

	// 发送欢迎邮件
	h.sendWelcomeEmailAsync(user.Email, user.Username)

	// 返回响应
	response := h.buildRegisterResponse(user)
	utils.Created(c, response)
}

// validateSendCodeRequest 验证发送验证码请求
func (h *UserRegisterHandler) validateSendCodeRequest(req *SendVerificationCodeRequest) error {
	// 验证邮箱格式
	if !h.isValidEmail(req.Email) {
		return fmt.Errorf("邮箱格式不正确: 请输入有效的邮箱地址")
	}

	// 验证验证码类型
	if err := h.validateCodeType(req.Type); err != nil {
		return fmt.Errorf("验证码类型不正确: %s", err.Error())
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	return nil
}

// checkEmailAvailability 检查邮箱是否可用（对于注册类型）
func (h *UserRegisterHandler) checkEmailAvailability(ctx context.Context, email, codeType string) error {
	// 对于注册验证码，检查用户是否已存在
	if codeType == "register" {
		exists, err := h.userService.CheckEmailExists(ctx, email)
		if err != nil {
			return fmt.Errorf("检查邮箱失败: %s", err.Error())
		}
		if exists {
			return fmt.Errorf("邮箱已被注册: 该邮箱已被其他用户使用")
		}
	}
	return nil
}

// generateAndStoreCode 生成并存储验证码
func (h *UserRegisterHandler) generateAndStoreCode(email, codeType string) (string, time.Duration, error) {
	// 生成验证码
	code := utils.GenerateRandomCode(6)

	// 保存验证码到缓存
	cacheKey := fmt.Sprintf("email_code:%s:%s", codeType, email)
	expiresIn := 10 * time.Minute // 验证码10分钟有效期

	if err := h.cacheManager.SetWithTTL(cacheKey, code, expiresIn); err != nil {
		return "", 0, fmt.Errorf("保存验证码失败: %s", err.Error())
	}

	return code, expiresIn, nil
}

// SendVerificationCode 发送邮箱验证码
// @Summary 发送邮箱验证码
// @Description 为注册或密码重置发送邮箱验证码
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body SendVerificationCodeRequest true "发送验证码请求参数"
// @Success 200 {object} utils.APIResponse{data=SendVerificationCodeResponse} "发送成功"
// @Failure 400 {object} utils.APIResponse{} "请求参数错误"
// @Failure 429 {object} utils.APIResponse{} "请求过于频繁"
// @Failure 500 {object} utils.APIResponse{} "内部服务器错误"
// @Router /api/v1/auth/send-code [post]
func (h *UserRegisterHandler) SendVerificationCode(c *gin.Context) {
	var req SendVerificationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorWithMessage(c, utils.CodeBadRequest, "参数格式错误: "+err.Error())
		return
	}

	// 验证请求参数
	if err := h.validateSendCodeRequest(&req); err != nil {
		utils.ErrorWithMessage(c, utils.CodeBadRequest, err.Error())
		return
	}

	// 检查发送频率限制
	if err := h.checkCodeSendLimit(c.Request.Context(), req.Email, req.Type); err != nil {
		utils.ErrorWithMessage(c, utils.CodeTooManyRequests, "发送过于频繁: "+err.Error())
		return
	}

	// 检查邮箱可用性
	if err := h.checkEmailAvailability(c.Request.Context(), req.Email, req.Type); err != nil {
		utils.ErrorWithMessage(c, utils.CodeDuplicateData, err.Error())
		return
	}

	// 生成并存储验证码
	code, expiresIn, err := h.generateAndStoreCode(req.Email, req.Type)
	if err != nil {
		utils.ErrorWithMessage(c, utils.CodeInternalError, err.Error())
		return
	}

	// 发送验证码邮件
	if err := h.emailService.SendVerificationCode(c.Request.Context(), req.Email, code); err != nil {
		utils.ErrorWithMessage(c, utils.CodeInternalError, "发送验证码失败: "+err.Error())
		return
	}

	// 记录发送时间（用于频率限制）
	rateLimitKey := fmt.Sprintf("email_send_limit:%s:%s", req.Type, req.Email)
	if err := h.cacheManager.SetWithTTL(rateLimitKey, fmt.Sprintf("%d", time.Now().Unix()), 1*time.Minute); err != nil {
		// 缓存设置失败，记录错误但不影响主流程
		_ = err // 明确忽略错误
	}

	response := SendVerificationCodeResponse{
		Email:     req.Email,
		ExpiresIn: int(expiresIn.Seconds()),
		Message:   "验证码已发送，请查收邮件",
	}

	utils.SuccessWithMessage(c, "验证码发送成功", response)
}

// validateRegisterRequest 验证注册请求参数
func (h *UserRegisterHandler) validateRegisterRequest(req *RegisterRequest) error {
	// 使用新的验证工具进行批量验证
	return utils.ValidateUserRegistration(
		req.Email,
		req.Username,
		req.Password,
		req.ConfirmPassword,
		req.DisplayName,
		req.AcceptTerms,
	)
}

// validatePasswordStrength 验证密码强度
func (h *UserRegisterHandler) validatePasswordStrength(password string) error {
	// 使用新的密码验证工具
	strength, err := utils.ValidatePasswordStrength(password)
	if err != nil {
		return err
	}

	// 要求至少中等强度
	if strength < utils.PasswordMedium {
		return fmt.Errorf("密码强度不足，请使用更复杂的密码")
	}

	return nil
}

// isValidEmail 验证邮箱格式
func (h *UserRegisterHandler) isValidEmail(email string) bool {
	return utils.ValidateEmail(email) == nil
}

// hashPassword 密码加密
func (h *UserRegisterHandler) hashPassword(password string) (string, error) {
	// 使用新的密码加密工具
	return utils.HashPassword(password)
}

// verifyEmailCode 验证邮箱验证码
func (h *UserRegisterHandler) verifyEmailCode(_ context.Context, email, code, codeType string) error {
	cacheKey := fmt.Sprintf("email_code:%s:%s", codeType, email)

	var storedCode string
	err := h.cacheManager.Get(cacheKey, &storedCode)
	if err != nil {
		return fmt.Errorf("验证码已过期或不存在")
	}

	if storedCode != code {
		return fmt.Errorf("验证码不正确")
	}

	return nil
}

// clearEmailCode 清除邮箱验证码
func (h *UserRegisterHandler) clearEmailCode(_ context.Context, email, codeType string) {
	cacheKey := fmt.Sprintf("email_code:%s:%s", codeType, email)
	if err := h.cacheManager.Delete(cacheKey); err != nil {
		// 缓存删除失败，记录错误但不影响主流程
		_ = err // 明确忽略错误
	}
}

// checkCodeSendLimit 检查验证码发送频率限制
func (h *UserRegisterHandler) checkCodeSendLimit(_ context.Context, email, codeType string) error {
	rateLimitKey := fmt.Sprintf("email_send_limit:%s:%s", codeType, email)

	var value string
	err := h.cacheManager.Get(rateLimitKey, &value)
	if err == nil {
		return fmt.Errorf("验证码发送过于频繁，请1分钟后再试")
	}

	return nil
}

// validateCodeType 验证验证码类型
func (h *UserRegisterHandler) validateCodeType(codeType string) error {
	// 使用utils包中的验证函数
	return utils.ValidateCodeType(codeType)
}

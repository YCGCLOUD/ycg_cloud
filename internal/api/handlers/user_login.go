package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"cloudpan/internal/pkg/errors"
	"cloudpan/internal/pkg/utils"
	"cloudpan/internal/repository/models"
	"cloudpan/internal/service/user"
)

// LoginRequest 登录请求结构体
type LoginRequest struct {
	// 登录标识符（可以是邮箱或用户名）
	Identifier string `json:"identifier" binding:"required" example:"user@example.com"`
	// 密码
	Password string `json:"password" binding:"required" example:"password123"`
	// 登录类型：email, username, phone
	LoginType string `json:"login_type,omitempty" example:"email"`
	// 记住我
	RememberMe bool `json:"remember_me,omitempty" example:"false"`
	// 验证码（可选，用于安全验证）
	VerificationCode string `json:"verification_code,omitempty" example:"123456"`
}

// LoginResponse 登录响应结构体
type LoginResponse struct {
	// 访问令牌
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	// 刷新令牌
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	// 令牌类型
	TokenType string `json:"token_type" example:"Bearer"`
	// 过期时间（秒）
	ExpiresIn int64 `json:"expires_in" example:"86400"`
	// 用户信息
	User *UserInfo `json:"user"`
}

// UserInfo 用户信息结构体
type UserInfo struct {
	ID          uint   `json:"id" example:"1"`
	UUID        string `json:"uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Username    string `json:"username" example:"johnDoe"`
	Email       string `json:"email" example:"john@example.com"`
	DisplayName string `json:"display_name" example:"John Doe"`
	Avatar      string `json:"avatar" example:"https://example.com/avatar.jpg"`
	Status      int    `json:"status" example:"1"`
	Role        string `json:"role" example:"user"`
	CreatedAt   string `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// RefreshTokenRequest 刷新令牌请求结构体
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIs..."`
}

// UserLoginHandler 用户登录处理器
type UserLoginHandler struct {
	userService user.UserService
	jwtManager  utils.JWTManager
	logger      *zap.Logger
	secretKey   string
}

// NewUserLoginHandler 创建新的用户登录处理器
func NewUserLoginHandler(userService user.UserService, logger *zap.Logger, secretKey string) (*UserLoginHandler, error) {
	if secretKey == "" {
		return nil, errors.NewValidationError("JWT secret key", "is required")
	}

	jwtManager, err := utils.NewDefaultJWTManager(secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT manager: %w", err)
	}

	return &UserLoginHandler{
		userService: userService,
		jwtManager:  jwtManager,
		logger:      logger,
		secretKey:   secretKey,
	}, nil
}

// Login 用户登录
//
// @Summary 用户登录
// @Description 支持邮箱、用户名登录，返回JWT访问令牌和刷新令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录请求"
// @Success 200 {object} utils.Response{data=LoginResponse} "登录成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 401 {object} utils.Response "认证失败"
// @Failure 429 {object} utils.Response "请求频率限制"
// @Failure 500 {object} utils.Response "内部服务器错误"
// @Router /api/v1/login [post]
func (h *UserLoginHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	// 解析请求参数
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid login request", zap.Error(err), zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeBadRequest, "请求参数格式错误")
		return
	}

	// 验证请求参数
	if err := h.validateLoginRequest(&req); err != nil {
		h.logger.Warn("Login request validation failed",
			zap.String("identifier", req.Identifier),
			zap.String("login_type", req.LoginType),
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeValidationError, err.Error())
		return
	}

	// 根据登录类型查找用户
	user, err := h.findUserByIdentifier(ctx, req.Identifier, req.LoginType)
	if err != nil {
		h.logger.Warn("User not found during login",
			zap.String("identifier", req.Identifier),
			zap.String("login_type", req.LoginType),
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeUnauthorized, "用户名或密码错误")
		return
	}

	// 验证密码
	if !utils.VerifyPassword(user.PasswordHash, req.Password) {
		h.logger.Warn("Password verification failed",
			zap.Uint("user_id", user.ID),
			zap.String("identifier", req.Identifier),
			zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeUnauthorized, "用户名或密码错误")
		return
	}

	// 检查用户状态
	if err := h.checkUserStatus(user); err != nil {
		h.logger.Warn("User status check failed",
			zap.Uint("user_id", user.ID),
			zap.String("status", user.Status),
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeUnauthorized, err.Error())
		return
	}

	// 生成JWT令牌
	response, err := h.generateTokens(user, req.RememberMe)
	if err != nil {
		h.logger.Error("Failed to generate tokens",
			zap.Uint("user_id", user.ID),
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		utils.InternalErrorWithMessage(c, "令牌生成失败")
		return
	}

	// 记录登录成功日志
	h.logger.Info("User login successful",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
		zap.String("ip", c.ClientIP()))

	utils.SuccessWithMessage(c, "登录成功", response)
}

// RefreshToken 刷新访问令牌
//
// @Summary 刷新访问令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "刷新令牌请求"
// @Success 200 {object} utils.Response{data=LoginResponse} "刷新成功"
// @Failure 400 {object} utils.Response "请求参数错误"
// @Failure 401 {object} utils.Response "令牌无效"
// @Failure 500 {object} utils.Response "内部服务器错误"
// @Router /api/v1/refresh [post]
func (h *UserLoginHandler) RefreshToken(c *gin.Context) {
	ctx := c.Request.Context()

	// 解析请求参数
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid refresh token request", zap.Error(err), zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeBadRequest, "请求参数格式错误")
		return
	}

	if req.RefreshToken == "" {
		utils.ErrorWithMessage(c, utils.CodeBadRequest, "刷新令牌不能为空")
		return
	}

	// 刷新令牌
	newAccessToken, newRefreshToken, err := h.jwtManager.RefreshToken(req.RefreshToken)
	if err != nil {
		h.logger.Warn("Token refresh failed", zap.Error(err), zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeUnauthorized, "刷新令牌无效或已过期")
		return
	}

	// 验证新的刷新令牌并获取用户信息
	claims, err := h.jwtManager.ValidateToken(newRefreshToken)
	if err != nil {
		h.logger.Error("Failed to validate new refresh token", zap.Error(err), zap.String("ip", c.ClientIP()))
		utils.InternalErrorWithMessage(c, "令牌验证失败")
		return
	}

	// 获取用户信息
	user, err := h.userService.GetUserByID(ctx, uint(claims.UserID))
	if err != nil {
		h.logger.Error("Failed to get user info during token refresh",
			zap.Uint64("user_id", claims.UserID),
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeUnauthorized, "用户不存在")
		return
	}

	// 检查用户状态
	if err := h.checkUserStatus(user); err != nil {
		h.logger.Warn("User status check failed during token refresh",
			zap.Uint("user_id", user.ID),
			zap.String("status", user.Status),
			zap.Error(err),
			zap.String("ip", c.ClientIP()))
		utils.ErrorWithMessage(c, utils.CodeUnauthorized, err.Error())
		return
	}

	// 构建响应
	response := &LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(24 * time.Hour.Seconds()), // 24小时
		User:         h.buildUserInfo(user),
	}

	h.logger.Info("Token refresh successful",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username),
		zap.String("ip", c.ClientIP()))

	utils.SuccessWithMessage(c, "令牌刷新成功", response)
}

// validateLoginRequest 验证登录请求参数
func (h *UserLoginHandler) validateLoginRequest(req *LoginRequest) error {
	// 验证登录标识符
	if req.Identifier == "" {
		return fmt.Errorf("登录标识符不能为空")
	}

	// 验证密码
	if req.Password == "" {
		return fmt.Errorf("密码不能为空")
	}

	// 自动检测登录类型
	if req.LoginType == "" {
		req.LoginType = h.detectLoginType(req.Identifier)
	}

	// 验证登录类型
	if req.LoginType != "email" && req.LoginType != "username" {
		return fmt.Errorf("不支持的登录类型")
	}

	// 验证邮箱格式
	if req.LoginType == "email" {
		if err := utils.ValidateEmail(req.Identifier); err != nil {
			return fmt.Errorf("邮箱格式不正确: %v", err)
		}
	}

	// 验证用户名格式
	if req.LoginType == "username" {
		if err := utils.ValidateUsername(req.Identifier); err != nil {
			return fmt.Errorf("用户名格式不正确: %v", err)
		}
	}

	return nil
}

// detectLoginType 自动检测登录类型
func (h *UserLoginHandler) detectLoginType(identifier string) string {
	// 检查是否为邮箱格式
	if strings.Contains(identifier, "@") {
		return "email"
	}
	// 默认为用户名
	return "username"
}

// findUserByIdentifier 根据标识符查找用户
func (h *UserLoginHandler) findUserByIdentifier(ctx context.Context, identifier, loginType string) (*models.User, error) {
	switch loginType {
	case "email":
		return h.userService.GetUserByEmail(ctx, identifier)
	case "username":
		return h.userService.GetUserByUsername(ctx, identifier)
	default:
		return nil, fmt.Errorf("不支持的登录类型")
	}
}

// checkUserStatus 检查用户状态
func (h *UserLoginHandler) checkUserStatus(user *models.User) error {
	switch user.Status {
	case "inactive": // 已禁用
		return fmt.Errorf("用户账户已被禁用")
	case "active": // 正常
		return nil
	case "suspended": // 已暂停
		return fmt.Errorf("用户账户已被暂停，请联系客服")
	case "deleted": // 已删除
		return fmt.Errorf("用户账户不存在")
	default:
		return fmt.Errorf("用户账户状态异常")
	}
}

// generateTokens 生成JWT令牌
func (h *UserLoginHandler) generateTokens(user *models.User, rememberMe bool) (*LoginResponse, error) {
	// 生成访问令牌
	accessToken, err := h.jwtManager.GenerateAccessToken(
		uint64(user.ID),
		user.Username,
		user.Email,
		"user", // 默认角色
	)
	if err != nil {
		return nil, fmt.Errorf("生成访问令牌失败: %w", err)
	}

	// 生成刷新令牌
	refreshToken, err := h.jwtManager.GenerateRefreshToken(
		uint64(user.ID),
		user.Username,
		user.Email,
		"user",
	)
	if err != nil {
		return nil, fmt.Errorf("生成刷新令牌失败: %w", err)
	}

	// 计算过期时间
	expiresIn := int64(24 * time.Hour.Seconds()) // 24小时
	if rememberMe {
		expiresIn = int64(7 * 24 * time.Hour.Seconds()) // 7天
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		User:         h.buildUserInfo(user),
	}, nil
}

// buildUserInfo 构建用户信息
func (h *UserLoginHandler) buildUserInfo(user *models.User) *UserInfo {
	displayName := ""
	if user.DisplayName != nil {
		displayName = *user.DisplayName
	}

	avatarURL := ""
	if user.AvatarURL != nil {
		avatarURL = *user.AvatarURL
	}

	// 将字符串状态转换为整数状态
	statusInt := 1 // 默认正常
	switch user.Status {
	case "inactive":
		statusInt = 0
	case "active":
		statusInt = 1
	case "suspended":
		statusInt = 3
	case "deleted":
		statusInt = 4
	}

	return &UserInfo{
		ID:          user.ID,
		UUID:        user.UUID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: displayName,
		Avatar:      avatarURL,
		Status:      statusInt,
		Role:        "user", // 默认角色，后续可从用户模型获取
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
	}
}

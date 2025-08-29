package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"cloudpan/internal/pkg/utils"
)

// AuthMiddleware JWT认证中间件配置
type AuthMiddleware struct {
	jwtManager utils.JWTManager
	logger     *zap.Logger
}

// NewAuthMiddleware 创建新的认证中间件
func NewAuthMiddleware(secretKey string, logger *zap.Logger) (*AuthMiddleware, error) {
	jwtManager, err := utils.NewDefaultJWTManager(secretKey)
	if err != nil {
		return nil, err
	}

	return &AuthMiddleware{
		jwtManager: jwtManager,
		logger:     logger,
	}, nil
}

// RequireAuth JWT认证中间件
//
// 验证请求头中的JWT Token，如果验证成功则将用户信息存储到上下文中
// 如果验证失败则返回401错误
func (auth *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取Token
		token := auth.extractToken(c)
		if token == "" {
			auth.logger.Warn("Missing authorization token", zap.String("ip", c.ClientIP()))
			utils.ErrorWithMessage(c, utils.CodeUnauthorized, "缺少认证令牌")
			c.Abort()
			return
		}

		// 验证Token
		claims, err := auth.jwtManager.ValidateToken(token)
		if err != nil {
			auth.logger.Warn("Invalid token",
				zap.Error(err),
				zap.String("ip", c.ClientIP()))
			utils.ErrorWithMessage(c, utils.CodeUnauthorized, "令牌无效或已过期")
			c.Abort()
			return
		}

		// 检查Token类型
		if claims.TokenType != "access" {
			auth.logger.Warn("Invalid token type",
				zap.String("token_type", claims.TokenType),
				zap.String("ip", c.ClientIP()))
			utils.ErrorWithMessage(c, utils.CodeUnauthorized, "令牌类型错误")
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// OptionalAuth 可选认证中间件
//
// 如果提供了有效的Token，则将用户信息存储到上下文中
// 如果没有提供Token或Token无效，则不进行处理，允许请求继续
func (auth *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取Token
		token := auth.extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		// 验证Token
		claims, err := auth.jwtManager.ValidateToken(token)
		if err != nil {
			// Token无效，但不阻止请求
			auth.logger.Debug("Invalid optional token",
				zap.Error(err),
				zap.String("ip", c.ClientIP()))
			c.Next()
			return
		}

		// 检查Token类型
		if claims.TokenType != "access" {
			c.Next()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// RequireRole 角色验证中间件
//
// 需要先使用RequireAuth中间件进行认证
// 验证用户是否具有指定角色
func (auth *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		role, exists := c.Get("role")
		if !exists {
			auth.logger.Warn("Missing user role in context", zap.String("ip", c.ClientIP()))
			utils.ErrorWithMessage(c, utils.CodeUnauthorized, "用户认证信息缺失")
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			auth.logger.Error("Invalid role type in context", zap.String("ip", c.ClientIP()))
			utils.ErrorWithMessage(c, utils.CodeInternalError, "用户角色信息错误")
			c.Abort()
			return
		}

		// 验证角色权限
		if !auth.hasRole(userRole, requiredRole) {
			userID, _ := c.Get("user_id")
			auth.logger.Warn("Insufficient role permissions",
				zap.Any("user_id", userID),
				zap.String("user_role", userRole),
				zap.String("required_role", requiredRole),
				zap.String("ip", c.ClientIP()))
			utils.ErrorWithMessage(c, utils.CodeForbidden, "权限不足")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole 要求具有任意一个角色
func (auth *AuthMiddleware) RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		role, exists := c.Get("role")
		if !exists {
			auth.logger.Warn("Missing user role in context", zap.String("ip", c.ClientIP()))
			utils.ErrorWithMessage(c, utils.CodeUnauthorized, "用户认证信息缺失")
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			auth.logger.Error("Invalid role type in context", zap.String("ip", c.ClientIP()))
			utils.ErrorWithMessage(c, utils.CodeInternalError, "用户角色信息错误")
			c.Abort()
			return
		}

		// 检查是否具有任意一个所需角色
		hasPermission := false
		for _, requiredRole := range roles {
			if auth.hasRole(userRole, requiredRole) {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			userID, _ := c.Get("user_id")
			auth.logger.Warn("Insufficient role permissions",
				zap.Any("user_id", userID),
				zap.String("user_role", userRole),
				zap.Strings("required_roles", roles),
				zap.String("ip", c.ClientIP()))
			utils.ErrorWithMessage(c, utils.CodeForbidden, "权限不足")
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractToken 从请求头中提取Token
func (auth *AuthMiddleware) extractToken(c *gin.Context) string {
	// 从Authorization头获取Token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// 检查Bearer前缀
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return ""
	}

	// 提取Token
	token := strings.TrimPrefix(authHeader, bearerPrefix)
	token = strings.TrimSpace(token)

	return token
}

// hasRole 检查用户是否具有指定角色
func (auth *AuthMiddleware) hasRole(userRole, requiredRole string) bool {
	// 定义角色层次结构
	roleHierarchy := map[string]int{
		"user":      1,
		"moderator": 2,
		"admin":     3,
		"superuser": 4,
	}

	// 获取用户角色等级
	userLevel, userExists := roleHierarchy[userRole]
	requiredLevel, requiredExists := roleHierarchy[requiredRole]

	// 如果角色不存在，则进行精确匹配
	if !userExists || !requiredExists {
		return userRole == requiredRole
	}

	// 用户角色等级必须大于等于所需角色等级
	return userLevel >= requiredLevel
}

// GetCurrentUser 获取当前用户信息的辅助函数
func GetCurrentUser(c *gin.Context) *utils.JWTClaims {
	claims, exists := c.Get("claims")
	if !exists {
		return nil
	}

	jwtClaims, ok := claims.(*utils.JWTClaims)
	if !ok {
		return nil
	}

	return jwtClaims
}

// GetCurrentUserID 获取当前用户ID的辅助函数
func GetCurrentUserID(c *gin.Context) (uint64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint64)
	return id, ok
}

// IsAuthenticated 检查用户是否已认证
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("user_id")
	return exists
}

package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5" // #nosec G501 - 仅用于文件校验，非安全用途
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// 密码强度等级常量
const (
	PasswordWeek   = 1 // 弱密码
	PasswordMedium = 2 // 中等强度密码
	PasswordStrong = 3 // 强密码
)

// BCrypt成本因子常量
const (
	MinCost     = 4  // 最小成本（测试用）
	DefaultCost = 12 // 默认成本（推荐）
	MaxCost     = 31 // 最大成本
)

// JWT相关常量
const (
	DefaultJWTExpiry     = 24 * time.Hour     // 默认JWT过期时间（24小时）
	DefaultRefreshExpiry = 7 * 24 * time.Hour // 默认刷新令牌过期时间（7天）
	MinSecretKeyLength   = 32                 // 最小密钥长度
)

// PasswordHasher 密码哈希器接口
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, plainPassword string) bool
	ValidatePasswordStrength(password string) (int, error)
	GenerateSecurePassword(length int) (string, error)
}

// JWTClaims JWT负载结构体
type JWTClaims struct {
	UserID    uint64 `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"` // "access" 或 "refresh"
	jwt.RegisteredClaims
}

// JWTManager JWT管理器接口
type JWTManager interface {
	GenerateAccessToken(userID uint64, username, email, role string) (string, error)
	GenerateRefreshToken(userID uint64, username, email, role string) (string, error)
	ValidateToken(tokenString string) (*JWTClaims, error)
	RefreshToken(refreshToken string) (string, string, error)
}

// AESCrypto AES加密接口
type AESCrypto interface {
	Encrypt(plaintext string, key string) (string, error)
	Decrypt(ciphertext string, key string) (string, error)
	GenerateKey() (string, error)
}

// bcryptHasher BCrypt密码哈希器实现
type bcryptHasher struct {
	cost int
}

// jwtManager JWT管理器实现
type jwtManager struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// aesCrypto AES加密实现
type aesCrypto struct{}

// NewPasswordHasher 创建新的密码哈希器
func NewPasswordHasher(cost int) PasswordHasher {
	if cost < MinCost || cost > MaxCost {
		cost = DefaultCost
	}
	return &bcryptHasher{cost: cost}
}

// NewDefaultPasswordHasher 创建默认的密码哈希器
func NewDefaultPasswordHasher() PasswordHasher {
	return &bcryptHasher{cost: DefaultCost}
}

// HashPassword 使用BCrypt加密密码
func (h *bcryptHasher) HashPassword(password string) (string, error) {
	if len(password) == 0 {
		return "", fmt.Errorf("密码不能为空")
	}

	// 使用BCrypt加密密码
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("密码加密失败: %w", err)
	}

	return string(bytes), nil
}

// VerifyPassword 验证密码是否正确
func (h *bcryptHasher) VerifyPassword(hashedPassword, plainPassword string) bool {
	if len(hashedPassword) == 0 || len(plainPassword) == 0 {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

// checkPasswordComplexity 检查密码复杂度
func checkPasswordComplexity(password string) int {
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

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
	return complexityCount
}

// checkWeakPasswords 检查常见弱密码
func checkWeakPasswords(password string) error {
	weakPasswords := []string{
		"password", "123456", "123456789", "qwerty", "abc123",
		"password123", "admin", "root", "user", "test", "guest",
		"111111", "000000", "654321", "123123", "987654321",
	}

	passwordLower := strings.ToLower(password)
	for _, weak := range weakPasswords {
		if strings.Contains(passwordLower, weak) {
			return fmt.Errorf("密码过于简单，不能包含常见的弱密码模式")
		}
	}
	return nil
}

// checkPasswordPatterns 检查密码模式
func checkPasswordPatterns(password string) error {
	// 检查重复字符
	if hasRepeatedChars(password, 3) {
		return fmt.Errorf("密码不能包含3个或更多连续相同的字符")
	}

	// 检查顺序字符
	if hasSequentialChars(password, 4) {
		return fmt.Errorf("密码不能包含4个或更多连续的顺序字符")
	}
	return nil
}

// calculatePasswordStrength 计算密码强度
func calculatePasswordStrength(password string, complexityCount int) int {
	// 根据复杂度和长度判断密码强度
	if len(password) >= 12 && complexityCount >= 3 {
		return PasswordStrong
	}
	if len(password) >= 8 && complexityCount >= 2 {
		return PasswordMedium
	}
	return PasswordWeek
}

// ValidatePasswordStrength 验证密码强度
func (h *bcryptHasher) ValidatePasswordStrength(password string) (int, error) {
	if len(password) < 6 {
		return PasswordWeek, fmt.Errorf("密码长度至少6位")
	}

	if len(password) > 128 {
		return PasswordWeek, fmt.Errorf("密码长度不能超过128位")
	}

	// 检查密码复杂度
	complexityCount := checkPasswordComplexity(password)

	// 检查常见弱密码
	if err := checkWeakPasswords(password); err != nil {
		return PasswordWeek, err
	}

	// 检查密码模式
	if err := checkPasswordPatterns(password); err != nil {
		return PasswordWeek, err
	}

	// 计算密码强度
	strength := calculatePasswordStrength(password, complexityCount)
	if strength == PasswordWeek {
		return PasswordWeek, fmt.Errorf("密码强度不足，建议包含大小写字母、数字和特殊字符")
	}

	return strength, nil
}

// generatePasswordCharsets 获取密码生成字符集
func generatePasswordCharsets() (string, string, string, string) {
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits := "0123456789"
	special := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	return lowercase, uppercase, digits, special
}

// addRequiredChars 添加必需的字符类型
func addRequiredChars(password *strings.Builder, lowercase, uppercase, digits, special string) error {
	// 至少一个小写字母
	if char, err := randomChar(lowercase); err == nil {
		password.WriteByte(char)
	} else {
		return fmt.Errorf("生成小写字母失败: %w", err)
	}

	// 至少一个大写字母
	if char, err := randomChar(uppercase); err == nil {
		password.WriteByte(char)
	} else {
		return fmt.Errorf("生成大写字母失败: %w", err)
	}

	// 至少一个数字
	if char, err := randomChar(digits); err == nil {
		password.WriteByte(char)
	} else {
		return fmt.Errorf("生成数字失败: %w", err)
	}

	// 至少一个特殊字符
	if char, err := randomChar(special); err == nil {
		password.WriteByte(char)
	} else {
		return fmt.Errorf("生成特殊字符失败: %w", err)
	}

	return nil
}

// fillRemainingChars 填充剩余字符
func fillRemainingChars(password *strings.Builder, length int, allChars string) error {
	for password.Len() < length {
		if char, err := randomChar(allChars); err == nil {
			password.WriteByte(char)
		} else {
			return fmt.Errorf("生成密码失败: %w", err)
		}
	}
	return nil
}

// shufflePassword 打乱密码字符顺序
func shufflePassword(password string) (string, error) {
	result := []byte(password)
	for i := len(result) - 1; i > 0; i-- {
		if j, err := randomInt(i + 1); err == nil {
			result[i], result[j] = result[j], result[i]
		} else {
			return "", fmt.Errorf("打乱密码顺序失败: %w", err)
		}
	}
	return string(result), nil
}

// GenerateSecurePassword 生成安全的密码
func (h *bcryptHasher) GenerateSecurePassword(length int) (string, error) {
	if length < 8 {
		length = 12 // 默认12位
	}
	if length > 128 {
		length = 128 // 最大128位
	}

	// 获取字符集
	lowercase, uppercase, digits, special := generatePasswordCharsets()

	// 确保至少包含每种字符类型
	var password strings.Builder
	if err := addRequiredChars(&password, lowercase, uppercase, digits, special); err != nil {
		return "", err
	}

	// 填充剩余长度
	allChars := lowercase + uppercase + digits + special
	if err := fillRemainingChars(&password, length, allChars); err != nil {
		return "", err
	}

	// 打乱字符顺序
	return shufflePassword(password.String())
}

// 全局便利函数

// HashPassword 加密密码（使用默认哈希器）
func HashPassword(password string) (string, error) {
	hasher := NewDefaultPasswordHasher()
	return hasher.HashPassword(password)
}

// VerifyPassword 验证密码（使用默认哈希器）
func VerifyPassword(hashedPassword, plainPassword string) bool {
	hasher := NewDefaultPasswordHasher()
	return hasher.VerifyPassword(hashedPassword, plainPassword)
}

// ValidatePasswordStrength 验证密码强度（使用默认哈希器）
func ValidatePasswordStrength(password string) (int, error) {
	hasher := NewDefaultPasswordHasher()
	return hasher.ValidatePasswordStrength(password)
}

// GenerateSecurePassword 生成安全密码（使用默认哈希器）
func GenerateSecurePassword(length int) (string, error) {
	hasher := NewDefaultPasswordHasher()
	return hasher.GenerateSecurePassword(length)
}

// ComparePasswords 安全地比较两个密码是否相同（防止时序攻击）
func ComparePasswords(password1, password2 string) bool {
	return subtle.ConstantTimeCompare([]byte(password1), []byte(password2)) == 1
}

// GenerateSalt 生成随机盐值
func GenerateSalt(length int) (string, error) {
	if length <= 0 {
		length = 32
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成盐值失败: %w", err)
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// 辅助函数

// randomChar 生成指定字符集中的随机字符
func randomChar(charset string) (byte, error) {
	if len(charset) == 0 {
		return 0, fmt.Errorf("字符集不能为空")
	}

	bytes := make([]byte, 1)
	if _, err := rand.Read(bytes); err != nil {
		return 0, err
	}

	return charset[int(bytes[0])%len(charset)], nil
}

// randomInt 生成0到max-1之间的随机整数
func randomInt(max int) (int, error) {
	if max <= 0 {
		return 0, fmt.Errorf("max必须大于0")
	}

	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return 0, err
	}

	// 将字节转换为无符号整数
	num := uint32(bytes[0])<<24 | uint32(bytes[1])<<16 | uint32(bytes[2])<<8 | uint32(bytes[3])
	return int(num) % max, nil
}

// hasRepeatedChars 检查是否有重复字符
func hasRepeatedChars(password string, maxRepeat int) bool {
	if maxRepeat <= 0 {
		return false
	}

	count := 1
	for i := 1; i < len(password); i++ {
		if password[i] == password[i-1] {
			count++
			if count >= maxRepeat {
				return true
			}
		} else {
			count = 1
		}
	}
	return false
}

// hasSequentialChars 检查是否有顺序字符
func hasSequentialChars(password string, maxSequential int) bool {
	if maxSequential <= 0 || len(password) < maxSequential {
		return false
	}

	for i := 0; i <= len(password)-maxSequential; i++ {
		// 检查递增序列
		isAscending := true
		for j := 1; j < maxSequential; j++ {
			if password[i+j] != password[i+j-1]+1 {
				isAscending = false
				break
			}
		}
		if isAscending {
			return true
		}

		// 检查递减序列
		isDescending := true
		for j := 1; j < maxSequential; j++ {
			if password[i+j] != password[i+j-1]-1 {
				isDescending = false
				break
			}
		}
		if isDescending {
			return true
		}
	}
	return false
}

// ==== JWT 相关实现 ====

// NewJWTManager 创建新的JWT管理器
func NewJWTManager(secretKey string, accessExpiry, refreshExpiry time.Duration) (JWTManager, error) {
	if len(secretKey) < MinSecretKeyLength {
		return nil, fmt.Errorf("密钥长度不能小于%d个字符", MinSecretKeyLength)
	}

	if accessExpiry <= 0 {
		accessExpiry = DefaultJWTExpiry
	}
	if refreshExpiry <= 0 {
		refreshExpiry = DefaultRefreshExpiry
	}

	return &jwtManager{
		secretKey:     []byte(secretKey),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}, nil
}

// NewDefaultJWTManager 创建默认的JWT管理器
func NewDefaultJWTManager(secretKey string) (JWTManager, error) {
	return NewJWTManager(secretKey, DefaultJWTExpiry, DefaultRefreshExpiry)
}

// GenerateAccessToken 生成访问令牌
func (j *jwtManager) GenerateAccessToken(userID uint64, username, email, role string) (string, error) {
	return j.generateToken(userID, username, email, role, "access", j.accessExpiry)
}

// GenerateRefreshToken 生成刷新令牌
func (j *jwtManager) GenerateRefreshToken(userID uint64, username, email, role string) (string, error) {
	return j.generateToken(userID, username, email, role, "refresh", j.refreshExpiry)
}

// generateToken 生成令牌（内部方法）
func (j *jwtManager) generateToken(userID uint64, username, email, role, tokenType string, expiry time.Duration) (string, error) {
	now := time.Now()

	// 生成唯一的JTI
	jti, err := GenerateRandomToken(16) // 16字节的随机令牌
	if err != nil {
		return "", fmt.Errorf("生成JTI失败: %w", err)
	}

	claims := &JWTClaims{
		UserID:    userID,
		Username:  username,
		Email:     email,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti, // 添加唯一标识符
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "cloudpan",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateToken 验证令牌
func (j *jwtManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("签名算法不支持: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("令牌解析失败: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("令牌无效")
	}

	return claims, nil
}

// RefreshToken 刷新令牌
func (j *jwtManager) RefreshToken(refreshToken string) (string, string, error) {
	claims, err := j.ValidateToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("刷新令牌无效: %w", err)
	}

	if claims.TokenType != "refresh" {
		return "", "", fmt.Errorf("令牌类型错误，期望刷新令牌")
	}

	// 生成新的访问令牌和刷新令牌
	newAccessToken, err := j.GenerateAccessToken(claims.UserID, claims.Username, claims.Email, claims.Role)
	if err != nil {
		return "", "", fmt.Errorf("生成访问令牌失败: %w", err)
	}

	newRefreshToken, err := j.GenerateRefreshToken(claims.UserID, claims.Username, claims.Email, claims.Role)
	if err != nil {
		return "", "", fmt.Errorf("生成刷新令牌失败: %w", err)
	}

	return newAccessToken, newRefreshToken, nil
}

// ==== 随机字符串生成工具 ====

// GenerateVerificationCode 生成数字验证码
func GenerateVerificationCode(length int) (string, error) {
	if length <= 0 {
		length = 6 // 默认6位
	}
	if length > 10 {
		length = 10 // 最多10位
	}

	digits := "0123456789"
	result := make([]byte, length)

	for i := 0; i < length; i++ {
		char, err := randomChar(digits)
		if err != nil {
			return "", fmt.Errorf("生成验证码失败: %w", err)
		}
		result[i] = char
	}

	return string(result), nil
}

// GenerateRandomToken 生成随机令牌（使用base64编码）
func GenerateRandomToken(byteLength int) (string, error) {
	if byteLength <= 0 {
		byteLength = 32 // 默认32字节
	}

	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成随机令牌失败: %w", err)
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// ==== AES加密解密实现 ====

// NewAESCrypto 创建新的AES加密器
func NewAESCrypto() AESCrypto {
	return &aesCrypto{}
}

// GenerateKey 生成AES密钥
func (a *aesCrypto) GenerateKey() (string, error) {
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("生成AES密钥失败: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// Encrypt AES加密
func (a *aesCrypto) Encrypt(plaintext string, key string) (string, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", fmt.Errorf("密钥解码失败: %w", err)
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("创建AES密码器失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM模式失败: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("生成随机数失败: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt AES解密
func (a *aesCrypto) Decrypt(ciphertext string, key string) (string, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", fmt.Errorf("密钥解码失败: %w", err)
	}

	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("密文解码失败: %w", err)
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("创建AES密码器失败: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM模式失败: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertextBytes) < nonceSize {
		return "", fmt.Errorf("密文长度不足")
	}

	nonce, ciphertext2 := ciphertextBytes[:nonceSize], ciphertextBytes[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext2, nil)
	if err != nil {
		return "", fmt.Errorf("解密失败: %w", err)
	}

	return string(plaintext), nil
}

// ==== 哈希算法实现 ====

// MD5Hash 计算MD5哈希值
// #nosec G401 - MD5仅用于文件完整性检查，非安全关键用途
func MD5Hash(data string) string {
	hash := md5.Sum([]byte(data)) // #nosec G401 - 非安全用途
	return hex.EncodeToString(hash[:])
}

// SHA256Hash 计算SHA256哈希值
func SHA256Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SHA256HashWithSalt 计算带盐的SHA256哈希值
func SHA256HashWithSalt(data, salt string) string {
	return SHA256Hash(data + salt)
}

// ==== 全局便利函数 ====

// EncryptAES AES加密（使用默认加密器）
func EncryptAES(plaintext, key string) (string, error) {
	crypto := NewAESCrypto()
	return crypto.Encrypt(plaintext, key)
}

// DecryptAES AES解密（使用默认加密器）
func DecryptAES(ciphertext, key string) (string, error) {
	crypto := NewAESCrypto()
	return crypto.Decrypt(ciphertext, key)
}

// GenerateAESKey 生成AES密钥（使用默认加密器）
func GenerateAESKey() (string, error) {
	crypto := NewAESCrypto()
	return crypto.GenerateKey()
}

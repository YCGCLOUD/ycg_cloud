package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	assert.NotNil(t, validator)
	assert.IsType(t, &defaultValidator{}, validator)
}

// Email验证测试
func TestValidateEmail(t *testing.T) {
	validator := NewValidator()

	t.Run("有效邮箱测试", func(t *testing.T) {
		validEmails := []string{
			"test@example.com",
			"user.name@domain.co.uk",
			"firstname+lastname@company.org",
			"test123@test-domain.com",
		}

		for _, email := range validEmails {
			err := validator.ValidateEmail(email)
			assert.NoError(t, err, "邮箱 %s 应该是有效的", email)
		}
	})

	t.Run("无效邮箱测试", func(t *testing.T) {
		invalidEmails := []string{
			"",                      // 空邮箱
			"invalid",               // 无@符号
			"@domain.com",           // 缺少本地部分
			"user@",                 // 缺少域名
			"user..name@domain.com", // 连续点
			".user@domain.com",      // 以点开头
			"user.@domain.com",      // 以点结尾
			"user@.com",             // 无效域名
		}

		for _, email := range invalidEmails {
			err := validator.ValidateEmail(email)
			assert.Error(t, err, "邮箱 %s 应该是无效的", email)
		}
	})

	t.Run("邮箱长度限制测试", func(t *testing.T) {
		// 超长邮箱
		longEmail := strings.Repeat("a", 250) + "@domain.com"
		err := validator.ValidateEmail(longEmail)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "邮箱长度不能超过254个字符")

		// 超长本地部分
		longLocal := strings.Repeat("a", 65) + "@domain.com"
		err = validator.ValidateEmail(longLocal)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "用户名部分长度必须在1-64个字符之间")
	})
}

func TestValidateEmailHelperFunctions(t *testing.T) {
	t.Run("validateEmailBasicFormat", func(t *testing.T) {
		err := validateEmailBasicFormat("test@example.com")
		assert.NoError(t, err)

		err = validateEmailBasicFormat("invalid-email")
		assert.Error(t, err)
	})

	t.Run("validateEmailParts", func(t *testing.T) {
		err := validateEmailParts("test@example.com")
		assert.NoError(t, err)

		err = validateEmailParts("test@@example.com")
		assert.Error(t, err)
	})

	t.Run("validateEmailSpecialChars", func(t *testing.T) {
		err := validateEmailSpecialChars("test@example.com")
		assert.NoError(t, err)

		err = validateEmailSpecialChars("te..st@example.com")
		assert.Error(t, err)
	})

	t.Run("validateEmailDomain", func(t *testing.T) {
		err := validateEmailDomain("test@example.com")
		assert.NoError(t, err)

		err = validateEmailDomain("test@-invalid.com")
		assert.Error(t, err)
	})
}

// Username验证测试
func TestValidateUsername(t *testing.T) {
	validator := NewValidator()

	t.Run("有效用户名测试", func(t *testing.T) {
		validUsernames := []string{
			"testuser",
			"test_user",
			"test-user",
			"user123",
			"abc",
		}

		for _, username := range validUsernames {
			err := validator.ValidateUsername(username)
			assert.NoError(t, err, "用户名 %s 应该是有效的", username)
		}
	})

	t.Run("无效用户名测试", func(t *testing.T) {
		invalidUsernames := []string{
			"",           // 空用户名
			"ab",         // 太短
			"123user",    // 以数字开头
			"-user",      // 以连字符开头
			"_user",      // 以下划线开头
			"user-",      // 以连字符结尾
			"user_",      // 以下划线结尾
			"user--name", // 连续连字符
			"user__name", // 连续下划线
			"admin",      // 保留名称
			"user@name",  // 包含非法字符
		}

		for _, username := range invalidUsernames {
			err := validator.ValidateUsername(username)
			assert.Error(t, err, "用户名 %s 应该是无效的", username)
		}
	})

	t.Run("用户名长度测试", func(t *testing.T) {
		// 超长用户名
		longUsername := strings.Repeat("a", 51)
		err := validator.ValidateUsername(longUsername)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "长度不能超过")
	})
}

func TestValidateUsernameHelperFunctions(t *testing.T) {
	t.Run("getReservedUsernames", func(t *testing.T) {
		reserved := getReservedUsernames()
		assert.NotEmpty(t, reserved)
		assert.Contains(t, reserved, "admin")
		assert.Contains(t, reserved, "root")
	})

	t.Run("validateUsernameFormat", func(t *testing.T) {
		err := validateUsernameFormat("validusername")
		assert.NoError(t, err)

		err = validateUsernameFormat("invalid@username")
		assert.Error(t, err)
	})

	t.Run("validateUsernameStartEnd", func(t *testing.T) {
		err := validateUsernameStartEnd("validusername")
		assert.NoError(t, err)

		err = validateUsernameStartEnd("123invalid")
		assert.Error(t, err)

		err = validateUsernameStartEnd("-invalid")
		assert.Error(t, err)
	})

	t.Run("validateUsernameConsecutiveChars", func(t *testing.T) {
		err := validateUsernameConsecutiveChars("validusername")
		assert.NoError(t, err)

		err = validateUsernameConsecutiveChars("invalid--name")
		assert.Error(t, err)
	})

	t.Run("validateUsernameReserved", func(t *testing.T) {
		err := validateUsernameReserved("validusername")
		assert.NoError(t, err)

		err = validateUsernameReserved("admin")
		assert.Error(t, err)
	})
}

// DisplayName验证测试
func TestValidateDisplayName(t *testing.T) {
	validator := NewValidator()

	t.Run("有效显示名称测试", func(t *testing.T) {
		validNames := []string{
			"",               // 空名称（允许）
			"张三",             // 中文名
			"John Doe",       // 英文名
			"User 123",       // 包含数字
			"Test-User_Name", // 包含特殊字符
		}

		for _, name := range validNames {
			err := validator.ValidateDisplayName(name)
			assert.NoError(t, err, "显示名称 %s 应该是有效的", name)
		}
	})

	t.Run("无效显示名称测试", func(t *testing.T) {
		invalidNames := []string{
			"   ",                      // 全空白字符
			string([]byte{0x01, 0x02}), // 控制字符
			strings.Repeat("a", 101),   // 超长名称
		}

		for _, name := range invalidNames {
			err := validator.ValidateDisplayName(name)
			assert.Error(t, err, "显示名称应该是无效的")
		}
	})
}

// 通用验证函数测试
func TestValidateRequired(t *testing.T) {
	validator := NewValidator()

	t.Run("有效必填字段测试", func(t *testing.T) {
		err := validator.ValidateRequired("test", "测试字段")
		assert.NoError(t, err)
	})

	t.Run("无效必填字段测试", func(t *testing.T) {
		invalidValues := []string{"", "   ", "\t\n"}

		for _, value := range invalidValues {
			err := validator.ValidateRequired(value, "测试字段")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "测试字段不能为空")
		}
	})
}

func TestValidateLength(t *testing.T) {
	validator := NewValidator()

	t.Run("长度验证测试", func(t *testing.T) {
		// 正常长度
		err := validator.ValidateLength("test", 2, 10, "测试字段")
		assert.NoError(t, err)

		// 太短
		err = validator.ValidateLength("a", 2, 10, "测试字段")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "长度不能少于")

		// 太长
		err = validator.ValidateLength("verylongstring", 2, 5, "测试字段")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "长度不能超过")
	})

	t.Run("UTF-8字符长度测试", func(t *testing.T) {
		// 中文字符
		err := validator.ValidateLength("你好世界", 2, 10, "测试字段")
		assert.NoError(t, err)

		// 中文字符太长
		err = validator.ValidateLength("你好世界测试长度", 2, 5, "测试字段")
		assert.Error(t, err)
	})

	t.Run("边界值测试", func(t *testing.T) {
		// 最小长度为0
		err := validator.ValidateLength("", 0, 5, "测试字段")
		assert.NoError(t, err)

		// 最大长度为0（忽略最大长度）
		err = validator.ValidateLength("test", 2, 0, "测试字段")
		assert.NoError(t, err)
	})
}

func TestValidatePattern(t *testing.T) {
	validator := NewValidator()

	t.Run("正则模式验证测试", func(t *testing.T) {
		// 匹配数字
		err := validator.ValidatePattern("123", `^[0-9]+$`, "数字字段")
		assert.NoError(t, err)

		// 不匹配数字
		err = validator.ValidatePattern("abc", `^[0-9]+$`, "数字字段")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "格式不正确")
	})

	t.Run("无效正则表达式测试", func(t *testing.T) {
		err := validator.ValidatePattern("test", `[`, "测试字段")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "验证测试字段时发生错误")
	})
}

// 特殊验证函数测试
func TestValidatePhoneNumber(t *testing.T) {
	t.Run("有效手机号测试", func(t *testing.T) {
		validPhones := []string{
			"",               // 空号码（允许）
			"13800138000",    // 中国手机号
			"+8613800138000", // 带国际区号
			"1234567890",     // 10位号码
			"123-456-7890",   // 带分隔符
			"(123) 456-7890", // 带括号
		}

		for _, phone := range validPhones {
			err := ValidatePhoneNumber(phone)
			assert.NoError(t, err, "手机号 %s 应该是有效的", phone)
		}
	})

	t.Run("无效手机号测试", func(t *testing.T) {
		invalidPhones := []string{
			"123abc",            // 包含字母
			"123456",            // 太短
			"12345678901234567", // 太长
			"abc-def-ghij",      // 全是字母
		}

		for _, phone := range invalidPhones {
			err := ValidatePhoneNumber(phone)
			assert.Error(t, err, "手机号 %s 应该是无效的", phone)
		}
	})
}

func TestValidateURL(t *testing.T) {
	t.Run("有效URL测试", func(t *testing.T) {
		validURLs := []string{
			"http://example.com",
			"https://www.example.com",
			"https://sub.domain.com:8080",
			"http://example.com/path",
			"https://example.com/path?query=value",
		}

		for _, url := range validURLs {
			err := ValidateURL(url)
			assert.NoError(t, err, "URL %s 应该是有效的", url)
		}
	})

	t.Run("无效URL测试", func(t *testing.T) {
		invalidURLs := []string{
			"",                  // 空URL
			"not-a-url",         // 不是URL
			"ftp://example.com", // 不支持的协议
			"http://",           // 缺少域名
			"https://.com",      // 无效域名
		}

		for _, url := range invalidURLs {
			err := ValidateURL(url)
			assert.Error(t, err, "URL %s 应该是无效的", url)
		}
	})
}

func TestValidateAge(t *testing.T) {
	t.Run("有效年龄测试", func(t *testing.T) {
		validAges := []int{0, 1, 18, 65, 100, 150}

		for _, age := range validAges {
			err := ValidateAge(age)
			assert.NoError(t, err, "年龄 %d 应该是有效的", age)
		}
	})

	t.Run("无效年龄测试", func(t *testing.T) {
		invalidAges := []int{-1, -10, 151, 200}

		for _, age := range invalidAges {
			err := ValidateAge(age)
			assert.Error(t, err, "年龄 %d 应该是无效的", age)
		}
	})
}

func TestValidateConfirmPassword(t *testing.T) {
	t.Run("密码匹配测试", func(t *testing.T) {
		err := ValidateConfirmPassword("password123", "password123")
		assert.NoError(t, err)
	})

	t.Run("密码不匹配测试", func(t *testing.T) {
		err := ValidateConfirmPassword("password123", "password456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "密码和确认密码不一致")
	})
}

func TestValidateAcceptTerms(t *testing.T) {
	t.Run("接受条款测试", func(t *testing.T) {
		err := ValidateAcceptTerms(true)
		assert.NoError(t, err)
	})

	t.Run("未接受条款测试", func(t *testing.T) {
		err := ValidateAcceptTerms(false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "必须接受服务条款")
	})
}

// 验证码相关测试
func TestValidateVerificationCode(t *testing.T) {
	t.Run("有效验证码测试", func(t *testing.T) {
		validCodes := []string{
			"123456",
			"000000",
			"999999",
		}

		for _, code := range validCodes {
			err := ValidateVerificationCode(code)
			assert.NoError(t, err, "验证码 %s 应该是有效的", code)
		}
	})

	t.Run("无效验证码测试", func(t *testing.T) {
		invalidCodes := []string{
			"",        // 空验证码
			"12345",   // 太短
			"1234567", // 太长
			"abcdef",  // 包含字母
			"12345a",  // 包含字母
		}

		for _, code := range invalidCodes {
			err := ValidateVerificationCode(code)
			assert.Error(t, err, "验证码 %s 应该是无效的", code)
		}
	})
}

func TestValidateCodeType(t *testing.T) {
	t.Run("有效验证码类型测试", func(t *testing.T) {
		validTypes := []string{
			"register",
			"password_reset",
			"login",
			"change_email",
		}

		for _, codeType := range validTypes {
			err := ValidateCodeType(codeType)
			assert.NoError(t, err, "验证码类型 %s 应该是有效的", codeType)
		}
	})

	t.Run("无效验证码类型测试", func(t *testing.T) {
		invalidTypes := []string{
			"",
			"invalid",
			"reg",
			"password",
			"unknown_type",
		}

		for _, codeType := range invalidTypes {
			err := ValidateCodeType(codeType)
			assert.Error(t, err, "验证码类型 %s 应该是无效的", codeType)
		}
	})
}

// 批量验证测试
func TestValidateUserRegistration(t *testing.T) {
	t.Run("有效注册数据测试", func(t *testing.T) {
		err := ValidateUserRegistration(
			"test@example.com",
			"testuser",
			"MySecure#Pass789!",
			"MySecure#Pass789!",
			"Test User",
			true,
		)
		assert.NoError(t, err)
	})

	t.Run("无效邮箱测试", func(t *testing.T) {
		err := ValidateUserRegistration(
			"invalid-email",
			"testuser",
			"StrongPassword123!",
			"StrongPassword123!",
			"Test User",
			true,
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "邮箱验证失败")
	})

	t.Run("无效用户名测试", func(t *testing.T) {
		err := ValidateUserRegistration(
			"test@example.com",
			"admin",
			"StrongPassword123!",
			"StrongPassword123!",
			"Test User",
			true,
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "用户名验证失败")
	})

	t.Run("弱密码测试", func(t *testing.T) {
		err := ValidateUserRegistration(
			"test@example.com",
			"testuser",
			"123456",
			"123456",
			"Test User",
			true,
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "密码验证失败")
	})

	t.Run("密码不匹配测试", func(t *testing.T) {
		err := ValidateUserRegistration(
			"test@example.com",
			"testuser",
			"MySecure#Pass789!",
			"DifferentPass#567!",
			"Test User",
			true,
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "确认密码验证失败")
	})

	t.Run("无效显示名称测试", func(t *testing.T) {
		err := ValidateUserRegistration(
			"test@example.com",
			"testuser",
			"MySecure#Pass789!",
			"MySecure#Pass789!",
			strings.Repeat("a", 101),
			true,
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "显示名称验证失败")
	})

	t.Run("未接受条款测试", func(t *testing.T) {
		err := ValidateUserRegistration(
			"test@example.com",
			"testuser",
			"MySecure#Pass789!",
			"MySecure#Pass789!",
			"Test User",
			false,
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "服务条款验证失败")
	})
}

// 辅助函数测试
func TestIsAlpha(t *testing.T) {
	t.Run("纯字母字符串测试", func(t *testing.T) {
		alphaStrings := []string{
			"abc",
			"ABC",
			"AbC",
			"你好",
		}

		for _, s := range alphaStrings {
			result := IsAlpha(s)
			assert.True(t, result, "字符串 %s 应该只包含字母", s)
		}
	})

	t.Run("非纯字母字符串测试", func(t *testing.T) {
		nonAlphaStrings := []string{
			"",
			"abc123",
			"abc!",
			"123",
			"abc ",
		}

		for _, s := range nonAlphaStrings {
			result := IsAlpha(s)
			assert.False(t, result, "字符串 %s 不应该只包含字母", s)
		}
	})
}

func TestIsAlphanumeric(t *testing.T) {
	t.Run("字母数字字符串测试", func(t *testing.T) {
		alphanumericStrings := []string{
			"abc",
			"123",
			"abc123",
			"ABC123",
		}

		for _, s := range alphanumericStrings {
			result := IsAlphanumeric(s)
			assert.True(t, result, "字符串 %s 应该只包含字母和数字", s)
		}
	})

	t.Run("非字母数字字符串测试", func(t *testing.T) {
		nonAlphanumericStrings := []string{
			"",
			"abc!",
			"abc 123",
			"abc-123",
		}

		for _, s := range nonAlphanumericStrings {
			result := IsAlphanumeric(s)
			assert.False(t, result, "字符串 %s 不应该只包含字母和数字", s)
		}
	})
}

func TestIsNumeric(t *testing.T) {
	t.Run("纯数字字符串测试", func(t *testing.T) {
		numericStrings := []string{
			"123",
			"0",
			"999",
		}

		for _, s := range numericStrings {
			result := IsNumeric(s)
			assert.True(t, result, "字符串 %s 应该只包含数字", s)
		}
	})

	t.Run("非纯数字字符串测试", func(t *testing.T) {
		nonNumericStrings := []string{
			"",
			"abc",
			"123abc",
			"12.3",
			"12 3",
		}

		for _, s := range nonNumericStrings {
			result := IsNumeric(s)
			assert.False(t, result, "字符串 %s 不应该只包含数字", s)
		}
	})
}

func TestContainsWhitespace(t *testing.T) {
	t.Run("包含空白字符的字符串测试", func(t *testing.T) {
		whitespaceStrings := []string{
			"hello world",
			" hello",
			"hello ",
			"hello\tworld",
			"hello\nworld",
		}

		for _, s := range whitespaceStrings {
			result := ContainsWhitespace(s)
			assert.True(t, result, "字符串 %s 应该包含空白字符", s)
		}
	})

	t.Run("不包含空白字符的字符串测试", func(t *testing.T) {
		noWhitespaceStrings := []string{
			"",
			"hello",
			"helloworld",
			"123",
			"abc123!@#",
		}

		for _, s := range noWhitespaceStrings {
			result := ContainsWhitespace(s)
			assert.False(t, result, "字符串 %s 不应该包含空白字符", s)
		}
	})
}

// 全局便利函数测试
func TestGlobalValidatorFunctions(t *testing.T) {
	t.Run("全局ValidateEmail函数", func(t *testing.T) {
		err := ValidateEmail("test@example.com")
		assert.NoError(t, err)

		err = ValidateEmail("invalid-email")
		assert.Error(t, err)
	})

	t.Run("全局ValidateUsername函数", func(t *testing.T) {
		err := ValidateUsername("testuser")
		assert.NoError(t, err)

		err = ValidateUsername("admin")
		assert.Error(t, err)
	})

	t.Run("全局ValidateDisplayName函数", func(t *testing.T) {
		err := ValidateDisplayName("Test User")
		assert.NoError(t, err)

		err = ValidateDisplayName(strings.Repeat("a", 101))
		assert.Error(t, err)
	})

	t.Run("全局ValidateRequired函数", func(t *testing.T) {
		err := ValidateRequired("test", "字段名")
		assert.NoError(t, err)

		err = ValidateRequired("", "字段名")
		assert.Error(t, err)
	})

	t.Run("全局ValidateLength函数", func(t *testing.T) {
		err := ValidateLength("test", 2, 10, "字段名")
		assert.NoError(t, err)

		err = ValidateLength("a", 2, 10, "字段名")
		assert.Error(t, err)
	})

	t.Run("全局ValidatePattern函数", func(t *testing.T) {
		err := ValidatePattern("123", `^[0-9]+$`, "字段名")
		assert.NoError(t, err)

		err = ValidatePattern("abc", `^[0-9]+$`, "字段名")
		assert.Error(t, err)
	})
}

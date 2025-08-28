package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomString(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		charset string
		wantErr bool
	}{
		{"valid alphanumeric", 10, Alphanumeric, false},
		{"valid digits", 5, Digits, false},
		{"valid letters", 8, Letters, false},
		{"zero length", 0, Alphanumeric, true},
		{"negative length", -1, Alphanumeric, true},
		{"empty charset", 5, "", false}, // 使用默认字符集
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateRandomString(tt.length, tt.charset)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				if tt.length > 0 {
					assert.Len(t, result, tt.length)
				}
			}
		})
	}
}

func TestGenerateAlphanumeric(t *testing.T) {
	result, err := GenerateAlphanumeric(10)
	assert.NoError(t, err)
	assert.Len(t, result, 10)

	// 验证只包含字母数字
	for _, char := range result {
		assert.True(t, strings.ContainsRune(Alphanumeric, char))
	}
}

func TestGenerateNumeric(t *testing.T) {
	result, err := GenerateNumeric(6)
	assert.NoError(t, err)
	assert.Len(t, result, 6)

	// 验证只包含数字
	for _, char := range result {
		assert.True(t, strings.ContainsRune(Digits, char))
	}
}

func TestGenerateHex(t *testing.T) {
	result, err := GenerateHex(8)
	assert.NoError(t, err)
	assert.Len(t, result, 8)

	// 验证只包含十六进制字符
	for _, char := range result {
		assert.True(t, strings.ContainsRune(HexChars, char))
	}
}

func TestGenerateSecureToken(t *testing.T) {
	result, err := GenerateSecureToken(16)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)

	// 生成两次应该不同
	result2, err := GenerateSecureToken(16)
	assert.NoError(t, err)
	assert.NotEqual(t, result, result2)
}

func TestGenerateUUID(t *testing.T) {
	result, err := GenerateUUID()
	assert.NoError(t, err)
	assert.Len(t, result, 36) // UUID标准长度
	assert.Contains(t, result, "-")

	// 生成两次应该不同
	result2, err := GenerateUUID()
	assert.NoError(t, err)
	assert.NotEqual(t, result, result2)
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", true},
		{"   ", true},
		{"\t\n", true},
		{"hello", false},
		{" hello ", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsEmpty(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsNotEmpty(t *testing.T) {
	assert.True(t, IsNotEmpty("hello"))
	assert.False(t, IsNotEmpty(""))
	assert.False(t, IsNotEmpty("   "))
}

func TestTrimAndLower(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{" HELLO ", "hello"},
		{"World", "world"},
		{"", ""},
		{"  MiXeD CaSe  ", "mixed case"},
	}

	for _, tt := range tests {
		result := TrimAndLower(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestTrimAndUpper(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{" hello ", "HELLO"},
		{"world", "WORLD"},
		{"", ""},
		{"  MiXeD CaSe  ", "MIXED CASE"},
	}

	for _, tt := range tests {
		result := TrimAndUpper(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"hello world", 5, "hello"},
		{"hello", 10, "hello"},
		{"", 5, ""},
		{"hello", 0, ""},
		{"hello", -1, ""},
		{"你好世界", 2, "你好"},
	}

	for _, tt := range tests {
		result := Truncate(tt.input, tt.maxLen)
		assert.Equal(t, tt.expected, result)
	}
}

func TestTruncateWithEllipsis(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"hello world", 8, "hello..."},
		{"hello", 10, "hello"},
		{"hello world", 3, "hel"},
		{"hello world", 0, ""},
		{"你好世界测试", 5, "你好..."},
	}

	for _, tt := range tests {
		result := TruncateWithEllipsis(tt.input, tt.maxLen)
		assert.Equal(t, tt.expected, result)
	}
}

func TestPadLeft(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		padChar  rune
		expected string
	}{
		{"123", 6, '0', "000123"},
		{"hello", 3, ' ', "hello"},
		{"", 3, 'x', "xxx"},
	}

	for _, tt := range tests {
		result := PadLeft(tt.input, tt.length, tt.padChar)
		assert.Equal(t, tt.expected, result)
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		padChar  rune
		expected string
	}{
		{"123", 6, '0', "123000"},
		{"hello", 3, ' ', "hello"},
		{"", 3, 'x', "xxx"},
	}

	for _, tt := range tests {
		result := PadRight(tt.input, tt.length, tt.padChar)
		assert.Equal(t, tt.expected, result)
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "olleh"},
		{"", ""},
		{"a", "a"},
		{"你好", "好你"},
		{"123", "321"},
	}

	for _, tt := range tests {
		result := Reverse(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"HelloWorld", "hello_world"},
		{"XMLHttpRequest", "xmlhttp_request"},
		{"userId", "user_id"},
		{"ID", "id"},
		{"", ""},
		{"hello", "hello"},
	}

	for _, tt := range tests {
		result := ToSnakeCase(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello_world", "helloWorld"},
		{"user-id", "userId"},
		{"first name", "firstName"},
		{"", ""},
		{"hello", "hello"},
		{"HELLO_WORLD", "helloWorld"},
	}

	for _, tt := range tests {
		result := ToCamelCase(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello_world", "HelloWorld"},
		{"user-id", "UserId"},
		{"first name", "FirstName"},
		{"", ""},
		{"hello", "Hello"},
	}

	for _, tt := range tests {
		result := ToPascalCase(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"Hello World", "WORLD", true},
		{"Hello World", "xyz", false},
		{"", "test", false},
		{"test", "", true},
	}

	for _, tt := range tests {
		result := ContainsIgnoreCase(tt.s, tt.substr)
		assert.Equal(t, tt.want, result)
	}
}

func TestStartsWithIgnoreCase(t *testing.T) {
	assert.True(t, StartsWithIgnoreCase("Hello World", "HELLO"))
	assert.False(t, StartsWithIgnoreCase("Hello World", "WORLD"))
}

func TestEndsWithIgnoreCase(t *testing.T) {
	assert.True(t, EndsWithIgnoreCase("Hello World", "WORLD"))
	assert.False(t, EndsWithIgnoreCase("Hello World", "HELLO"))
}

func TestRemoveNonAlphanumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello-world_123!", "helloworld123"},
		{"用户123", "123"}, // 中文字符被移除
		{"", ""},
		{"abc123", "abc123"},
	}

	for _, tt := range tests {
		result := RemoveNonAlphanumeric(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello<world>.txt", "hello_world_.txt"},
		{"file|name?.txt", "file_name_.txt"},
		{"", "unnamed"},
		{"...", "unnamed"},
		{". test .", "test"},
	}

	for _, tt := range tests {
		result := SanitizeFilename(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"test@example.com", true},
		{"user.name+tag@domain.co.uk", true},
		{"invalid-email", false},
		{"@domain.com", false},
		{"user@", false},
		{"", false},
	}

	for _, tt := range tests {
		result := IsValidEmail(tt.email)
		assert.Equal(t, tt.valid, result, "Email: %s", tt.email)
	}
}

func TestIsValidUsername(t *testing.T) {
	tests := []struct {
		username string
		valid    bool
	}{
		{"user123", true},
		{"user_name", true},
		{"user-name", true},
		{"ab", false},                    // too short
		{"user@name", false},             // invalid char
		{"", false},                      // empty
		{strings.Repeat("a", 33), false}, // too long
	}

	for _, tt := range tests {
		result := IsValidUsername(tt.username)
		assert.Equal(t, tt.valid, result, "Username: %s", tt.username)
	}
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		email    string
		expected string
	}{
		{"test@example.com", "t**t@example.com"},
		{"a@example.com", "*@example.com"},
		{"ab@example.com", "**@example.com"},
		{"invalid-email", "invalid-email"}, // 无效邮箱不处理
	}

	for _, tt := range tests {
		result := MaskEmail(tt.email)
		assert.Equal(t, tt.expected, result)
	}
}

func TestMaskPhone(t *testing.T) {
	tests := []struct {
		phone    string
		expected string
	}{
		{"13800138000", "138****8000"},
		{"1234567", "1****67"},
		{"12345", "12345"}, // 太短不处理
	}

	for _, tt := range tests {
		result := MaskPhone(tt.phone)
		assert.Equal(t, tt.expected, result)
	}
}

func TestStringToInt(t *testing.T) {
	tests := []struct {
		input        string
		defaultValue int
		expected     int
	}{
		{"123", 0, 123},
		{"invalid", 999, 999},
		{"", 0, 0},
		{" 456 ", 0, 456},
	}

	for _, tt := range tests {
		result := StringToInt(tt.input, tt.defaultValue)
		assert.Equal(t, tt.expected, result)
	}
}

func TestStringToInt64(t *testing.T) {
	result := StringToInt64("123", 0)
	assert.Equal(t, int64(123), result)

	result = StringToInt64("invalid", 999)
	assert.Equal(t, int64(999), result)
}

func TestStringToFloat64(t *testing.T) {
	result := StringToFloat64("123.45", 0)
	assert.Equal(t, 123.45, result)

	result = StringToFloat64("invalid", 999.0)
	assert.Equal(t, 999.0, result)
}

func TestStringToBool(t *testing.T) {
	tests := []struct {
		input        string
		defaultValue bool
		expected     bool
	}{
		{"true", false, true},
		{"1", false, true},
		{"yes", false, true},
		{"false", true, false},
		{"0", true, false},
		{"no", true, false},
		{"invalid", true, true},
		{"invalid", false, false},
	}

	for _, tt := range tests {
		result := StringToBool(tt.input, tt.defaultValue)
		assert.Equal(t, tt.expected, result)
	}
}

func TestJoinNonEmpty(t *testing.T) {
	result := JoinNonEmpty(", ", "hello", "", "world", "")
	assert.Equal(t, "hello, world", result)

	result = JoinNonEmpty("|", "", "", "")
	assert.Equal(t, "", result)
}

func TestSplitAndTrim(t *testing.T) {
	tests := []struct {
		input    string
		sep      string
		expected []string
	}{
		{"a, b, c", ",", []string{"a", "b", "c"}},
		{" a , , b ", ",", []string{"a", "b"}},
		{"", ",", []string{}},
		{"single", ",", []string{"single"}},
	}

	for _, tt := range tests {
		result := SplitAndTrim(tt.input, tt.sep)
		assert.Equal(t, tt.expected, result)
	}
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<script>alert('xss')</script>", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
		{"Hello & World", "Hello &amp; World"},
		{"", ""},
	}

	for _, tt := range tests {
		result := EscapeHTML(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestUnescapeHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"&lt;script&gt;", "<script>"},
		{"Hello &amp; World", "Hello & World"},
		{"", ""},
	}

	for _, tt := range tests {
		result := UnescapeHTML(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestToHex(t *testing.T) {
	result := ToHex("hello")
	assert.Equal(t, "68656c6c6f", result)

	result = ToHex("")
	assert.Equal(t, "", result)
}

func TestFromHex(t *testing.T) {
	result, err := FromHex("68656c6c6f")
	assert.NoError(t, err)
	assert.Equal(t, "hello", result)

	_, err = FromHex("invalid")
	assert.Error(t, err)
}

func TestToBase64(t *testing.T) {
	result := ToBase64("hello")
	assert.Equal(t, "aGVsbG8=", result)
}

func TestFromBase64(t *testing.T) {
	result, err := FromBase64("aGVsbG8=")
	assert.NoError(t, err)
	assert.Equal(t, "hello", result)

	_, err = FromBase64("invalid!")
	assert.Error(t, err)
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "abcdefghijklmnopqrstuvwxyz", LettersLowercase)
	assert.Equal(t, "ABCDEFGHIJKLMNOPQRSTUVWXYZ", LettersUppercase)
	assert.Equal(t, LettersLowercase+LettersUppercase, Letters)
	assert.Equal(t, "0123456789", Digits)
	assert.Contains(t, SpecialChars, "!")
	assert.Contains(t, SpecialChars, "@")
	assert.Equal(t, "0123456789abcdef", HexChars)
}

// Benchmark tests
func BenchmarkGenerateAlphanumeric(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateAlphanumeric(10)
	}
}

func BenchmarkToSnakeCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ToSnakeCase("HelloWorldTestString")
	}
}

func BenchmarkToCamelCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ToCamelCase("hello_world_test_string")
	}
}

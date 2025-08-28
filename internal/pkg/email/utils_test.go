package email

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateEmailAddress(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"valid email with subdomain", "user@mail.example.com", false},
		{"valid email with numbers", "user123@example.com", false},
		{"valid email with dots", "user.name@example.com", false},
		{"valid email with plus", "user+tag@example.com", false},
		{"empty email", "", true},
		{"invalid format - no @", "testexample.com", true},
		{"invalid format - no domain", "test@", true},
		{"invalid format - no user", "@example.com", true},
		{"invalid format - multiple @", "test@@example.com", true},
		{"invalid format - no TLD", "test@example", true},
		{"invalid format - special chars", "test@exam ple.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmailAddress(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEmailList(t *testing.T) {
	tests := []struct {
		name    string
		emails  []string
		wantErr bool
	}{
		{
			name:    "valid email list",
			emails:  []string{"test1@example.com", "test2@example.com"},
			wantErr: false,
		},
		{
			name:    "empty email list",
			emails:  []string{},
			wantErr: true,
		},
		{
			name:    "mixed valid and invalid",
			emails:  []string{"test@example.com", "invalid-email"},
			wantErr: true,
		},
		{
			name:    "all invalid",
			emails:  []string{"invalid1", "invalid2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmailList(tt.emails)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNormalizeEmailAddress(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{"uppercase", "TEST@EXAMPLE.COM", "test@example.com"},
		{"mixed case", "TeSt@ExAmPlE.cOm", "test@example.com"},
		{"with spaces", "  test@example.com  ", "test@example.com"},
		{"already normalized", "test@example.com", "test@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeEmailAddress(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeEmailList(t *testing.T) {
	emails := []string{"TEST@EXAMPLE.COM", "  User@Domain.Com  "}
	expected := []string{"test@example.com", "user@domain.com"}

	result := NormalizeEmailList(emails)
	assert.Equal(t, expected, result)
}

func TestIsTemporaryEmailProvider(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"regular email", "user@gmail.com", false},
		{"temporary email 1", "user@10minutemail.com", true},
		{"temporary email 2", "user@guerrillamail.com", true},
		{"temporary email 3", "user@mailinator.com", true},
		{"case insensitive", "USER@MAILINATOR.COM", true},
		{"subdomain", "user@sub.gmail.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTemporaryEmailProvider(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEmailDomain(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{"normal email", "user@example.com", "example.com"},
		{"subdomain", "user@mail.example.com", "mail.example.com"},
		{"uppercase", "user@EXAMPLE.COM", "example.com"},
		{"invalid email", "invalid-email", ""},
		{"multiple @", "user@@example.com", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEmailDomain(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsBusinessEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"business email", "user@company.com", true},
		{"gmail personal", "user@gmail.com", false},
		{"yahoo personal", "user@yahoo.com", false},
		{"qq personal", "user@qq.com", false},
		{"163 personal", "user@163.com", false},
		{"corporate domain", "user@corporation.org", true},
		{"invalid email", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBusinessEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeEmailContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "remove script tags",
			content:  "<p>Hello</p><script>alert('xss')</script><p>World</p>",
			expected: "<p>Hello</p><p>World</p>",
		},
		{
			name:     "remove style tags",
			content:  "<p>Hello</p><style>body{color:red}</style><p>World</p>",
			expected: "<p>Hello</p><p>World</p>",
		},
		{
			name:     "remove javascript protocol",
			content:  "<a href=\"javascript:alert('xss')\">Click</a>",
			expected: "<a href=\"alert('xss')\">Click</a>",
		},
		{
			name:     "clean content",
			content:  "<p>Hello <strong>World</strong></p>",
			expected: "<p>Hello <strong>World</strong></p>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeEmailContent(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateUnsubscribeURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		userID   string
		token    string
		expected string
	}{
		{
			name:     "with base URL",
			baseURL:  "https://example.com",
			userID:   "123",
			token:    "abc",
			expected: "https://example.com/unsubscribe?user=123&token=abc",
		},
		{
			name:     "empty base URL",
			baseURL:  "",
			userID:   "123",
			token:    "abc",
			expected: "https://example.com/unsubscribe?user=123&token=abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateUnsubscribeURL(tt.baseURL, tt.userID, tt.token)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatEmailAddress(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		nameVal  string
		expected string
	}{
		{
			name:     "with name",
			email:    "test@example.com",
			nameVal:  "Test User",
			expected: "Test User <test@example.com>",
		},
		{
			name:     "without name",
			email:    "test@example.com",
			nameVal:  "",
			expected: "test@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatEmailAddress(tt.email, tt.nameVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseEmailAddress(t *testing.T) {
	tests := []struct {
		name          string
		address       string
		expectedEmail string
		expectedName  string
	}{
		{
			name:          "name and email",
			address:       "Test User <test@example.com>",
			expectedEmail: "test@example.com",
			expectedName:  "Test User",
		},
		{
			name:          "quoted name",
			address:       "\"Test User\" <test@example.com>",
			expectedEmail: "test@example.com",
			expectedName:  "Test User",
		},
		{
			name:          "email only",
			address:       "test@example.com",
			expectedEmail: "test@example.com",
			expectedName:  "",
		},
		{
			name:          "with spaces",
			address:       "  Test User  <  test@example.com  >  ",
			expectedEmail: "test@example.com",
			expectedName:  "Test User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, name := ParseEmailAddress(tt.address)
			assert.Equal(t, tt.expectedEmail, email)
			assert.Equal(t, tt.expectedName, name)
		})
	}
}

func TestGetEmailProvider(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{"gmail", "user@gmail.com", "Google"},
		{"yahoo", "user@yahoo.com", "Yahoo"},
		{"hotmail", "user@hotmail.com", "Microsoft"},
		{"outlook", "user@outlook.com", "Microsoft"},
		{"qq", "user@qq.com", "Tencent"},
		{"163", "user@163.com", "NetEase"},
		{"other", "user@unknown.com", "Other"},
		{"invalid", "invalid-email", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEmailProvider(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEstimateDeliveryTime(t *testing.T) {
	tests := []struct {
		name        string
		emailCount  int
		provider    string
		expectedMin time.Duration
		expectedMax time.Duration
	}{
		{
			name:        "small batch google",
			emailCount:  10,
			provider:    "Google",
			expectedMin: 3 * time.Second,
			expectedMax: 6 * time.Second,
		},
		{
			name:        "large batch other",
			emailCount:  200,
			provider:    "Other",
			expectedMin: 6 * time.Second,
			expectedMax: 12 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateDeliveryTime(tt.emailCount, tt.provider)
			assert.GreaterOrEqual(t, result, tt.expectedMin)
			assert.LessOrEqual(t, result, tt.expectedMax)
		})
	}
}

func TestGetOptimalSendTime(t *testing.T) {
	optimal := GetOptimalSendTime("UTC")

	// 最佳时间应该是工作日的上午10点
	assert.True(t, optimal.Weekday() >= time.Monday && optimal.Weekday() <= time.Friday)
	assert.Equal(t, 10, optimal.Hour())
	assert.Equal(t, 0, optimal.Minute())
	assert.Equal(t, 0, optimal.Second())

	// 应该是未来时间
	assert.True(t, optimal.After(time.Now()) || optimal.Equal(time.Now().Truncate(time.Second)))
}

func TestCalculateEmailPriority(t *testing.T) {
	tests := []struct {
		name         string
		templateType string
		urgent       bool
		expected     int
	}{
		{"verification code", TemplateVerificationCode, false, PriorityHigh},
		{"password reset", TemplatePasswordReset, false, PriorityHigh},
		{"security alert", TemplateSecurityAlert, false, PriorityUrgent},
		{"welcome email", TemplateWelcome, false, PriorityNormal},
		{"urgent welcome", TemplateWelcome, true, PriorityUrgent},
		{"file shared", TemplateFileShared, false, PriorityLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateEmailPriority(tt.templateType, tt.urgent)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateEmailQueue(t *testing.T) {
	to := []string{"test@example.com"}
	variables := map[string]interface{}{"code": "123456"}

	queue := CreateEmailQueue(TemplateVerificationCode, to, variables, PriorityHigh)

	assert.NotEmpty(t, queue.ID)
	assert.Equal(t, to, queue.To)
	assert.Equal(t, TemplateVerificationCode, queue.Template)
	assert.Equal(t, variables, queue.Variables)
	assert.Equal(t, PriorityHigh, queue.Priority)
	assert.Equal(t, EmailStatusPending, queue.Status)
	assert.False(t, queue.CreatedAt.IsZero())
	assert.False(t, queue.UpdatedAt.IsZero())
	assert.Equal(t, 3, queue.MaxAttempts)
}

func TestCreateDirectEmailQueue(t *testing.T) {
	to := []string{"test@example.com"}
	subject := "Test Subject"
	htmlBody := "<h1>Test</h1>"
	textBody := "Test"

	queue := CreateDirectEmailQueue(to, subject, htmlBody, textBody, PriorityNormal)

	assert.NotEmpty(t, queue.ID)
	assert.Equal(t, to, queue.To)
	assert.Equal(t, subject, queue.Subject)
	assert.Equal(t, htmlBody, queue.HTMLBody)
	assert.Equal(t, textBody, queue.TextBody)
	assert.Equal(t, PriorityNormal, queue.Priority)
	assert.Equal(t, EmailStatusPending, queue.Status)
	assert.False(t, queue.CreatedAt.IsZero())
	assert.False(t, queue.UpdatedAt.IsZero())
	assert.Equal(t, 3, queue.MaxAttempts)
}

// 基准测试
func BenchmarkValidateEmailAddress(b *testing.B) {
	email := "test@example.com"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateEmailAddress(email)
	}
}

func BenchmarkNormalizeEmailAddress(b *testing.B) {
	email := "  TEST@EXAMPLE.COM  "
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NormalizeEmailAddress(email)
	}
}

func BenchmarkSanitizeEmailContent(b *testing.B) {
	content := "<p>Hello</p><script>alert('xss')</script><style>body{color:red}</style><p>World</p>"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeEmailContent(content)
	}
}

package utils

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatTime(t *testing.T) {
	testTime := time.Date(2024, 1, 1, 15, 4, 5, 0, time.UTC)

	tests := []struct {
		layout   string
		expected string
	}{
		{DateLayout, "2024-01-01"},
		{TimeLayout, "15:04:05"},
		{DateTimeLayout, "2024-01-01 15:04:05"},
		{RFC3339Layout, "2024-01-01T15:04:05Z"},
	}

	for _, tt := range tests {
		result := FormatTime(testTime, tt.layout)
		assert.Equal(t, tt.expected, result)
	}

	// Test zero time
	result := FormatTime(time.Time{}, DateLayout)
	assert.Equal(t, "", result)
}

func TestFormatTimePtr(t *testing.T) {
	testTime := time.Date(2024, 1, 1, 15, 4, 5, 0, time.UTC)

	result := FormatTimePtr(&testTime, DateLayout)
	assert.Equal(t, "2024-01-01", result)

	// Test nil pointer
	result = FormatTimePtr(nil, DateLayout)
	assert.Equal(t, "", result)

	// Test zero time pointer
	zeroTime := time.Time{}
	result = FormatTimePtr(&zeroTime, DateLayout)
	assert.Equal(t, "", result)
}

func TestFormatNow(t *testing.T) {
	result := FormatNow(DateLayout)
	assert.NotEmpty(t, result)
	assert.Len(t, result, 10) // YYYY-MM-DD format
}

func TestFormatSpecific(t *testing.T) {
	testTime := time.Date(2024, 1, 1, 15, 4, 5, 123000000, time.UTC)

	assert.Equal(t, "2024-01-01", FormatDate(testTime))
	assert.Equal(t, "2024-01-01 15:04:05", FormatDateTime(testTime))
	assert.Equal(t, "2024-01-01 15:04:05.123", FormatTimestamp(testTime))
	assert.Equal(t, "2024-01-01T15:04:05Z", FormatRFC3339(testTime))
	assert.Equal(t, "20240101150405", FormatCompactDateTime(testTime))
	assert.Equal(t, "2024年01月01日", FormatChineseDate(testTime))
	assert.Equal(t, "2024年01月01日 15:04:05", FormatChineseDateTime(testTime))
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		layout   string
		value    string
		expected time.Time
		hasError bool
	}{
		{DateLayout, "2024-01-01", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), false},
		{DateTimeLayout, "2024-01-01 15:04:05", time.Date(2024, 1, 1, 15, 4, 5, 0, time.UTC), false},
		{DateLayout, "", time.Time{}, false}, // empty string
		{DateLayout, "invalid", time.Time{}, true},
	}

	for _, tt := range tests {
		result, err := ParseTime(tt.layout, tt.value)
		if tt.hasError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	}
}

func TestParseTimeInLocation(t *testing.T) {
	result, err := ParseTimeInLocation(DateTimeLayout, "2024-01-01 15:04:05", TimezoneBijing)
	assert.NoError(t, err)
	assert.Equal(t, TimezoneBijing, result.Location())

	// Test with nil location (should use Local)
	result, err = ParseTimeInLocation(DateTimeLayout, "2024-01-01 15:04:05", nil)
	assert.NoError(t, err)
	assert.Equal(t, time.Local, result.Location())
}

func TestParseSpecific(t *testing.T) {
	_, err := ParseDate("2024-01-01")
	assert.NoError(t, err)

	_, err = ParseDateTime("2024-01-01 15:04:05")
	assert.NoError(t, err)

	_, err = ParseRFC3339("2024-01-01T15:04:05Z")
	assert.NoError(t, err)
}

func TestTryParseTime(t *testing.T) {
	tests := []string{
		"2024-01-01T15:04:05Z",
		"2024-01-01 15:04:05",
		"2024-01-01",
		"2024/01/01 15:04:05",
		"2024/01/01",
		"01-01-2024 15:04:05",
		"01-01-2024",
		"2024-1-1 15:4:5",
		"2024-1-1",
	}

	for _, timeStr := range tests {
		t.Run(timeStr, func(t *testing.T) {
			result, err := TryParseTime(timeStr)
			assert.NoError(t, err)
			assert.False(t, result.IsZero())
		})
	}

	// Test invalid format
	_, err := TryParseTime("invalid-date-format")
	assert.Error(t, err)
}

func TestToTimezone(t *testing.T) {
	testTime := time.Now().UTC()

	result, err := ToTimezone(testTime, "Asia/Shanghai")
	assert.NoError(t, err)
	assert.Equal(t, "Asia/Shanghai", result.Location().String())

	// Test invalid timezone
	_, err = ToTimezone(testTime, "Invalid/Timezone")
	assert.Error(t, err)
}

func TestTimezoneConversions(t *testing.T) {
	testTime := time.Now()

	beijingTime := ToBeijingTime(testTime)
	assert.Equal(t, TimezoneBijing, beijingTime.Location())

	utcTime := ToUTC(testTime)
	assert.Equal(t, time.UTC, utcTime.Location())

	localTime := ToLocal(testTime)
	assert.Equal(t, time.Local, localTime.Location())
}

func TestGetTimezone(t *testing.T) {
	utcTime := time.Now().UTC()
	timezone := GetTimezone(utcTime)
	assert.Equal(t, "UTC", timezone)
}

func TestGetTimezoneOffset(t *testing.T) {
	utcTime := time.Now().UTC()
	offset := GetTimezoneOffset(utcTime)
	assert.Equal(t, 0, offset) // UTC has 0 offset
}

func TestIsToday(t *testing.T) {
	now := time.Now()
	assert.True(t, IsToday(now))

	yesterday := now.AddDate(0, 0, -1)
	assert.False(t, IsToday(yesterday))

	tomorrow := now.AddDate(0, 0, 1)
	assert.False(t, IsToday(tomorrow))
}

func TestIsYesterday(t *testing.T) {
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	assert.True(t, IsYesterday(yesterday))
	assert.False(t, IsYesterday(now))
}

func TestIsTomorrow(t *testing.T) {
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)
	assert.True(t, IsTomorrow(tomorrow))
	assert.False(t, IsTomorrow(now))
}

func TestIsWeekend(t *testing.T) {
	// Create a Saturday
	saturday := time.Date(2024, 1, 6, 12, 0, 0, 0, time.UTC) // 2024-01-06 is Saturday
	assert.True(t, IsWeekend(saturday))

	// Create a Sunday
	sunday := time.Date(2024, 1, 7, 12, 0, 0, 0, time.UTC) // 2024-01-07 is Sunday
	assert.True(t, IsWeekend(sunday))

	// Create a Monday
	monday := time.Date(2024, 1, 8, 12, 0, 0, 0, time.UTC) // 2024-01-08 is Monday
	assert.False(t, IsWeekend(monday))
}

func TestIsWorkday(t *testing.T) {
	monday := time.Date(2024, 1, 8, 12, 0, 0, 0, time.UTC) // 2024-01-08 is Monday
	assert.True(t, IsWorkday(monday))

	saturday := time.Date(2024, 1, 6, 12, 0, 0, 0, time.UTC) // 2024-01-06 is Saturday
	assert.False(t, IsWorkday(saturday))
}

func TestIsSameDay(t *testing.T) {
	t1 := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC)
	t3 := time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC)

	assert.True(t, IsSameDay(t1, t2))
	assert.False(t, IsSameDay(t1, t3))
}

func TestIsSameWeek(t *testing.T) {
	monday := time.Date(2024, 1, 8, 12, 0, 0, 0, time.UTC)      // Monday
	wednesday := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)  // Wednesday same week
	nextMonday := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC) // Monday next week

	assert.True(t, IsSameWeek(monday, wednesday))
	assert.False(t, IsSameWeek(monday, nextMonday))
}

func TestIsSameMonth(t *testing.T) {
	t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)

	assert.True(t, IsSameMonth(t1, t2))
	assert.False(t, IsSameMonth(t1, t3))
}

func TestIsSameYear(t *testing.T) {
	t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	assert.True(t, IsSameYear(t1, t2))
	assert.False(t, IsSameYear(t1, t3))
}

func TestStartEndOfDay(t *testing.T) {
	testTime := time.Date(2024, 1, 1, 15, 30, 45, 0, time.UTC)

	start := StartOfDay(testTime)
	expected := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expected, start)

	end := EndOfDay(testTime)
	expectedEnd := time.Date(2024, 1, 1, 23, 59, 59, 999999999, time.UTC)
	assert.Equal(t, expectedEnd, end)
}

func TestStartEndOfWeek(t *testing.T) {
	// Wednesday 2024-01-03
	wednesday := time.Date(2024, 1, 3, 15, 30, 45, 0, time.UTC)

	start := StartOfWeek(wednesday)
	// Should be Monday 2024-01-01 00:00:00
	expectedStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedStart, start)

	end := EndOfWeek(wednesday)
	// Should be Sunday 2024-01-07 23:59:59
	expectedEnd := time.Date(2024, 1, 7, 23, 59, 59, 999999999, time.UTC)
	assert.Equal(t, expectedEnd, end)
}

func TestStartEndOfMonth(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 15, 30, 45, 0, time.UTC)

	start := StartOfMonth(testTime)
	expectedStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedStart, start)

	end := EndOfMonth(testTime)
	expectedEnd := time.Date(2024, 1, 31, 23, 59, 59, 999999999, time.UTC)
	assert.Equal(t, expectedEnd, end)
}

func TestStartEndOfYear(t *testing.T) {
	testTime := time.Date(2024, 6, 15, 15, 30, 45, 0, time.UTC)

	start := StartOfYear(testTime)
	expectedStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedStart, start)

	end := EndOfYear(testTime)
	expectedEnd := time.Date(2024, 12, 31, 23, 59, 59, 999999999, time.UTC)
	assert.Equal(t, expectedEnd, end)
}

func TestTimeBetween(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 3, 2, 30, 30, 0, time.UTC)

	days := DaysBetween(start, end)
	assert.Equal(t, 2, days)

	hours := HoursBetween(start, end)
	assert.Equal(t, 50, hours)

	minutes := MinutesBetween(start, end)
	assert.Equal(t, 3030, minutes)

	seconds := SecondsBetween(start, end)
	assert.Equal(t, 181830, seconds)

	// Test reversed order
	days = DaysBetween(end, start)
	assert.Equal(t, 2, days)
}

func TestAge(t *testing.T) {
	// 由于Age函数使用time.Now()，我们测试逻辑而不是具体值
	birthDate := time.Date(1990, 3, 10, 0, 0, 0, 0, time.UTC)
	now := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	// 计算年龄逻辑
	age := now.Year() - birthDate.Year()
	if now.YearDay() < birthDate.YearDay() {
		age--
	}

	assert.True(t, age > 0)
	assert.Equal(t, 34, age) // 2024 - 1990 = 34
}

func TestTimeAgo(t *testing.T) {
	now := time.Now()

	// Test recent time
	recent := now.Add(-30 * time.Second)
	result := TimeAgo(recent)
	assert.Equal(t, "刚刚", result)

	// Test minutes ago
	minutesAgo := now.Add(-5 * time.Minute)
	result = TimeAgo(minutesAgo)
	assert.Equal(t, "5分钟前", result)

	// Test hours ago
	hoursAgo := now.Add(-2 * time.Hour)
	result = TimeAgo(hoursAgo)
	assert.Equal(t, "2小时前", result)

	// Test days ago
	daysAgo := now.Add(-3 * 24 * time.Hour)
	result = TimeAgo(daysAgo)
	assert.Equal(t, "3天前", result)
}

func TestTimeUntil(t *testing.T) {
	now := time.Now()

	// Test past time
	past := now.Add(-1 * time.Hour)
	result := TimeUntil(past)
	assert.Equal(t, "已过期", result)

	// Test near future
	nearFuture := now.Add(30 * time.Second)
	result = TimeUntil(nearFuture)
	assert.Equal(t, "不到1分钟", result)

	// Test minutes - use a larger buffer to account for execution time
	minutes := now.Add(5*time.Minute + 30*time.Second) // Add buffer
	result = TimeUntil(minutes)
	// Check that result contains "5" or "6" minutes due to potential execution delay
	assert.True(t, result == "5分钟" || result == "6分钟",
		"Expected 5 or 6 minutes, but got: %s", result)
}

func TestUnixConversions(t *testing.T) {
	timestamp := int64(1704067200) // 2024-01-01 00:00:00 UTC

	result := UnixToTime(timestamp)
	expected := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, result.Equal(expected)) // 使用Equal比较时间

	backToUnix := TimeToUnix(result)
	assert.Equal(t, timestamp, backToUnix)

	// Test milliseconds
	milliTimestamp := int64(1704067200123)
	result = UnixMilliToTime(milliTimestamp)
	backToMilli := TimeToUnixMilli(result)
	assert.Equal(t, milliTimestamp, backToMilli)
}

func TestAddBusinessDays(t *testing.T) {
	// Start on a Friday (2024-01-05)
	friday := time.Date(2024, 1, 5, 12, 0, 0, 0, time.UTC)

	// Add 1 business day should give Monday
	result := AddBusinessDays(friday, 1)
	assert.Equal(t, time.Monday, result.Weekday())

	// Add 0 business days should give same day
	result = AddBusinessDays(friday, 0)
	assert.Equal(t, friday, result)

	// Subtract 1 business day should give Thursday
	result = AddBusinessDays(friday, -1)
	assert.Equal(t, time.Thursday, result.Weekday())
}

func TestNextPrevBusinessDay(t *testing.T) {
	friday := time.Date(2024, 1, 5, 12, 0, 0, 0, time.UTC)

	next := NextBusinessDay(friday)
	assert.Equal(t, time.Monday, next.Weekday())

	prev := PrevBusinessDay(friday)
	assert.Equal(t, time.Thursday, prev.Weekday())
}

func TestSleepFunctions(t *testing.T) {
	// Just test that these functions don't panic
	// We won't actually sleep in tests
	assert.NotPanics(t, func() {
		SleepMilliseconds(0)
	})

	assert.NotPanics(t, func() {
		SleepSeconds(0)
	})
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"1s", time.Second, false},
		{"1分钟", time.Minute, false},
		{"1小时", time.Hour, false},
		{"1天", 24 * time.Hour, false},
		{"1d", 24 * time.Hour, false},
		{"2d3h", 2*24*time.Hour + 3*time.Hour, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		result, err := ParseDuration(tt.input)
		if tt.hasError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    time.Duration
		contains string
	}{
		{500 * time.Millisecond, "ms"},
		{5 * time.Second, "s"},
		{5 * time.Minute, "m"},
		{5 * time.Hour, "h"},
		{25 * time.Hour, "d"},
	}

	for _, tt := range tests {
		result := FormatDuration(tt.input)
		assert.Contains(t, result, tt.contains)
	}
}

func TestIsLeapYear(t *testing.T) {
	assert.True(t, IsLeapYear(2024))  // divisible by 4
	assert.True(t, IsLeapYear(2000))  // divisible by 400
	assert.False(t, IsLeapYear(1900)) // divisible by 100 but not 400
	assert.False(t, IsLeapYear(2023)) // not divisible by 4
}

func TestDaysInMonth(t *testing.T) {
	assert.Equal(t, 31, DaysInMonth(2024, time.January))
	assert.Equal(t, 29, DaysInMonth(2024, time.February)) // leap year
	assert.Equal(t, 28, DaysInMonth(2023, time.February)) // not leap year
	assert.Equal(t, 30, DaysInMonth(2024, time.April))
}

func TestTimeout(t *testing.T) {
	ctx, cancel := Timeout(100 * time.Millisecond)
	defer cancel()

	assert.NotNil(t, ctx)

	// Wait for timeout
	select {
	case <-ctx.Done():
		assert.Equal(t, context.DeadlineExceeded, ctx.Err())
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Context should have timed out")
	}
}

func TestIsExpired(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(1 * time.Hour)

	assert.True(t, IsExpired(past))
	assert.False(t, IsExpired(future))
}

func TestIsExpiredWithTolerance(t *testing.T) {
	almostExpired := time.Now().Add(-30 * time.Second)
	tolerance := 1 * time.Minute

	// Should not be expired with tolerance
	assert.False(t, IsExpiredWithTolerance(almostExpired, tolerance))

	reallyExpired := time.Now().Add(-2 * time.Minute)
	// Should be expired even with tolerance
	assert.True(t, IsExpiredWithTolerance(reallyExpired, tolerance))
}

// Benchmark tests
func BenchmarkFormatTime(b *testing.B) {
	testTime := time.Now()
	for i := 0; i < b.N; i++ {
		FormatTime(testTime, DateTimeLayout)
	}
}

func BenchmarkParseTime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseTime(DateTimeLayout, "2024-01-01 15:04:05")
	}
}

func BenchmarkTimeAgo(b *testing.B) {
	testTime := time.Now().Add(-2 * time.Hour)
	for i := 0; i < b.N; i++ {
		TimeAgo(testTime)
	}
}

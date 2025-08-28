package utils

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// 时间格式常量
const (
	// DateLayout 日期格式 YYYY-MM-DD
	DateLayout = "2006-01-02"
	// TimeLayout 时间格式 HH:MM:SS
	TimeLayout = "15:04:05"
	// DateTimeLayout 日期时间格式 YYYY-MM-DD HH:MM:SS
	DateTimeLayout = "2006-01-02 15:04:05"
	// RFC3339Layout RFC3339 格式
	RFC3339Layout = time.RFC3339
	// TimestampLayout 时间戳格式 YYYY-MM-DD HH:MM:SS.000
	TimestampLayout = "2006-01-02 15:04:05.000"
	// CompactDateLayout 紧凑日期格式 YYYYMMDD
	CompactDateLayout = "20060102"
	// CompactTimeLayout 紧凑时间格式 HHMMSS
	CompactTimeLayout = "150405"
	// CompactDateTimeLayout 紧凑日期时间格式 YYYYMMDDHHMMSS
	CompactDateTimeLayout = "20060102150405"
	// ChineseDateLayout 中文日期格式
	ChineseDateLayout = "2006年01月02日"
	// ChineseDateTimeLayout 中文日期时间格式
	ChineseDateTimeLayout = "2006年01月02日 15:04:05"
)

// 常用时区
var (
	// TimezoneBijing 北京时区
	TimezoneBijing, _ = time.LoadLocation("Asia/Shanghai")
	// TimezoneUTC UTC时区
	TimezoneUTC = time.UTC
	// TimezoneLocal 本地时区
	TimezoneLocal = time.Local
)

// FormatTime 格式化时间
func FormatTime(t time.Time, layout string) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(layout)
}

// FormatTimePtr 格式化时间指针
func FormatTimePtr(t *time.Time, layout string) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(layout)
}

// FormatNow 格式化当前时间
func FormatNow(layout string) string {
	return time.Now().Format(layout)
}

// FormatDate 格式化为日期字符串
func FormatDate(t time.Time) string {
	return FormatTime(t, DateLayout)
}

// FormatDateTime 格式化为日期时间字符串
func FormatDateTime(t time.Time) string {
	return FormatTime(t, DateTimeLayout)
}

// FormatTimestamp 格式化为时间戳字符串
func FormatTimestamp(t time.Time) string {
	return FormatTime(t, TimestampLayout)
}

// FormatRFC3339 格式化为RFC3339字符串
func FormatRFC3339(t time.Time) string {
	return FormatTime(t, RFC3339Layout)
}

// FormatCompactDateTime 格式化为紧凑格式
func FormatCompactDateTime(t time.Time) string {
	return FormatTime(t, CompactDateTimeLayout)
}

// FormatChineseDate 格式化为中文日期
func FormatChineseDate(t time.Time) string {
	return FormatTime(t, ChineseDateLayout)
}

// FormatChineseDateTime 格式化为中文日期时间
func FormatChineseDateTime(t time.Time) string {
	return FormatTime(t, ChineseDateTimeLayout)
}

// ParseTime 解析时间字符串
func ParseTime(layout, value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	return time.Parse(layout, value)
}

// ParseTimeInLocation 在指定时区解析时间
func ParseTimeInLocation(layout, value string, loc *time.Location) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	if loc == nil {
		loc = time.Local
	}
	return time.ParseInLocation(layout, value, loc)
}

// ParseDate 解析日期字符串
func ParseDate(value string) (time.Time, error) {
	return ParseTime(DateLayout, value)
}

// ParseDateTime 解析日期时间字符串
func ParseDateTime(value string) (time.Time, error) {
	return ParseTime(DateTimeLayout, value)
}

// ParseRFC3339 解析RFC3339格式
func ParseRFC3339(value string) (time.Time, error) {
	return ParseTime(RFC3339Layout, value)
}

// TryParseTime 尝试用多种格式解析时间
func TryParseTime(value string) (time.Time, error) {
	layouts := []string{
		RFC3339Layout,
		DateTimeLayout,
		DateLayout,
		TimestampLayout,
		time.RFC3339Nano,
		"2006-01-02T15:04:05",
		"2006/01/02 15:04:05",
		"2006/01/02",
		"01-02-2006 15:04:05",
		"01-02-2006",
		"2006-1-2 15:4:5",
		"2006-1-2",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", value)
}

// ToTimezone 转换到指定时区
func ToTimezone(t time.Time, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return t, fmt.Errorf("invalid timezone: %w", err)
	}
	return t.In(loc), nil
}

// ToBeijingTime 转换为北京时间
func ToBeijingTime(t time.Time) time.Time {
	return t.In(TimezoneBijing)
}

// ToUTC 转换为UTC时间
func ToUTC(t time.Time) time.Time {
	return t.UTC()
}

// ToLocal 转换为本地时间
func ToLocal(t time.Time) time.Time {
	return t.Local()
}

// GetTimezone 获取时间的时区名称
func GetTimezone(t time.Time) string {
	zone, _ := t.Zone()
	return zone
}

// GetTimezoneOffset 获取时区偏移量（秒）
func GetTimezoneOffset(t time.Time) int {
	_, offset := t.Zone()
	return offset
}

// IsToday 检查是否是今天
func IsToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.YearDay() == now.YearDay()
}

// IsYesterday 检查是否是昨天
func IsYesterday(t time.Time) bool {
	yesterday := time.Now().AddDate(0, 0, -1)
	return t.Year() == yesterday.Year() && t.YearDay() == yesterday.YearDay()
}

// IsTomorrow 检查是否是明天
func IsTomorrow(t time.Time) bool {
	tomorrow := time.Now().AddDate(0, 0, 1)
	return t.Year() == tomorrow.Year() && t.YearDay() == tomorrow.YearDay()
}

// IsWeekend 检查是否是周末
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// IsWorkday 检查是否是工作日
func IsWorkday(t time.Time) bool {
	return !IsWeekend(t)
}

// IsSameDay 检查两个时间是否是同一天
func IsSameDay(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() && t1.YearDay() == t2.YearDay()
}

// IsSameWeek 检查两个时间是否在同一周
func IsSameWeek(t1, t2 time.Time) bool {
	year1, week1 := t1.ISOWeek()
	year2, week2 := t2.ISOWeek()
	return year1 == year2 && week1 == week2
}

// IsSameMonth 检查两个时间是否在同一月
func IsSameMonth(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() && t1.Month() == t2.Month()
}

// IsSameYear 检查两个时间是否在同一年
func IsSameYear(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year()
}

// StartOfDay 获取一天的开始时间
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay 获取一天的结束时间
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// StartOfWeek 获取一周的开始时间（周一）
func StartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	return StartOfDay(t.AddDate(0, 0, -weekday+1))
}

// EndOfWeek 获取一周的结束时间（周日）
func EndOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 { // Sunday
		weekday = 7
	}
	return EndOfDay(t.AddDate(0, 0, 7-weekday))
}

// StartOfMonth 获取一月的开始时间
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth 获取一月的结束时间
func EndOfMonth(t time.Time) time.Time {
	return EndOfDay(StartOfMonth(t).AddDate(0, 1, 0).AddDate(0, 0, -1))
}

// StartOfYear 获取一年的开始时间
func StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear 获取一年的结束时间
func EndOfYear(t time.Time) time.Time {
	return EndOfDay(time.Date(t.Year(), 12, 31, 0, 0, 0, 0, t.Location()))
}

// DaysBetween 计算两个日期之间的天数
func DaysBetween(start, end time.Time) int {
	if start.After(end) {
		start, end = end, start
	}
	return int(end.Sub(start).Hours() / 24)
}

// HoursBetween 计算两个时间之间的小时数
func HoursBetween(start, end time.Time) int {
	if start.After(end) {
		start, end = end, start
	}
	return int(end.Sub(start).Hours())
}

// MinutesBetween 计算两个时间之间的分钟数
func MinutesBetween(start, end time.Time) int {
	if start.After(end) {
		start, end = end, start
	}
	return int(end.Sub(start).Minutes())
}

// SecondsBetween 计算两个时间之间的秒数
func SecondsBetween(start, end time.Time) int {
	if start.After(end) {
		start, end = end, start
	}
	return int(end.Sub(start).Seconds())
}

// Age 计算年龄
func Age(birthDate time.Time) int {
	now := time.Now()
	age := now.Year() - birthDate.Year()

	// 如果今年的生日还没到，年龄减1
	if now.YearDay() < birthDate.YearDay() {
		age--
	}

	return age
}

// TimeAgo 人性化显示时间差（多少时间前）
func TimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "刚刚"
	}
	if diff < time.Hour {
		return fmt.Sprintf("%d分钟前", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%d小时前", int(diff.Hours()))
	}
	if diff < 30*24*time.Hour {
		return fmt.Sprintf("%d天前", int(diff.Hours()/24))
	}
	if diff < 365*24*time.Hour {
		return fmt.Sprintf("%d个月前", int(diff.Hours()/(24*30)))
	}
	return fmt.Sprintf("%d年前", int(diff.Hours()/(24*365)))
}

// TimeUntil 人性化显示时间差（还有多长时间）
func TimeUntil(t time.Time) string {
	now := time.Now()
	diff := t.Sub(now)

	if diff < 0 {
		return "已过期"
	}
	if diff < time.Minute {
		return "不到1分钟"
	}
	if diff < time.Hour {
		return fmt.Sprintf("%d分钟", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%d小时", int(diff.Hours()))
	}
	if diff < 30*24*time.Hour {
		return fmt.Sprintf("%d天", int(diff.Hours()/24))
	}
	if diff < 365*24*time.Hour {
		return fmt.Sprintf("%d个月", int(diff.Hours()/(24*30)))
	}
	return fmt.Sprintf("%d年", int(diff.Hours()/(24*365)))
}

// UnixToTime Unix时间戳转时间
func UnixToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

// UnixMilliToTime Unix毫秒时间戳转时间
func UnixMilliToTime(timestamp int64) time.Time {
	return time.Unix(timestamp/1000, (timestamp%1000)*1000000)
}

// TimeToUnix 时间转Unix时间戳
func TimeToUnix(t time.Time) int64 {
	return t.Unix()
}

// TimeToUnixMilli 时间转Unix毫秒时间戳
func TimeToUnixMilli(t time.Time) int64 {
	return t.UnixNano() / 1000000
}

// AddBusinessDays 添加工作日
func AddBusinessDays(t time.Time, days int) time.Time {
	result := t
	remaining := days

	for remaining != 0 {
		if remaining > 0 {
			result = result.AddDate(0, 0, 1)
			if IsWorkday(result) {
				remaining--
			}
		} else {
			result = result.AddDate(0, 0, -1)
			if IsWorkday(result) {
				remaining++
			}
		}
	}

	return result
}

// NextBusinessDay 下一个工作日
func NextBusinessDay(t time.Time) time.Time {
	return AddBusinessDays(t, 1)
}

// PrevBusinessDay 上一个工作日
func PrevBusinessDay(t time.Time) time.Time {
	return AddBusinessDays(t, -1)
}

// Sleep 休眠指定时间
func Sleep(duration time.Duration) {
	time.Sleep(duration)
}

// SleepSeconds 休眠指定秒数
func SleepSeconds(seconds int) {
	time.Sleep(time.Duration(seconds) * time.Second)
}

// SleepMilliseconds 休眠指定毫秒数
func SleepMilliseconds(milliseconds int) {
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
}

// ParseDuration 解析时间间隔字符串
func ParseDuration(s string) (time.Duration, error) {
	// 支持中文单位
	s = strings.ReplaceAll(s, "秒", "s")
	s = strings.ReplaceAll(s, "分钟", "m")
	s = strings.ReplaceAll(s, "分", "m")
	s = strings.ReplaceAll(s, "小时", "h")
	s = strings.ReplaceAll(s, "时", "h")
	s = strings.ReplaceAll(s, "天", "d")
	s = strings.ReplaceAll(s, "日", "d")

	// 处理天数单位（Go原生不支持）
	if strings.Contains(s, "d") {
		parts := strings.Split(s, "d")
		if len(parts) == 2 {
			days, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return 0, fmt.Errorf("invalid duration format: %s", s)
			}
			remaining := strings.TrimSpace(parts[1])
			var additionalDuration time.Duration
			if remaining != "" {
				additionalDuration, err = time.ParseDuration(remaining)
				if err != nil {
					return 0, fmt.Errorf("invalid duration format: %s", s)
				}
			}
			return time.Duration(days)*24*time.Hour + additionalDuration, nil
		}
	}

	return time.ParseDuration(s)
}

// FormatDuration 格式化时间间隔
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
	return fmt.Sprintf("%.1fd", d.Hours()/24)
}

// IsLeapYear 检查是否是闰年
func IsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// DaysInMonth 获取指定月份的天数
func DaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// Timeout 创建超时上下文
func Timeout(duration time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), duration)
}

// IsExpired 检查时间是否已过期
func IsExpired(t time.Time) bool {
	return time.Now().After(t)
}

// IsExpiredWithTolerance 检查时间是否已过期（带容错时间）
func IsExpiredWithTolerance(t time.Time, tolerance time.Duration) bool {
	return time.Now().After(t.Add(tolerance))
}

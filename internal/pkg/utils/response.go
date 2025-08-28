package utils

import (
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ResponseCode 响应状态码
type ResponseCode int

// 业务状态码定义
const (
	// 成功响应码
	CodeSuccess ResponseCode = 200 // 成功

	// 客户端错误 (400-499)
	CodeBadRequest       ResponseCode = 400 // 请求参数错误
	CodeUnauthorized     ResponseCode = 401 // 未认证
	CodeForbidden        ResponseCode = 403 // 权限不足
	CodeNotFound         ResponseCode = 404 // 资源不存在
	CodeMethodNotAllowed ResponseCode = 405 // 方法不允许
	CodeConflict         ResponseCode = 409 // 资源冲突
	CodeTooManyRequests  ResponseCode = 429 // 请求过于频繁

	// 服务端错误 (500-599)
	CodeInternalError      ResponseCode = 500 // 服务器内部错误
	CodeBadGateway         ResponseCode = 502 // 网关错误
	CodeServiceUnavailable ResponseCode = 503 // 服务不可用
	CodeGatewayTimeout     ResponseCode = 504 // 网关超时

	// 自定义业务错误码 (1000+)
	CodeValidationError    ResponseCode = 1001 // 数据验证失败
	CodeDuplicateData      ResponseCode = 1002 // 数据重复
	CodeDataNotFound       ResponseCode = 1003 // 数据不存在
	CodeOperationFailed    ResponseCode = 1004 // 操作失败
	CodeQuotaExceeded      ResponseCode = 1005 // 配额超出
	CodeInvalidToken       ResponseCode = 1006 // 无效令牌
	CodeTokenExpired       ResponseCode = 1007 // 令牌过期
	CodePermissionDenied   ResponseCode = 1008 // 权限被拒绝
	CodeAccountLocked      ResponseCode = 1009 // 账户被锁定
	CodePasswordWrong      ResponseCode = 1010 // 密码错误
	CodeCaptchaRequired    ResponseCode = 1011 // 需要验证码
	CodeCaptchaWrong       ResponseCode = 1012 // 验证码错误
	CodeEmailNotVerified   ResponseCode = 1013 // 邮箱未验证
	CodePhoneNotVerified   ResponseCode = 1014 // 手机号未验证
	CodeFileUploadFailed   ResponseCode = 1015 // 文件上传失败
	CodeFileNotFound       ResponseCode = 1016 // 文件不存在
	CodeFileTypeNotAllowed ResponseCode = 1017 // 文件类型不允许
	CodeFileSizeExceeded   ResponseCode = 1018 // 文件大小超出限制
	CodeStorageQuotaFull   ResponseCode = 1019 // 存储配额已满
	CodeNetworkError       ResponseCode = 1020 // 网络错误
	CodeDatabaseError      ResponseCode = 1021 // 数据库错误
	CodeCacheError         ResponseCode = 1022 // 缓存错误
	CodeConfigError        ResponseCode = 1023 // 配置错误
)

// ResponseCodeMessages 响应码对应的消息
var ResponseCodeMessages = map[ResponseCode]string{
	CodeSuccess:            "操作成功",
	CodeBadRequest:         "请求参数错误",
	CodeUnauthorized:       "未认证",
	CodeForbidden:          "权限不足",
	CodeNotFound:           "资源不存在",
	CodeMethodNotAllowed:   "方法不允许",
	CodeConflict:           "资源冲突",
	CodeTooManyRequests:    "请求过于频繁",
	CodeInternalError:      "服务器内部错误",
	CodeBadGateway:         "网关错误",
	CodeServiceUnavailable: "服务不可用",
	CodeGatewayTimeout:     "网关超时",
	CodeValidationError:    "数据验证失败",
	CodeDuplicateData:      "数据重复",
	CodeDataNotFound:       "数据不存在",
	CodeOperationFailed:    "操作失败",
	CodeQuotaExceeded:      "配额超出",
	CodeInvalidToken:       "无效令牌",
	CodeTokenExpired:       "令牌过期",
	CodePermissionDenied:   "权限被拒绝",
	CodeAccountLocked:      "账户被锁定",
	CodePasswordWrong:      "密码错误",
	CodeCaptchaRequired:    "需要验证码",
	CodeCaptchaWrong:       "验证码错误",
	CodeEmailNotVerified:   "邮箱未验证",
	CodePhoneNotVerified:   "手机号未验证",
	CodeFileUploadFailed:   "文件上传失败",
	CodeFileNotFound:       "文件不存在",
	CodeFileTypeNotAllowed: "文件类型不允许",
	CodeFileSizeExceeded:   "文件大小超出限制",
	CodeStorageQuotaFull:   "存储配额已满",
	CodeNetworkError:       "网络错误",
	CodeDatabaseError:      "数据库错误",
	CodeCacheError:         "缓存错误",
	CodeConfigError:        "配置错误",
}

// Response 标准响应结构
type Response struct {
	Code      ResponseCode `json:"code"`           // 业务状态码
	Message   string       `json:"message"`        // 响应消息
	Data      interface{}  `json:"data,omitempty"` // 响应数据
	RequestID string       `json:"request_id"`     // 请求ID
	Timestamp int64        `json:"timestamp"`      // 时间戳
}

// ListResponse 列表响应结构
type ListResponse struct {
	Code       ResponseCode `json:"code"`       // 业务状态码
	Message    string       `json:"message"`    // 响应消息
	Data       interface{}  `json:"data"`       // 响应数据列表
	Pagination *Pagination  `json:"pagination"` // 分页信息
	RequestID  string       `json:"request_id"` // 请求ID
	Timestamp  int64        `json:"timestamp"`  // 时间戳
}

// Pagination 分页信息
type Pagination struct {
	CurrentPage  int   `json:"current_page"`            // 当前页码
	PageSize     int   `json:"page_size"`               // 每页大小
	TotalCount   int64 `json:"total_count"`             // 总记录数
	TotalPages   int   `json:"total_pages"`             // 总页数
	HasPrevious  bool  `json:"has_previous"`            // 是否有上一页
	HasNext      bool  `json:"has_next"`                // 是否有下一页
	PreviousPage int   `json:"previous_page,omitempty"` // 上一页页码
	NextPage     int   `json:"next_page,omitempty"`     // 下一页页码
}

// PageRequest 分页请求参数
type PageRequest struct {
	Page     int    `form:"page" json:"page" binding:"min=1"`           // 页码，默认1
	PageSize int    `form:"page_size" json:"page_size" binding:"min=1"` // 每页大小，默认20
	SortBy   string `form:"sort_by" json:"sort_by"`                     // 排序字段
	SortDir  string `form:"sort_dir" json:"sort_dir"`                   // 排序方向 asc/desc
}

// GetMessage 获取响应码对应的消息
func (code ResponseCode) GetMessage() string {
	if msg, exists := ResponseCodeMessages[code]; exists {
		return msg
	}
	return "未知错误"
}

// GetHTTPStatus 获取响应码对应的HTTP状态码
func (code ResponseCode) GetHTTPStatus() int {
	// 直接映射的HTTP状态码
	if code >= 400 && code < 600 {
		return int(code)
	}

	// 成功状态
	if code == CodeSuccess {
		return http.StatusOK
	}

	// 业务错误码映射
	return getBusinessErrorHTTPStatus(code)
}

// getBusinessErrorHTTPStatus 获取业务错误码对应的HTTP状态码
func getBusinessErrorHTTPStatus(code ResponseCode) int {
	switch code {
	case CodeValidationError:
		return http.StatusBadRequest
	case CodeDuplicateData:
		return http.StatusConflict
	case CodeDataNotFound:
		return http.StatusNotFound
	case CodeInvalidToken, CodeTokenExpired:
		return http.StatusUnauthorized
	case CodePermissionDenied, CodeQuotaExceeded:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

// getRequestID 从Gin上下文获取请求ID
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if rid, ok := requestID.(string); ok {
			return rid
		}
	}
	return "unknown"
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	response := Response{
		Code:      CodeSuccess,
		Message:   CodeSuccess.GetMessage(),
		Data:      data,
		RequestID: getRequestID(c),
		Timestamp: time.Now().Unix(),
	}
	c.JSON(CodeSuccess.GetHTTPStatus(), response)
}

// SuccessWithMessage 成功响应（自定义消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	response := Response{
		Code:      CodeSuccess,
		Message:   message,
		Data:      data,
		RequestID: getRequestID(c),
		Timestamp: time.Now().Unix(),
	}
	c.JSON(CodeSuccess.GetHTTPStatus(), response)
}

// Error 错误响应
func Error(c *gin.Context, code ResponseCode) {
	response := Response{
		Code:      code,
		Message:   code.GetMessage(),
		RequestID: getRequestID(c),
		Timestamp: time.Now().Unix(),
	}
	c.JSON(code.GetHTTPStatus(), response)
}

// ErrorWithMessage 错误响应（自定义消息）
func ErrorWithMessage(c *gin.Context, code ResponseCode, message string) {
	response := Response{
		Code:      code,
		Message:   message,
		RequestID: getRequestID(c),
		Timestamp: time.Now().Unix(),
	}
	c.JSON(code.GetHTTPStatus(), response)
}

// ErrorWithData 错误响应（包含数据）
func ErrorWithData(c *gin.Context, code ResponseCode, message string, data interface{}) {
	response := Response{
		Code:      code,
		Message:   message,
		Data:      data,
		RequestID: getRequestID(c),
		Timestamp: time.Now().Unix(),
	}
	c.JSON(code.GetHTTPStatus(), response)
}

// ValidationError 验证错误响应
func ValidationError(c *gin.Context, errors interface{}) {
	ErrorWithData(c, CodeValidationError, "数据验证失败", errors)
}

// Unauthorized 未认证响应
func Unauthorized(c *gin.Context) {
	Error(c, CodeUnauthorized)
}

// Forbidden 权限不足响应
func Forbidden(c *gin.Context) {
	Error(c, CodeForbidden)
}

// NotFound 资源不存在响应
func NotFound(c *gin.Context) {
	Error(c, CodeNotFound)
}

// NotFoundWithMessage 资源不存在响应（自定义消息）
func NotFoundWithMessage(c *gin.Context, message string) {
	ErrorWithMessage(c, CodeNotFound, message)
}

// InternalError 服务器错误响应
func InternalError(c *gin.Context) {
	Error(c, CodeInternalError)
}

// InternalErrorWithMessage 服务器错误响应（自定义消息）
func InternalErrorWithMessage(c *gin.Context, message string) {
	ErrorWithMessage(c, CodeInternalError, message)
}

// SuccessList 成功列表响应
func SuccessList(c *gin.Context, data interface{}, pagination *Pagination) {
	response := ListResponse{
		Code:       CodeSuccess,
		Message:    CodeSuccess.GetMessage(),
		Data:       data,
		Pagination: pagination,
		RequestID:  getRequestID(c),
		Timestamp:  time.Now().Unix(),
	}
	c.JSON(CodeSuccess.GetHTTPStatus(), response)
}

// SuccessListWithMessage 成功列表响应（自定义消息）
func SuccessListWithMessage(c *gin.Context, message string, data interface{}, pagination *Pagination) {
	response := ListResponse{
		Code:       CodeSuccess,
		Message:    message,
		Data:       data,
		Pagination: pagination,
		RequestID:  getRequestID(c),
		Timestamp:  time.Now().Unix(),
	}
	c.JSON(CodeSuccess.GetHTTPStatus(), response)
}

// NewPagination 创建分页信息
func NewPagination(currentPage, pageSize int, totalCount int64) *Pagination {
	if currentPage < 1 {
		currentPage = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))
	if totalPages < 1 {
		totalPages = 1
	}

	pagination := &Pagination{
		CurrentPage: currentPage,
		PageSize:    pageSize,
		TotalCount:  totalCount,
		TotalPages:  totalPages,
		HasPrevious: currentPage > 1,
		HasNext:     currentPage < totalPages,
	}

	if pagination.HasPrevious {
		pagination.PreviousPage = currentPage - 1
	}
	if pagination.HasNext {
		pagination.NextPage = currentPage + 1
	}

	return pagination
}

// DefaultPageRequest 获取默认分页请求参数
func DefaultPageRequest() PageRequest {
	return PageRequest{
		Page:     1,
		PageSize: 20,
		SortBy:   "id",
		SortDir:  "desc",
	}
}

// ParsePageRequest 解析分页请求参数
func ParsePageRequest(c *gin.Context) PageRequest {
	var req PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req = DefaultPageRequest()
	}

	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 100 { // 限制最大每页数量
		req.PageSize = 100
	}
	if req.SortBy == "" {
		req.SortBy = "id"
	}
	if req.SortDir != "asc" && req.SortDir != "desc" {
		req.SortDir = "desc"
	}

	return req
}

// GetOffset 计算数据库查询偏移量
func (pr PageRequest) GetOffset() int {
	return (pr.Page - 1) * pr.PageSize
}

// GetLimit 获取查询限制数量
func (pr PageRequest) GetLimit() int {
	return pr.PageSize
}

// GetOrderBy 获取排序字符串
func (pr PageRequest) GetOrderBy() string {
	return pr.SortBy + " " + pr.SortDir
}

// ValidateSortField 验证排序字段是否合法
func (pr PageRequest) ValidateSortField(allowedFields []string) bool {
	for _, field := range allowedFields {
		if pr.SortBy == field {
			return true
		}
	}
	return false
}

// FileResponse 文件响应结构
type FileResponse struct {
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	ContentType string `json:"content_type"`
	DownloadURL string `json:"download_url,omitempty"`
	PreviewURL  string `json:"preview_url,omitempty"`
	Checksum    string `json:"checksum,omitempty"`
}

// UploadResponse 上传响应结构
type UploadResponse struct {
	FileID      string `json:"file_id"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	ContentType string `json:"content_type"`
	URL         string `json:"url,omitempty"`
	Checksum    string `json:"checksum,omitempty"`
}

// SuccessFile 文件操作成功响应
func SuccessFile(c *gin.Context, fileResp *FileResponse) {
	Success(c, fileResp)
}

// SuccessUpload 上传成功响应
func SuccessUpload(c *gin.Context, uploadResp *UploadResponse) {
	SuccessWithMessage(c, "文件上传成功", uploadResp)
}

// AuthResponse 认证响应结构
type AuthResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token,omitempty"`
	TokenType    string      `json:"token_type"`
	ExpiresIn    int64       `json:"expires_in"`
	User         interface{} `json:"user,omitempty"`
}

// SuccessAuth 认证成功响应
func SuccessAuth(c *gin.Context, authResp *AuthResponse) {
	SuccessWithMessage(c, "登录成功", authResp)
}

// NoContent 无内容响应
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Created 创建成功响应
func Created(c *gin.Context, data interface{}) {
	response := Response{
		Code:      CodeSuccess,
		Message:   "创建成功",
		Data:      data,
		RequestID: getRequestID(c),
		Timestamp: time.Now().Unix(),
	}
	c.JSON(http.StatusCreated, response)
}

// Updated 更新成功响应
func Updated(c *gin.Context, data interface{}) {
	SuccessWithMessage(c, "更新成功", data)
}

// Deleted 删除成功响应
func Deleted(c *gin.Context) {
	SuccessWithMessage(c, "删除成功", nil)
}

// TooManyRequests 请求过于频繁响应
func TooManyRequests(c *gin.Context) {
	Error(c, CodeTooManyRequests)
}

// ServiceUnavailable 服务不可用响应
func ServiceUnavailable(c *gin.Context) {
	Error(c, CodeServiceUnavailable)
}

// ConvertToMap 将响应转换为map（用于日志记录）
func (r *Response) ConvertToMap() map[string]interface{} {
	return map[string]interface{}{
		"code":       r.Code,
		"message":    r.Message,
		"data":       r.Data,
		"request_id": r.RequestID,
		"timestamp":  r.Timestamp,
	}
}

// ConvertToMap 将列表响应转换为map（用于日志记录）
func (r *ListResponse) ConvertToMap() map[string]interface{} {
	return map[string]interface{}{
		"code":       r.Code,
		"message":    r.Message,
		"data":       r.Data,
		"pagination": r.Pagination,
		"request_id": r.RequestID,
		"timestamp":  r.Timestamp,
	}
}

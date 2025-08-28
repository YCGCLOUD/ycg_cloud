# HXLOS Cloud 云盘系统使用示例

## 概述

本文档提供了HXLOS Cloud云盘系统的详细使用示例，包括常见场景的完整代码实现。

## 目录

- [快速开始](#快速开始)
- [用户管理示例](#用户管理示例)
- [文件操作示例](#文件操作示例)
- [缓存使用示例](#缓存使用示例)
- [完整业务场景](#完整业务场景)

---

## 快速开始

### 项目初始化

```go
package main

import (
    "log"
    "cloudpan/internal/pkg/config"
    "cloudpan/internal/pkg/database"
    "cloudpan/internal/pkg/cache"
)

func main() {
    // 1. 初始化配置
    if err := config.Init(); err != nil {
        log.Fatalf("配置初始化失败: %v", err)
    }
    
    // 2. 初始化数据库
    if err := database.InitMySQL(); err != nil {
        log.Fatalf("数据库初始化失败: %v", err)
    }
    
    // 3. 初始化Redis缓存
    if err := cache.InitRedis(); err != nil {
        log.Fatalf("Redis初始化失败: %v", err)
    }
    
    // 4. 执行数据库迁移
    if err := database.AutoMigrate(); err != nil {
        log.Printf("数据库迁移警告: %v", err)
    }
    
    log.Println("系统初始化完成")
}
```

### 基础模型定义

```go
package models

import (
    "time"
    "cloudpan/internal/pkg/database/models"
)

// User 用户模型
type User struct {
    models.AuditModel
    Username    string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
    Email       string    `gorm:"uniqueIndex;size:100;not null" json:"email"`
    Password    string    `gorm:"size:255;not null" json:"-"`
    Nickname    string    `gorm:"size:100" json:"nickname"`
    Status      string    `gorm:"size:20;default:'active'" json:"status"`
    LastLoginAt *time.Time `json:"last_login_at"`
    
    Profile *UserProfile `gorm:"foreignKey:UserID" json:"profile,omitempty"`
    Files   []File       `gorm:"foreignKey:OwnerID" json:"files,omitempty"`
}

// UserProfile 用户配置
type UserProfile struct {
    models.BaseModel
    UserID      uint   `gorm:"uniqueIndex;not null" json:"user_id"`
    StorageUsed int64  `gorm:"default:0" json:"storage_used"`
    StorageMax  int64  `gorm:"default:1073741824" json:"storage_max"` // 1GB
    Language    string `gorm:"size:10;default:'zh-CN'" json:"language"`
}

// File 文件模型
type File struct {
    models.AuditModel
    Name        string `gorm:"size:255;not null" json:"name"`
    Path        string `gorm:"size:1000;not null" json:"path"`
    Size        int64  `gorm:"not null" json:"size"`
    MimeType    string `gorm:"size:100" json:"mime_type"`
    Hash        string `gorm:"size:64;index" json:"hash"`
    OwnerID     uint   `gorm:"not null;index" json:"owner_id"`
    IsDir       bool   `gorm:"default:false" json:"is_dir"`
    
    Owner User `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
}
```

---

## 用户管理示例

### 用户注册

```go
package service

import (
    "context"
    "fmt"
    "time"
    "golang.org/x/crypto/bcrypt"
    "cloudpan/internal/pkg/database"
    "cloudpan/internal/pkg/cache"
    "cloudpan/internal/pkg/errors"
)

type UserService struct {
    db    *gorm.DB
    cache cache.Manager
}

func NewUserService() *UserService {
    return &UserService{
        db:    database.GetDB(),
        cache: cache.Cache,
    }
}

// RegisterUser 用户注册
func (s *UserService) RegisterUser(ctx context.Context, req *RegisterRequest) (*User, error) {
    // 1. 验证输入
    if err := s.validateRegisterRequest(req); err != nil {
        return nil, err
    }
    
    // 2. 检查用户名和邮箱是否已存在
    exists, err := database.Exists(s.db, &User{}, "username = ? OR email = ?", req.Username, req.Email)
    if err != nil {
        return nil, errors.WrapError(err, "failed to check user existence")
    }
    if exists {
        return nil, errors.ErrResourceExists
    }
    
    // 3. 加密密码
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, errors.WrapError(err, "failed to hash password")
    }
    
    // 4. 在事务中创建用户和配置
    var user *User
    err = database.Transaction(func(tx *gorm.DB) error {
        user = &User{
            Username: req.Username,
            Email:    req.Email,
            Password: string(hashedPassword),
            Nickname: req.Nickname,
            Status:   "active",
        }
        
        if err := tx.Create(user).Error; err != nil {
            return err
        }
        
        profile := &UserProfile{
            UserID:     user.ID,
            StorageMax: 1024 * 1024 * 1024, // 1GB
            Language:   "zh-CN",
        }
        
        if err := tx.Create(profile).Error; err != nil {
            return err
        }
        
        user.Profile = profile
        return nil
    })
    
    if err != nil {
        return nil, errors.WrapError(err, "failed to create user")
    }
    
    // 5. 缓存用户信息
    cacheKey := cache.Keys.UserProfile(fmt.Sprintf("%d", user.ID))
    s.cache.Set(cacheKey, user, 1*time.Hour)
    
    return user, nil
}

type RegisterRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Nickname string `json:"nickname" validate:"max=100"`
}

func (s *UserService) validateRegisterRequest(req *RegisterRequest) error {
    if req.Username == "" {
        return errors.NewValidationError("username", "username is required")
    }
    
    if len(req.Username) < 3 || len(req.Username) > 50 {
        return errors.NewValidationError("username", "username must be 3-50 characters")
    }
    
    if req.Email == "" {
        return errors.NewValidationError("email", "email is required")
    }
    
    if len(req.Password) < 8 {
        return errors.NewValidationError("password", "password must be at least 8 characters")
    }
    
    return nil
}
```

### 用户登录

```go
// LoginUser 用户登录
func (s *UserService) LoginUser(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
    // 1. 验证输入
    if req.Email == "" || req.Password == "" {
        return nil, errors.NewValidationError("credentials", "email and password are required")
    }
    
    // 2. 查找用户
    var user User
    err := s.db.Preload("Profile").Where("email = ? AND status = ?", req.Email, "active").First(&user).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, errors.ErrResourceNotFound
        }
        return nil, errors.WrapError(err, "failed to find user")
    }
    
    // 3. 验证密码
    err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
    if err != nil {
        return nil, errors.ErrPermissionDenied
    }
    
    // 4. 生成会话token
    token, err := s.generateSessionToken(&user)
    if err != nil {
        return nil, errors.WrapError(err, "failed to generate session token")
    }
    
    // 5. 缓存会话
    sessionData := &SessionData{
        UserID:   user.ID,
        Username: user.Username,
        Email:    user.Email,
        LoginAt:  time.Now(),
    }
    
    sessionKey := cache.Keys.UserSession(token)
    err = cache.UserSession.Set(token, sessionData, 24*time.Hour)
    if err != nil {
        return nil, errors.WrapError(err, "failed to cache session")
    }
    
    // 6. 更新最后登录时间
    now := time.Now()
    s.db.Model(&user).Update("last_login_at", now)
    
    return &LoginResponse{
        Token: token,
        User:  s.maskUserSensitiveData(&user),
    }, nil
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Token string `json:"token"`
    User  *User  `json:"user"`
}

type SessionData struct {
    UserID   uint      `json:"user_id"`
    Username string    `json:"username"`
    Email    string    `json:"email"`
    LoginAt  time.Time `json:"login_at"`
}
```

---

## 文件操作示例

### 文件上传

```go
package service

type FileService struct {
    db        *gorm.DB
    cache     cache.Manager
    userSvc   *UserService
}

func NewFileService(userSvc *UserService) *FileService {
    return &FileService{
        db:      database.GetDB(),
        cache:   cache.Cache,
        userSvc: userSvc,
    }
}

// UploadFile 上传文件
func (s *FileService) UploadFile(ctx context.Context, userID uint, req *UploadRequest) (*File, error) {
    // 1. 检查用户权限和配额
    if err := s.checkUploadPermission(userID, req.Size); err != nil {
        return nil, err
    }
    
    // 2. 计算文件哈希
    hash, err := s.calculateFileHash(req.Content)
    if err != nil {
        return nil, errors.WrapError(err, "failed to calculate file hash")
    }
    
    // 3. 检查文件是否已存在（去重）
    var existingFile File
    err = s.db.Where("hash = ? AND owner_id = ?", hash, userID).First(&existingFile).Error
    if err == nil {
        return &existingFile, nil
    }
    
    // 4. 保存文件到存储系统
    filePath, err := s.saveFileToStorage(userID, req.Name, req.Content)
    if err != nil {
        return nil, errors.WrapError(err, "failed to save file")
    }
    
    // 5. 在事务中创建文件记录并更新用户配额
    var file *File
    err = database.Transaction(func(tx *gorm.DB) error {
        file = &File{
            Name:     req.Name,
            Path:     filePath,
            Size:     req.Size,
            MimeType: req.MimeType,
            Hash:     hash,
            OwnerID:  userID,
            IsDir:    false,
        }
        
        if err := tx.Create(file).Error; err != nil {
            return err
        }
        
        // 更新用户存储使用量
        return tx.Model(&UserProfile{}).
            Where("user_id = ?", userID).
            Update("storage_used", gorm.Expr("storage_used + ?", req.Size)).Error
    })
    
    if err != nil {
        os.Remove(filePath) // 清理已保存的文件
        return nil, errors.WrapError(err, "failed to create file record")
    }
    
    return file, nil
}

type UploadRequest struct {
    Name     string `json:"name"`
    Size     int64  `json:"size"`
    MimeType string `json:"mime_type"`
    Content  []byte `json:"-"`
}
```

### 缓存使用示例

```go
// 多级缓存策略
func (s *FileService) GetFileWithCache(fileID uint) (*File, error) {
    cacheKey := cache.Keys.FileInfo(fmt.Sprintf("%d", fileID))
    
    // 1. 尝试从缓存获取
    var file File
    err := s.cache.Get(cacheKey, &file)
    if err == nil {
        return &file, nil
    }
    
    // 2. 缓存未命中，从数据库获取
    err = s.db.Where("id = ?", fileID).First(&file).Error
    if err != nil {
        return nil, err
    }
    
    // 3. 写入缓存
    s.cache.Set(cacheKey, &file, 30*time.Minute)
    
    return &file, nil
}

// 批量缓存操作
func (s *FileService) BatchCacheFiles(files []File) error {
    batch := s.cache.Batch()
    
    for _, file := range files {
        key := cache.Keys.FileInfo(fmt.Sprintf("%d", file.ID))
        batch.Set(key, &file, 30*time.Minute)
    }
    
    return batch.Execute()
}
```

---

## 完整业务场景

### 文件分享场景

```go
// CreateFileShare 创建文件分享
func (s *FileService) CreateFileShare(ctx context.Context, userID uint, req *ShareRequest) (*FileShare, error) {
    // 1. 检查文件权限
    file, err := s.checkFilePermission(userID, req.FileID, "share")
    if err != nil {
        return nil, err
    }
    
    // 2. 生成分享token
    token, err := s.generateShareToken()
    if err != nil {
        return nil, err
    }
    
    // 3. 创建分享记录
    share := &FileShare{
        FileID:    req.FileID,
        OwnerID:   userID,
        Token:     token,
        Password:  req.Password,
        ExpiresAt: req.ExpiresAt,
        MaxViews:  req.MaxViews,
    }
    
    if err := s.db.Create(share).Error; err != nil {
        return nil, errors.WrapError(err, "failed to create file share")
    }
    
    // 4. 缓存分享信息
    shareKey := cache.Keys.FileShare(token)
    s.cache.Set(shareKey, share, time.Until(req.ExpiresAt))
    
    return share, nil
}

type ShareRequest struct {
    FileID    uint       `json:"file_id"`
    Password  string     `json:"password,omitempty"`
    ExpiresAt time.Time  `json:"expires_at"`
    MaxViews  int        `json:"max_views,omitempty"`
}

type FileShare struct {
    models.BaseModel
    FileID    uint       `json:"file_id"`
    OwnerID   uint       `json:"owner_id"`
    Token     string     `json:"token"`
    Password  string     `json:"-"`
    ExpiresAt time.Time  `json:"expires_at"`
    MaxViews  int        `json:"max_views"`
    ViewCount int        `json:"view_count"`
}
```

### 用户配额管理

```go
// UpdateUserQuota 更新用户配额
func (s *UserService) UpdateUserQuota(userID uint, sizeChange int64) error {
    return database.Transaction(func(tx *gorm.DB) error {
        var profile UserProfile
        err := tx.Where("user_id = ?", userID).First(&profile).Error
        if err != nil {
            return err
        }
        
        newUsed := profile.StorageUsed + sizeChange
        if newUsed < 0 {
            newUsed = 0
        }
        
        if newUsed > profile.StorageMax {
            return errors.ErrQuotaExceeded
        }
        
        return tx.Model(&profile).Update("storage_used", newUsed).Error
    })
}

// GetUserQuotaInfo 获取用户配额信息
func (s *UserService) GetUserQuotaInfo(userID uint) (*QuotaInfo, error) {
    cacheKey := cache.Keys.UserQuota(fmt.Sprintf("%d", userID))
    
    var quota QuotaInfo
    err := s.cache.Get(cacheKey, &quota)
    if err == nil {
        return &quota, nil
    }
    
    var profile UserProfile
    err = s.db.Where("user_id = ?", userID).First(&profile).Error
    if err != nil {
        return nil, err
    }
    
    quota = QuotaInfo{
        Used:       profile.StorageUsed,
        Max:        profile.StorageMax,
        Available:  profile.StorageMax - profile.StorageUsed,
        Percentage: float64(profile.StorageUsed) / float64(profile.StorageMax) * 100,
    }
    
    s.cache.Set(cacheKey, &quota, 5*time.Minute)
    return &quota, nil
}

type QuotaInfo struct {
    Used       int64   `json:"used"`
    Max        int64   `json:"max"`
    Available  int64   `json:"available"`
    Percentage float64 `json:"percentage"`
}
```

---

## 注意事项

1. **错误处理**: 所有示例都包含完整的错误处理
2. **事务管理**: 涉及多个操作的场景使用数据库事务
3. **缓存策略**: 合理使用缓存提高性能
4. **安全考虑**: 验证用户权限和输入参数
5. **性能优化**: 使用批量操作和预加载减少数据库查询

这些示例展示了系统的实际使用方法，可以作为开发参考。
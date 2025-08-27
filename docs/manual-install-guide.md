# 代码质量工具手动安装指南

## 网络问题解决方案

### 方案1：使用Go代理
```bash
# 设置Go代理为国内镜像
set GOPROXY=https://goproxy.cn,direct
go install github.com/securecodewarrior/gosec/v2@latest
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
```

### 方案2：使用七牛云代理
```bash
set GOPROXY=https://goproxy.qiniu.com,direct
go install github.com/securecodewarrior/gosec/v2@latest
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
```

### 方案3：使用阿里云代理
```bash
set GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
go install github.com/securecodewarrior/gosec/v2@latest
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
```

### 方案4：手动下载安装
1. 访问 https://github.com/securecodewarrior/gosec/releases
2. 下载对应系统的二进制文件
3. 将其放入 $GOPATH/bin 目录

### 方案5：使用替代工具

#### 安全检查替代方案：
- **staticcheck**: `go install honnef.co/go/tools/cmd/staticcheck@latest`
- **go-critic**: `go install github.com/go-critic/go-critic/cmd/gocritic@latest`
- **errcheck**: `go install github.com/kisielk/errcheck@latest`

#### 圈复杂度检查替代方案：
- **gocognit**: `go install github.com/uudashr/gocognit/cmd/gocognit@latest`
- **cyclop**: `go install github.com/bkielbasa/cyclop/cmd/cyclop@latest`

## 当前可用的检查方案

即使没有gosec和gocyclo，我们仍然可以通过以下方式进行质量检查：

### 安全检查：
1. **go vet** - 内置静态分析
2. **go mod verify** - 依赖安全检查
3. **人工代码审查** - 重点关注输入验证、SQL注入、XSS等

### 复杂度检查：
1. **人工审查** - 函数长度、嵌套深度
2. **代码审查清单** - 制定复杂度检查标准
3. **功能拆分原则** - 单一职责原则

## 临时解决方案

在工具安装成功前，我们可以：
1. 加强代码审查流程
2. 制定代码复杂度人工检查标准
3. 使用IDE插件进行实时检查
4. 定期进行安全代码审查
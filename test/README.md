# 测试目录

本目录包含项目的集成测试和端到端测试。

## 目录结构

```
test/
├── README.md              # 测试目录说明
├── integration_test.go    # 集成测试
└── (未来可能添加的其他测试)
```

## 测试类型

### 集成测试
- **文件**: `integration_test.go`
- **描述**: 测试多个组件之间的协同工作
- **覆盖范围**: 
  - 日志系统集成
  - 中间件集成
  - HTTP响应格式
  - 错误处理和恢复
  - 工具函数集成

### 运行测试

#### 运行所有集成测试
```bash
# 从项目根目录运行
go test ./test -v

# 或者进入test目录运行
cd test
go test -v
```

#### 运行特定测试
```bash
go test ./test -run TestIntegration -v
go test ./test -run TestLoggerIntegration -v
go test ./test -run TestUtilsIntegration -v
```

#### 运行性能基准测试
```bash
go test ./test -bench=BenchmarkIntegration -v
```

## 测试组织原则

1. **单元测试**: 与被测试代码放在同一目录，以 `_test.go` 结尾
2. **集成测试**: 放在 `/test` 目录中
3. **端到端测试**: 如有需要，也放在 `/test` 目录中

## 测试环境

集成测试使用独立的测试配置：
- 测试日志输出到 `test_logs/` 目录
- 使用Gin的测试模式
- 模拟HTTP请求和响应
- 自动清理测试文件

## 注意事项

- 运行测试前确保项目依赖已安装 (`go mod download`)
- 测试会创建临时日志文件，测试结束后自动清理
- 如遇到文件锁定问题，可手动删除 `test_logs/` 目录
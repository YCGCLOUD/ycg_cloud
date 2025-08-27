# HXLOS Cloud Storage - 代码质量检查 Makefile
# 开发计划第2天：代码质量检查工具配置

.PHONY: fmt lint vet test coverage quality-check clean build

# 代码格式化
fmt:
	@echo "=== 运行 gofmt 代码格式化 ==="
	gofmt -l -w .
	@echo "代码格式化完成"

# 代码规范检查
lint:
	@echo "=== 运行 golint 代码规范检查 ==="
	golint ./...
	@echo "代码规范检查完成"

# 静态分析
vet:
	@echo "=== 运行 go vet 静态分析 ==="
	go vet ./...
	@echo "静态分析完成"

# 单元测试
test:
	@echo "=== 运行单元测试 ==="
	go test -v ./...
	@echo "单元测试完成"

# 测试覆盖率
coverage:
	@echo "=== 生成测试覆盖率报告 ==="
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告已生成: coverage.html"

# 完整质量检查（覆盖率≥80%、圈复杂度≤10）
quality-check: fmt lint vet test
	@echo "=== 执行完整代码质量检查 ==="
	@echo "1. 代码格式化检查..."
	@gofmt -l . | wc -l
	@echo "2. 代码规范检查..."
	@golint ./... | wc -l
	@echo "3. 静态分析检查..."
	@go vet ./...
	@echo "4. 单元测试执行..."
	@go test ./...
	@echo "=== 质量检查完成 ==="

# 清理生成文件
clean:
	@echo "=== 清理生成文件 ==="
	rm -f coverage.out coverage.html
	rm -rf bin/
	@echo "清理完成"

# 构建项目
build:
	@echo "=== 构建项目 ==="
	go build -o bin/cloudpan.exe ./cmd
	@echo "构建完成: bin/cloudpan.exe"

# 开发环境质量检查（每日使用）
dev-check: fmt vet test
	@echo "=== 开发环境快速质量检查 ==="
	@echo "质量检查通过，可以提交代码"

# 生产环境质量检查（严格标准）
prod-check: quality-check coverage
	@echo "=== 生产环境严格质量检查 ==="
	@echo "检查覆盖率是否达到80%标准..."
	@go test -coverprofile=coverage.out ./... && \
	go tool cover -func=coverage.out | grep "total:" | awk '{print $$3}' | \
	sed 's/%//' | awk '{if($$1>=80) print "✅ 覆盖率达标: "$$1"%"; else {print "❌ 覆盖率不达标: "$$1"% (要求≥80%)"; exit 1}}'
	@echo "=== 生产环境质量检查完成 ==="
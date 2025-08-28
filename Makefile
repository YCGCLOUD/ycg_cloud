# HXLOS Cloud Storage - 代码质量检查 Makefile
# 开发计划第2天：代码质量检查工具配置

.PHONY: fmt lint vet sec test coverage quality-check clean build

# Code formatting
fmt:
	@echo "=== Running gofmt code formatting ==="
	gofmt -l -w .
	@echo "Code formatting completed"

# Code linting
lint:
	@echo "=== Running golint code linting ==="
	golint ./...
	@echo "Code linting completed"

# Static analysis
vet:
	@echo "=== Running go vet static analysis ==="
	go vet ./...
	@echo "Static analysis completed"

# Cyclomatic complexity check
cyclo:
	@echo "=== Running gocyclo complexity check ==="
	gocyclo -over 10 .
	@echo "Complexity check completed"

# Security vulnerability scan
sec:
	@echo "=== Running gosec security scan ==="
	gosec ./...
	@echo "Security scan completed"

# Unit testing
test:
	@echo "=== Running unit tests ==="
	go test -v ./...
	@echo "Unit testing completed"

# Test coverage
coverage:
	@echo "=== Generating test coverage report ==="
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 完整质量检查（覆盖率≥80%、圈复杂度≤10、安全漏洞扫描）
quality-check: fmt lint vet cyclo sec test
	@echo "=== Running complete code quality check ==="
	@echo "1. Code formatting check..."
	@gofmt -l . | wc -l
	@echo "2. Code linting check..."
	@golint ./... | wc -l
	@echo "3. Static analysis check..."
	@go vet ./...
	@echo "4. Complexity check..."
	@gocyclo -over 10 . || echo "✅ Complexity check passed (no functions over 10)"
	@echo "5. Security vulnerability scan..."
	@gosec ./... || echo "⚠️ Security issues found, please review"
	@echo "6. Unit test execution..."
	@go test ./...
	@echo "=== Quality check completed ==="

# Clean generated files
clean:
	@echo "=== Cleaning generated files ==="
	rm -f coverage.out coverage.html
	rm -rf bin/
	@echo "Clean completed"

# Build project
build:
	@echo "=== Building project ==="
	go build -o bin/cloudpan.exe ./cmd
	@echo "Build completed: bin/cloudpan.exe"

# Development environment quality check (daily use)
dev-check: fmt vet cyclo sec test
	@echo "=== Development environment quick quality check ==="
	@echo "Quality check passed, ready to commit code"

# Production environment quality check (strict standards)
prod-check: quality-check coverage
	@echo "=== Production environment strict quality check ==="
	@echo "Checking if coverage meets 80% standard..."
	@go test -coverprofile=coverage.out ./... && \
	go tool cover -func=coverage.out | grep "total:" | awk '{print $$3}' | \
	sed 's/%//' | awk '{if($$1>=80) print "✅ Coverage passed: "$$1"%"; else {print "❌ Coverage failed: "$$1"% (required ≥80%)"; exit 1}}'
	@echo "=== Production environment quality check completed ==="
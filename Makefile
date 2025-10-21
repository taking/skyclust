# SkyClust Cloud Management Portal Makefile

# ============================================================================
# Variables
# ============================================================================

# Application
APP_NAME := skyclust-cmp
BUILD_DIR := bin
BINARY_NAME := cmp-server
DOCKER_IMAGE := skyclust-cmp-server
DOCKER_TAG := latest
GO_VERSION := 1.24

# Build Info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go Build Flags
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"
GCFLAGS := -gcflags="-e"

# Colors
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
CYAN := \033[36m
RESET := \033[0m

# ============================================================================
# Phony Targets
# ============================================================================

.PHONY: help build clean docker lint format deps test \
        proto-gen proto-lint proto-breaking \
        build-providers providers-up providers-down \
        compose-up compose-down dev run

# ============================================================================
# Default Target
# ============================================================================

all: clean deps lint build ## 전체 빌드 (clean, deps, lint, build)

# ============================================================================
# Help
# ============================================================================

help: ## 도움말 메시지 표시
	@echo "$(BLUE)╔═══════════════════════════════════════════════════════════╗$(RESET)"
	@echo "$(BLUE)║  SkyClust Cloud Management Portal - Makefile Help        ║$(RESET)"
	@echo "$(BLUE)╚═══════════════════════════════════════════════════════════╝$(RESET)"
	@echo ""
	@echo "$(CYAN)Core Commands:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)  %-24s$(RESET) %s\n", $$1, $$2}' | \
		sort

# ============================================================================
# Build Targets
# ============================================================================

build: ## 메인 서버 빌드
	@echo "$(YELLOW)▶ Building $(APP_NAME)...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 go build $(GCFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server/main.go
	@echo "$(GREEN)✓ Built $(BUILD_DIR)/$(BINARY_NAME)$(RESET)"

build-optimized: ## 최적화된 바이너리 빌드
	@echo "$(YELLOW)▶ Building optimized $(APP_NAME)...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-optimized ./cmd/server/main.go
	@echo "$(GREEN)✓ Built $(BUILD_DIR)/$(BINARY_NAME)-optimized$(RESET)"

build-all-platforms: build-linux build-darwin build-windows ## 모든 플랫폼용 빌드

build-linux: ## Linux AMD64 빌드
	@echo "$(YELLOW)▶ Building for Linux AMD64...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/server/main.go
	@echo "$(GREEN)✓ Built $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64$(RESET)"

build-darwin: ## macOS 빌드 (AMD64 + ARM64)
	@echo "$(YELLOW)▶ Building for macOS...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/server/main.go
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/server/main.go
	@echo "$(GREEN)✓ Built $(BUILD_DIR)/$(BINARY_NAME)-darwin-*$(RESET)"

build-windows: ## Windows AMD64 빌드
	@echo "$(YELLOW)▶ Building for Windows AMD64...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/server/main.go
	@echo "$(GREEN)✓ Built $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe$(RESET)"

# ============================================================================
# Proto Targets
# ============================================================================

proto-gen: ## gRPC 코드 생성
	@echo "$(YELLOW)▶ Generating gRPC code from proto files...$(RESET)"
	@if command -v buf >/dev/null 2>&1; then \
		buf generate; \
		echo "$(GREEN)✓ gRPC code generated in api/gen$(RESET)"; \
	else \
		echo "$(RED)✗ buf not installed. Install: go install github.com/bufbuild/buf/cmd/buf@latest$(RESET)"; \
		exit 1; \
	fi

proto-lint: ## Proto 파일 lint
	@echo "$(YELLOW)▶ Linting proto files...$(RESET)"
	@if command -v buf >/dev/null 2>&1; then \
		buf lint; \
		echo "$(GREEN)✓ Proto files linted$(RESET)"; \
	else \
		echo "$(RED)✗ buf not installed$(RESET)"; \
		exit 1; \
	fi

proto-breaking: ## Proto breaking changes 검사
	@echo "$(YELLOW)▶ Checking for breaking changes...$(RESET)"
	@if command -v buf >/dev/null 2>&1; then \
		buf breaking --against '.git#branch=main' || echo "$(YELLOW)⚠ No baseline to compare$(RESET)"; \
	else \
		echo "$(RED)✗ buf not installed$(RESET)"; \
		exit 1; \
	fi

proto-clean: ## Generated proto 파일 삭제
	@echo "$(YELLOW)▶ Cleaning generated proto files...$(RESET)"
	@rm -rf api/gen
	@echo "$(GREEN)✓ Cleaned api/gen$(RESET)"

# ============================================================================
# Provider Targets
# ============================================================================

build-providers: ## 모든 Provider 서버 빌드
	@echo "$(YELLOW)▶ Building gRPC Provider servers...$(RESET)"
	@mkdir -p $(BUILD_DIR)/providers
	@echo "  • AWS Provider..."
	@cd providers/aws && go build -o ../../$(BUILD_DIR)/providers/aws-provider . 2>/dev/null || echo "$(YELLOW)    ⚠ AWS build skipped (missing dependencies)$(RESET)"
	@echo "  • GCP Provider..."
	@cd providers/gcp && go build -o ../../$(BUILD_DIR)/providers/gcp-provider . 2>/dev/null || echo "$(YELLOW)    ⚠ GCP build skipped (missing dependencies)$(RESET)"
	@echo "  • Azure Provider..."
	@cd providers/azure && go build -o ../../$(BUILD_DIR)/providers/azure-provider . 2>/dev/null || echo "$(YELLOW)    ⚠ Azure build skipped (missing dependencies)$(RESET)"
	@echo "  • OpenStack Provider..."
	@cd providers/openstack && go build -o ../../$(BUILD_DIR)/providers/openstack-provider . 2>/dev/null || echo "$(YELLOW)    ⚠ OpenStack build skipped (missing dependencies)$(RESET)"
	@echo "  • Proxmox Provider..."
	@cd providers/proxmox && go build -o ../../$(BUILD_DIR)/providers/proxmox-provider . 2>/dev/null || echo "$(YELLOW)    ⚠ Proxmox build skipped (missing dependencies)$(RESET)"
	@echo "$(GREEN)✓ Provider servers built$(RESET)"

build-provider-aws: ## AWS Provider 빌드
	@echo "$(YELLOW)▶ Building AWS Provider...$(RESET)"
	@mkdir -p $(BUILD_DIR)/providers
	@cd providers/aws && go build -o ../../$(BUILD_DIR)/providers/aws-provider .
	@echo "$(GREEN)✓ AWS Provider built$(RESET)"

build-provider-gcp: ## GCP Provider 빌드
	@echo "$(YELLOW)▶ Building GCP Provider...$(RESET)"
	@mkdir -p $(BUILD_DIR)/providers
	@cd providers/gcp && go build -o ../../$(BUILD_DIR)/providers/gcp-provider .
	@echo "$(GREEN)✓ GCP Provider built$(RESET)"

docker-build-providers: ## 모든 Provider Docker 이미지 빌드
	@echo "$(YELLOW)▶ Building Provider Docker images...$(RESET)"
	@docker build -t skyclust-aws-provider:latest ./providers/aws 2>/dev/null || echo "$(YELLOW)  ⚠ AWS image skipped$(RESET)"
	@docker build -t skyclust-gcp-provider:latest ./providers/gcp 2>/dev/null || echo "$(YELLOW)  ⚠ GCP image skipped$(RESET)"
	@docker build -t skyclust-azure-provider:latest ./providers/azure 2>/dev/null || echo "$(YELLOW)  ⚠ Azure image skipped$(RESET)"
	@docker build -t skyclust-openstack-provider:latest ./providers/openstack 2>/dev/null || echo "$(YELLOW)  ⚠ OpenStack image skipped$(RESET)"
	@docker build -t skyclust-proxmox-provider:latest ./providers/proxmox 2>/dev/null || echo "$(YELLOW)  ⚠ Proxmox image skipped$(RESET)"
	@echo "$(GREEN)✓ Provider Docker images built$(RESET)"

providers-up: ## Provider 서비스 시작
	@echo "$(YELLOW)▶ Starting Provider services...$(RESET)"
	@docker-compose -f docker-compose.providers.yml up -d
	@echo "$(GREEN)✓ Provider services started$(RESET)"

providers-down: ## Provider 서비스 중지
	@echo "$(YELLOW)▶ Stopping Provider services...$(RESET)"
	@docker-compose -f docker-compose.providers.yml down
	@echo "$(GREEN)✓ Provider services stopped$(RESET)"

providers-logs: ## Provider 서비스 로그
	@docker-compose -f docker-compose.providers.yml logs -f

providers-restart: providers-down providers-up ## Provider 서비스 재시작

# ============================================================================
# Development Targets
# ============================================================================

dev: ## 개발 모드 실행
	@echo "$(YELLOW)▶ Starting development server...$(RESET)"
	@echo "$(CYAN)  API: http://localhost:8081$(RESET)"
	@echo "$(CYAN)  Health: http://localhost:8081/health$(RESET)"
	@CONFIG_PATH=configs/config.dev.yaml go run ./cmd/server/main.go

run: build ## 빌드 후 실행
	@echo "$(YELLOW)▶ Running $(APP_NAME)...$(RESET)"
	@./$(BUILD_DIR)/$(BINARY_NAME)

run-config: build ## 설정 파일과 함께 실행
	@echo "$(YELLOW)▶ Running with config...$(RESET)"
	@CONFIG_PATH=configs/config.dev.yaml ./$(BUILD_DIR)/$(BINARY_NAME)

# ============================================================================
# Code Quality Targets
# ============================================================================

lint: ## Lint 실행
	@echo "$(YELLOW)▶ Running linter...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
		echo "$(GREEN)✓ Lint passed$(RESET)"; \
	else \
		echo "$(RED)✗ golangci-lint not installed$(RESET)"; \
		echo "$(YELLOW)  Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(RESET)"; \
		exit 1; \
	fi

format: ## 코드 포맷팅
	@echo "$(YELLOW)▶ Formatting code...$(RESET)"
	@go fmt ./...
	@go mod tidy
	@echo "$(GREEN)✓ Code formatted$(RESET)"

# ============================================================================
# Docker Targets
# ============================================================================

docker-build: ## Main 서버 Docker 이미지 빌드
	@echo "$(YELLOW)▶ Building Docker image...$(RESET)"
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(GREEN)✓ Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)$(RESET)"

docker-run: docker-build ## Docker 컨테이너 실행
	@echo "$(YELLOW)▶ Running Docker container...$(RESET)"
	@docker run --rm -p 8081:8081 $(DOCKER_IMAGE):$(DOCKER_TAG)

# ============================================================================
# Docker Compose Targets
# ============================================================================

compose-up: ## Docker Compose 시작 (전체 스택)
	@echo "$(YELLOW)▶ Starting services with docker-compose...$(RESET)"
	@docker-compose -f docker-compose.dev.yml up -d
	@echo "$(GREEN)✓ Services started$(RESET)"

compose-down: ## Docker Compose 중지
	@echo "$(YELLOW)▶ Stopping services...$(RESET)"
	@docker-compose -f docker-compose.dev.yml down
	@echo "$(GREEN)✓ Services stopped$(RESET)"

compose-logs: ## Docker Compose 로그
	@docker-compose -f docker-compose.dev.yml logs -f

compose-restart: compose-down compose-up ## Docker Compose 재시작

# ============================================================================
# Dependency Targets
# ============================================================================

deps: ## 의존성 다운로드
	@echo "$(YELLOW)▶ Downloading dependencies...$(RESET)"
	@go mod download
	@go mod verify
	@echo "$(GREEN)✓ Dependencies downloaded$(RESET)"

deps-update: ## 의존성 업데이트
	@echo "$(YELLOW)▶ Updating dependencies...$(RESET)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(RESET)"

deps-clean: ## 의존성 캐시 정리
	@echo "$(YELLOW)▶ Cleaning dependency cache...$(RESET)"
	@go clean -modcache
	@echo "$(GREEN)✓ Cache cleaned$(RESET)"

# ============================================================================
# Test Targets
# ============================================================================

test: ## 모든 테스트 실행
	@echo "$(YELLOW)▶ Running tests...$(RESET)"
	@go test ./... -v
	@echo "$(GREEN)✓ Tests completed$(RESET)"

test-unit: ## 단위 테스트 실행
	@echo "$(YELLOW)▶ Running unit tests...$(RESET)"
	@go test -v -short ./internal/...
	@echo "$(GREEN)✓ Unit tests completed$(RESET)"

test-coverage: ## 테스트 커버리지
	@echo "$(YELLOW)▶ Running tests with coverage...$(RESET)"
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage: coverage.html$(RESET)"

test-bench: ## 벤치마크 테스트
	@echo "$(YELLOW)▶ Running benchmark tests...$(RESET)"
	@go test -v -bench=. -benchmem ./...
	@echo "$(GREEN)✓ Benchmarks completed$(RESET)"

# ============================================================================
# Clean Targets
# ============================================================================

clean: ## 빌드 아티팩트 정리
	@echo "$(YELLOW)▶ Cleaning build artifacts...$(RESET)"
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Cleaned$(RESET)"

clean-all: clean proto-clean ## 모든 아티팩트 정리
	@echo "$(YELLOW)▶ Cleaning all artifacts...$(RESET)"
	@rm -rf vendor
	@echo "$(GREEN)✓ All cleaned$(RESET)"

clean-docker: ## Docker 이미지 정리
	@echo "$(YELLOW)▶ Cleaning Docker artifacts...$(RESET)"
	@docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true
	@docker system prune -f
	@echo "$(GREEN)✓ Docker cleaned$(RESET)"

# ============================================================================
# Setup Targets
# ============================================================================

setup: ## 프로젝트 초기 설정
	@echo "$(YELLOW)▶ Setting up project...$(RESET)"
	@mkdir -p $(BUILD_DIR)/providers
	@mkdir -p api/gen
	@echo "$(GREEN)✓ Project setup completed$(RESET)"

install-tools: ## 개발 도구 설치
	@echo "$(YELLOW)▶ Installing development tools...$(RESET)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/bufbuild/buf/cmd/buf@latest
	@go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	@echo "$(GREEN)✓ Tools installed$(RESET)"

# ============================================================================
# Info Targets
# ============================================================================

version: ## 버전 정보
	@echo "$(CYAN)App Name:$(RESET)        $(APP_NAME)"
	@echo "$(CYAN)Version:$(RESET)         $(VERSION)"
	@echo "$(CYAN)Commit:$(RESET)          $(COMMIT)"
	@echo "$(CYAN)Build Time:$(RESET)      $(BUILD_TIME)"
	@echo "$(CYAN)Go Version:$(RESET)      $(shell go version)"

status: ## 프로젝트 상태
	@echo "$(BLUE)╔═══════════════════════════════════════╗$(RESET)"
	@echo "$(BLUE)║  Project Status                       ║$(RESET)"
	@echo "$(BLUE)╚═══════════════════════════════════════╝$(RESET)"
	@echo "$(CYAN)Git Branch:$(RESET)      $(shell git branch --show-current 2>/dev/null || echo 'unknown')"
	@echo "$(CYAN)Modified Files:$(RESET)  $(shell git status --porcelain 2>/dev/null | wc -l | tr -d ' ')"
	@echo "$(CYAN)Go Version:$(RESET)      $(shell go version | cut -d' ' -f3)"
	@echo "$(CYAN)Dependencies:$(RESET)    $(shell go list -m all 2>/dev/null | wc -l | tr -d ' ') modules"

# ============================================================================
# Release Targets
# ============================================================================

release: clean deps lint test build-all-platforms ## 릴리스 준비
	@echo "$(GREEN)✓ Release ready$(RESET)"


# ============================================================================
# Quick Commands
# ============================================================================

quick-start: deps proto-gen build ## 빠른 시작 (deps + proto + build)
	@echo "$(GREEN)✓ Quick start completed$(RESET)"
	@echo "$(CYAN)  Run: make dev$(RESET)"

full-build: clean deps proto-gen lint build build-providers ## 전체 빌드
	@echo "$(GREEN)✓ Full build completed$(RESET)"

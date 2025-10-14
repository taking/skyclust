# SkyClust 설정 가이드

이 문서는 SkyClust 애플리케이션의 설정 방법과 옵션들을 설명합니다.

## 설정 개요

SkyClust는 다양한 설정 방법을 지원합니다:

1. **환경 변수** (최우선)
2. **설정 파일** (YAML)
3. **명령행 인수**
4. **기본값**

## 환경 변수 설정

### 필수 환경 변수

```bash
# 데이터베이스 설정
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=skyclust
export DB_PASSWORD=password
export DB_NAME=skyclust
export DB_SSLMODE=disable

# Redis 설정
export REDIS_URL=redis://localhost:6379

# JWT 설정
export JWT_SECRET=your-jwt-secret-key-here
export JWT_EXPIRY=24h

# 암호화 설정
export ENCRYPTION_KEY=your-32-byte-encryption-key-here

# 서버 설정
export SERVER_PORT=8080
export SERVER_HOST=0.0.0.0
```

### 선택적 환경 변수

```bash
# 로깅 설정
export LOG_LEVEL=info
export LOG_FORMAT=json

# 캐시 설정
export CACHE_TTL=3600
export CACHE_MAX_SIZE=1000

# 메시징 설정
export NATS_URL=nats://localhost:4222
export NATS_CLUSTER=skyclust-cluster

# 모니터링 설정
export METRICS_ENABLED=true
export METRICS_PORT=9090

# 보안 설정
export CORS_ORIGINS=http://localhost:3000
export RATE_LIMIT=1000
```

## 설정 파일 (YAML)

### 기본 설정 파일

```yaml
# config.yaml
server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s

database:
  host: "localhost"
  port: 5432
  user: "skyclust"
  password: "password"
  name: "skyclust"
  sslmode: "disable"
  max_open_conns: 25
  max_idle_conns: 25
  conn_max_lifetime: 5m

redis:
  url: "redis://localhost:6379"
  password: ""
  db: 0
  pool_size: 10
  min_idle_conns: 5

jwt:
  secret: "your-jwt-secret-key-here"
  expiry: "24h"
  issuer: "skyclust"

encryption:
  key: "your-32-byte-encryption-key-here"
  algorithm: "aes-256-gcm"

logging:
  level: "info"
  format: "json"
  output: "stdout"
  file: ""

cache:
  ttl: 3600
  max_size: 1000
  cleanup_interval: 300s

messaging:
  nats:
    url: "nats://localhost:4222"
    cluster: "skyclust-cluster"
    max_reconnects: 5
    reconnect_wait: 2s

monitoring:
  enabled: true
  port: 9090
  path: "/metrics"

security:
  cors:
    origins: ["http://localhost:3000"]
    methods: ["GET", "POST", "PUT", "DELETE"]
    headers: ["Content-Type", "Authorization"]
  rate_limit:
    requests: 1000
    window: "1h"
  password_policy:
    min_length: 8
    require_uppercase: true
    require_lowercase: true
    require_numbers: true
    require_symbols: false

plugins:
  directory: "plugins"
  auto_load: true
  timeout: 30s

cloud_providers:
  aws:
    enabled: true
    default_region: "us-east-1"
    max_retries: 3
    timeout: 30s
  gcp:
    enabled: true
    default_project: ""
    max_retries: 3
    timeout: 30s
  azure:
    enabled: true
    default_subscription: ""
    max_retries: 3
    timeout: 30s
```

### 환경별 설정 파일

#### 개발 환경 (config.dev.yaml)

```yaml
server:
  port: 8080
  host: "localhost"

database:
  host: "localhost"
  sslmode: "disable"

logging:
  level: "debug"
  format: "console"

security:
  cors:
    origins: ["http://localhost:3000", "http://localhost:3001"]
```

#### 프로덕션 환경 (config.prod.yaml)

```yaml
server:
  port: 8080
  host: "0.0.0.0"

database:
  sslmode: "require"
  max_open_conns: 100
  max_idle_conns: 50

logging:
  level: "info"
  format: "json"
  file: "/var/log/skyclust/app.log"

security:
  cors:
    origins: ["https://skyclust.com"]
  rate_limit:
    requests: 10000
    window: "1h"
```

## 클라우드 프로바이더 설정

### AWS 설정

```bash
# AWS 자격증명
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_REGION=us-east-1

# 또는 AWS 프로파일 사용
export AWS_PROFILE=skyclust
```

### GCP 설정

```bash
# 서비스 계정 키 파일
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json
export GCP_PROJECT_ID=your-project-id
export GCP_REGION=us-central1
```

### Azure 설정

```bash
# Azure 자격증명
export AZURE_CLIENT_ID=your-client-id
export AZURE_CLIENT_SECRET=your-client-secret
export AZURE_TENANT_ID=your-tenant-id
export AZURE_SUBSCRIPTION_ID=your-subscription-id
```

## 데이터베이스 설정

### PostgreSQL 설정

```sql
-- 데이터베이스 생성
CREATE DATABASE skyclust;

-- 사용자 생성
CREATE USER skyclust WITH PASSWORD 'password';

-- 권한 부여
GRANT ALL PRIVILEGES ON DATABASE skyclust TO skyclust;

-- 확장 기능 설치
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";
```

### 연결 풀 설정

```yaml
database:
  max_open_conns: 25      # 최대 연결 수
  max_idle_conns: 25      # 최대 유휴 연결 수
  conn_max_lifetime: 5m   # 연결 최대 수명
  conn_max_idle_time: 1m  # 유휴 연결 최대 시간
```

## Redis 설정

### 기본 설정

```yaml
redis:
  url: "redis://localhost:6379"
  password: ""
  db: 0
  pool_size: 10
  min_idle_conns: 5
  max_retries: 3
  dial_timeout: 5s
  read_timeout: 3s
  write_timeout: 3s
```

### 클러스터 설정

```yaml
redis:
  cluster:
    nodes:
      - "redis://node1:6379"
      - "redis://node2:6379"
      - "redis://node3:6379"
    password: ""
    pool_size: 10
```

## 로깅 설정

### 로그 레벨

```yaml
logging:
  level: "info"  # debug, info, warn, error
  format: "json" # json, console
  output: "stdout" # stdout, stderr, file
  file: "/var/log/skyclust/app.log"
```

### 구조화된 로깅

```go
// 로그 필드 추가
logger.Info("User action",
    zap.String("user_id", userID),
    zap.String("action", "create_vm"),
    zap.String("resource_id", resourceID),
    zap.Duration("duration", time.Since(start)),
)
```

## 보안 설정

### CORS 설정

```yaml
security:
  cors:
    origins: ["https://skyclust.com", "https://app.skyclust.com"]
    methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    headers: ["Content-Type", "Authorization", "X-Requested-With"]
    credentials: true
    max_age: 86400
```

### Rate Limiting

```yaml
security:
  rate_limit:
    requests: 1000
    window: "1h"
    burst: 100
    skip_successful: false
    skip_failed: false
```

### 비밀번호 정책

```yaml
security:
  password_policy:
    min_length: 8
    require_uppercase: true
    require_lowercase: true
    require_numbers: true
    require_symbols: false
    max_length: 128
    history_count: 5
```

## 모니터링 설정

### 메트릭 설정

```yaml
monitoring:
  enabled: true
  port: 9090
  path: "/metrics"
  namespace: "skyclust"
  subsystem: "api"
```

### 헬스 체크

```yaml
health:
  enabled: true
  path: "/health"
  timeout: 5s
  interval: 30s
  checks:
    database: true
    redis: true
    nats: true
```

## 플러그인 설정

### 플러그인 디렉토리

```yaml
plugins:
  directory: "plugins"
  auto_load: true
  timeout: 30s
  max_plugins: 50
```

### 플러그인별 설정

```yaml
plugins:
  aws:
    enabled: true
    config:
      region: "us-east-1"
      max_retries: 3
      timeout: 30s
  gcp:
    enabled: true
    config:
      project_id: "your-project"
      region: "us-central1"
```

## Docker 설정

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_URL=redis://redis:6379
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=skyclust
      - POSTGRES_USER=skyclust
      - POSTGRES_PASSWORD=password
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

### 환경 변수 파일

```bash
# .env
DB_HOST=postgres
DB_PORT=5432
DB_USER=skyclust
DB_PASSWORD=password
DB_NAME=skyclust
REDIS_URL=redis://redis:6379
JWT_SECRET=your-jwt-secret
ENCRYPTION_KEY=your-encryption-key
```

## 설정 검증

### 설정 파일 검증

```bash
# 설정 파일 문법 검사
go run cmd/server/main.go --config config.yaml --validate

# 환경 변수 검증
go run cmd/server/main.go --validate-env
```

### 설정 테스트

```bash
# 설정 로드 테스트
go run cmd/server/main.go --config config.yaml --dry-run

# 데이터베이스 연결 테스트
go run cmd/server/main.go --test-db

# Redis 연결 테스트
go run cmd/server/main.go --test-redis
```

## 설정 우선순위

1. **명령행 인수** (최우선)
2. **환경 변수**
3. **설정 파일**
4. **기본값** (최하위)

### 예시

```bash
# 명령행 인수가 최우선
go run cmd/server/main.go --port 8081

# 환경 변수가 설정 파일보다 우선
export SERVER_PORT=8082
go run cmd/server/main.go --config config.yaml

# 설정 파일의 기본값
# config.yaml에서 port: 8080
```

## 설정 모범 사례

### 1. 보안

- **프로덕션 환경**: 환경 변수 사용
- **민감한 정보**: 암호화하여 저장
- **기본값**: 보안이 강화된 기본값 사용

### 2. 성능

- **연결 풀**: 적절한 크기 설정
- **캐시**: TTL과 크기 조정
- **로깅**: 프로덕션에서는 info 레벨

### 3. 모니터링

- **메트릭**: 모든 주요 지표 수집
- **로그**: 구조화된 로그 사용
- **알림**: 임계값 설정

### 4. 확장성

- **설정 분리**: 환경별 설정 파일
- **동적 설정**: 런타임 설정 변경
- **설정 검증**: 시작 시 설정 검증

## 문제 해결

### 일반적인 설정 오류

#### 데이터베이스 연결 실패
```bash
# 연결 정보 확인
echo $DB_HOST $DB_PORT $DB_USER $DB_NAME

# 연결 테스트
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;"
```

#### Redis 연결 실패
```bash
# Redis URL 확인
echo $REDIS_URL

# 연결 테스트
redis-cli -u $REDIS_URL ping
```

#### JWT 토큰 오류
```bash
# JWT 시크릿 확인
echo $JWT_SECRET

# 토큰 검증
go run cmd/server/main.go --validate-jwt
```

### 설정 디버깅

```bash
# 설정 값 출력
go run cmd/server/main.go --config config.yaml --print-config

# 환경 변수 출력
go run cmd/server/main.go --print-env

# 설정 검증
go run cmd/server/main.go --config config.yaml --validate
```
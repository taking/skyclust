# SkyClust - 멀티 클라우드 관리 플랫폼

SkyClust는 플러그인 기반 아키텍처를 통해 여러 클라우드 프로바이더를 통합 관리하는 포괄적인 멀티 클라우드 관리 플랫폼입니다. Go와 React로 구축되어 워크스페이스 관리, VM 오케스트레이션, 비용 분석, 실시간 모니터링 등 엔터프라이즈급 기능을 제공합니다.

## 주요 기능

### 핵심 기능
- **멀티 클라우드 지원**: AWS, GCP, Azure, OpenStack, Proxmox VE
- **플러그인 아키텍처**: 클라우드 프로바이더 플러그인의 동적 로딩
- **워크스페이스 관리**: 멀티 테넌트 워크스페이스 격리
- **VM 오케스트레이션**: 프로바이더 간 인스턴스 생명주기 관리
- **비용 분석**: 실시간 비용 추적 및 최적화
- **실시간 모니터링**: WebSocket/SSE 기반 라이브 업데이트
- **인프라 코드**: IaC 관리를 위한 OpenTofu 통합
- **Kubernetes 관리**: 클러스터 및 리소스 관리
- **자격증명 관리**: 암호화된 클라우드 자격증명 저장
- **감사 로깅**: 포괄적인 활동 추적

### 엔터프라이즈 기능
- **RBAC**: 역할 기반 접근 제어
- **감사 추적**: 완전한 활동 로깅
- **성능 최적화**: 쿼리 최적화 및 캐싱
- **구조화된 로깅**: 포괄적인 로깅 시스템
- **입력 검증**: 강화된 보안 검증
- **API 문서화**: OpenAPI/Swagger 문서

## 아키텍처

### 백엔드 (Go)
- **프레임워크**: Gin (HTTP), GORM (ORM)
- **데이터베이스**: 쿼리 최적화가 적용된 PostgreSQL
- **메시징**: NATS (이벤트 버스)
- **인증**: RBAC가 적용된 JWT 기반
- **암호화**: 민감한 데이터를 위한 AES 암호화
- **로깅**: Zap을 사용한 구조화된 로깅
- **모니터링**: 성능 추적 및 메트릭

### 프론트엔드 (React + TypeScript)
- **프레임워크**: App Router가 적용된 Next.js 14
- **UI 라이브러리**: shadcn/ui 컴포넌트가 적용된 Tailwind CSS
- **상태 관리**: 서버 상태를 위한 React Query
- **HTTP 클라이언트**: 인터셉터가 적용된 Axios
- **실시간**: WebSocket 및 SSE 통합
- **접근성**: WCAG 2.1 준수 컴포넌트

### 인프라
- **컨테이너화**: Docker + Docker Compose
- **데이터베이스**: 최적화가 적용된 PostgreSQL 15
- **메시징**: NATS 2.10
- **캐싱**: 세션 및 쿼리 캐싱을 위한 Redis
- **모니터링**: 내장 성능 메트릭

## 프로젝트 구조

```
skyclust/
├── cmd/server/              # 애플리케이션 진입점
├── internal/
│   ├── api/                 # HTTP 핸들러 및 라우트
│   │   ├── common/          # 공유 유틸리티 (로깅, 검증, 최적화)
│   │   ├── auth/            # 인증 핸들러
│   │   ├── workspace/       # 워크스페이스 관리
│   │   ├── notification/     # 알림 시스템
│   │   ├── credential/       # 자격증명 관리
│   │   ├── audit/            # 감사 로깅
│   │   └── ...              # 기타 API 모듈
│   ├── domain/              # 비즈니스 로직 및 엔티티
│   ├── usecase/             # 애플리케이션 유스케이스
│   ├── repository/           # 데이터 접근 계층
│   ├── di/                  # 의존성 주입
│   └── utils/               # 유틸리티 함수
├── pkg/                     # 공유 패키지
│   ├── cache/               # 캐싱 유틸리티
│   ├── config/              # 설정 관리
│   ├── middleware/          # HTTP 미들웨어
│   ├── plugin/              # 플러그인 인터페이스
│   ├── security/            # 보안 유틸리티
│   └── telemetry/           # 모니터링 및 메트릭
├── frontend/                # React 프론트엔드 애플리케이션
├── plugins/                 # 클라우드 프로바이더 플러그인
│   ├── public/              # 공개 클라우드 프로바이더
│   └── private/             # 프라이빗 클라우드 프로바이더
├── docs/                    # 문서
└── examples/                # 사용 예제
```

## 빠른 시작

### 필수 요구사항

- Go 1.21 이상
- Node.js 18 이상
- PostgreSQL 15 이상
- Redis 7 이상
- Docker & Docker Compose

### 설치

1. **저장소 클론:**
```bash
git clone <repository-url>
cd skyclust
```

2. **의존성 설치:**
```bash
# 백엔드
go mod tidy

# 프론트엔드
cd frontend
npm install
```

3. **애플리케이션 설정:**
```bash
# 환경 변수 파일 복사
cp .env.example .env

# 설정 편집
nano .env
```

4. **Docker Compose로 실행:**
```bash
# 개발 환경
docker-compose -f docker-compose.dev.yml up -d

# 프로덕션 환경
docker-compose up -d
```

### 수동 설정

1. **의존성 시작:**
```bash
# PostgreSQL
docker run -d --name postgres -e POSTGRES_PASSWORD=password -p 5432:5432 postgres:15

# Redis
docker run -d --name redis -p 6379:6379 redis:7
```

2. **백엔드 실행:**
```bash
go run cmd/server/main.go
```

3. **프론트엔드 실행:**
```bash
cd frontend
npm run dev
```

## API 문서

### 인증
```bash
# 사용자 등록
POST /api/v1/auth/register
{
  "username": "user",
  "email": "user@example.com",
  "password": "password"
}

# 로그인
POST /api/v1/auth/login
{
  "email": "user@example.com",
  "password": "password"
}
```

### 워크스페이스 관리
```bash
# 워크스페이스 목록 조회
GET /api/v1/workspaces

# 워크스페이스 생성
POST /api/v1/workspaces
{
  "name": "My Workspace",
  "description": "Workspace description"
}
```

### VM 관리
```bash
# VM 목록 조회
GET /api/v1/vms

# VM 생성
POST /api/v1/vms
{
  "name": "my-vm",
  "provider": "aws",
  "instance_type": "t3.micro",
  "region": "us-east-1"
}
```

### 비용 분석
```bash
# 비용 분석 조회
GET /api/v1/cost-analysis?period=30d

# 비용 세부 분석
GET /api/v1/cost-analysis/breakdown?provider=aws
```

## 설정

### 환경 변수

| 변수 | 설명 | 기본값 |
|------|------|--------|
| `DB_HOST` | 데이터베이스 호스트 | `localhost` |
| `DB_PORT` | 데이터베이스 포트 | `5432` |
| `DB_USER` | 데이터베이스 사용자 | `skyclust` |
| `DB_PASSWORD` | 데이터베이스 비밀번호 | `password` |
| `DB_NAME` | 데이터베이스 이름 | `skyclust` |
| `REDIS_URL` | Redis 연결 URL | `redis://localhost:6379` |
| `JWT_SECRET` | JWT 서명 시크릿 | 생성됨 |
| `ENCRYPTION_KEY` | 데이터 암호화 키 | 생성됨 |

### 클라우드 프로바이더 설정

#### AWS
```bash
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_REGION=us-east-1
```

#### GCP
```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json
export GCP_PROJECT_ID=your-project-id
```

#### Azure
```bash
export AZURE_CLIENT_ID=your-client-id
export AZURE_CLIENT_SECRET=your-client-secret
export AZURE_TENANT_ID=your-tenant-id
```

## 개발

### 테스트 실행
```bash
# 단위 테스트
make test-unit

# 통합 테스트
make test-integration

# 커버리지 리포트
make test-coverage

# 벤치마크 테스트
make test-benchmark
```

### 코드 품질
```bash
# 코드 포맷팅
make fmt

# 코드 린팅
make lint

# 보안 스캔
make security
```

### 빌드
```bash
# 백엔드 빌드
make build

# 프론트엔드 빌드
make build-frontend

# 전체 빌드
make build-all
```

## 플러그인 개발

### 새 플러그인 생성

1. **플러그인 디렉토리 생성:**
```bash
mkdir plugins/your-provider
cd plugins/your-provider
```

2. **플러그인 인터페이스 구현:**
```go
package main

import "skyclust/pkg/plugin"

type YourProvider struct {
    config map[string]interface{}
}

func New() plugin.CloudProvider {
    return &YourProvider{}
}

func (p *YourProvider) GetName() string {
    return "YourProvider"
}

func (p *YourProvider) Initialize(config map[string]interface{}) error {
    p.config = config
    return nil
}

// 다른 필수 메서드들 구현...
```

3. **플러그인 빌드:**
```bash
go build -buildmode=plugin -o your-provider.so main.go
```

## 모니터링 및 관찰성

### 메트릭
- 요청 지연시간 및 처리량
- 데이터베이스 쿼리 성능
- 메모리 및 CPU 사용량
- 에러율 및 성공률

### 로깅
- 구조화된 JSON 로깅
- 요청/응답 로깅
- 에러 추적
- 감사 추적 로깅

### 헬스 체크
```bash
# 애플리케이션 헬스
GET /health

# 데이터베이스 헬스
GET /health/db

# Redis 헬스
GET /health/redis
```

## 보안

### 인증
- JWT 기반 인증
- 역할 기반 접근 제어 (RBAC)
- 세션 관리
- 비밀번호 암호화

### 데이터 보호
- 민감한 데이터를 위한 AES 암호화
- 안전한 자격증명 저장
- 입력 검증 및 살균
- SQL 인젝션 방지

### 감사 로깅
- 완전한 활동 추적
- 보안 이벤트 로깅
- 데이터 접근 로깅
- 컴플라이언스 보고

## 기여하기

1. 저장소 포크
2. 기능 브랜치 생성 (`git checkout -b feature/amazing-feature`)
3. 변경사항 커밋 (`git commit -m 'Add amazing feature'`)
4. 브랜치에 푸시 (`git push origin feature/amazing-feature`)
5. Pull Request 생성

## 라이선스

이 프로젝트는 MIT 라이선스 하에 있습니다. 자세한 내용은 [LICENSE](LICENSE) 파일을 참조하세요.

## 지원

- 문서: [docs/](docs/)
- 이슈: [GitHub Issues](https://github.com/taking/skyclust/issues)
- 토론: [GitHub Discussions](https://github.com/taking/skyclust/discussions)

## 로드맵

- [ ] 추가 클라우드 프로바이더 (DigitalOcean, Linode)
- [ ] 고급 비용 최적화
- [ ] 멀티 리전 배포
- [ ] 플러그인 마켓플레이스
- [ ] 고급 모니터링 및 알림
- [ ] 인프라 템플릿
- [ ] 컴플라이언스 프레임워크 (SOC2, ISO27001)
# SkyClust 개발 가이드

이 문서는 SkyClust 프로젝트의 개발 환경 설정, 코딩 규칙, 테스트 방법 등을 설명합니다.

## 개발 환경 설정

### 필수 요구사항

- **Go**: 1.21 이상
- **Node.js**: 18 이상
- **PostgreSQL**: 15 이상
- **Redis**: 7 이상
- **Docker & Docker Compose**: 최신 버전

### 개발 환경 구성

#### 1. 저장소 클론

```bash
git clone <repository-url>
cd skyclust
```

#### 2. 백엔드 의존성 설치

```bash
go mod tidy
```

#### 3. 프론트엔드 의존성 설치

```bash
cd frontend
npm install
```

#### 4. 환경 변수 설정

```bash
# 환경 변수 파일 복사
cp .env.example .env

# 환경 변수 편집
nano .env
```

#### 5. 데이터베이스 설정

```bash
# PostgreSQL 컨테이너 실행
docker run -d --name postgres-dev \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=skyclust \
  -p 5432:5432 \
  postgres:15

# Redis 컨테이너 실행
docker run -d --name redis-dev \
  -p 6379:6379 \
  redis:7
```

## 프로젝트 구조

```
skyclust/
├── cmd/                    # 애플리케이션 진입점
│   └── server/
│       └── main.go
├── internal/               # 내부 패키지
│   ├── api/               # HTTP 핸들러
│   │   ├── common/        # 공통 유틸리티
│   │   ├── auth/          # 인증 관련
│   │   ├── workspace/     # 워크스페이스 관리
│   │   ├── notification/   # 알림 시스템
│   │   ├── credential/     # 자격증명 관리
│   │   └── audit/          # 감사 로그
│   ├── domain/            # 도메인 로직
│   ├── usecase/           # 애플리케이션 유스케이스
│   ├── repository/         # 데이터 접근 계층
│   ├── di/                # 의존성 주입
│   └── utils/             # 유틸리티 함수
├── pkg/                   # 공유 패키지
│   ├── cache/             # 캐싱 유틸리티
│   ├── config/            # 설정 관리
│   ├── middleware/         # HTTP 미들웨어
│   ├── plugin/            # 플러그인 인터페이스
│   ├── security/          # 보안 유틸리티
│   └── telemetry/         # 모니터링 및 메트릭
├── frontend/              # React 프론트엔드
├── plugins/               # 클라우드 프로바이더 플러그인
├── docs/                  # 문서
└── examples/              # 사용 예제
```

## 코딩 규칙

### Go 코딩 규칙

#### 1. 패키지 구조

- **Clean Architecture** 원칙 준수
- **Domain-Driven Design** 적용
- **의존성 역전** 원칙 적용

#### 2. 네이밍 규칙

```go
// 패키지명: 소문자, 단어 구분 없음
package auth

// 타입명: PascalCase
type UserHandler struct{}

// 함수명: PascalCase (public), camelCase (private)
func (h *UserHandler) CreateUser() {}
func (h *UserHandler) validateInput() {}

// 상수: PascalCase
const MaxRetries = 3

// 변수: camelCase
var userCount int
```

#### 3. 에러 처리

```go
// 도메인 에러 사용
if err != nil {
    if domain.IsDomainError(err) {
        return domain.GetDomainError(err)
    }
    return domain.NewDomainError(
        domain.ErrCodeInternalError,
        "Failed to process request",
        http.StatusInternalServerError,
    )
}
```

#### 4. 로깅

```go
// 구조화된 로깅 사용
logger.Info("User created successfully",
    zap.String("user_id", user.ID.String()),
    zap.String("username", user.Username),
    zap.Duration("duration", time.Since(start)),
)
```

#### 5. 테스트

```go
func TestCreateUser(t *testing.T) {
    // Given
    mockService := new(MockUserService)
    handler := NewHandler(mockService)
    
    // When
    result, err := handler.CreateUser(createUserRequest)
    
    // Then
    assert.NoError(t, err)
    assert.NotNil(t, result)
    mockService.AssertExpectations(t)
}
```

### React/TypeScript 코딩 규칙

#### 1. 컴포넌트 구조

```typescript
// 컴포넌트 파일 구조
interface ComponentProps {
  title: string;
  onClose: () => void;
}

export const Component: React.FC<ComponentProps> = ({ title, onClose }) => {
  // 훅 사용
  const [state, setState] = useState<string>('');
  
  // 이벤트 핸들러
  const handleClick = useCallback(() => {
    onClose();
  }, [onClose]);
  
  return (
    <div>
      <h1>{title}</h1>
      <button onClick={handleClick}>Close</button>
    </div>
  );
};
```

#### 2. 타입 정의

```typescript
// API 응답 타입
interface APIResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

// 도메인 타입
interface User {
  id: string;
  username: string;
  email: string;
  createdAt: string;
}
```

#### 3. 훅 사용

```typescript
// 커스텀 훅
export const useUsers = () => {
  const { data, isLoading, error } = useQuery({
    queryKey: ['users'],
    queryFn: () => api.users.list(),
  });
  
  return { users: data, isLoading, error };
};
```

## 개발 워크플로우

### 1. 브랜치 전략

```
main
├── develop
│   ├── feature/user-authentication
│   ├── feature/vm-management
│   └── bugfix/login-error
└── release/v1.0.0
```

### 2. 커밋 메시지 규칙

```
<type>: <short description>

<optional detailed description>

Examples:
feat: add user authentication
fix: resolve login validation error
docs: update API documentation
refactor: improve error handling
test: add unit tests for user service
```

### 3. Pull Request 규칙

- **제목**: 명확하고 간결한 설명
- **설명**: 변경 사항과 이유 설명
- **테스트**: 테스트 케이스 포함
- **리뷰**: 최소 1명 이상의 리뷰 필요

## 테스트

### 1. 단위 테스트

```bash
# 모든 단위 테스트 실행
make test-unit

# 특정 패키지 테스트
go test ./internal/api/auth/...

# 커버리지 포함 테스트
make test-coverage
```

### 2. 통합 테스트

```bash
# 통합 테스트 실행
make test-integration

# 데이터베이스 테스트
make test-db
```

### 3. E2E 테스트

```bash
# 프론트엔드 E2E 테스트
cd frontend
npm run test:e2e
```

### 4. 벤치마크 테스트

```bash
# 벤치마크 테스트 실행
make test-benchmark
```

## 코드 품질

### 1. 린팅

```bash
# Go 코드 린팅
make lint

# 프론트엔드 린팅
cd frontend
npm run lint
```

### 2. 포맷팅

```bash
# Go 코드 포맷팅
make fmt

# 프론트엔드 포맷팅
cd frontend
npm run format
```

### 3. 보안 검사

```bash
# 보안 취약점 검사
make security

# 의존성 검사
make deps-check
```

## 빌드 및 배포

### 1. 개발 빌드

```bash
# 백엔드 빌드
make build

# 프론트엔드 빌드
make build-frontend

# 전체 빌드
make build-all
```

### 2. Docker 빌드

```bash
# 개발 이미지 빌드
docker-compose -f docker-compose.dev.yml build

# 프로덕션 이미지 빌드
docker build -t skyclust:latest .
```

### 3. 배포

```bash
# 개발 환경 배포
docker-compose -f docker-compose.dev.yml up -d

# 프로덕션 환경 배포
docker-compose up -d
```

## 디버깅

### 1. 로그 확인

```bash
# 애플리케이션 로그
docker logs skyclust-server

# 데이터베이스 로그
docker logs postgres-dev

# Redis 로그
docker logs redis-dev
```

### 2. 데이터베이스 디버깅

```bash
# PostgreSQL 연결
docker exec -it postgres-dev psql -U skyclust -d skyclust

# 테이블 확인
\dt

# 쿼리 실행
SELECT * FROM users LIMIT 10;
```

### 3. 성능 프로파일링

```bash
# CPU 프로파일링
go tool pprof http://localhost:8080/debug/pprof/profile

# 메모리 프로파일링
go tool pprof http://localhost:8080/debug/pprof/heap
```

## 모니터링

### 1. 메트릭 확인

```bash
# 애플리케이션 메트릭
curl http://localhost:8080/metrics

# 헬스 체크
curl http://localhost:8080/health
```

### 2. 로그 모니터링

```bash
# 실시간 로그 확인
docker logs -f skyclust-server

# 특정 로그 필터링
docker logs skyclust-server | grep ERROR
```

## 문제 해결

### 1. 일반적인 문제

#### 데이터베이스 연결 오류
```bash
# 데이터베이스 상태 확인
docker ps | grep postgres

# 연결 테스트
docker exec -it postgres-dev psql -U skyclust -d skyclust -c "SELECT 1;"
```

#### Redis 연결 오류
```bash
# Redis 상태 확인
docker ps | grep redis

# 연결 테스트
docker exec -it redis-dev redis-cli ping
```

#### 포트 충돌
```bash
# 포트 사용 확인
netstat -tulpn | grep :8080

# 프로세스 종료
kill -9 <PID>
```

### 2. 성능 문제

#### 느린 쿼리
```bash
# 쿼리 최적화 확인
docker exec -it postgres-dev psql -U skyclust -d skyclust -c "
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;"
```

#### 메모리 사용량
```bash
# 메모리 사용량 확인
docker stats skyclust-server

# 가비지 컬렉션 강제 실행
curl http://localhost:8080/debug/pprof/heap?gc=1
```

## 개발 도구

### 1. IDE 설정

#### VS Code
```json
{
  "go.toolsManagement.checkForUpdates": "local",
  "go.useLanguageServer": true,
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint"
}
```

#### GoLand
- Go 모듈 지원 활성화
- 자동 포맷팅 설정
- 린팅 도구 설정

### 2. 유용한 확장 프로그램

- **Go**: Go 언어 지원
- **Prettier**: 코드 포맷팅
- **ESLint**: JavaScript/TypeScript 린팅
- **Docker**: Docker 지원
- **GitLens**: Git 통합

### 3. 개발 스크립트

```bash
# 개발 서버 시작
make dev

# 테스트 실행
make test

# 코드 품질 검사
make quality

# 전체 빌드
make build-all
```

## 기여 가이드

### 1. 이슈 생성

- **버그 리포트**: 명확한 재현 단계 포함
- **기능 요청**: 사용 사례와 이점 설명
- **문서 개선**: 구체적인 개선 사항 제시

### 2. Pull Request

- **작은 단위**: 하나의 기능이나 버그 수정
- **테스트 포함**: 새로운 기능에 대한 테스트
- **문서 업데이트**: 필요한 경우 문서 수정
- **리뷰 요청**: 적절한 리뷰어 지정

### 3. 코드 리뷰

- **구조적 검토**: 아키텍처와 설계 검토
- **코드 품질**: 가독성과 유지보수성 검토
- **성능 검토**: 성능에 미치는 영향 검토
- **보안 검토**: 보안 취약점 검토

## 참고 자료

- [Go 공식 문서](https://golang.org/doc/)
- [React 공식 문서](https://reactjs.org/docs/)
- [Next.js 공식 문서](https://nextjs.org/docs)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)

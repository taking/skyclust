# SkyClust 서비스 가이드

이 문서는 SkyClust 프로젝트의 각 서비스별 사용법을 설명합니다.

## 📋 목차

- [인증 서비스 (Auth Service)](#인증-서비스-auth-service)
- [사용자 서비스 (User Service)](#사용자-서비스-user-service)
- [워크스페이스 서비스 (Workspace Service)](#워크스페이스-서비스-workspace-service)
- [자격증명 서비스 (Credential Service)](#자격증명-서비스-credential-service)
- [VM 서비스 (VM Service)](#vm-서비스-vm-service)
- [클라우드 프로바이더 서비스 (Cloud Provider Service)](#클라우드-프로바이더-서비스-cloud-provider-service)
- [IaC 서비스 (Infrastructure as Code Service)](#iac-서비스-infrastructure-as-code-service)
- [캐시 서비스 (Cache Service)](#캐시-서비스-cache-service)
- [이벤트 서비스 (Event Service)](#이벤트-서비스-event-service)
- [OIDC 서비스 (OpenID Connect Service)](#oidc-서비스-openid-connect-service)

---

## 인증 서비스 (Auth Service)

JWT 기반 인증을 제공하는 서비스입니다.

### 주요 기능
- 사용자 회원가입
- 사용자 로그인
- JWT 토큰 검증
- 사용자 로그아웃

### 사용법

```go
// 서비스 초기화
authService := usecase.NewAuthService(
    userRepo,
    auditLogRepo,
    jwtSecret,
    tokenExpiry,
)

// 사용자 회원가입
user, err := authService.Register(registerReq)
if err != nil {
    // 에러 처리
}

// 사용자 로그인
token, err := authService.Login(loginReq)
if err != nil {
    // 에러 처리
}

// 토큰 검증
userID, err := authService.ValidateToken(token)
if err != nil {
    // 에러 처리
}
```

### API 엔드포인트
- `POST /api/v1/auth/register` - 회원가입
- `POST /api/v1/auth/login` - 로그인
- `POST /api/v1/auth/logout` - 로그아웃

---

## 사용자 서비스 (User Service)

사용자 관리 기능을 제공하는 서비스입니다.

### 주요 기능
- 사용자 생성
- 사용자 조회
- 사용자 정보 수정
- 사용자 삭제

### 사용법

```go
// 서비스 초기화
userService := usecase.NewUserService(
    userRepo,
    auditLogRepo,
)

// 사용자 생성
user, err := userService.CreateUser(createReq)
if err != nil {
    // 에러 처리
}

// 사용자 조회
user, err := userService.GetUser(userID)
if err != nil {
    // 에러 처리
}
```

### API 엔드포인트
- `GET /api/v1/users/:id` - 사용자 조회
- `PUT /api/v1/users/:id` - 사용자 정보 수정
- `DELETE /api/v1/users/:id` - 사용자 삭제

---

## 워크스페이스 서비스 (Workspace Service)

워크스페이스 관리 기능을 제공하는 서비스입니다.

### 주요 기능
- 워크스페이스 생성
- 워크스페이스 조회
- 워크스페이스 수정
- 워크스페이스 삭제
- 사용자 멤버십 관리

### 사용법

```go
// 서비스 초기화
workspaceService := usecase.NewWorkspaceService(
    workspaceRepo,
    userRepo,
    eventBus,
    auditLogRepo,
)

// 워크스페이스 생성
workspace, err := workspaceService.CreateWorkspace(createReq)
if err != nil {
    // 에러 처리
}

// 사용자를 워크스페이스에 추가
err = workspaceService.AddUserToWorkspace(workspaceID, userID)
if err != nil {
    // 에러 처리
}
```

### API 엔드포인트
- `POST /api/v1/workspaces` - 워크스페이스 생성
- `GET /api/v1/workspaces/:id` - 워크스페이스 조회
- `PUT /api/v1/workspaces/:id` - 워크스페이스 수정
- `DELETE /api/v1/workspaces/:id` - 워크스페이스 삭제
- `POST /api/v1/workspaces/:id/members` - 멤버 추가
- `DELETE /api/v1/workspaces/:id/members/:userID` - 멤버 제거

---

## 자격증명 서비스 (Credential Service)

클라우드 자격증명 관리 기능을 제공하는 서비스입니다.

### 주요 기능
- 자격증명 생성 (암호화 저장)
- 자격증명 조회
- 자격증명 수정
- 자격증명 삭제
- 자격증명 암호화/복호화

### 사용법

```go
// 서비스 초기화
credentialService := usecase.NewCredentialService(
    credentialRepo,
    auditLogRepo,
    encryptor,
)

// 자격증명 생성
credential, err := credentialService.CreateCredential(createReq)
if err != nil {
    // 에러 처리
}

// 자격증명 조회
credentials, err := credentialService.GetCredentials(workspaceID)
if err != nil {
    // 에러 처리
}
```

### API 엔드포인트
- `POST /api/v1/credentials` - 자격증명 생성
- `GET /api/v1/credentials` - 자격증명 목록 조회
- `GET /api/v1/credentials/:id` - 자격증명 조회
- `PUT /api/v1/credentials/:id` - 자격증명 수정
- `DELETE /api/v1/credentials/:id` - 자격증명 삭제

---

## VM 서비스 (VM Service)

가상머신 관리 기능을 제공하는 서비스입니다.

### 주요 기능
- VM 생성
- VM 조회
- VM 수정
- VM 삭제
- VM 시작/중지
- VM 상태 조회

### 사용법

```go
// 서비스 초기화
vmService := usecase.NewVMService(
    vmRepo,
    workspaceRepo,
    cloudProvider,
    eventBus,
    auditLogRepo,
)

// VM 생성
vm, err := vmService.CreateVM(createReq)
if err != nil {
    // 에러 처리
}

// VM 시작
err = vmService.StartVM(vmID)
if err != nil {
    // 에러 처리
}
```

### API 엔드포인트
- `POST /api/v1/vms` - VM 생성
- `GET /api/v1/vms` - VM 목록 조회
- `GET /api/v1/vms/:id` - VM 조회
- `PUT /api/v1/vms/:id` - VM 수정
- `DELETE /api/v1/vms/:id` - VM 삭제
- `POST /api/v1/vms/:id/start` - VM 시작
- `POST /api/v1/vms/:id/stop` - VM 중지

---

## 클라우드 프로바이더 서비스 (Cloud Provider Service)

클라우드 프로바이더와의 상호작용을 관리하는 서비스입니다.

### 주요 기능
- 클라우드 인스턴스 생성
- 클라우드 인스턴스 조회
- 클라우드 인스턴스 삭제
- 클라우드 인스턴스 상태 조회

### 사용법

```go
// 서비스 초기화
cloudProviderService := usecase.NewCloudProviderService(pluginManager)

// 인스턴스 생성
instance, err := cloudProviderService.CreateInstance(ctx, "aws", createReq)
if err != nil {
    // 에러 처리
}

// 인스턴스 조회
instance, err := cloudProviderService.GetInstance(ctx, "aws", instanceID)
if err != nil {
    // 에러 처리
}
```

---

## IaC 서비스 (Infrastructure as Code Service)

OpenTofu를 사용한 인프라 관리 기능을 제공하는 서비스입니다.

### 주요 기능
- OpenTofu 계획 (Plan)
- OpenTofu 적용 (Apply)
- OpenTofu 파괴 (Destroy)
- 실행 상태 관리

### 사용법

```go
// 서비스 초기화
iacService := iac.NewService(db, eventBus)

// OpenTofu 계획
execution, err := iacService.Plan(ctx, workspaceID, config)
if err != nil {
    // 에러 처리
}

// OpenTofu 적용
execution, err := iacService.Apply(ctx, workspaceID, config)
if err != nil {
    // 에러 처리
}
```

### API 엔드포인트
- `POST /api/v1/iac/plan` - OpenTofu 계획
- `POST /api/v1/iac/apply` - OpenTofu 적용
- `POST /api/v1/iac/destroy` - OpenTofu 파괴
- `GET /api/v1/iac/executions` - 실행 목록 조회
- `GET /api/v1/iac/executions/:id` - 실행 조회

---

## 캐시 서비스 (Cache Service)

Redis를 사용한 캐싱 기능을 제공하는 서비스입니다.

### 주요 기능
- 캐시 데이터 저장
- 캐시 데이터 조회
- 캐시 데이터 삭제
- 캐시 전체 삭제

### 사용법

```go
// 서비스 초기화
cacheService := usecase.NewCacheService(redisService)

// 데이터 저장
err := cacheService.Set("key", "value", time.Hour)
if err != nil {
    // 에러 처리
}

// 데이터 조회
value, err := cacheService.Get("key")
if err != nil {
    // 에러 처리
}
```

---

## 이벤트 서비스 (Event Service)

NATS를 사용한 이벤트 처리 기능을 제공하는 서비스입니다.

### 주요 기능
- 이벤트 발행
- 이벤트 구독
- 워크스페이스별 이벤트
- 사용자별 이벤트

### 사용법

```go
// 서비스 초기화
eventService := usecase.NewEventService(natsService)

// 이벤트 발행
err := eventService.Publish("topic", eventData)
if err != nil {
    // 에러 처리
}

// 이벤트 구독
err := eventService.Subscribe("topic", handler)
if err != nil {
    // 에러 처리
}
```

---

## OIDC 서비스 (OpenID Connect Service)

OIDC 기반 소셜 로그인 기능을 제공하는 서비스입니다.

### 주요 기능
- OIDC 인증 URL 생성
- OIDC 코드 교환
- 소셜 로그인 지원 (Google, GitHub, Azure AD)

### 사용법

```go
// 서비스 초기화
oidcService := usecase.NewOIDCService(
    userRepo,
    auditLogRepo,
    authService,
)

// 인증 URL 생성
authURL, err := oidcService.GetAuthURL(ctx, "google")
if err != nil {
    // 에러 처리
}

// 코드 교환
token, err := oidcService.ExchangeCode(ctx, "google", code)
if err != nil {
    // 에러 처리
}
```

### API 엔드포인트
- `GET /api/v1/oidc/:provider/auth` - OIDC 인증 URL
- `POST /api/v1/oidc/:provider/callback` - OIDC 콜백 처리

---

## 🔧 설정 및 초기화

### 환경 변수
```bash
# 데이터베이스
DB_HOST=localhost
DB_PORT=5432
DB_USER=skyclust
DB_PASSWORD=password
DB_NAME=skyclust

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# NATS
NATS_URL=nats://localhost:4222

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRATION=24h
```

### 의존성 주입
```go
// 컨테이너 초기화
container := container.NewContainer(ctx, config)

// 서비스 사용
authService := container.AuthService
userService := container.UserService
workspaceService := container.WorkspaceService
```

---

## 📝 참고사항

- 모든 서비스는 의존성 주입을 통해 초기화됩니다.
- 에러 처리는 도메인별 에러 타입을 사용합니다.
- 모든 API 호출은 감사 로그에 기록됩니다.
- 캐싱과 이벤트 처리는 비동기로 동작합니다.

# SkyClust - 멀티 클라우드 관리 플랫폼

SkyClust는 Clean Architecture 기반의 멀티 클라우드 통합 관리 플랫폼입니다. Go 백엔드와 Next.js 프론트엔드로 구축되어 워크스페이스 기반 멀티 테넌트, Kubernetes 클러스터 관리, 비용 분석, 실시간 모니터링 등 엔터프라이즈급 기능을 제공합니다.

## 주요 기능

### 핵심 기능
- **멀티 클라우드 지원**: AWS, GCP, Azure, NCP (Naver Cloud Platform)
- **워크스페이스 기반 멀티 테넌트**: 완전한 워크스페이스 격리 및 자원 관리
- **Kubernetes 클러스터 관리**: AWS EKS, GCP GKE, Azure AKS 통합 관리
- **VM 오케스트레이션**: 프로바이더 간 인스턴스 생명주기 관리
- **네트워크 관리**: VPC/VNet, 서브넷, 보안 그룹/NSG 자동 생성/관리 (AWS, GCP, Azure)
- **리소스 그룹 관리**: Azure Resource Group 관리
- **비용 분석**: AWS Cost Explorer, GCP Cloud Billing API 통합, VM 및 Kubernetes 비용 통합 분석
- **자격증명 관리**: 워크스페이스 기반 자격증명 관리 및 AES 암호화 (JSON Input 및 File Upload 지원)
- **OIDC 인증**: 사용자 정의 OIDC 프로바이더 등록 및 SSO 지원
- **실시간 모니터링**: Server-Sent Events (SSE) 기반 라이브 업데이트 (압축 및 배치 지원)
- **알림 시스템**: 실시간 알림 및 사용자 설정 관리
- **감사 로깅**: 포괄적인 활동 추적 및 내보내기
- **데이터 내보내기**: CSV, JSON 형식 지원

### 엔터프라이즈 기능
- **RBAC**: 역할 기반 접근 제어
- **감사 추적**: 완전한 활동 로깅 및 통계
- **성능 최적화**: 쿼리 최적화 및 캐싱
- **구조화된 로깅**: Zap 기반 구조화된 로깅
- **입력 검증**: 강화된 보안 검증
- **RESTful API**: 일관된 REST API 설계

## 아키텍처

### 백엔드 (Go)
- **프레임워크**: Gin (HTTP), GORM (ORM)
- **데이터베이스**: PostgreSQL 15
- **메시징**: NATS (이벤트 버스)
- **캐싱**: Redis
- **인증**: JWT 기반 인증 및 RBAC
- **암호화**: AES 암호화를 통한 민감 데이터 보호
- **로깅**: Zap 구조화 로깅
- **모니터링**: OpenTelemetry 기반 추적 및 메트릭

### 프론트엔드 (Next.js + TypeScript)
- **프레임워크**: Next.js 14 (App Router)
- **UI 라이브러리**: shadcn/ui + Tailwind CSS
- **상태 관리**: React Query (서버 상태)
- **HTTP 클라이언트**: Axios (인터셉터 포함)
- **실시간 통신**: Server-Sent Events (SSE)
- **접근성**: WCAG 2.1 준수

### 인프라
- **컨테이너화**: Docker + Docker Compose
- **데이터베이스**: PostgreSQL 15
- **메시징**: NATS 2.10
- **캐싱**: Redis 7

## 프로젝트 구조

```
skyclust/
├── cmd/server/              # 애플리케이션 진입점
├── internal/
│   ├── application/
│   │   ├── handlers/       # HTTP 핸들러 (RESTful API)
│   │   │   ├── admin/      # 관리자 기능
│   │   │   ├── audit/      # 감사 로그
│   │   │   ├── auth/       # 인증
│   │   │   ├── cost_analysis/ # 비용 분석
│   │   │   ├── credential/ # 자격증명 관리
│   │   │   ├── export/     # 데이터 내보내기
│   │   │   ├── kubernetes/ # Kubernetes 관리
│   │   │   ├── network/    # 네트워크 관리
│   │   │   ├── notification/ # 알림 시스템
│   │   │   ├── oidc/       # OIDC 인증
│   │   │   ├── sse/        # Server-Sent Events
│   │   │   ├── system/    # 시스템 모니터링
│   │   │   └── workspace/ # 워크스페이스 관리
│   │   └── services/       # 비즈니스 로직 서비스
│   ├── domain/             # 도메인 엔티티 및 인터페이스
│   ├── infrastructure/     # 인프라 계층 (DB, 외부 서비스)
│   ├── routes/             # 라우트 관리
│   ├── di/                  # 의존성 주입
│   └── shared/              # 공유 유틸리티
├── pkg/                     # 공유 패키지
│   ├── auth/               # 인증 유틸리티
│   ├── cache/               # 캐싱
│   ├── config/              # 설정 관리
│   ├── middleware/          # HTTP 미들웨어
│   ├── security/            # 보안 유틸리티
│   └── telemetry/           # 모니터링
├── frontend/                # Next.js 프론트엔드
├── docs/                    # 문서
└── .bruno/                  # Bruno API 테스트 컬렉션
```

## 빠른 시작

### 필수 요구사항
- Go 1.24 이상
- Node.js 18 이상
- PostgreSQL 15 이상
- Redis 7 이상
- Docker & Docker Compose

### 설치 및 실행

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

3. **환경 변수 설정:**
```bash
cp .env.sample .env
# .env 파일 편집
```

4. **Docker Compose로 실행:**
```bash
# 개발 환경
docker-compose -f docker-compose.dev.yml up -d

# 또는 Makefile 사용
make compose-up
```

5. **개발 서버 실행:**
```bash
# 백엔드
make dev

# 프론트엔드
cd frontend
npm run dev
```

## API 구조

### 공개 엔드포인트
- `GET /health` - 헬스 체크
- `POST /api/v1/auth/register` - 사용자 등록
- `POST /api/v1/auth/login` - 로그인
- `GET /api/v1/oidc/providers/types` - OIDC 프로바이더 타입 목록
- `GET /api/v1/system/status` - 시스템 상태

### 인증 필요 엔드포인트
- `GET /api/v1/auth/sessions/me` - 현재 세션 정보
- `DELETE /api/v1/auth/sessions/me` - 로그아웃
- `GET /api/v1/auth/me` - 현재 사용자 정보

### 주요 기능별 엔드포인트

**워크스페이스 관리:**
- `GET /api/v1/workspaces` - 워크스페이스 목록
- `POST /api/v1/workspaces` - 워크스페이스 생성
- `GET /api/v1/workspaces/:id` - 워크스페이스 상세

**자격증명 관리:**
- `GET /api/v1/credentials` - 자격증명 목록 (workspace_id 필수)
- `POST /api/v1/credentials` - 자격증명 생성
- `GET /api/v1/credentials/:id` - 자격증명 상세

**Kubernetes 관리:**
- `GET /api/v1/aws/kubernetes/clusters` - EKS 클러스터 목록
- `POST /api/v1/aws/kubernetes/clusters` - EKS 클러스터 생성
- `GET /api/v1/gcp/kubernetes/clusters` - GKE 클러스터 목록
- `POST /api/v1/gcp/kubernetes/clusters` - GKE 클러스터 생성

**비용 분석:**
- `GET /api/v1/cost-analysis/workspaces/:workspaceId/summary` - 비용 요약
- `GET /api/v1/cost-analysis/workspaces/:workspaceId/predictions` - 비용 예측
- `GET /api/v1/cost-analysis/workspaces/:workspaceId/trend` - 비용 트렌드
- `GET /api/v1/cost-analysis/workspaces/:workspaceId/breakdown` - 비용 세부 분석

**알림:**
- `GET /api/v1/notifications` - 알림 목록
- `PATCH /api/v1/notifications/:id` - 알림 읽음 처리
- `PATCH /api/v1/notifications` - 알림 일괄 읽음 처리

**감사 로그:**
- `GET /api/v1/admin/audit-logs` - 감사 로그 목록
- `GET /api/v1/admin/audit-logs?aggregate=stats` - 감사 로그 통계
- `GET /api/v1/admin/audit-logs?format=summary` - 감사 로그 요약
- `DELETE /api/v1/admin/audit-logs?retention_days=90` - 감사 로그 정리

**내보내기:**
- `POST /api/v1/exports` - 내보내기 생성
- `GET /api/v1/exports/:id` - 내보내기 상태 조회
- `GET /api/v1/exports/:id/file` - 내보내기 파일 다운로드

**OIDC:**
- `GET /api/v1/oidc/providers` - 사용자 등록 OIDC 프로바이더 목록
- `POST /api/v1/oidc/providers` - OIDC 프로바이더 등록
- `POST /api/v1/auth/oidc/sessions` - OIDC 로그인
- `DELETE /api/v1/auth/oidc/sessions/me` - OIDC 로그아웃

상세한 API 문서는 `.bruno/` 폴더의 Bruno 컬렉션을 참조하세요.

## 설정

### 환경 변수

| 변수 | 설명 | 기본값 |
|------|------|--------|
| `DB_HOST` | 데이터베이스 호스트 | `localhost` |
| `DB_PORT` | 데이터베이스 포트 | `5432` |
| `DB_USER` | 데이터베이스 사용자 | `skyclust` |
| `DB_PASSWORD` | 데이터베이스 비밀번호 | - |
| `DB_NAME` | 데이터베이스 이름 | `skyclust` |
| `REDIS_URL` | Redis 연결 URL | `redis://localhost:6379` |
| `JWT_SECRET` | JWT 서명 시크릿 | - |
| `ENCRYPTION_KEY` | 데이터 암호화 키 | - |

### 클라우드 프로바이더 설정

클라우드 자격증명은 워크스페이스 기반으로 관리되며, API를 통해 등록합니다:

**AWS 자격증명:**
```json
{
  "workspace_id": "workspace-uuid",
  "name": "AWS Production",
  "provider": "aws",
  "data": {
    "access_key": "AKIA...",
    "secret_key": "...",
    "region": "ap-northeast-2"
  }
}
```

**GCP 자격증명:**
```json
{
  "workspace_id": "workspace-uuid",
  "name": "GCP Production",
  "provider": "gcp",
  "data": {
    "project_id": "my-project",
    "service_account_key": {...}
  }
}
```

## 개발

### 코드 품질
```bash
# 코드 포맷팅
make format

# 린팅
make lint

# 테스트
make test
make test-coverage
```

### 빌드
```bash
# 백엔드 빌드
make build

# 모든 플랫폼 빌드
make build-all-platforms

# Docker 이미지 빌드
make docker-build
```

### Makefile 명령어
```bash
make help          # 도움말
make dev           # 개발 서버 실행
make compose-up    # Docker Compose 시작
make compose-down  # Docker Compose 중지
make clean         # 빌드 아티팩트 정리
```

## 아키텍처 원칙

### Clean Architecture
프로젝트는 Clean Architecture 원칙을 따릅니다:
- **Domain Layer**: 비즈니스 로직 및 엔티티
- **Application Layer**: 유스케이스 및 서비스
- **Infrastructure Layer**: 데이터베이스, 외부 API 통합
- **Presentation Layer**: HTTP 핸들러 및 라우트

### RESTful API 설계
- Kebab-case URL 사용
- 복수형 리소스 이름
- 적절한 HTTP 메서드 사용
- 중첩 리소스는 쿼리 파라미터로 처리

### 도메인 타입과 DTO 분리
- `domain/`: 핵심 비즈니스 엔티티
- `internal/application/handlers/<feature>/types.go`: API DTO
- `internal/api/common/types.go`: 공통 DTO

## 보안

### 인증 및 인가
- JWT 기반 인증
- RBAC (역할 기반 접근 제어)
- 세션 관리 (RESTful 세션 엔드포인트)
- OIDC SSO 지원

### 데이터 보호
- AES 암호화를 통한 민감 데이터 보호
- 워크스페이스 기반 자격증명 격리
- 안전한 자격증명 저장
- 입력 검증 및 SQL 인젝션 방지

### 감사 로깅
- 완전한 활동 추적
- 보안 이벤트 로깅
- 데이터 접근 로깅
- 컴플라이언스 보고

## 비용 분석

### 지원 기능
- AWS Cost Explorer API 통합
- GCP Cloud Billing API 통합
- VM 비용 계산 및 추적
- Kubernetes 클러스터 비용 (EKS, GKE)
- 비용 예측 (선형 회귀)
- 예산 알림
- 비용 트렌드 분석
- 리소스 타입별 필터링 (vm, cluster, node_group, node_pool)

### 리소스 타입 필터
비용 분석 API는 `resource_types` 쿼리 파라미터를 지원합니다:
- `all`: 모든 리소스 타입 (기본값)
- `vm`: VM만
- `cluster`: Kubernetes 클러스터만
- `vm,cluster`: VM과 클러스터 함께

### 경고 정보
비용 분석 API는 다음 상황에서 경고를 반환합니다:
- API 권한 부족 (`API_PERMISSION_DENIED`)
- API 미활성화 (`API_NOT_ENABLED`)
- 비용 계산 실패 (`VM_COST_CALCULATION_FAILED`, `KUBERNETES_COST_CALCULATION_FAILED`)
- 자격증명 오류 (`CREDENTIAL_ERROR`)

## 모니터링

### 메트릭
- 요청 지연시간 및 처리량
- 데이터베이스 쿼리 성능
- 메모리 및 CPU 사용량
- 에러율 및 성공률

### 로깅
- 구조화된 JSON 로깅 (Zap)
- 요청/응답 로깅
- 에러 추적
- 감사 추적 로깅

### 헬스 체크
- `GET /health` - 애플리케이션 헬스
- `GET /api/v1/system/status` - 시스템 상태
- `GET /api/v1/system/metrics` - 시스템 메트릭

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
  - [자격증명 설정 가이드](docs/credential_setup_guide.md) - AWS/GCP IAM 설정 가이드
  - [시스템 인터페이스 및 API 목록](docs/system_interfaces_apis_dtos_summary.md)
  - [기술 설계 문서](docs/technical_design_document.md)
- API 테스트: [.bruno/](.bruno/)
- 이슈: [GitHub Issues](https://github.com/taking/skyclust/issues)

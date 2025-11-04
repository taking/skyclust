# SkyClust 시스템 인터페이스, REST API, DTO 목록 정리

## 전체 시스템 개요

### 통계 요약
- **인터페이스 총 개수**: 20개 이상
- **REST API 엔드포인트 총 개수**: 150+ 개
- **DTO 총 개수**: 50+ 개
- **지원 클라우드 제공업체**: AWS, GCP, Azure, NCP
- **지원 Kubernetes**: AWS EKS, GCP GKE

---

## 1. 인터페이스 목록

### 1.1 Service 인터페이스

| 번호 | 인터페이스명 | 파일 위치 | 주요 기능 |
|------|-------------|-----------|-----------|
| 1 | `VMService` | `internal/interfaces/services/vm_service.go` | VM 생명주기 관리, 상태 관리, 접근 제어 |
| 2 | `UserService` | `internal/interfaces/services/user_service.go` | 사용자 관리, 인증, 프로필 관리 |
| 3 | `AuthService` | `internal/interfaces/services/auth_service.go` | 인증, 토큰 관리, 세션 관리 |
| 4 | `CredentialService` | `internal/interfaces/services/credential_service.go` | 클라우드 자격증명 관리, 암호화 (워크스페이스 기반) |
| 5 | `WorkspaceService` | `internal/interfaces/services/workspace_service.go` | 워크스페이스 관리, 멀티 테넌트 |
| 6 | `CloudProviderService` | `internal/interfaces/services/cloud_provider_service.go` | 클라우드 제공업체 통합 관리 |
| 7 | `RBACService` | `internal/interfaces/services/rbac_service.go` | 역할 기반 접근 제어 |
| 8 | `AuditLogService` | `internal/interfaces/services/audit_log_service.go` | 감사 로그 관리 |
| 9 | `NotificationService` | `internal/interfaces/services/notification_service.go` | 알림 관리 |
| 10 | `ExportService` | `internal/interfaces/services/export_service.go` | 데이터 내보내기 |
| 11 | `KubernetesService` | `internal/application/services/kubernetes_service.go` | Kubernetes 클러스터 관리 |
| 12 | `NetworkService` | `internal/application/services/network_service.go` | 네트워크 리소스 관리 |
| 13 | `CostAnalysisService` | `internal/interfaces/services/cost_analysis_service.go` | 비용 분석 및 예측 |
| 14 | `OIDCService` | `internal/interfaces/services/oidc_service.go` | OIDC 인증 관리 |

### 1.2 Repository 인터페이스

| 번호 | 인터페이스명 | 파일 위치 | 주요 기능 |
|------|-------------|-----------|-----------|
| 1 | `VMRepository` | `internal/interfaces/repositories/vm_repository.go` | VM 데이터 영속성 |
| 2 | `UserRepository` | `internal/interfaces/repositories/user_repository.go` | 사용자 데이터 영속성 |
| 3 | `WorkspaceRepository` | `internal/interfaces/repositories/workspace_repository.go` | 워크스페이스 데이터 영속성 |
| 4 | `CredentialRepository` | `internal/interfaces/repositories/credential_repository.go` | 자격증명 데이터 영속성 (워크스페이스 기반) |
| 5 | `AuditLogRepository` | `internal/interfaces/repositories/audit_log_repository.go` | 감사 로그 데이터 영속성 |

### 1.3 Handler 인터페이스

| 번호 | 인터페이스명 | 파일 위치 | 주요 기능 |
|------|-------------|-----------|-----------|
| 1 | `HTTPHandler` | `internal/interfaces/handlers/http_handler.go` | HTTP 핸들러 기본 인터페이스 |
| 2 | `KubernetesHandler` | `internal/application/handlers/kubernetes/` | Kubernetes 리소스 관리 |
| 3 | `NetworkHandler` | `internal/application/handlers/network/` | 네트워크 리소스 관리 |

---

## 2. REST API 엔드포인트 목록

### 2.1 인증 및 사용자 관리

**공개 엔드포인트:**
```
POST   /api/v1/auth/register              # 사용자 등록
POST   /api/v1/auth/login                 # 로그인
```

**인증 필요 엔드포인트:**
```
GET    /api/v1/auth/sessions/me          # 현재 세션 정보
DELETE /api/v1/auth/sessions/me          # 로그아웃
GET    /api/v1/auth/me                   # 현재 사용자 정보

GET    /api/v1/users                      # 사용자 목록
GET    /api/v1/users/:id                  # 사용자 상세
PUT    /api/v1/users/:id                  # 사용자 수정
DELETE /api/v1/users/:id                  # 사용자 삭제
```

### 2.2 OIDC 인증

**공개 엔드포인트:**
```
GET    /api/v1/oidc/providers/types       # OIDC 프로바이더 타입 목록
GET    /api/v1/auth/oidc/:provider/auth-url    # OIDC 인증 URL 생성
POST   /api/v1/auth/oidc/callback         # OIDC 콜백 처리
GET    /api/v1/auth/oidc/:provider/logout-url  # OIDC 로그아웃 URL
```

**인증 필요 엔드포인트:**
```
POST   /api/v1/auth/oidc/sessions         # OIDC 세션 생성
DELETE /api/v1/auth/oidc/sessions/me       # OIDC 로그아웃
GET    /api/v1/oidc/providers             # 사용자 등록 OIDC 프로바이더 목록
POST   /api/v1/oidc/providers             # OIDC 프로바이더 등록
GET    /api/v1/oidc/providers/:id         # OIDC 프로바이더 상세
PUT    /api/v1/oidc/providers/:id         # OIDC 프로바이더 수정
DELETE /api/v1/oidc/providers/:id         # OIDC 프로바이더 삭제
```

### 2.3 자격증명 관리 (워크스페이스 기반)

```
GET    /api/v1/credentials?workspace_id=:id  # 자격증명 목록
POST   /api/v1/credentials                    # 자격증명 생성 (workspace_id 필수)
GET    /api/v1/credentials/:id?workspace_id=:id  # 자격증명 상세
PUT    /api/v1/credentials/:id?workspace_id=:id  # 자격증명 수정
DELETE /api/v1/credentials/:id?workspace_id=:id  # 자격증명 삭제
GET    /api/v1/credentials?provider=:provider&workspace_id=:id  # 프로바이더별 조회
POST   /api/v1/credentials/upload            # 파일 업로드로 자격증명 생성
```

### 2.4 워크스페이스 관리

```
GET    /api/v1/workspaces                   # 워크스페이스 목록
POST   /api/v1/workspaces                   # 워크스페이스 생성
GET    /api/v1/workspaces/:id               # 워크스페이스 상세
PUT    /api/v1/workspaces/:id               # 워크스페이스 수정
DELETE /api/v1/workspaces/:id               # 워크스페이스 삭제
```

### 2.5 Kubernetes 클러스터 관리

**AWS EKS:**
```
GET    /api/v1/aws/kubernetes/clusters     # EKS 클러스터 목록
POST   /api/v1/aws/kubernetes/clusters     # EKS 클러스터 생성
GET    /api/v1/aws/kubernetes/clusters/:name  # EKS 클러스터 상세
DELETE /api/v1/aws/kubernetes/clusters/:name  # EKS 클러스터 삭제
GET    /api/v1/aws/kubernetes/clusters/:name/kubeconfig  # Kubeconfig 다운로드

GET    /api/v1/aws/kubernetes/clusters/:name/node-groups  # 노드 그룹 목록
POST   /api/v1/aws/kubernetes/clusters/:name/node-groups  # 노드 그룹 생성
GET    /api/v1/aws/kubernetes/clusters/:name/node-groups/:ngName  # 노드 그룹 상세
DELETE /api/v1/aws/kubernetes/clusters/:name/node-groups/:ngName  # 노드 그룹 삭제
```

**GCP GKE:**
```
GET    /api/v1/gcp/kubernetes/clusters     # GKE 클러스터 목록
POST   /api/v1/gcp/kubernetes/clusters     # GKE 클러스터 생성 (Standard/Autopilot/고급 모드)
GET    /api/v1/gcp/kubernetes/clusters/:name  # GKE 클러스터 상세
DELETE /api/v1/gcp/kubernetes/clusters/:name  # GKE 클러스터 삭제
GET    /api/v1/gcp/kubernetes/clusters/:name/kubeconfig  # Kubeconfig 다운로드

GET    /api/v1/gcp/kubernetes/clusters/:name/node-pools  # 노드 풀 목록
POST   /api/v1/gcp/kubernetes/clusters/:name/node-pools  # 노드 풀 생성
GET    /api/v1/gcp/kubernetes/clusters/:name/node-pools/:npName  # 노드 풀 상세
PUT    /api/v1/gcp/kubernetes/clusters/:name/node-pools/:npName/scale  # 노드 풀 스케일링
DELETE /api/v1/gcp/kubernetes/clusters/:name/node-pools/:npName  # 노드 풀 삭제
```

### 2.6 네트워크 관리

**AWS:**
```
GET    /api/v1/aws/network/vpcs           # VPC 목록
POST   /api/v1/aws/network/vpcs            # VPC 생성
GET    /api/v1/aws/network/vpcs/:id        # VPC 상세
PUT    /api/v1/aws/network/vpcs/:id        # VPC 수정
DELETE /api/v1/aws/network/vpcs/:id        # VPC 삭제

GET    /api/v1/aws/network/subnets         # 서브넷 목록
POST   /api/v1/aws/network/subnets         # 서브넷 생성
GET    /api/v1/aws/network/subnets/:id      # 서브넷 상세
PUT    /api/v1/aws/network/subnets/:id     # 서브넷 수정
DELETE /api/v1/aws/network/subnets/:id     # 서브넷 삭제

GET    /api/v1/aws/network/security-groups  # 보안 그룹 목록
POST   /api/v1/aws/network/security-groups  # 보안 그룹 생성
GET    /api/v1/aws/network/security-groups/:id  # 보안 그룹 상세
PUT    /api/v1/aws/network/security-groups/:id  # 보안 그룹 수정
DELETE /api/v1/aws/network/security-groups/:id  # 보안 그룹 삭제
POST   /api/v1/aws/network/security-groups/:id/rules  # 보안 그룹 규칙 추가
DELETE /api/v1/aws/network/security-groups/:id/rules  # 보안 그룹 규칙 삭제
PUT    /api/v1/aws/network/security-groups/:id/rules  # 보안 그룹 규칙 일괄 업데이트
```

**GCP:**
```
GET    /api/v1/gcp/network/vpcs           # VPC 목록
POST   /api/v1/gcp/network/vpcs           # VPC 생성
GET    /api/v1/gcp/network/vpcs/:id       # VPC 상세
PUT    /api/v1/gcp/network/vpcs/:id       # VPC 수정
DELETE /api/v1/gcp/network/vpcs/:id       # VPC 삭제

GET    /api/v1/gcp/network/subnets        # 서브넷 목록
POST   /api/v1/gcp/network/subnets        # 서브넷 생성
GET    /api/v1/gcp/network/subnets/:id    # 서브넷 상세
PUT    /api/v1/gcp/network/subnets/:id    # 서브넷 수정
DELETE /api/v1/gcp/network/subnets/:id    # 서브넷 삭제

GET    /api/v1/gcp/network/firewall-rules  # 방화벽 규칙 목록
POST   /api/v1/gcp/network/firewall-rules  # 방화벽 규칙 생성
GET    /api/v1/gcp/network/firewall-rules/:id  # 방화벽 규칙 상세
PUT    /api/v1/gcp/network/firewall-rules/:id  # 방화벽 규칙 수정
DELETE /api/v1/gcp/network/firewall-rules/:id  # 방화벽 규칙 삭제
POST   /api/v1/gcp/network/firewall-rules/:id/rules  # 방화벽 규칙 개별 추가
DELETE /api/v1/gcp/network/firewall-rules/:id/rules  # 방화벽 규칙 개별 삭제
```

### 2.7 비용 분석

```
GET    /api/v1/cost-analysis/workspaces/:workspaceId/summary?period=30d&resource_types=all  # 비용 요약
GET    /api/v1/cost-analysis/workspaces/:workspaceId/predictions?days=30&resource_types=all  # 비용 예측
GET    /api/v1/cost-analysis/workspaces/:workspaceId/trend?period=30d&resource_types=all    # 비용 트렌드
GET    /api/v1/cost-analysis/workspaces/:workspaceId/breakdown?period=30d&dimension=service&resource_types=all  # 비용 세부 분석
GET    /api/v1/cost-analysis/workspaces/:workspaceId/comparison?current_period=30d&compare_period=previous  # 비용 비교
GET    /api/v1/cost-analysis/workspaces/:workspaceId/budget-alerts  # 예산 알림
```

**리소스 타입 필터:**
- `all`: 모든 리소스 (VM + Kubernetes 클러스터)
- `vm`: VM만
- `cluster`: Kubernetes 클러스터만
- `vm,cluster`: VM과 클러스터 함께

### 2.8 알림 시스템

```
GET    /api/v1/notifications              # 알림 목록
GET    /api/v1/notifications/:id          # 알림 상세
PATCH  /api/v1/notifications/:id          # 알림 읽음 처리 ({"read": true})
PATCH  /api/v1/notifications              # 알림 일괄 읽음 처리
DELETE /api/v1/notifications/:id          # 알림 삭제
DELETE /api/v1/notifications              # 알림 일괄 삭제
GET    /api/v1/notifications/stats        # 알림 통계
GET    /api/v1/notifications/preferences  # 알림 설정 조회
PUT    /api/v1/notifications/preferences  # 알림 설정 업데이트
POST   /api/v1/notifications/test          # 테스트 알림 발송
```

### 2.9 감사 로그

```
GET    /api/v1/admin/audit-logs           # 감사 로그 목록
GET    /api/v1/admin/audit-logs/:id       # 감사 로그 상세
GET    /api/v1/admin/audit-logs?aggregate=stats  # 감사 로그 통계
GET    /api/v1/admin/audit-logs?format=summary   # 감사 로그 요약
GET    /api/v1/admin/audit-logs/export?format=csv|json  # 감사 로그 내보내기
DELETE /api/v1/admin/audit-logs?retention_days=90  # 감사 로그 정리
```

### 2.10 데이터 내보내기

```
POST   /api/v1/exports                    # 내보내기 생성
GET    /api/v1/exports/:id                # 내보내기 상태 조회
GET    /api/v1/exports/:id/file           # 내보내기 파일 다운로드
```

### 2.11 실시간 통신 (SSE)

```
GET    /api/v1/sse/monitoring             # 모니터링 스트림
GET    /api/v1/sse/notifications          # 알림 스트림
```

### 2.12 시스템 모니터링

**공개:**
```
GET    /health                            # 헬스 체크
GET    /api/v1/system/status              # 시스템 상태
```

**인증 필요:**
```
GET    /api/v1/system/metrics             # 시스템 메트릭
GET    /api/v1/admin/system/config        # 시스템 설정
```

---

## 3. 주요 DTO 구조

### 3.1 자격증명 DTO

**CreateCredentialRequest (워크스페이스 기반):**
```go
type CreateCredentialRequest struct {
    WorkspaceID string                 `json:"workspace_id" binding:"required"`
    Name      string                  `json:"name" binding:"required"`
    Provider  string                  `json:"provider" binding:"required,oneof=aws gcp azure ncp"`
    Data      map[string]interface{}   `json:"data" binding:"required"`
}
```

### 3.2 Kubernetes DTO

**CreateClusterRequest:**
```go
type CreateClusterRequest struct {
    Name         string `json:"name" binding:"required"`
    Region       string `json:"region" binding:"required"`
    Version      string `json:"version,omitempty"`
    NodeGroup    NodeGroupConfig `json:"node_group"`
    // ... 기타 설정
}
```

### 3.3 비용 분석 DTO

**CostSummary:**
```go
type CostSummary struct {
    TotalCost   float64            `json:"total_cost"`
    Currency    string             `json:"currency"`
    Period      string             `json:"period"`
    StartDate   time.Time          `json:"start_date"`
    EndDate     time.Time          `json:"end_date"`
    ByProvider  map[string]float64 `json:"by_provider"`
    Warnings    []CostWarning      `json:"warnings,omitempty"`
}
```

**CostWarning:**
```go
type CostWarning struct {
    Code         string `json:"code"`
    Message      string `json:"message"`
    Provider     string `json:"provider,omitempty"`
    ResourceType string `json:"resource_type,omitempty"`
}
```

### 3.4 알림 DTO

**UpdateNotificationRequest:**
```go
type UpdateNotificationRequest struct {
    Read *bool `json:"read" binding:"required"`
}
```

**UpdateNotificationsRequest:**
```go
type UpdateNotificationsRequest struct {
    Read            *bool    `json:"read" binding:"required"`
    NotificationIDs []string `json:"notification_ids,omitempty"`
}
```

---

## 4. 주요 변경 사항

### 4.1 자격증명 관리 (Workspace 기반 전환)
- 이전: User 기반 자격증명 관리
- 현재: Workspace 기반 자격증명 관리
- 모든 자격증명 API에 `workspace_id` 파라미터 필요

### 4.2 RESTful API 개선
- URL 일관성: kebab-case 사용
- HTTP 메서드: PATCH 사용 (부분 업데이트)
- 세션 관리: `/auth/sessions/me`로 통합

### 4.3 비용 분석 개선
- VM과 Kubernetes 클러스터 비용 통합 지원
- AWS Cost Explorer API 통합
- GCP Cloud Billing API 통합
- 리소스 타입별 필터링 지원
- 경고 정보 포함

### 4.4 알림 시스템 개선
- PATCH 메서드로 읽음 처리
- 일괄 업데이트 지원

---

## 5. API 버전 관리

현재 버전: `v1`

모든 API는 `/api/v1/` 접두사를 사용합니다.

---

## 참고 문서

- [Kubernetes API 목록](kubernetes_api_list.md)
- [Kubernetes Service 인터페이스](kubernetes_service_interfaces_apis_dtos_summary.md)
- [기술 설계 문서](technical_design_document.md)
- [아키텍처 다이어그램](architecture_diagrams.md)

# SkyClust 시스템 인터페이스, REST API, DTO 목록 정리

## 📊 **전체 시스템 개요**

### **📈 통계 요약**
- **인터페이스 총 개수**: 20개
- **REST API 엔드포인트 총 개수**: 150+ 개
- **DTO 총 개수**: 50+ 개

---

## 🔌 **1. 인터페이스 목록 (총 20개)**

### **1.1 Service 인터페이스 (10개)**
| 번호 | 인터페이스명 | 파일 위치 | 주요 기능 |
|------|-------------|-----------|-----------|
| 1 | `VMService` | `internal/interfaces/services/vm_service.go` | VM 생명주기 관리, 상태 관리, 접근 제어 |
| 2 | `UserService` | `internal/interfaces/services/user_service.go` | 사용자 관리, 인증, 프로필 관리 |
| 3 | `AuthService` | `internal/interfaces/services/auth_service.go` | 인증, 토큰 관리, 세션 관리 |
| 4 | `CredentialService` | `internal/interfaces/services/credential_service.go` | 클라우드 자격증명 관리, 암호화 |
| 5 | `WorkspaceService` | `internal/interfaces/services/workspace_service.go` | 워크스페이스 관리, 멀티 테넌트 |
| 6 | `CloudProviderService` | `internal/interfaces/services/cloud_provider_service.go` | 클라우드 제공업체 통합 관리 |
| 7 | `RBACService` | `internal/interfaces/services/rbac_service.go` | 역할 기반 접근 제어 |
| 8 | `AuditLogService` | `internal/interfaces/services/audit_log_service.go` | 감사 로그 관리 |
| 9 | `NotificationService` | `internal/interfaces/services/notification_service.go` | 알림 관리 |
| 10 | `ExportService` | `internal/interfaces/services/export_service.go` | 데이터 내보내기 |

### **1.2 Repository 인터페이스 (5개)**
| 번호 | 인터페이스명 | 파일 위치 | 주요 기능 |
|------|-------------|-----------|-----------|
| 1 | `VMRepository` | `internal/interfaces/repositories/vm_repository.go` | VM 데이터 영속성 |
| 2 | `UserRepository` | `internal/interfaces/repositories/user_repository.go` | 사용자 데이터 영속성 |
| 3 | `WorkspaceRepository` | `internal/interfaces/repositories/workspace_repository.go` | 워크스페이스 데이터 영속성 |
| 4 | `CredentialRepository` | `internal/interfaces/repositories/credential_repository.go` | 자격증명 데이터 영속성 |
| 5 | `AuditLogRepository` | `internal/interfaces/repositories/audit_log_repository.go` | 감사 로그 데이터 영속성 |

### **1.3 Handler 인터페이스 (5개)**
| 번호 | 인터페이스명 | 파일 위치 | 주요 기능 |
|------|-------------|-----------|-----------|
| 1 | `HTTPHandler` | `internal/interfaces/handlers/http_handler.go` | HTTP 핸들러 기본 인터페이스 |
| 2 | `KubernetesHandler` | `internal/application/handlers/kubernetes/` | Kubernetes 리소스 관리 |
| 3 | `NetworkHandler` | `internal/application/handlers/network/` | 네트워크 리소스 관리 |
| 4 | `AuthHandler` | `internal/application/handlers/auth/` | 인증 관련 처리 |
| 5 | `ProviderHandler` | `internal/application/handlers/provider/` | 클라우드 제공업체 관리 |

---

## 🌐 **2. REST API 엔드포인트 목록 (총 150+ 개)**

### **2.1 인증 및 사용자 관리 (15개)**
```
/api/v1/auth/
├── POST   /register              # 사용자 등록
├── POST   /login                 # 로그인
├── POST   /logout                # 로그아웃
├── GET    /me                    # 현재 사용자 정보
├── POST   /refresh               # 토큰 갱신
└── GET    /profile               # 프로필 조회

/api/v1/users/
├── GET    /                      # 사용자 목록
├── GET    /:id                   # 사용자 상세
├── PUT    /:id                   # 사용자 수정
├── DELETE /:id                   # 사용자 삭제
└── POST   /:id/change-password   # 비밀번호 변경

/api/v1/auth/oidc/
├── GET    /providers             # OIDC 제공업체 목록
├── POST   /login/:provider       # OIDC 로그인
└── GET    /callback/:provider    # OIDC 콜백
```

### **2.2 자격증명 관리 (8개)**
```
/api/v1/credentials/
├── GET    /                      # 자격증명 목록
├── POST   /                      # 자격증명 생성
├── GET    /:id                   # 자격증명 상세
├── PUT    /:id                   # 자격증명 수정
├── DELETE /:id                   # 자격증명 삭제
├── POST   /:id/test              # 자격증명 테스트
├── POST   /:id/rotate            # 자격증명 회전
└── GET    /:id/masked            # 마스킹된 자격증명 조회
```

### **2.3 워크스페이스 관리 (10개)**
```
/api/v1/workspaces/
├── GET    /                      # 워크스페이스 목록
├── POST   /                      # 워크스페이스 생성
├── GET    /:id                   # 워크스페이스 상세
├── PUT    /:id                   # 워크스페이스 수정
├── DELETE /:id                   # 워크스페이스 삭제
├── GET    /:id/members           # 멤버 목록
├── POST   /:id/members           # 멤버 추가
├── DELETE /:id/members/:user_id  # 멤버 제거
├── GET    /:id/resources         # 리소스 목록
└── GET    /:id/usage             # 사용량 조회
```

### **2.4 Kubernetes 관리 (40개)**

#### **AWS EKS (20개)**
```
/api/v1/aws/kubernetes/
├── clusters/
│   ├── POST   /                  # 클러스터 생성
│   ├── GET    /                  # 클러스터 목록
│   ├── GET    /:name             # 클러스터 상세
│   ├── DELETE /:name             # 클러스터 삭제
│   └── GET    /:name/kubeconfig  # Kubeconfig 생성
├── clusters/:name/nodepools/
│   ├── POST   /                  # 노드풀 생성
│   ├── GET    /                  # 노드풀 목록
│   ├── GET    /:nodepool         # 노드풀 상세
│   ├── DELETE /:nodepool         # 노드풀 삭제
│   └── PUT    /:nodepool/scale   # 노드풀 스케일링
├── clusters/:name/nodegroups/
│   ├── POST   /                  # 노드그룹 생성
│   ├── GET    /                  # 노드그룹 목록
│   ├── GET    /:nodegroup        # 노드그룹 상세
│   └── DELETE /:nodegroup        # 노드그룹 삭제
└── clusters/:name/
    ├── POST   /upgrade            # 클러스터 업그레이드
    ├── GET    /upgrade/status     # 업그레이드 상태
    ├── GET    /nodes              # 노드 목록
    └── GET    /nodes/:node        # 노드 상세
```

#### **GCP GKE (20개)**
```
/api/v1/gcp/kubernetes/
├── clusters/
│   ├── POST   /                  # 클러스터 생성
│   ├── GET    /                  # 클러스터 목록
│   ├── GET    /:name             # 클러스터 상세
│   ├── DELETE /:name             # 클러스터 삭제
│   └── GET    /:name/kubeconfig  # Kubeconfig 생성
├── clusters/:name/nodepools/
│   ├── POST   /                  # 노드풀 생성
│   ├── GET    /                  # 노드풀 목록
│   ├── GET    /:nodepool         # 노드풀 상세
│   ├── DELETE /:nodepool         # 노드풀 삭제
│   └── PUT    /:nodepool/scale   # 노드풀 스케일링
├── clusters/:name/nodes/
│   ├── GET    /                  # 노드 목록
│   ├── GET    /:node             # 노드 상세
│   ├── POST   /:node/drain       # 노드 드레인
│   ├── POST   /:node/cordon      # 노드 코돈
│   └── POST   /:node/uncordon    # 노드 언코돈
└── clusters/:name/
    ├── POST   /upgrade            # 클러스터 업그레이드
    ├── GET    /upgrade/status     # 업그레이드 상태
    └── GET    /nodes/:node/ssh    # SSH 설정
```

### **2.5 네트워크 관리 (60개)**

#### **AWS 네트워크 (30개)**
```
/api/v1/aws/network/
├── vpcs/
│   ├── GET    /                  # VPC 목록
│   ├── POST   /                  # VPC 생성
│   ├── GET    /:id               # VPC 상세
│   ├── PUT    /:id               # VPC 수정
│   └── DELETE /:id               # VPC 삭제
├── subnets/
│   ├── GET    /                  # 서브넷 목록
│   ├── POST   /                  # 서브넷 생성
│   ├── GET    /:id               # 서브넷 상세
│   ├── PUT    /:id               # 서브넷 수정
│   └── DELETE /:id               # 서브넷 삭제
└── security-groups/
    ├── GET    /                  # 보안그룹 목록
    ├── POST   /                  # 보안그룹 생성
    ├── GET    /:id               # 보안그룹 상세
    ├── PUT    /:id               # 보안그룹 수정
    ├── DELETE /:id               # 보안그룹 삭제
    ├── POST   /:id/rules         # 규칙 추가
    ├── DELETE /:id/rules         # 규칙 삭제
    └── PUT    /:id/rules         # 규칙 수정
```

#### **GCP 네트워크 (30개)**
```
/api/v1/gcp/network/
├── vpcs/
│   ├── GET    /                  # VPC 목록
│   ├── POST   /                  # VPC 생성
│   ├── GET    /:id               # VPC 상세
│   ├── PUT    /:id               # VPC 수정
│   └── DELETE /:id               # VPC 삭제
├── subnets/
│   ├── GET    /                  # 서브넷 목록
│   ├── POST   /                  # 서브넷 생성
│   ├── GET    /:id               # 서브넷 상세
│   ├── PUT    /:id               # 서브넷 수정
│   └── DELETE /:id               # 서브넷 삭제
└── firewall-rules/
    ├── GET    /                  # 방화벽 규칙 목록
    ├── POST   /                  # 방화벽 규칙 생성
    ├── GET    /:id               # 방화벽 규칙 상세
    ├── PUT    /:id               # 방화벽 규칙 수정
    ├── DELETE /:id               # 방화벽 규칙 삭제
    ├── POST   /:id/ports         # 포트 추가
    └── DELETE /:id/ports         # 포트 삭제
```

### **2.6 기타 관리 기능 (17개)**
```
/api/v1/providers/
├── GET    /                      # 제공업체 목록
├── GET    /:name                 # 제공업체 상세
├── GET    /:name/instances       # 인스턴스 목록
├── GET    /:name/instances/:id   # 인스턴스 상세
├── POST   /:name/instances       # 인스턴스 생성
├── DELETE /:name/instances/:id   # 인스턴스 삭제
├── GET    /:name/regions         # 리전 목록
├── GET    /:name/cost-estimates  # 비용 추정
└── POST   /:name/cost-estimates  # 비용 추정 생성

/api/v1/cost-analysis/
├── GET    /                      # 비용 분석 목록
├── POST   /                      # 비용 분석 생성
└── GET    /:id                   # 비용 분석 상세

/api/v1/notifications/
├── GET    /                      # 알림 목록
├── POST   /                      # 알림 생성
└── PUT    /:id/read              # 알림 읽음 처리

/api/v1/exports/
├── GET    /                      # 내보내기 목록
├── POST   /                      # 내보내기 생성
└── GET    /:id/download          # 내보내기 다운로드

/api/v1/sse/
└── GET    /events                # SSE 이벤트 스트림
```

---

## 📦 **3. DTO 목록 (총 50+ 개)**

### **3.1 VM 관련 DTO (4개)**
| 번호 | DTO명 | 파일 위치 | 용도 |
|------|-------|-----------|------|
| 1 | `VMDTO` | `internal/application/dto/vm_dto.go` | VM 응답 데이터 |
| 2 | `CreateVMRequest` | `internal/application/dto/vm_dto.go` | VM 생성 요청 |
| 3 | `UpdateVMRequest` | `internal/application/dto/vm_dto.go` | VM 수정 요청 |
| 4 | `VMOperationRequest` | `internal/application/dto/vm_dto.go` | VM 작업 요청 |

### **3.2 Kubernetes 관련 DTO (15개)**
| 번호 | DTO명 | 파일 위치 | 용도 |
|------|-------|-----------|------|
| 1 | `CreateClusterRequest` | `internal/application/dto/kubernetes_dto.go` | 클러스터 생성 요청 |
| 2 | `CreateClusterResponse` | `internal/application/dto/kubernetes_dto.go` | 클러스터 생성 응답 |
| 3 | `ClusterInfo` | `internal/application/dto/kubernetes_dto.go` | 클러스터 정보 |
| 4 | `NodePoolInfo` | `internal/application/dto/kubernetes_dto.go` | 노드풀 정보 |
| 5 | `CreateGKEClusterRequest` | `internal/application/dto/kubernetes_dto.go` | GKE 클러스터 생성 요청 |
| 6 | `GKEConfig` | `internal/application/dto/kubernetes_dto.go` | GKE 설정 |
| 7 | `NetworkConfigInfo` | `internal/application/dto/kubernetes_dto.go` | 네트워크 설정 정보 |
| 8 | `NodePoolSummaryInfo` | `internal/application/dto/kubernetes_dto.go` | 노드풀 요약 정보 |
| 9 | `SecurityConfigInfo` | `internal/application/dto/kubernetes_dto.go` | 보안 설정 정보 |
| 10 | `KubeconfigResponse` | `internal/application/dto/kubernetes_dto.go` | Kubeconfig 응답 |
| 11 | `CreateNodePoolRequest` | `internal/application/dto/kubernetes_dto.go` | 노드풀 생성 요청 |
| 12 | `ScaleNodePoolRequest` | `internal/application/dto/kubernetes_dto.go` | 노드풀 스케일링 요청 |
| 13 | `UpgradeClusterRequest` | `internal/application/dto/kubernetes_dto.go` | 클러스터 업그레이드 요청 |
| 14 | `NodeInfo` | `internal/application/dto/kubernetes_dto.go` | 노드 정보 |
| 15 | `ClusterMetrics` | `internal/application/dto/kubernetes_dto.go` | 클러스터 메트릭 |

### **3.3 네트워크 관련 DTO (20개)**
| 번호 | DTO명 | 파일 위치 | 용도 |
|------|-------|-----------|------|
| 1 | `VPCInfo` | `internal/application/dto/network_dto.go` | VPC 정보 |
| 2 | `SubnetInfo` | `internal/application/dto/network_dto.go` | 서브넷 정보 |
| 3 | `SecurityGroupInfo` | `internal/application/dto/network_dto.go` | 보안그룹 정보 |
| 4 | `CreateVPCRequest` | `internal/application/dto/network_dto.go` | VPC 생성 요청 |
| 5 | `CreateSubnetRequest` | `internal/application/dto/network_dto.go` | 서브넷 생성 요청 |
| 6 | `CreateSecurityGroupRequest` | `internal/application/dto/network_dto.go` | 보안그룹 생성 요청 |
| 7 | `UpdateVPCRequest` | `internal/application/dto/network_dto.go` | VPC 수정 요청 |
| 8 | `UpdateSubnetRequest` | `internal/application/dto/network_dto.go` | 서브넷 수정 요청 |
| 9 | `UpdateSecurityGroupRequest` | `internal/application/dto/network_dto.go` | 보안그룹 수정 요청 |
| 10 | `SecurityGroupRuleInfo` | `internal/application/dto/network_dto.go` | 보안그룹 규칙 정보 |
| 11 | `GatewayInfo` | `internal/application/dto/network_dto.go` | 게이트웨이 정보 |
| 12 | `ListVPCsRequest` | `internal/application/dto/network_dto.go` | VPC 목록 조회 요청 |
| 13 | `ListSubnetsRequest` | `internal/application/dto/network_dto.go` | 서브넷 목록 조회 요청 |
| 14 | `ListSecurityGroupsRequest` | `internal/application/dto/network_dto.go` | 보안그룹 목록 조회 요청 |
| 15 | `GetVPCRequest` | `internal/application/dto/network_dto.go` | VPC 상세 조회 요청 |
| 16 | `GetSubnetRequest` | `internal/application/dto/network_dto.go` | 서브넷 상세 조회 요청 |
| 17 | `GetSecurityGroupRequest` | `internal/application/dto/network_dto.go` | 보안그룹 상세 조회 요청 |
| 18 | `DeleteVPCRequest` | `internal/application/dto/network_dto.go` | VPC 삭제 요청 |
| 19 | `DeleteSubnetRequest` | `internal/application/dto/network_dto.go` | 서브넷 삭제 요청 |
| 20 | `DeleteSecurityGroupRequest` | `internal/application/dto/network_dto.go` | 보안그룹 삭제 요청 |

### **3.4 GCP 특화 DTO (5개)**
| 번호 | DTO명 | 파일 위치 | 용도 |
|------|-------|-----------|------|
| 1 | `GCPCredentialData` | `internal/application/dto/gcp_dto.go` | GCP 자격증명 데이터 |
| 2 | `GCPProjectInfo` | `internal/application/dto/gcp_dto.go` | GCP 프로젝트 정보 |
| 3 | `GCPRegionInfo` | `internal/application/dto/gcp_dto.go` | GCP 리전 정보 |
| 4 | `GCPZoneInfo` | `internal/application/dto/gcp_dto.go` | GCP 존 정보 |
| 5 | `GCPInstanceType` | `internal/application/dto/gcp_dto.go` | GCP 인스턴스 타입 |

### **3.5 사용자 및 워크스페이스 DTO (6개)**
| 번호 | DTO명 | 파일 위치 | 용도 |
|------|-------|-----------|------|
| 1 | `UserDTO` | `internal/application/dto/user_dto.go` | 사용자 정보 |
| 2 | `CreateUserRequest` | `internal/application/dto/user_dto.go` | 사용자 생성 요청 |
| 3 | `UpdateUserRequest` | `internal/application/dto/user_dto.go` | 사용자 수정 요청 |
| 4 | `WorkspaceDTO` | `internal/application/dto/workspace_dto.go` | 워크스페이스 정보 |
| 5 | `CreateWorkspaceRequest` | `internal/application/dto/workspace_dto.go` | 워크스페이스 생성 요청 |
| 6 | `UpdateWorkspaceRequest` | `internal/application/dto/workspace_dto.go` | 워크스페이스 수정 요청 |

---

## 📊 **4. 가상 자원 관리 시스템 상세 분석**

### **4.1 VM 관련 인터페이스 상세**

#### **VMService 인터페이스 메서드 (13개)**
```go
type VMService interface {
    // VM 생명주기 관리 (5개)
    CreateVM(ctx context.Context, req domain.CreateVMRequest) (*domain.VM, error)
    GetVM(ctx context.Context, id string) (*domain.VM, error)
    UpdateVM(ctx context.Context, id string, req domain.UpdateVMRequest) (*domain.VM, error)
    DeleteVM(ctx context.Context, id string) error
    ListVMs(ctx context.Context, limit, offset int) ([]*domain.VM, error)
    
    // 워크스페이스별 VM 관리 (2개)
    GetWorkspaceVMs(ctx context.Context, workspaceID string) ([]*domain.VM, error)
    GetVMsByStatus(ctx context.Context, status domain.VMStatus, limit, offset int) ([]*domain.VM, error)
    
    // VM 운영 관리 (4개)
    StartVM(ctx context.Context, id string) error
    StopVM(ctx context.Context, id string) error
    RestartVM(ctx context.Context, id string) error
    TerminateVM(ctx context.Context, id string) error
    
    // 상태 관리 (2개)
    UpdateVMStatus(ctx context.Context, id string, status domain.VMStatus) error
    GetVMStatus(ctx context.Context, id string) (domain.VMStatus, error)
    
    // 접근 제어 (1개)
    CheckVMAccess(ctx context.Context, userID string, vmID string) (bool, error)
}
```

#### **VMRepository 인터페이스 메서드 (12개)**
```go
type VMRepository interface {
    // 기본 CRUD 작업 (4개)
    Create(vm *domain.VM) error
    GetByID(ctx context.Context, id string) (*domain.VM, error)
    Update(vm *domain.VM) error
    Delete(ctx context.Context, id string) error
    
    // 목록 작업 (3개)
    GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*domain.VM, error)
    List(limit, offset int) ([]*domain.VM, error)
    Count() (int64, error)
    
    // 검색 작업 (3개)
    Search(query string, limit, offset int) ([]*domain.VM, error)
    GetByStatus(status domain.VMStatus, limit, offset int) ([]*domain.VM, error)
    GetByProvider(provider string, limit, offset int) ([]*domain.VM, error)
    
    // 상태 작업 (2개)
    UpdateStatus(id string, status domain.VMStatus) error
    GetStatus(id string) (domain.VMStatus, error)
}
```

### **4.2 VM REST API 엔드포인트 (예상 8개)**
```
/api/v1/vms/
├── GET    /                      # VM 목록 조회
├── POST   /                      # VM 생성
├── GET    /:id                   # VM 상세 조회
├── PUT    /:id                   # VM 수정
├── DELETE /:id                   # VM 삭제
├── POST   /:id/start             # VM 시작
├── POST   /:id/stop              # VM 중지
├── POST   /:id/restart           # VM 재시작
└── POST   /:id/terminate         # VM 종료
```

### **4.3 VM 관련 DTO 상세 (4개)**
```go
// 1. VMDTO - VM 응답 데이터
type VMDTO struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    WorkspaceID string            `json:"workspace_id"`
    Provider    string            `json:"provider"`
    Type        string            `json:"type"`
    Region      string            `json:"region"`
    ImageID     string            `json:"image_id"`
    Status      string            `json:"status"`
    InstanceID  string            `json:"instance_id,omitempty"`
    Metadata    map[string]string `json:"metadata,omitempty"`
    CreatedAt   time.Time         `json:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at"`
}

// 2. CreateVMRequest - VM 생성 요청
type CreateVMRequest struct {
    Name        string            `json:"name" validate:"required,min=3,max=100"`
    WorkspaceID string            `json:"workspace_id" validate:"required"`
    Provider    string            `json:"provider" validate:"required"`
    Type        string            `json:"type" validate:"required"`
    Region      string            `json:"region" validate:"required"`
    ImageID     string            `json:"image_id" validate:"required"`
    Metadata    map[string]string `json:"metadata,omitempty"`
}

// 3. UpdateVMRequest - VM 수정 요청
type UpdateVMRequest struct {
    Name     string            `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
    Type     string            `json:"type,omitempty"`
    Region   string            `json:"region,omitempty"`
    ImageID  string            `json:"image_id,omitempty"`
    Metadata map[string]string `json:"metadata,omitempty"`
}

// 4. VMOperationRequest - VM 작업 요청
type VMOperationRequest struct {
    Operation string `json:"operation" validate:"required,oneof=start stop restart terminate"`
}
```

---

## 🎯 **5. 시스템 확장성 분석**

### **5.1 현재 구현 상태**
- ✅ **Kubernetes 관리**: AWS EKS, GCP GKE 완전 구현
- ✅ **네트워크 관리**: AWS VPC, GCP VPC 완전 구현
- ✅ **인증 및 권한**: JWT, RBAC 완전 구현
- ✅ **자격증명 관리**: 암호화된 저장 완전 구현
- 🔄 **VM 관리**: 인터페이스 정의 완료, 구현 진행 중
- 🔄 **Azure 지원**: 계획 단계
- 🔄 **NCP 지원**: 계획 단계

### **5.2 확장 계획**
- **VM 핸들러 구현**: REST API 엔드포인트 구현 예정
- **Azure AKS 지원**: Azure Kubernetes Service 통합
- **Azure VM 지원**: Azure Virtual Machine 통합
- **NCP NKS 지원**: Naver Cloud Platform Kubernetes Service 통합
- **실시간 모니터링**: WebSocket 기반 실시간 상태 업데이트
- **비용 최적화**: AI 기반 리소스 최적화 추천

---

## 📈 **6. 성능 및 확장성 지표**

### **6.1 API 성능 목표**
- **응답 시간**: 평균 150ms 이하
- **동시 처리**: 1000+ 요청/초
- **가용성**: 99.9% 이상
- **확장성**: 수평 확장 지원

### **6.2 데이터 처리 능력**
- **VM 관리**: 10,000+ VM 동시 관리
- **클러스터 관리**: 1,000+ 클러스터 동시 관리
- **사용자 관리**: 100,000+ 사용자 지원
- **워크스페이스**: 10,000+ 워크스페이스 지원

이러한 체계적인 인터페이스, API, DTO 설계를 통해 **확장 가능하고 유지보수가 용이한 멀티 클라우드 관리 시스템**을 구축할 수 있습니다.

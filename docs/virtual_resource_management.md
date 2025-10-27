# 가상 자원 관리 시스템 기능 설명

## 📋 **개요**
SkyClust의 가상 자원 관리 시스템은 멀티 클라우드 환경에서 VM(가상머신) 인스턴스를 통합 관리하는 핵심 기능입니다. AWS EC2, GCP Compute Engine, Azure VM 등 다양한 클라우드 제공업체의 VM을 단일 인터페이스로 관리할 수 있습니다.

---

## 🏗️ **1. 가상 자원 관리 모델 설계**

### **1.1 도메인 모델**

#### **VM 엔티티 구조**
```go
type VM struct {
    ID          string                 // VM 고유 식별자
    Name        string                 // VM 이름
    WorkspaceID string                 // 워크스페이스 ID
    Provider    string                 // 클라우드 제공업체 (aws, gcp, azure)
    InstanceID  string                 // 클라우드 인스턴스 ID
    Status      VMStatus               // VM 상태
    Type        string                 // 인스턴스 타입
    Region      string                 // 리전
    ImageID     string                 // 이미지 ID
    CPUs        int                    // CPU 코어 수
    Memory      int                    // 메모리 (MB)
    Storage     int                    // 스토리지 (GB)
    CreatedAt   time.Time              // 생성 시간
    UpdatedAt   time.Time              // 수정 시간
    Metadata    map[string]interface{} // 메타데이터
}
```

#### **VM 상태 관리**
```go
type VMStatus string

const (
    VMStatusPending    VMStatus = "pending"    // 대기 중
    VMStatusRunning    VMStatus = "running"    // 실행 중
    VMStatusStopped    VMStatus = "stopped"    // 중지됨
    VMStatusStopping   VMStatus = "stopping"   // 중지 중
    VMStatusStarting   VMStatus = "starting"   // 시작 중
    VMStatusTerminated VMStatus = "terminated" // 종료됨
    VMStatusError      VMStatus = "error"      // 오류
)
```

### **1.2 비즈니스 로직 모델**

#### **VM 상태 전이 규칙**
- **시작 가능**: `stopped`, `error` 상태에서만 시작 가능
- **중지 가능**: `running` 상태에서만 중지 가능
- **재시작 가능**: `running`, `stopped` 상태에서 재시작 가능
- **종료 가능**: `terminated` 상태가 아닌 모든 상태에서 종료 가능

#### **워크스페이스 기반 격리**
- 각 VM은 특정 워크스페이스에 속함
- 워크스페이스별로 VM 접근 권한 제어
- 멀티 테넌트 환경 지원

---

## 🔌 **2. 가상 자원 관리 인터페이스 설계**

### **2.1 서비스 인터페이스**

#### **VMService 인터페이스**
```go
type VMService interface {
    // VM 생명주기 관리
    CreateVM(ctx context.Context, req CreateVMRequest) (*VM, error)
    GetVM(ctx context.Context, id string) (*VM, error)
    UpdateVM(ctx context.Context, id string, req UpdateVMRequest) (*VM, error)
    DeleteVM(ctx context.Context, id string) error
    ListVMs(ctx context.Context, limit, offset int) ([]*VM, error)
    
    // 워크스페이스별 VM 관리
    GetWorkspaceVMs(ctx context.Context, workspaceID string) ([]*VM, error)
    GetVMsByStatus(ctx context.Context, status VMStatus, limit, offset int) ([]*VM, error)
    
    // VM 운영 관리
    StartVM(ctx context.Context, id string) error
    StopVM(ctx context.Context, id string) error
    RestartVM(ctx context.Context, id string) error
    TerminateVM(ctx context.Context, id string) error
    
    // 상태 관리
    UpdateVMStatus(ctx context.Context, id string, status VMStatus) error
    GetVMStatus(ctx context.Context, id string) (VMStatus, error)
    
    // 접근 제어
    CheckVMAccess(ctx context.Context, userID string, vmID string) (bool, error)
}
```

### **2.2 클라우드 제공업체 인터페이스**

#### **CloudProviderService 인터페이스**
```go
type CloudProviderService interface {
    // 인스턴스 생명주기 관리
    CreateInstance(ctx context.Context, provider string, req CreateInstanceRequest) (*CloudInstance, error)
    GetInstance(ctx context.Context, provider, instanceID string) (*CloudInstance, error)
    DeleteInstance(ctx context.Context, provider, instanceID string) error
    
    // 인스턴스 운영 관리
    StartInstance(ctx context.Context, provider, instanceID string) error
    StopInstance(ctx context.Context, provider, instanceID string) error
    GetInstanceStatus(ctx context.Context, provider, instanceID string) (string, error)
}
```

### **2.3 REST API 인터페이스**

#### **API 엔드포인트 구조**
```
/api/v1/vms/
├── GET    /                    # VM 목록 조회
├── POST   /                    # VM 생성
├── GET    /:id                 # VM 상세 조회
├── PUT    /:id                 # VM 수정
├── DELETE /:id                 # VM 삭제
├── POST   /:id/start           # VM 시작
├── POST   /:id/stop            # VM 중지
├── POST   /:id/restart         # VM 재시작
└── POST   /:id/terminate       # VM 종료
```

#### **요청/응답 DTO**
```go
// VM 생성 요청
type CreateVMRequest struct {
    Name        string            `json:"name" validate:"required,min=3,max=100"`
    WorkspaceID string            `json:"workspace_id" validate:"required"`
    Provider    string            `json:"provider" validate:"required"`
    Type        string            `json:"type" validate:"required"`
    Region      string            `json:"region" validate:"required"`
    ImageID     string            `json:"image_id" validate:"required"`
    Metadata    map[string]string `json:"metadata,omitempty"`
}

// VM 응답
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
```

---

## ⚙️ **3. 가상 자원 관리 기능 개발**

### **3.1 핵심 기능**

#### **VM 생성 기능**
```go
func (s *VMService) CreateVM(ctx context.Context, req CreateVMRequest) (*VM, error) {
    // 1. 워크스페이스 권한 확인
    if err := s.checkWorkspaceAccess(ctx, req.WorkspaceID); err != nil {
        return nil, err
    }
    
    // 2. VM 엔티티 생성
    vm := &domain.VM{
        ID:          uuid.New().String(),
        Name:        req.Name,
        WorkspaceID: req.WorkspaceID,
        Provider:    req.Provider,
        Status:      domain.VMStatusPending,
        Type:        req.Type,
        Region:      req.Region,
        ImageID:     req.ImageID,
        Metadata:    req.Metadata,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
    
    // 3. 클라우드 인스턴스 생성
    instanceReq := CreateInstanceRequest{
        Name:     req.Name,
        Type:     req.Type,
        Region:   req.Region,
        ImageID:  req.ImageID,
        Metadata: req.Metadata,
    }
    
    instance, err := s.cloudProvider.CreateInstance(ctx, req.Provider, instanceReq)
    if err != nil {
        return nil, fmt.Errorf("failed to create cloud instance: %w", err)
    }
    
    // 4. VM 정보 업데이트
    vm.InstanceID = instance.ID
    vm.Status = domain.VMStatusStarting
    
    // 5. 데이터베이스 저장
    if err := s.vmRepo.Create(ctx, vm); err != nil {
        return nil, fmt.Errorf("failed to save VM: %w", err)
    }
    
    // 6. 감사 로그 기록
    s.auditLogRepo.LogAction(ctx, vm.WorkspaceID, "CREATE", "vm", vm.ID, req.Provider, req)
    
    return vm, nil
}
```

#### **VM 상태 관리 기능**
```go
func (s *VMService) StartVM(ctx context.Context, id string) error {
    // 1. VM 조회
    vm, err := s.vmRepo.GetByID(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to get VM: %w", err)
    }
    
    // 2. 상태 검증
    if !vm.CanStart() {
        return fmt.Errorf("VM cannot be started in current status: %s", vm.Status)
    }
    
    // 3. 클라우드 인스턴스 시작
    if err := s.cloudProvider.StartInstance(ctx, vm.Provider, vm.InstanceID); err != nil {
        return fmt.Errorf("failed to start cloud instance: %w", err)
    }
    
    // 4. 상태 업데이트
    vm.Status = domain.VMStatusStarting
    vm.UpdatedAt = time.Now()
    
    if err := s.vmRepo.Update(ctx, vm); err != nil {
        return fmt.Errorf("failed to update VM status: %w", err)
    }
    
    // 5. 이벤트 발행
    s.eventBus.Publish("vm.started", map[string]interface{}{
        "vm_id": vm.ID,
        "workspace_id": vm.WorkspaceID,
        "provider": vm.Provider,
    })
    
    return nil
}
```

### **3.2 멀티 클라우드 지원**

#### **Provider별 Dispatch 패턴**
```go
func (s *VMService) createCloudInstance(ctx context.Context, provider string, req CreateInstanceRequest) (*CloudInstance, error) {
    switch provider {
    case "aws":
        return s.createAWSInstance(ctx, req)
    case "gcp":
        return s.createGCPInstance(ctx, req)
    case "azure":
        return s.createAzureInstance(ctx, req)
    default:
        return nil, fmt.Errorf("unsupported provider: %s", provider)
    }
}

func (s *VMService) createAWSInstance(ctx context.Context, req CreateInstanceRequest) (*CloudInstance, error) {
    // AWS EC2 인스턴스 생성 로직
    ec2Client := s.getEC2Client(req.Region)
    
    input := &ec2.RunInstancesInput{
        ImageId:      aws.String(req.ImageID),
        InstanceType: ec2Types.InstanceType(req.Type),
        MinCount:     aws.Int32(1),
        MaxCount:     aws.Int32(1),
        TagSpecifications: []ec2Types.TagSpecification{
            {
                ResourceType: ec2Types.ResourceTypeInstance,
                Tags: []ec2Types.Tag{
                    {Key: aws.String("Name"), Value: aws.String(req.Name)},
                },
            },
        },
    }
    
    result, err := ec2Client.RunInstances(ctx, input)
    if err != nil {
        return nil, err
    }
    
    instance := result.Instances[0]
    return &CloudInstance{
        ID:       *instance.InstanceId,
        Status:   string(instance.State.Name),
        Type:     string(instance.InstanceType),
        Region:   req.Region,
        ImageID:  *instance.ImageId,
        Metadata: req.Metadata,
    }, nil
}
```

### **3.3 실시간 상태 동기화**

#### **상태 폴링 및 업데이트**
```go
func (s *VMService) SyncVMStatus(ctx context.Context, vmID string) error {
    vm, err := s.vmRepo.GetByID(ctx, vmID)
    if err != nil {
        return err
    }
    
    // 클라우드에서 실제 상태 조회
    cloudStatus, err := s.cloudProvider.GetInstanceStatus(ctx, vm.Provider, vm.InstanceID)
    if err != nil {
        return err
    }
    
    // 상태 매핑
    mappedStatus := s.mapCloudStatusToVMStatus(cloudStatus)
    
    // 상태가 변경된 경우에만 업데이트
    if vm.Status != mappedStatus {
        vm.Status = mappedStatus
        vm.UpdatedAt = time.Now()
        
        if err := s.vmRepo.Update(ctx, vm); err != nil {
            return err
        }
        
        // 상태 변경 이벤트 발행
        s.eventBus.Publish("vm.status.changed", map[string]interface{}{
            "vm_id": vm.ID,
            "old_status": vm.Status,
            "new_status": mappedStatus,
        })
    }
    
    return nil
}
```

### **3.4 워크스페이스 기반 접근 제어**

#### **권한 검증**
```go
func (s *VMService) CheckVMAccess(ctx context.Context, userID string, vmID string) (bool, error) {
    vm, err := s.vmRepo.GetByID(ctx, vmID)
    if err != nil {
        return false, err
    }
    
    // 사용자의 워크스페이스 접근 권한 확인
    hasAccess, err := s.workspaceRepo.CheckUserAccess(ctx, userID, vm.WorkspaceID)
    if err != nil {
        return false, err
    }
    
    return hasAccess, nil
}
```

---

## 📊 **주요 특징 및 장점**

### **3.1 통합 관리**
- **단일 인터페이스**: AWS, GCP, Azure VM을 동일한 API로 관리
- **일관된 응답**: 클라우드별 차이점을 추상화한 통합 응답 형식
- **자동 변환**: 클라우드별 상태 및 메타데이터 자동 매핑

### **3.2 실시간 모니터링**
- **상태 동기화**: 클라우드 인스턴스 상태를 실시간으로 동기화
- **이벤트 기반**: 상태 변경 시 실시간 이벤트 발행
- **WebSocket 지원**: 프론트엔드에서 실시간 상태 업데이트

### **3.3 보안 및 격리**
- **워크스페이스 격리**: 멀티 테넌트 환경에서 VM 격리
- **RBAC 지원**: 역할 기반 접근 제어
- **감사 로깅**: 모든 VM 작업 추적 및 로깅

### **3.4 확장성**
- **Provider 확장**: 새로운 클라우드 제공업체 쉽게 추가
- **마이크로서비스**: VM 서비스 독립적 확장 가능
- **이벤트 기반**: 느슨한 결합으로 유연한 아키텍처

---

## 🎯 **사용 사례**

### **개발 환경 관리**
- 개발팀별 독립적인 VM 환경 제공
- 필요 시 VM 자동 생성 및 배포
- 개발 완료 후 자동 정리

### **테스트 환경 오케스트레이션**
- 테스트 시나리오별 VM 클러스터 구성
- 멀티 클라우드 테스트 환경 구축
- 테스트 완료 후 리소스 자동 해제

### **프로덕션 환경 관리**
- 고가용성을 위한 멀티 클라우드 배포
- 클라우드별 장애 대응 및 페일오버
- 비용 최적화를 위한 동적 스케일링

이러한 가상 자원 관리 시스템을 통해 **멀티 클라우드 환경의 복잡성을 단순화**하고, **운영 효율성을 극대화**할 수 있습니다.

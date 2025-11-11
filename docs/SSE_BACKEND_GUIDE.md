# Backend SSE 적용 가이드

이 문서는 SkyClust 백엔드에서 새로운 API나 기능을 추가할 때 SSE(Server-Sent Events)를 적용하는 방법을 설명합니다.

## 📋 목차

1. [개요](#개요)
2. [이벤트 발행 방법](#이벤트-발행-방법)
3. [이벤트 구조 및 토픽 형식](#이벤트-구조-및-토픽-형식)
4. [예제: 리소스 생성/수정/삭제 시 이벤트 발행](#예제-리소스-생성수정삭제-시-이벤트-발행)
5. [SSE 핸들러에서 이벤트 구독](#sse-핸들러에서-이벤트-구독)
6. [이벤트 데이터 구조](#이벤트-데이터-구조)
7. [모범 사례](#모범-사례)

---

## 개요

SkyClust 백엔드는 NATS를 사용하여 리소스 변경 이벤트를 발행하고, SSE 핸들러가 이러한 이벤트를 구독하여 클라이언트에 전달합니다.

### 아키텍처 흐름

```
Service Layer (리소스 생성/수정/삭제)
    ↓
Event Publisher (messaging.Publisher)
    ↓
NATS Bus (messaging.NATSService)
    ↓
NATS Subject (cmp.events.{eventType})
    ↓
SSE Handler (구독 및 브로드캐스트)
    ↓
Frontend (EventSource)
```

---

## 이벤트 발행 방법

### 1. Event Publisher 초기화

Service 생성 시 `messaging.Publisher`를 초기화합니다:

```go
import (
    "skyclust/internal/infrastructure/messaging"
)

type Service struct {
    // ... 기타 필드
    eventPublisher *messaging.Publisher
}

func NewService(
    // ... 기타 의존성
    eventBus messaging.Bus,
    logger *zap.Logger,
) *Service {
    eventPublisher := messaging.NewPublisher(eventBus, logger)
    
    return &Service{
        // ... 기타 필드
        eventPublisher: eventPublisher,
    }
}
```

### 2. 리소스별 이벤트 발행 메서드

`messaging.Publisher`는 리소스별 이벤트 발행 메서드를 제공합니다:

#### Kubernetes 리소스
- `PublishKubernetesClusterEvent(ctx, provider, credentialID, region, action, data)`
- `PublishKubernetesNodePoolEvent(ctx, provider, credentialID, clusterName, action, data)`
- `PublishKubernetesNodeEvent(ctx, provider, credentialID, clusterName, action, data)`

#### Network 리소스
- `PublishVPCEvent(ctx, provider, credentialID, region, action, data)`
- `PublishSubnetEvent(ctx, provider, credentialID, vpcID, action, data)`
- `PublishSecurityGroupEvent(ctx, provider, credentialID, region, action, data)`

#### VM 리소스
- `PublishVMEvent(ctx, provider, workspaceID, vmID, action, data)`

#### 일반 이벤트
- `PublishKubernetesEvent(ctx, provider, credentialID, region, resource, action, data)`
- `PublishNetworkEvent(ctx, provider, credentialID, region, resource, action, data)`

---

## 이벤트 구조 및 토픽 형식

### NATS Subject 형식

모든 이벤트는 `cmp.events.{eventType}` 형식의 NATS subject로 발행됩니다.

#### Kubernetes 이벤트 토픽
```
kubernetes.{provider}.{credential_id}.{region}.{resource}.{action}
```

예시:
- `kubernetes.aws.cred-123.ap-northeast-2.clusters.created`
- `kubernetes.gcp.cred-456.asia-northeast3.clusters.updated`
- `kubernetes.azure.cred-789.koreacentral.clusters.deleted`

#### Network 이벤트 토픽
```
network.{provider}.{credential_id}.{region}.{resource}.{action}
```

예시:
- `network.aws.cred-123.ap-northeast-2.vpcs.created`
- `network.gcp.cred-456.asia-northeast3.subnets.updated`
- `network.azure.cred-789.koreacentral.security-groups.deleted`

#### VM 이벤트 토픽
```
vm.{provider}.{workspace_id}.{action}
```

예시:
- `vm.aws.ws-123.created`
- `vm.gcp.ws-456.updated`
- `vm.azure.ws-789.deleted`

### Action 타입

- `created`: 리소스 생성
- `updated`: 리소스 수정
- `deleted`: 리소스 삭제
- `list`: 리소스 목록 갱신 (동기화 워커 등에서 사용)

---

## 예제: 리소스 생성/수정/삭제 시 이벤트 발행

### 예제 1: Kubernetes 클러스터 생성

```go
func (s *Service) CreateCluster(ctx context.Context, req CreateClusterRequest) (*ClusterResponse, error) {
    // 1. 클러스터 생성 로직
    cluster, err := s.createClusterInternal(ctx, req)
    if err != nil {
        return nil, err
    }

    // 2. 이벤트 데이터 준비
    credentialID := credential.ID.String()
    clusterData := map[string]interface{}{
        "cluster_id":   cluster.ID,
        "cluster_name": cluster.Name,
        "name":         cluster.Name, // frontend 호환성
        "version":      cluster.Version,
        "status":       cluster.Status,
        "region":       cluster.Region,
        "provider":     credential.Provider,
        "credential_id": credentialID,
    }

    // 3. 이벤트 발행
    if err := s.eventPublisher.PublishKubernetesClusterEvent(
        ctx,
        credential.Provider,
        credentialID,
        req.Region,
        "created",
        clusterData,
    ); err != nil {
        // 이벤트 발행 실패는 치명적이지 않으므로 경고만 로깅
        s.logger.Warn("Failed to publish Kubernetes cluster created event",
            zap.String("provider", credential.Provider),
            zap.String("credential_id", credentialID),
            zap.String("cluster_name", cluster.Name),
            zap.Error(err))
    }

    return cluster, nil
}
```

### 예제 2: VPC 생성

```go
func (s *Service) CreateVPC(ctx context.Context, credential *domain.Credential, req CreateVPCRequest) (*VPCInfo, error) {
    // 1. VPC 생성 로직
    vpc, err := s.createVPCInternal(ctx, credential, req)
    if err != nil {
        return nil, err
    }

    // 2. 이벤트 데이터 준비
    credentialID := credential.ID.String()
    vpcData := map[string]interface{}{
        "vpc_id":        vpc.ID,
        "name":          vpc.Name,
        "state":         vpc.State,
        "region":        vpc.Region,
        "provider":      credential.Provider,
        "credential_id": credentialID,
    }

    // 3. 이벤트 발행
    if err := s.eventPublisher.PublishVPCEvent(
        ctx,
        credential.Provider,
        credentialID,
        req.Region,
        "created",
        vpcData,
    ); err != nil {
        s.logger.Warn("Failed to publish VPC created event",
            zap.String("provider", credential.Provider),
            zap.String("credential_id", credentialID),
            zap.String("vpc_id", vpc.ID),
            zap.Error(err))
    }

    return vpc, nil
}
```

### 예제 3: VM 생성

```go
func (s *Service) CreateVM(ctx context.Context, req CreateVMRequest) (*domain.VM, error) {
    // 1. VM 생성 로직
    vm, err := s.createVMInternal(ctx, req)
    if err != nil {
        return nil, err
    }

    // 2. 이벤트 데이터 준비
    vmData := map[string]interface{}{
        "vm_id":        vm.ID,
        "workspace_id": vm.WorkspaceID,
        "provider":     vm.Provider,
        "name":         vm.Name,
        "status":       string(vm.Status),
        "region":       vm.Region,
        "type":         vm.Type,
    }

    // 3. 이벤트 발행
    if err := s.eventService.Publish(ctx, domain.EventVMCreated, vmData); err != nil {
        s.logger.Error("Failed to publish VM created event", zap.Error(err))
    }

    return vm, nil
}
```

### 예제 4: 리소스 수정

```go
func (s *Service) UpdateCluster(ctx context.Context, req UpdateClusterRequest) (*ClusterResponse, error) {
    // 1. 클러스터 수정 로직
    cluster, err := s.updateClusterInternal(ctx, req)
    if err != nil {
        return nil, err
    }

    // 2. 이벤트 데이터 준비
    credentialID := credential.ID.String()
    clusterData := map[string]interface{}{
        "cluster_id":   cluster.ID,
        "cluster_name": cluster.Name,
        "name":         cluster.Name,
        "version":      cluster.Version,
        "status":       cluster.Status,
        "region":       cluster.Region,
        "provider":     credential.Provider,
        "credential_id": credentialID,
    }

    // 3. 이벤트 발행 (action: "updated")
    if err := s.eventPublisher.PublishKubernetesClusterEvent(
        ctx,
        credential.Provider,
        credentialID,
        cluster.Region,
        "updated",
        clusterData,
    ); err != nil {
        s.logger.Warn("Failed to publish Kubernetes cluster updated event",
            zap.String("provider", credential.Provider),
            zap.String("credential_id", credentialID),
            zap.String("cluster_name", cluster.Name),
            zap.Error(err))
    }

    return cluster, nil
}
```

### 예제 5: 리소스 삭제

```go
func (s *Service) DeleteCluster(ctx context.Context, req DeleteClusterRequest) error {
    // 1. 삭제 전 정보 저장 (이벤트 발행용)
    cluster, err := s.getCluster(ctx, req.ClusterID)
    if err != nil {
        return err
    }

    credentialID := credential.ID.String()
    clusterData := map[string]interface{}{
        "cluster_id":   cluster.ID,
        "cluster_name": cluster.Name,
        "name":         cluster.Name,
        "region":       cluster.Region,
        "provider":     credential.Provider,
        "credential_id": credentialID,
    }

    // 2. 클러스터 삭제 로직
    if err := s.deleteClusterInternal(ctx, req); err != nil {
        return err
    }

    // 3. 이벤트 발행 (action: "deleted")
    if err := s.eventPublisher.PublishKubernetesClusterEvent(
        ctx,
        credential.Provider,
        credentialID,
        cluster.Region,
        "deleted",
        clusterData,
    ); err != nil {
        s.logger.Warn("Failed to publish Kubernetes cluster deleted event",
            zap.String("provider", credential.Provider),
            zap.String("credential_id", credentialID),
            zap.String("cluster_name", cluster.Name),
            zap.Error(err))
    }

    return nil
}
```

---

## SSE 핸들러에서 이벤트 구독

SSE 핸들러는 자동으로 NATS 이벤트를 구독합니다. 추가 설정이 필요하지 않습니다.

### 구독 패턴

SSE 핸들러는 다음 패턴으로 NATS subject를 구독합니다:

```go
// Kubernetes 클러스터 이벤트
cmp.events.kubernetes.*.*.*.clusters.created
cmp.events.kubernetes.*.*.*.clusters.updated
cmp.events.kubernetes.*.*.*.clusters.deleted

// Network VPC 이벤트
cmp.events.network.*.*.*.vpcs.created
cmp.events.network.*.*.*.vpcs.updated
cmp.events.network.*.*.*.vpcs.deleted
```

### 필터링

클라이언트는 `POST /api/v1/sse/subscribe` 엔드포인트를 통해 특정 이벤트만 구독할 수 있습니다:

```json
{
  "event_type": "kubernetes.aws.cred-123.ap-northeast-2.clusters.created",
  "filters": {
    "credential_ids": ["cred-123"],
    "regions": ["ap-northeast-2"],
    "providers": ["aws"]
  }
}
```

---

## 이벤트 데이터 구조

### 표준 이벤트 데이터 필드

모든 리소스 이벤트는 다음 필드를 포함해야 합니다:

#### 필수 필드
- `provider`: 클라우드 프로바이더 (aws, gcp, azure)
- `credential_id`: 자격증명 ID
- `region`: 리전 (VM의 경우 `workspace_id` 사용)

#### 리소스별 필수 필드

**Kubernetes Cluster:**
- `cluster_id` 또는 `cluster_name`: 클러스터 식별자
- `name`: 클러스터 이름 (frontend 호환성)
- `version`: Kubernetes 버전
- `status`: 클러스터 상태

**VPC:**
- `vpc_id`: VPC ID
- `name`: VPC 이름
- `state`: VPC 상태

**Subnet:**
- `subnet_id`: Subnet ID
- `vpc_id`: VPC ID
- `name`: Subnet 이름
- `cidr_block`: CIDR 블록

**Security Group:**
- `security_group_id`: Security Group ID
- `name`: Security Group 이름
- `vpc_id`: VPC ID (선택)

**VM:**
- `vm_id`: VM ID
- `workspace_id`: 워크스페이스 ID
- `name`: VM 이름
- `status`: VM 상태
- `region`: 리전

### 이벤트 데이터 예시

```go
// Kubernetes Cluster Created
clusterData := map[string]interface{}{
    "provider":      "aws",
    "credential_id": "cred-123",
    "region":        "ap-northeast-2",
    "cluster_id":    "cluster-abc",
    "cluster_name":  "my-cluster",
    "name":          "my-cluster", // frontend 호환성
    "version":       "1.28",
    "status":        "active",
}

// VPC Created
vpcData := map[string]interface{}{
    "provider":      "aws",
    "credential_id": "cred-123",
    "region":        "ap-northeast-2",
    "vpc_id":       "vpc-abc123",
    "name":         "my-vpc",
    "state":        "available",
    "cidr_block":   "10.0.0.0/16",
}

// VM Created
vmData := map[string]interface{}{
    "provider":     "aws",
    "workspace_id": "ws-123",
    "vm_id":       "vm-abc",
    "name":        "my-vm",
    "status":      "running",
    "region":      "ap-northeast-2",
    "type":        "t3.medium",
}
```

---

## 모범 사례

### 1. 이벤트 발행 실패 처리

이벤트 발행 실패는 치명적이지 않으므로, 리소스 생성/수정/삭제 작업을 중단하지 않습니다:

```go
if err := s.eventPublisher.PublishKubernetesClusterEvent(...); err != nil {
    // 경고만 로깅하고 계속 진행
    s.logger.Warn("Failed to publish event", zap.Error(err))
    // return err 하지 않음!
}
```

### 2. 이벤트 데이터에 리소스 객체 포함 (선택)

Frontend에서 실시간 업데이트를 위해 이벤트 데이터에 전체 리소스 객체를 포함할 수 있습니다:

```go
clusterData := map[string]interface{}{
    // ... 기본 필드
    "cluster": cluster, // 전체 클러스터 객체 (선택)
}
```

### 3. 일관된 필드명 사용

Frontend 호환성을 위해 다음 필드명을 일관되게 사용합니다:

- `credential_id` (snake_case) - Backend 표준
- `credentialId` (camelCase) - Frontend 표준 (필터링용)

SSE 핸들러는 두 형식을 모두 지원합니다.

### 4. 동기화 워커에서 List 이벤트 발행

주기적으로 리소스를 동기화하는 워커에서는 `list` action을 사용합니다:

```go
// 동기화 워커에서
clusterData := map[string]interface{}{
    "cluster_id": cluster.ID,
    "name":       cluster.Name,
    // ...
}
_ = w.eventPublisher.PublishKubernetesClusterEvent(
    ctx,
    credential.Provider,
    credentialID,
    region,
    "list", // list action
    clusterData,
)
```

### 5. 이벤트 발행 타이밍

- **생성**: 리소스 생성 성공 후 즉시 발행
- **수정**: 리소스 수정 성공 후 즉시 발행
- **삭제**: 리소스 삭제 전 정보를 저장하고, 삭제 성공 후 발행

---

## 참고 자료

- [NATS Service 구현](../internal/infrastructure/messaging/nats_service.go)
- [Event Publisher 구현](../internal/infrastructure/messaging/publisher.go)
- [SSE Handler 구현](../internal/application/handlers/sse/handler.go)
- [Frontend SSE 활용 가이드](./SSE_FRONTEND_GUIDE.md)


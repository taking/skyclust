# Azure 기능 개발 브리핑

## 📋 개요
기존 AWS, GCP 구현을 참고하여 Azure 자격증명, AKS 클러스터, 네트워크 기능 개발이 필요한 부분을 분석한 결과입니다.

---

## ✅ 현재 상태

### 1. Azure 자격증명 (Credential)
**상태**: ✅ **기본 검증 로직 완료**

- **구현 완료**:
  - `internal/application/services/common/credential_validator.go`: `validateAzureCredentials` 구현됨
  - 필수 필드 검증: `subscription_id`, `client_id`, `client_secret`, `tenant_id`
  - Frontend validation: `create-credential-use-case.ts`에 Azure 검증 로직 포함

- **추가 필요 사항**: 없음 (기본 자격증명 등록/조회는 이미 지원됨)

---

### 2. Azure AKS 클러스터 (Kubernetes)
**상태**: ❌ **전체 미구현**

#### 2.1 Service Layer (`internal/application/services/kubernetes/service.go`)

**미구현 기능 목록**:

1. **클러스터 관리**
   - ❌ `listAzureAKSClusters`: AKS 클러스터 목록 조회 (line 519)
   - ❌ `getAzureAKSCluster`: AKS 클러스터 상세 조회 (line 647)
   - ❌ `createAzureAKSCluster`: AKS 클러스터 생성 (미구현)
   - ❌ `deleteAzureAKSCluster`: AKS 클러스터 삭제 (line 728)
   - ❌ `getAzureAKSKubeconfig`: AKS kubeconfig 생성 (line 818)

2. **노드 풀/노드 그룹 관리**
   - ❌ `listAzureNodeGroups`: 노드 그룹 목록 조회 (line 1101)
   - ❌ `getAzureNodeGroup`: 노드 그룹 상세 조회 (line 1295)
   - ❌ `createAzureNodePool`: 노드 풀 생성 (미구현)
   - ❌ `deleteAzureNodePool`: 노드 풀 삭제 (미구현)
   - ❌ `scaleAzureNodePool`: 노드 풀 스케일링 (미구현)

3. **헬퍼 함수**
   - ❌ `extractAzureCredentials`: Azure 자격증명 추출 (AWS `extractAWSCredentials` 참고)
   - ❌ `createAzureContainerServiceClient`: Azure Container Service 클라이언트 생성 (GCP `createGCPContainerService` 참고)
   - ❌ `handleAzureError`: Azure SDK 에러 처리 (AWS `handleAWSError` 참고)

#### 2.2 Handler Layer (`internal/application/handlers/kubernetes/providers/azure_handler.go`)

**현재 상태**: 모든 메서드가 `NotImplemented` 반환

**구현 필요 메서드**:
- `CreateCluster`: AKS 클러스터 생성
- `ListClusters`: AKS 클러스터 목록 조회
- `GetCluster`: AKS 클러스터 상세 조회
- `DeleteCluster`: AKS 클러스터 삭제
- `GetKubeconfig`: kubeconfig 생성
- `CreateNodePool`: 노드 풀 생성
- `ListNodePools`: 노드 풀 목록 조회
- `GetNodePool`: 노드 풀 상세 조회
- `DeleteNodePool`: 노드 풀 삭제
- `ScaleNodePool`: 노드 풀 스케일링
- `CreateNodeGroup`: 노드 그룹 생성 (AKS는 Node Pool 사용)
- `ListNodeGroups`: 노드 그룹 목록 조회
- `GetNodeGroup`: 노드 그룹 상세 조회
- `DeleteNodeGroup`: 노드 그룹 삭제

**참고 구현**:
- AWS: `internal/application/handlers/kubernetes/providers/aws_handler.go`
- GCP: `internal/application/handlers/kubernetes/providers/gcp_handler.go`

#### 2.3 DTO (Data Transfer Objects)

**필요한 Azure 전용 DTO**:
- `CreateAKSClusterRequest`: AKS 클러스터 생성 요청 (AWS `CreateClusterRequest`, GCP `CreateGKEClusterRequest` 참고)
- `CreateAKSClusterResponse`: AKS 클러스터 생성 응답
- `AKSNodePoolConfig`: AKS 노드 풀 설정 (GCP `GKENodePoolConfig` 참고)
- `AKSNetworkConfig`: AKS 네트워크 설정 (GCP `GKENetworkConfig` 참고)
- `AKSSecurityConfig`: AKS 보안 설정 (GCP `GKESecurityConfig` 참고)

**공통 DTO 재사용**:
- `ListClustersResponse`, `ClusterInfo`, `NodeGroupInfo` 등은 기존 구조 재사용 가능

---

### 3. Azure 네트워크 (Network)
**상태**: ❌ **전체 미구현**

#### 3.1 Service Layer (`internal/application/services/network/service.go`)

**미구현 기능 목록**:

1. **VPC (Virtual Network) 관리**
   - ❌ `listAzureVPCs`: Virtual Network 목록 조회 (line 902, 빈 배열 반환)
   - ❌ `getAzureVPC`: Virtual Network 상세 조회 (line 914)
   - ❌ `createAzureVPC`: Virtual Network 생성 (line 925)
   - ❌ `updateAzureVPC`: Virtual Network 업데이트 (line 943)
   - ❌ `deleteAzureVPC`: Virtual Network 삭제 (line 955)

2. **서브넷 (Subnet) 관리**
   - ❌ `listAzureSubnets`: 서브넷 목록 조회 (line 3947)
   - ❌ `getAzureSubnet`: 서브넷 상세 조회 (line 3953)
   - ❌ `createAzureSubnet`: 서브넷 생성 (line 3959)
   - ❌ `updateAzureSubnet`: 서브넷 업데이트 (line 3965)
   - ❌ `deleteAzureSubnet`: 서브넷 삭제 (line 3971)

3. **보안 그룹 (Network Security Group) 관리**
   - ❌ `listAzureSecurityGroups`: NSG 목록 조회 (line 1330)
   - ❌ `getAzureSecurityGroup`: NSG 상세 조회 (미구현)
   - ❌ `createAzureSecurityGroup`: NSG 생성 (미구현)
   - ❌ `updateAzureSecurityGroup`: NSG 업데이트 (미구현)
   - ❌ `deleteAzureSecurityGroup`: NSG 삭제 (미구현)
   - ❌ `addAzureSecurityGroupRule`: NSG 규칙 추가 (미구현)
   - ❌ `removeAzureSecurityGroupRule`: NSG 규칙 제거 (미구현)

4. **헬퍼 함수**
   - ❌ `createAzureNetworkClient`: Azure Network Management 클라이언트 생성
   - ❌ `extractAzureCredentials`: Azure 자격증명 추출 (Kubernetes와 공유 가능)
   - ❌ `handleAzureNetworkError`: Azure Network API 에러 처리

#### 3.2 Handler Layer (`internal/application/handlers/network/providers/azure_handler.go`)

**현재 상태**: 모든 메서드가 `NotImplemented` 반환

**구현 필요 메서드**:
- `ListVPCs`: Virtual Network 목록 조회
- `CreateVPC`: Virtual Network 생성
- `GetVPC`: Virtual Network 상세 조회
- `UpdateVPC`: Virtual Network 업데이트
- `DeleteVPC`: Virtual Network 삭제
- `ListSubnets`: 서브넷 목록 조회
- `CreateSubnet`: 서브넷 생성
- `GetSubnet`: 서브넷 상세 조회
- `UpdateSubnet`: 서브넷 업데이트
- `DeleteSubnet`: 서브넷 삭제
- `ListSecurityGroups`: NSG 목록 조회
- `CreateSecurityGroup`: NSG 생성
- `GetSecurityGroup`: NSG 상세 조회
- `UpdateSecurityGroup`: NSG 업데이트
- `DeleteSecurityGroup`: NSG 삭제
- `AddSecurityGroupRule`: NSG 규칙 추가
- `RemoveSecurityGroupRule`: NSG 규칙 제거
- `UpdateSecurityGroupRules`: NSG 규칙 일괄 업데이트

**참고 구현**:
- AWS: `internal/application/handlers/network/providers/aws_handler.go`
- GCP: `internal/application/handlers/network/providers/gcp_handler.go`

#### 3.3 DTO (Data Transfer Objects)

**공통 DTO 재사용 가능**:
- `ListVPCsRequest`, `ListVPCsResponse`, `VPCInfo`
- `ListSubnetsRequest`, `ListSubnetsResponse`, `SubnetInfo`
- `ListSecurityGroupsRequest`, `ListSecurityGroupsResponse`, `SecurityGroupInfo`
- `CreateVPCRequest`, `UpdateVPCRequest`, `DeleteVPCRequest`
- `CreateSubnetRequest`, `UpdateSubnetRequest`, `DeleteSubnetRequest`
- `CreateSecurityGroupRequest`, `UpdateSecurityGroupRequest`, `DeleteSecurityGroupRequest`

**Azure 전용 필드 추가 고려**:
- `CreateVPCRequest`에 `ResourceGroup` 필드 추가 필요 (Azure는 리소스 그룹 필수)
- `CreateSubnetRequest`에 `ResourceGroup` 필드 추가 필요

---

## 🔧 필요한 작업

### 1. 의존성 추가 (go.mod)

**필요한 Azure SDK 패키지**:
```go
// Azure Container Service (AKS)
github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v5

// Azure Network Management
github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5

// Azure Identity (인증)
github.com/Azure/azure-sdk-for-go/sdk/azidentity

// Azure Core
github.com/Azure/azure-sdk-for-go/sdk/azcore
```

**설치 명령**:
```bash
go get github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v5
go get github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5
go get github.com/Azure/azure-sdk-for-go/sdk/azidentity
go get github.com/Azure/azure-sdk-for-go/sdk/azcore
```

### 2. Azure 인증 헬퍼 구현

**파일**: `internal/application/services/kubernetes/azure_helpers.go` (신규 생성)
**파일**: `internal/application/services/network/azure_helpers.go` (신규 생성)

**필요 함수**:
- `extractAzureCredentials`: 자격증명 추출 및 검증
- `createAzureClient`: Azure SDK 클라이언트 생성 (Service Principal 인증)
- `handleAzureError`: Azure SDK 에러를 도메인 에러로 변환

**참고**:
- AWS: `internal/application/services/kubernetes/aws_helpers.go`
- GCP: `internal/application/services/kubernetes/service.go`의 `createGCPContainerService` 함수

### 3. AKS 클러스터 서비스 구현

**파일**: `internal/application/services/kubernetes/service.go`

**구현 순서**:
1. `createAzureAKSCluster`: 클러스터 생성
2. `listAzureAKSClusters`: 클러스터 목록 조회
3. `getAzureAKSCluster`: 클러스터 상세 조회
4. `deleteAzureAKSCluster`: 클러스터 삭제
5. `getAzureAKSKubeconfig`: kubeconfig 생성
6. `createAzureNodePool`: 노드 풀 생성
7. `listAzureNodePools`: 노드 풀 목록 조회
8. `getAzureNodePool`: 노드 풀 상세 조회
9. `deleteAzureNodePool`: 노드 풀 삭제
10. `scaleAzureNodePool`: 노드 풀 스케일링

**참고 구현**:
- AWS EKS: `createAWSEKSCluster`, `listAWSEKSClusters`, `getAWSEKSCluster`, `deleteAWSEKSCluster`
- GCP GKE: `createGCPGKEClusterWithAdvanced`, `listGCPGKEClusters`, `getGCPGKECluster`, `deleteGCPGKECluster`

### 4. Azure 네트워크 서비스 구현

**파일**: `internal/application/services/network/service.go`

**구현 순서**:
1. `listAzureVPCs`: Virtual Network 목록 조회
2. `getAzureVPC`: Virtual Network 상세 조회
3. `createAzureVPC`: Virtual Network 생성
4. `updateAzureVPC`: Virtual Network 업데이트
5. `deleteAzureVPC`: Virtual Network 삭제
6. `listAzureSubnets`: 서브넷 목록 조회
7. `getAzureSubnet`: 서브넷 상세 조회
8. `createAzureSubnet`: 서브넷 생성
9. `updateAzureSubnet`: 서브넷 업데이트
10. `deleteAzureSubnet`: 서브넷 삭제
11. `listAzureSecurityGroups`: NSG 목록 조회
12. `getAzureSecurityGroup`: NSG 상세 조회
13. `createAzureSecurityGroup`: NSG 생성
14. `updateAzureSecurityGroup`: NSG 업데이트
15. `deleteAzureSecurityGroup`: NSG 삭제
16. `addAzureSecurityGroupRule`: NSG 규칙 추가
17. `removeAzureSecurityGroupRule`: NSG 규칙 제거

**참고 구현**:
- AWS: `listAWSVPCs`, `createAWSVPC`, `listAWSSecurityGroups`, `createAWSSecurityGroup`
- GCP: `listGCPVPCs`, `listGCPSecurityGroups`, `createGCPSecurityGroup`

### 5. Handler 구현

**파일**: `internal/application/handlers/kubernetes/providers/azure_handler.go`
**파일**: `internal/application/handlers/network/providers/azure_handler.go`

**구현 방식**:
- AWS/GCP Handler와 동일한 패턴 사용
- `BaseHandler`의 `Compose`, `StandardCRUDDecorators` 활용
- 에러 처리, 로깅, 감사로그는 `BaseHandler` 메서드 활용

### 6. DTO 확장

**파일**: `internal/application/services/kubernetes/dto.go`
**파일**: `internal/application/services/network/dto.go`

**Azure 전용 DTO 추가**:
- `CreateAKSClusterRequest`: AKS 클러스터 생성 요청
- `AKSNodePoolConfig`: AKS 노드 풀 설정
- `AKSNetworkConfig`: AKS 네트워크 설정
- `AKSSecurityConfig`: AKS 보안 설정

**기존 DTO 확장**:
- `CreateVPCRequest`에 `ResourceGroup` 필드 추가
- `CreateSubnetRequest`에 `ResourceGroup` 필드 추가

---

## 📐 일관성 및 최적화 고려사항

### 1. 아키텍처 일관성
- ✅ **Service Layer**: AWS/GCP와 동일한 패턴 유지
- ✅ **Handler Layer**: `BaseHandler` 활용, Decorator 패턴 사용
- ✅ **에러 처리**: `domain.NewDomainError` 사용, 일관된 에러 코드
- ✅ **로깅**: `zap.Logger` 사용, 구조화된 로깅
- ✅ **캐싱**: `cache.Cache` 활용, TTL 설정
- ✅ **이벤트 발행**: `messaging.Publisher` 사용, NATS 이벤트 발행
- ✅ **감사로그**: `common.LogAction` 사용

### 2. 코드 구조 일관성
- ✅ **함수 네이밍**: `createAzureAKSCluster`, `listAzureVPCs` 등 일관된 네이밍
- ✅ **에러 처리**: `handleAzureError` 헬퍼 함수로 통일
- ✅ **자격증명 추출**: `extractAzureCredentials` 헬퍼 함수로 통일
- ✅ **클라이언트 생성**: `createAzureContainerServiceClient`, `createAzureNetworkClient` 등

### 3. 최적화 고려사항
- ✅ **캐싱 전략**: AWS/GCP와 동일한 캐시 키 구조 사용
- ✅ **병렬 처리**: 여러 리소스 조회 시 goroutine 활용 (필요시)
- ✅ **에러 재시도**: Azure SDK의 재시도 정책 활용
- ✅ **리소스 정리**: Context 취소 시 리소스 정리

### 4. Azure 특수 고려사항
- ⚠️ **리소스 그룹**: Azure는 리소스 그룹이 필수이므로 모든 요청에 포함
- ⚠️ **Subscription ID**: 자격증명에 포함되어 있지만 명시적으로 전달 필요
- ⚠️ **Location/Region**: Azure는 "location" 용어 사용, "region"과 매핑 필요
- ⚠️ **비동기 작업**: Azure는 대부분 비동기 작업이므로 Operation ID 추적 필요

---

## 🎯 구현 우선순위

### Phase 1: 기본 인프라 (필수)
1. Azure SDK 의존성 추가
2. Azure 인증 헬퍼 구현 (`extractAzureCredentials`, `createAzureClient`)
3. Azure 에러 처리 헬퍼 구현 (`handleAzureError`)

### Phase 2: AKS 클러스터 기본 기능 (핵심)
1. AKS 클러스터 생성 (`createAzureAKSCluster`)
2. AKS 클러스터 목록 조회 (`listAzureAKSClusters`)
3. AKS 클러스터 상세 조회 (`getAzureAKSCluster`)
4. AKS Handler 기본 메서드 구현

### Phase 3: AKS 노드 풀 관리
1. 노드 풀 생성 (`createAzureNodePool`)
2. 노드 풀 목록 조회 (`listAzureNodePools`)
3. 노드 풀 상세 조회 (`getAzureNodePool`)
4. 노드 풀 삭제 (`deleteAzureNodePool`)
5. 노드 풀 스케일링 (`scaleAzureNodePool`)

### Phase 4: 네트워크 기본 기능
1. Virtual Network 목록 조회 (`listAzureVPCs`)
2. Virtual Network 생성 (`createAzureVPC`)
3. Virtual Network 상세 조회 (`getAzureVPC`)
4. 서브넷 목록 조회 (`listAzureSubnets`)
5. 서브넷 생성 (`createAzureSubnet`)

### Phase 5: 네트워크 고급 기능
1. Network Security Group 관리 (생성, 조회, 삭제)
2. NSG 규칙 관리 (추가, 제거, 업데이트)
3. 네트워크 Handler 구현

### Phase 6: 완성도 향상
1. kubeconfig 생성 (`getAzureAKSKubeconfig`)
2. 클러스터 삭제 (`deleteAzureAKSCluster`)
3. Virtual Network 업데이트/삭제
4. 서브넷 업데이트/삭제
5. 에러 처리 개선
6. 캐싱 최적화

---

## 📝 참고 자료

### Azure SDK 문서
- [Azure Container Service SDK](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice)
- [Azure Network SDK](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork)
- [Azure Identity SDK](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity)

### Azure REST API 문서
- [AKS Management API](https://learn.microsoft.com/en-us/rest/api/aks/)
- [Azure Virtual Network API](https://learn.microsoft.com/en-us/rest/api/virtualnetwork/)
- [Network Security Group API](https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-security-groups)

### 기존 구현 참고
- AWS EKS: `internal/application/services/kubernetes/service.go` (line 387-807)
- GCP GKE: `internal/application/services/kubernetes/service.go` (line 101-384)
- AWS Network: `internal/application/services/network/service.go` (line 112-1000)
- GCP Network: `internal/application/services/network/service.go` (line 446-900)

---

## ⚠️ 주의사항

1. **자격증명 구조**: Azure는 `subscription_id`, `client_id`, `client_secret`, `tenant_id` 필요
2. **리소스 그룹**: 모든 Azure 리소스는 리소스 그룹에 속해야 함
3. **비동기 작업**: Azure API는 대부분 비동기이므로 Operation ID 추적 필요
4. **에러 코드**: Azure SDK 에러를 도메인 에러로 적절히 변환
5. **캐싱**: Azure 리소스 조회 결과는 적절한 TTL로 캐싱
6. **이벤트 발행**: 리소스 생성/수정/삭제 시 NATS 이벤트 발행
7. **감사로그**: 모든 작업에 대해 감사로그 기록

---

## ✅ 체크리스트

### 개발 전
- [ ] Azure SDK 의존성 추가
- [ ] Azure 자격증명 테스트 계정 준비
- [ ] Azure 리소스 그룹 생성

### 개발 중
- [ ] Azure 인증 헬퍼 구현 및 테스트
- [ ] AKS 클러스터 기본 기능 구현
- [ ] AKS 노드 풀 관리 구현
- [ ] 네트워크 기본 기능 구현
- [ ] 네트워크 고급 기능 구현
- [ ] Handler 구현
- [ ] 에러 처리 및 로깅
- [ ] 캐싱 적용
- [ ] 이벤트 발행
- [ ] 감사로그 기록

### 개발 후
- [ ] 단위 테스트 작성
- [ ] 통합 테스트 작성
- [ ] 문서화
- [ ] Frontend 연동 확인

---

**작성일**: 2025-01-XX
**작성자**: AI Assistant
**버전**: 1.0


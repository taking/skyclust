# GCP GKE 클러스터 생성 완전 가이드

## 🚀 실행 순서 (필수)

### 방법 1: 자동 서브넷 모드 (권장)

#### 1단계: VPC 생성
```bash
POST /api/v1/gcp/network/vpcs
```
- **파일**: `VPC 생성.bru`
- **목적**: GKE 클러스터를 위한 네트워크 생성
- **결과**: `skyclust-vpc` 생성
- **설정**: `auto_create_subnets: true` (GCP SDK 제한)

#### 2단계: 서브넷 생성 (선택사항)
```bash
POST /api/v1/gcp/network/subnets
```
- **파일**: `서브넷 생성.bru`
- **목적**: GKE 노드가 사용할 커스텀 서브넷 생성
- **결과**: `skyclust-subnet` 생성
- **고급 기능**: `private_ip_google_access`, `flow_logs`
- **참고**: 자동 생성된 서브넷 사용 시 이 단계 생략 가능

#### 3단계: 방화벽 규칙 생성
```bash
POST /api/v1/gcp/network/firewall-rules
```
- **파일**: `방화벽 규칙 생성 (예제).bru`
- **목적**: GKE 노드와의 통신을 위한 보안 규칙
- **결과**: `skyclust-firewall` 생성
- **고급 기능**: `priority`, `direction`, `allowed`, `denied`

#### 4단계: GKE 클러스터 생성
```bash
POST /api/v1/gcp/kubernetes/clusters
```
- **파일**: `클러스터 생성 (완전 예제).bru`
- **목적**: GKE 클러스터 생성
- **결과**: `skyclust-gke-cluster` 생성

### 방법 2: 수동 서브넷 모드 (고급 설정)

#### ⚠️ GCP SDK 제한사항
- **현재 지원 불가**: `auto_create_subnets: false`는 GCP SDK에서 지원하지 않음
- **대안**: `auto_create_subnets: true`로 VPC 생성 후 불필요한 서브넷 삭제
- **권장**: 자동 서브넷 모드 사용 (방법 1)

#### 1단계: VPC 생성 (자동 서브넷)
```bash
POST /api/v1/gcp/network/vpcs
```
- **파일**: `VPC 생성 - Auto Subnets.bru`
- **목적**: 모든 리전에 자동으로 서브넷 생성
- **결과**: `skyclust-vpc-auto` + 모든 리전 서브넷 자동 생성
- **설정**: `auto_create_subnets: true`

#### 2단계: 방화벽 규칙 생성
```bash
POST /api/v1/gcp/network/firewall-rules
```
- **파일**: `방화벽 규칙 생성 (예제).bru`
- **목적**: GKE 노드와의 통신을 위한 보안 규칙
- **결과**: `skyclust-firewall` 생성

#### 3단계: GKE 클러스터 생성
```bash
POST /api/v1/gcp/kubernetes/clusters
```
- **파일**: `클러스터 생성 (완전 예제).bru`
- **목적**: GKE 클러스터 생성
- **결과**: `skyclust-gke-cluster` 생성

## 📋 사용된 더미 데이터

### 프로젝트 정보
- **Project ID**: `leafy-environs-445206-d2`
- **Region**: `asia-northeast3`
- **Zone**: `asia-northeast3-a`

### 네트워크 정보
- **VPC**: `skyclust-vpc` (10.0.0.0/16)
- **서브넷**: `skyclust-subnet` (10.0.0.0/24)
- **방화벽**: `skyclust-firewall`

### 클러스터 정보
- **이름**: `skyclust-gke-cluster`
- **버전**: `1.28`
- **노드 풀**: `default-pool`
- **머신 타입**: `e2-medium`

## ✅ 고급 기능 지원 완료!

### VPC 생성에서 지원되는 고급 기능
- ✅ `description` - VPC 설명 추가
- ✅ `auto_create_subnets` - 자동 서브넷 생성 옵션
- ✅ `routing_mode` - 라우팅 모드 설정 (REGIONAL/GLOBAL)
- ✅ `mtu` - MTU 크기 설정
- ✅ `region` - 선택사항 (VPC는 Global 리소스)

#### VPC는 Global 리소스
- **VPC 자체**: Global 리소스 (특정 리전에 속하지 않음)
- **`region` 필드**: 선택사항 (VPC 생성 시 불필요)
- **서브넷**: Regional 리소스 (각 리전에 생성)

#### `auto_create_subnets: true` 동작
- **자동 생성**: 모든 활성 리전에 서브넷 자동 생성
- **서브넷 이름**: 리전 이름 (예: `asia-northeast3`, `us-central1`)
- **IP 범위**: GCP가 자동 할당 (예: `10.0.0.0/20`, `10.1.0.0/20`)
- **설정**: GCP 기본값 사용 (사용자 정의 불가)
- **용도**: 빠른 프로토타이핑, 기본 네트워킹

### 서브넷 생성에서 지원되는 고급 기능
- ✅ `description` - 서브넷 설명 추가
- ✅ `private_ip_google_access` - Google API 접근 허용
- ✅ `flow_logs` - VPC Flow Logs 활성화

### 방화벽 생성에서 지원되는 고급 기능
- ✅ `priority` - 방화벽 규칙 우선순위
- ✅ `direction` - 트래픽 방향 (INGRESS/EGRESS)
- ✅ `action` - 동작 (ALLOW/DENY)
- ✅ `source_ranges` - 소스 IP 범위
- ✅ `target_tags` - 대상 태그
- ✅ `allowed` - 허용 규칙
- ✅ `denied` - 거부 규칙

## 🔧 환경 변수 설정

Bruno에서 다음 환경 변수를 설정하세요:

```javascript
{
  "baseUrl": "http://localhost:8080",
  "apiVersion": "v1",
  "token": "your-jwt-token",
  "credentialsId": "edb33f37-4f8b-4307-a9d4-3f147ab09a2f",
  "gcp_project_id": "leafy-environs-445206-d2",
  "gcpZone": "asia-northeast3-a"
}
```

## 🔄 Auto Create Subnets vs Manual Subnets

### `auto_create_subnets: true` (자동 모드)
```json
{
  "auto_create_subnets": true,
  "routing_mode": "REGIONAL",
  "mtu": 1460
}
```

**장점:**
- ✅ **빠른 설정**: 모든 리전에 자동으로 서브넷 생성
- ✅ **간편함**: 별도 서브넷 생성 작업 불필요
- ✅ **일관성**: 모든 리전에 동일한 네트워크 구조

**단점:**
- ❌ **제어 불가**: 서브넷 이름, IP 범위, 설정 변경 불가
- ❌ **비용**: 불필요한 리전의 서브넷도 생성
- ❌ **보안**: 기본 설정만 사용 가능

### `auto_create_subnets: false` (수동 모드) - ⚠️ 현재 지원 불가
```json
{
  "auto_create_subnets": false,
  "routing_mode": "REGIONAL",
  "mtu": 1460
}
```

**⚠️ GCP SDK 제한사항:**
- ❌ **현재 지원 불가**: GCP SDK에서 `auto_create_subnets: false` 지원하지 않음
- ❌ **Legacy Mode 금지**: GCP가 Legacy Mode 네트워크 생성 완전 금지
- ❌ **API 레벨 차단**: `auto_create_subnets: false`는 API에서 거부됨

**대안 방법:**
- ✅ **자동 서브넷 생성**: `auto_create_subnets: true`로 VPC 생성
- ✅ **불필요한 서브넷 삭제**: 생성 후 불필요한 서브넷 삭제
- ✅ **커스텀 서브넷 생성**: 필요한 서브넷을 수동으로 생성

## 📝 실제 Request Body 예시

### 1단계: VPC 생성 (자동 서브넷 모드) - 권장
```json
{
  "credential_id": "edb33f37-4f8b-4307-a9d4-3f147ab09a2f",
  "name": "skyclust-vpc",
  "description": "VPC for SkyClust GKE cluster with auto-created subnets",
  "project_id": "leafy-environs-445206-d2",
  "auto_create_subnets": true,
  "routing_mode": "REGIONAL",
  "mtu": 1460,
  "tags": {
    "Environment": "development",
    "Project": "skyclust",
    "CreatedBy": "bruno-api-test"
  }
}
```

### 1단계: VPC 생성 (자동 서브넷 모드) - 대안
```json
{
  "credential_id": "edb33f37-4f8b-4307-a9d4-3f147ab09a2f",
  "name": "skyclust-vpc-auto",
  "description": "VPC with auto-created subnets for each region",
  "project_id": "leafy-environs-445206-d2",
  "auto_create_subnets": true,
  "routing_mode": "REGIONAL",
  "mtu": 1460,
  "tags": {
    "Environment": "development",
    "Project": "skyclust",
    "CreatedBy": "bruno-api-test",
    "AutoSubnets": "true"
  }
}
```

### 2단계: 서브넷 생성 (고급 기능 포함) - 선택사항
```json
{
  "credential_id": "edb33f37-4f8b-4307-a9d4-3f147ab09a2f",
  "name": "skyclust-subnet",
  "description": "Custom subnet for SkyClust GKE cluster with advanced features",
  "project_id": "leafy-environs-445206-d2",
  "region": "asia-northeast3",
  "vpc_id": "projects/leafy-environs-445206-d2/global/networks/skyclust-vpc",
  "cidr_block": "10.0.0.0/24",
  "zone": "asia-northeast3-a",
  "private_ip_google_access": true,
  "flow_logs": true,
  "tags": {
    "Environment": "development",
    "Project": "skyclust",
    "CreatedBy": "bruno-api-test"
  }
}
```

### 3단계: 방화벽 규칙 생성 (고급 기능 포함)
```json
{
  "credential_id": "edb33f37-4f8b-4307-a9d4-3f147ab09a2f",
  "name": "skyclust-firewall",
  "description": "Firewall rules for SkyClust GKE cluster with advanced features",
  "project_id": "leafy-environs-445206-d2",
  "vpc_id": "projects/leafy-environs-445206-d2/global/networks/skyclust-vpc",
  "region": "asia-northeast3",
  "priority": 1000,
  "direction": "INGRESS",
  "action": "ALLOW",
  "source_ranges": ["0.0.0.0/0"],
  "target_tags": ["gke-node"],
  "allowed": [
    {
      "protocol": "tcp",
      "ports": ["22", "80", "443", "8080", "30000-32767"]
    },
    {
      "protocol": "icmp"
    }
  ],
  "tags": {
    "Environment": "development",
    "Project": "skyclust",
    "CreatedBy": "bruno-api-test"
  }
}
```

## ⚠️ 주의사항

1. **순서 준수**: VPC → 서브넷 → 방화벽 → 클러스터 순서로 실행
2. **시간 소요**: 클러스터 생성은 5-10분 소요
3. **비용**: GKE 클러스터는 실행 중 비용이 발생
4. **리소스 정리**: 테스트 완료 후 리소스 삭제 권장
5. **구현 제한**: 현재 기본 CRUD만 지원, 고급 GCP 기능은 미구현

## 🧹 정리 순서 (역순)

1. **GKE 클러스터 삭제**
2. **방화벽 규칙 삭제**
3. **서브넷 삭제**
4. **VPC 삭제**

## 📊 예상 비용

- **GKE 클러스터**: 월 $73 (기본)
- **노드 풀**: 월 $24 (e2-medium 1개)
- **네트워크**: 월 $5 (VPC, 서브넷)
- **총 예상 비용**: 월 $102

## 🔍 문제 해결

### 일반적인 오류
1. **VPC 없음**: VPC를 먼저 생성하세요
2. **서브넷 없음**: 서브넷을 먼저 생성하세요
3. **권한 부족**: GCP 서비스 계정 권한 확인
4. **할당량 초과**: GCP 할당량 확인
5. **지원되지 않는 필드**: README의 구현 제한사항 확인

### 로그 확인
```bash
# 서버 로그 확인
tail -f server.log

# GCP 콘솔에서 리소스 상태 확인
https://console.cloud.google.com/
```

## 🎉 고급 기능 구현 완료!

### VPC 고급 기능 구현 완료
- ✅ `description` 필드 지원
- ✅ `auto_create_subnets` 옵션
- ✅ `routing_mode` 설정 (REGIONAL/GLOBAL)
- ✅ `mtu` 설정

### 서브넷 고급 기능 구현 완료
- ✅ `description` 필드 지원
- ✅ `private_ip_google_access` 옵션
- ✅ `flow_logs` 설정

### 방화벽 고급 기능 구현 완료
- ✅ `priority` 설정
- ✅ `direction` 설정 (INGRESS/EGRESS)
- ✅ `action` 설정 (ALLOW/DENY)
- ✅ `source_ranges` 설정
- ✅ `target_tags` 설정
- ✅ `allowed` 규칙 설정
- ✅ `denied` 규칙 설정

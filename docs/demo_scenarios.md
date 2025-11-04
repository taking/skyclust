# SkyClust 데모 시나리오 및 실행 가이드

## 🎯 **데모 목표**
멀티 클라우드 Kubernetes 관리 플랫폼의 핵심 기능을 실제로 시연하여 비즈니스 가치를 보여줍니다.

---

## 🚀 **데모 시나리오 1: 멀티 클라우드 클러스터 생성**

### **시나리오 개요**
AWS EKS와 GCP GKE 클러스터를 동시에 생성하여 멀티 클라우드 환경 구축

### **실행 단계**

#### **1단계: 자격증명 등록 (워크스페이스 기반)**
```bash
# AWS 자격증명 등록
curl -X POST http://localhost:8080/api/v1/credentials \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "workspace_id": "workspace-uuid",
    "name": "AWS Production",
    "provider": "aws",
    "data": {
      "access_key": "AKIA...",
      "secret_key": "secret...",
      "region": "us-west-2"
    }
  }'

# GCP 자격증명 등록
curl -X POST http://localhost:8080/api/v1/credentials \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "workspace_id": "workspace-uuid",
    "name": "GCP Production",
    "provider": "gcp",
    "data": {
      "type": "service_account",
      "project_id": "my-project",
      "private_key_id": "key-id",
      "private_key": "-----BEGIN PRIVATE KEY-----\n...",
      "client_email": "service@my-project.iam.gserviceaccount.com",
      "client_id": "123456789",
      "auth_uri": "https://accounts.google.com/o/oauth2/auth",
      "token_uri": "https://oauth2.googleapis.com/token",
      "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
      "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/service%40my-project.iam.gserviceaccount.com",
      "universe_domain": "googleapis.com"
    }
  }'
```

#### **2단계: 네트워크 인프라 생성**
```bash
# AWS VPC 생성
curl -X POST http://localhost:8080/api/v1/aws/network/vpcs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "credential_id": "aws-credential-id",
    "name": "skyclust-aws-vpc",
    "cidr_block": "10.0.0.0/16",
    "region": "us-west-2",
    "tags": {
      "Environment": "production",
      "Project": "skyclust"
    }
  }'

# GCP VPC 생성
curl -X POST http://localhost:8080/api/v1/gcp/network/vpcs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "credential_id": "gcp-credential-id",
    "name": "skyclust-gcp-vpc",
    "project_id": "my-project",
    "auto_create_subnets": true,
    "routing_mode": "REGIONAL",
    "mtu": 1460,
    "tags": {
      "Environment": "production",
      "Project": "skyclust"
    }
  }'
```

#### **3단계: Kubernetes 클러스터 생성**
```bash
# AWS EKS 클러스터 생성
curl -X POST http://localhost:8080/api/v1/aws/kubernetes/clusters \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "credential_id": "aws-credential-id",
    "name": "skyclust-eks-demo",
    "version": "1.28",
    "region": "us-west-2",
    "vpc_id": "vpc-12345",
    "subnet_ids": ["subnet-12345", "subnet-67890"],
    "tags": {
      "Environment": "production",
      "Project": "skyclust"
    },
    "aws_config": {
      "cluster_type": "standard",
      "node_group": {
        "name": "demo-node-group",
        "instance_types": ["t3.medium"],
        "disk_size": 20,
        "node_count": 3,
        "auto_scaling": {
          "enabled": true,
          "min_size": 1,
          "max_size": 5
        }
      }
    }
  }'

# GCP GKE 클러스터 생성
curl -X POST http://localhost:8080/api/v1/gcp/kubernetes/clusters \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "credential_id": "gcp-credential-id",
    "name": "skyclust-gke-demo",
    "version": "1.28",
    "region": "asia-northeast3",
    "zone": "asia-northeast3-a",
    "project_id": "my-project",
    "vpc_id": "projects/my-project/global/networks/skyclust-gcp-vpc",
    "subnet_ids": ["projects/my-project/regions/asia-northeast3/subnetworks/skyclust-gcp-vpc"],
    "tags": {
      "Environment": "production",
      "Project": "skyclust"
    },
    "gcp_config": {
      "cluster_type": "standard",
      "network_config": {
        "private_nodes": false,
        "private_endpoint": false,
        "master_authorized_networks": ["0.0.0.0/0"],
        "pod_cidr": "10.0.0.0/16",
        "service_cidr": "10.1.0.0/16"
      },
      "node_pool": {
        "name": "demo-pool",
        "machine_type": "e2-medium",
        "disk_size_gb": 20,
        "disk_type": "pd-standard",
        "node_count": 3,
        "auto_scaling": {
          "enabled": true,
          "min_node_count": 1,
          "max_node_count": 5
        }
      }
    }
  }'
```

#### **4단계: 클러스터 상태 확인**
```bash
# AWS EKS 클러스터 목록 조회
curl -X GET "http://localhost:8080/api/v1/aws/kubernetes/clusters?credential_id=aws-credential-id&region=us-west-2" \
  -H "Authorization: Bearer $JWT_TOKEN"

# GCP GKE 클러스터 목록 조회
curl -X GET "http://localhost:8080/api/v1/gcp/kubernetes/clusters?credential_id=gcp-credential-id&region=asia-northeast3" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

#### **5단계: Kubeconfig 생성**
```bash
# AWS EKS Kubeconfig 생성
curl -X GET "http://localhost:8080/api/v1/aws/kubernetes/clusters/skyclust-eks-demo/kubeconfig?credential_id=aws-credential-id&region=us-west-2" \
  -H "Authorization: Bearer $JWT_TOKEN"

# GCP GKE Kubeconfig 생성
curl -X GET "http://localhost:8080/api/v1/gcp/kubernetes/clusters/skyclust-gke-demo/kubeconfig?credential_id=gcp-credential-id&region=asia-northeast3" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

---

## 🌐 **데모 시나리오 2: 통합 네트워크 관리**

### **시나리오 개요**
멀티 클라우드 환경에서 네트워크 리소스를 통합 관리하는 기능 시연

### **실행 단계**

#### **1단계: 네트워크 리소스 목록 조회**
```bash
# AWS VPC 목록 조회
curl -X GET "http://localhost:8080/api/v1/aws/network/vpcs?credential_id=aws-credential-id&region=us-west-2" \
  -H "Authorization: Bearer $JWT_TOKEN"

# GCP VPC 목록 조회
curl -X GET "http://localhost:8080/api/v1/gcp/network/vpcs?credential_id=gcp-credential-id" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

#### **2단계: 서브넷 생성**
```bash
# AWS 서브넷 생성
curl -X POST http://localhost:8080/api/v1/aws/network/subnets \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "credential_id": "aws-credential-id",
    "vpc_id": "vpc-12345",
    "name": "skyclust-subnet-1",
    "cidr_block": "10.0.1.0/24",
    "availability_zone": "us-west-2a",
    "region": "us-west-2",
    "tags": {
      "Environment": "production",
      "Tier": "web"
    }
  }'

# GCP 서브넷 생성
curl -X POST http://localhost:8080/api/v1/gcp/network/subnets \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "credential_id": "gcp-credential-id",
    "vpc_id": "projects/my-project/global/networks/skyclust-gcp-vpc",
    "name": "skyclust-subnet-1",
    "cidr_block": "10.0.1.0/24",
    "region": "asia-northeast3",
    "private_ip_google_access": true,
    "flow_logs": false,
    "description": "Production subnet for web tier"
  }'
```

#### **3단계: 보안 그룹/방화벽 규칙 생성**
```bash
# AWS 보안 그룹 생성
curl -X POST http://localhost:8080/api/v1/aws/network/security-groups \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "credential_id": "aws-credential-id",
    "name": "skyclust-web-sg",
    "description": "Security group for web servers",
    "vpc_id": "vpc-12345",
    "region": "us-west-2",
    "rules": [
      {
        "type": "ingress",
        "protocol": "tcp",
        "from_port": 80,
        "to_port": 80,
        "cidr_blocks": ["0.0.0.0/0"],
        "description": "HTTP access"
      },
      {
        "type": "ingress",
        "protocol": "tcp",
        "from_port": 443,
        "to_port": 443,
        "cidr_blocks": ["0.0.0.0/0"],
        "description": "HTTPS access"
      }
    ],
    "tags": {
      "Environment": "production",
      "Tier": "web"
    }
  }'

# GCP 방화벽 규칙 생성
curl -X POST http://localhost:8080/api/v1/gcp/network/firewall-rules \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "credential_id": "gcp-credential-id",
    "name": "skyclust-web-firewall",
    "description": "Allow HTTP and HTTPS traffic",
    "vpc_id": "projects/my-project/global/networks/skyclust-gcp-vpc",
    "region": "asia-northeast3",
    "project_id": "my-project",
    "direction": "INGRESS",
    "priority": 1000,
    "action": "ALLOW",
    "protocol": "tcp",
    "ports": ["80", "443"],
    "source_ranges": ["0.0.0.0/0"],
    "target_tags": ["web-server"],
    "tags": {
      "Environment": "production",
      "Tier": "web"
    }
  }'
```

---

## 📊 **데모 시나리오 3: 실시간 모니터링**

### **시나리오 개요**
생성된 클러스터와 네트워크 리소스의 실시간 모니터링 기능 시연

### **실행 단계**

#### **1단계: 클러스터 상세 정보 조회**
```bash
# AWS EKS 클러스터 상세 조회
curl -X GET "http://localhost:8080/api/v1/aws/kubernetes/clusters/skyclust-eks-demo?credential_id=aws-credential-id&region=us-west-2" \
  -H "Authorization: Bearer $JWT_TOKEN"

# GCP GKE 클러스터 상세 조회
curl -X GET "http://localhost:8080/api/v1/gcp/kubernetes/clusters/skyclust-gke-demo?credential_id=gcp-credential-id&region=asia-northeast3" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

#### **2단계: 네트워크 리소스 상세 조회**
```bash
# AWS VPC 상세 조회
curl -X GET "http://localhost:8080/api/v1/aws/network/vpcs/skyclust-aws-vpc?credential_id=aws-credential-id&region=us-west-2" \
  -H "Authorization: Bearer $JWT_TOKEN"

# GCP VPC 상세 조회
curl -X GET "http://localhost:8080/api/v1/gcp/network/vpcs/skyclust-gcp-vpc?credential_id=gcp-credential-id" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

#### **3단계: 서브넷 목록 조회**
```bash
# AWS 서브넷 목록 조회
curl -X GET "http://localhost:8080/api/v1/aws/network/subnets?credential_id=aws-credential-id&vpc_id=vpc-12345&region=us-west-2" \
  -H "Authorization: Bearer $JWT_TOKEN"

# GCP 서브넷 목록 조회
curl -X GET "http://localhost:8080/api/v1/gcp/network/subnets?credential_id=gcp-credential-id&vpc_id=projects/my-project/global/networks/skyclust-gcp-vpc&region=asia-northeast3" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

---

## 🧹 **데모 시나리오 4: 리소스 정리**

### **시나리오 개요**
생성된 모든 리소스를 안전하게 삭제하는 기능 시연

### **실행 단계**

#### **1단계: Kubernetes 클러스터 삭제**
```bash
# AWS EKS 클러스터 삭제
curl -X DELETE "http://localhost:8080/api/v1/aws/kubernetes/clusters/skyclust-eks-demo?credential_id=aws-credential-id&region=us-west-2" \
  -H "Authorization: Bearer $JWT_TOKEN"

# GCP GKE 클러스터 삭제
curl -X DELETE "http://localhost:8080/api/v1/gcp/kubernetes/clusters/skyclust-gke-demo?credential_id=gcp-credential-id&region=asia-northeast3" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

#### **2단계: 네트워크 리소스 삭제**
```bash
# AWS 서브넷 삭제
curl -X DELETE "http://localhost:8080/api/v1/aws/network/subnets/skyclust-subnet-1?credential_id=aws-credential-id&region=us-west-2" \
  -H "Authorization: Bearer $JWT_TOKEN"

# AWS 보안 그룹 삭제
curl -X DELETE "http://localhost:8080/api/v1/aws/network/security-groups/skyclust-web-sg?credential_id=aws-credential-id&region=us-west-2" \
  -H "Authorization: Bearer $JWT_TOKEN"

# AWS VPC 삭제
curl -X DELETE "http://localhost:8080/api/v1/aws/network/vpcs/skyclust-aws-vpc?credential_id=aws-credential-id&region=us-west-2" \
  -H "Authorization: Bearer $JWT_TOKEN"

# GCP 서브넷 삭제
curl -X DELETE "http://localhost:8080/api/v1/gcp/network/subnets/skyclust-subnet-1?credential_id=gcp-credential-id&region=asia-northeast3" \
  -H "Authorization: Bearer $JWT_TOKEN"

# GCP 방화벽 규칙 삭제
curl -X DELETE "http://localhost:8080/api/v1/gcp/network/firewall-rules/skyclust-web-firewall?credential_id=gcp-credential-id" \
  -H "Authorization: Bearer $JWT_TOKEN"

# GCP VPC 삭제
curl -X DELETE "http://localhost:8080/api/v1/gcp/network/vpcs/skyclust-gcp-vpc?credential_id=gcp-credential-id" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

---

## 🎯 **데모 핵심 포인트**

### **1. 통합 관리의 편의성**
- **단일 인터페이스**: AWS와 GCP를 동일한 API로 관리
- **일관된 응답**: 클라우드별 차이점을 추상화한 통합 응답
- **자동 변환**: 클라우드별 특성을 자동으로 처리

### **2. 자동화의 강력함**
- **의존성 관리**: VPC 삭제 시 관련 리소스 자동 정리
- **순서 보장**: 리소스 생성/삭제 순서 자동 관리
- **오류 처리**: 클라우드별 오류를 통합된 형태로 처리

### **3. 보안의 신뢰성**
- **암호화된 자격증명**: 민감한 정보 보호
- **RBAC**: 세밀한 권한 제어
- **감사 로깅**: 모든 작업 추적

### **4. 확장성의 유연성**
- **새로운 클라우드 추가**: Azure, NCP 등 쉽게 확장 가능
- **마이크로서비스**: 서비스별 독립적 확장
- **이벤트 기반**: 느슨한 결합으로 유연한 아키텍처

---

## 📈 **예상 데모 결과**

### **성능 지표**
- **클러스터 생성 시간**: AWS EKS ~3분, GCP GKE ~2분
- **API 응답 시간**: 평균 150ms 이하
- **동시 처리**: 100개 이상의 동시 요청 처리 가능

### **비즈니스 가치**
- **운영 효율성**: 수동 작업 90% 감소
- **비용 절감**: AWS Cost Explorer, GCP Cloud Billing API 통합으로 정확한 비용 분석 및 20% 비용 절약
- **개발 생산성**: 인프라 관리 시간 80% 단축
- **워크스페이스 격리**: 멀티 테넌트 환경에서 완전한 자원 격리

---

## 🎉 **데모 마무리**

### **핵심 메시지**
1. **멀티 클라우드 복잡성 해결**: 여러 클라우드를 하나의 플랫폼에서 관리
2. **자동화의 가치**: 수동 작업을 최소화하고 일관된 배포 프로세스 제공
3. **확장 가능한 아키텍처**: 새로운 클라우드 제공업체 쉽게 추가
4. **엔터프라이즈급 보안**: 암호화, RBAC, 감사 로깅으로 보안 강화

### **다음 단계**
- **Pilot 프로젝트**: 실제 환경에서 테스트
- **사용자 교육**: 팀원 대상 교육 프로그램
- **점진적 도입**: 기존 시스템과의 통합 계획

이를 통해 **클라우드 복잡성에서 해방**되어 **애플리케이션 개발에 집중**할 수 있는 환경을 제공합니다.

# Kubernetes 서비스 관련 API 목록

## 📋 **API 개요**
- **총 API 개수**: 50개
- **지원 클라우드**: AWS EKS, GCP GKE
- **기본 URL**: `/api/v1/{provider}/kubernetes`

---

## 🔌 **AWS EKS API (25개)**

### **클러스터 관리 (5개)**
| Method | URL | 설명 |
|--------|-----|------|
| `POST` | `/api/v1/aws/kubernetes/clusters` | AWS EKS 클러스터 생성 |
| `GET` | `/api/v1/aws/kubernetes/clusters` | AWS EKS 클러스터 목록 조회 |
| `GET` | `/api/v1/aws/kubernetes/clusters/:name` | AWS EKS 클러스터 상세 조회 |
| `DELETE` | `/api/v1/aws/kubernetes/clusters/:name` | AWS EKS 클러스터 삭제 |
| `GET` | `/api/v1/aws/kubernetes/clusters/:name/kubeconfig` | AWS EKS Kubeconfig 생성 |

### **노드풀 관리 (5개)**
| Method | URL | 설명 |
|--------|-----|------|
| `POST` | `/api/v1/aws/kubernetes/clusters/:name/nodepools` | 노드풀 생성 |
| `GET` | `/api/v1/aws/kubernetes/clusters/:name/nodepools` | 노드풀 목록 조회 |
| `GET` | `/api/v1/aws/kubernetes/clusters/:name/nodepools/:nodepool` | 노드풀 상세 조회 |
| `DELETE` | `/api/v1/aws/kubernetes/clusters/:name/nodepools/:nodepool` | 노드풀 삭제 |
| `PUT` | `/api/v1/aws/kubernetes/clusters/:name/nodepools/:nodepool/scale` | 노드풀 스케일링 |

### **노드그룹 관리 (4개)**
| Method | URL | 설명 |
|--------|-----|------|
| `POST` | `/api/v1/aws/kubernetes/clusters/:name/nodegroups` | 노드그룹 생성 |
| `GET` | `/api/v1/aws/kubernetes/clusters/:name/nodegroups` | 노드그룹 목록 조회 |
| `GET` | `/api/v1/aws/kubernetes/clusters/:name/nodegroups/:nodegroup` | 노드그룹 상세 조회 |
| `DELETE` | `/api/v1/aws/kubernetes/clusters/:name/nodegroups/:nodegroup` | 노드그룹 삭제 |

### **클러스터 운영 (2개)**
| Method | URL | 설명 |
|--------|-----|------|
| `POST` | `/api/v1/aws/kubernetes/clusters/:name/upgrade` | 클러스터 업그레이드 |
| `GET` | `/api/v1/aws/kubernetes/clusters/:name/upgrade/status` | 업그레이드 상태 조회 |

### **노드 관리 (9개)**
| Method | URL | 설명 |
|--------|-----|------|
| `GET` | `/api/v1/aws/kubernetes/clusters/:name/nodes` | 노드 목록 조회 |
| `GET` | `/api/v1/aws/kubernetes/clusters/:name/nodes/:node` | 노드 상세 조회 |
| `POST` | `/api/v1/aws/kubernetes/clusters/:name/nodes/:node/drain` | 노드 드레인 |
| `POST` | `/api/v1/aws/kubernetes/clusters/:name/nodes/:node/cordon` | 노드 코돈 |
| `POST` | `/api/v1/aws/kubernetes/clusters/:name/nodes/:node/uncordon` | 노드 언코돈 |
| `GET` | `/api/v1/aws/kubernetes/clusters/:name/nodes/:node/logs` | 노드 로그 조회 |
| `GET` | `/api/v1/aws/kubernetes/clusters/:name/nodes/:node/metrics` | 노드 메트릭 조회 |
| `POST` | `/api/v1/aws/kubernetes/clusters/:name/nodes/:node/restart` | 노드 재시작 |
| `GET` | `/api/v1/aws/kubernetes/clusters/:name/nodes/:node/ssh` | SSH 접근 설정 |

---

## 🌐 **GCP GKE API (25개)**

### **클러스터 관리 (5개)**
| Method | URL | 설명 |
|--------|-----|------|
| `POST` | `/api/v1/gcp/kubernetes/clusters` | GCP GKE 클러스터 생성 |
| `GET` | `/api/v1/gcp/kubernetes/clusters` | GCP GKE 클러스터 목록 조회 |
| `GET` | `/api/v1/gcp/kubernetes/clusters/:name` | GCP GKE 클러스터 상세 조회 |
| `DELETE` | `/api/v1/gcp/kubernetes/clusters/:name` | GCP GKE 클러스터 삭제 |
| `GET` | `/api/v1/gcp/kubernetes/clusters/:name/kubeconfig` | GCP GKE Kubeconfig 생성 |

### **노드풀 관리 (5개)**
| Method | URL | 설명 |
|--------|-----|------|
| `POST` | `/api/v1/gcp/kubernetes/clusters/:name/nodepools` | GKE 노드풀 생성 |
| `GET` | `/api/v1/gcp/kubernetes/clusters/:name/nodepools` | GKE 노드풀 목록 조회 |
| `GET` | `/api/v1/gcp/kubernetes/clusters/:name/nodepools/:nodepool` | GKE 노드풀 상세 조회 |
| `DELETE` | `/api/v1/gcp/kubernetes/clusters/:name/nodepools/:nodepool` | GKE 노드풀 삭제 |
| `PUT` | `/api/v1/gcp/kubernetes/clusters/:name/nodepools/:nodepool/scale` | GKE 노드풀 스케일링 |

### **클러스터 운영 (2개)**
| Method | URL | 설명 |
|--------|-----|------|
| `POST` | `/api/v1/gcp/kubernetes/clusters/:name/upgrade` | GKE 클러스터 업그레이드 |
| `GET` | `/api/v1/gcp/kubernetes/clusters/:name/upgrade/status` | GKE 업그레이드 상태 조회 |

### **노드 관리 (8개)**
| Method | URL | 설명 |
|--------|-----|------|
| `GET` | `/api/v1/gcp/kubernetes/clusters/:name/nodes` | GKE 노드 목록 조회 |
| `GET` | `/api/v1/gcp/kubernetes/clusters/:name/nodes/:node` | GKE 노드 상세 조회 |
| `POST` | `/api/v1/gcp/kubernetes/clusters/:name/nodes/:node/drain` | GKE 노드 드레인 |
| `POST` | `/api/v1/gcp/kubernetes/clusters/:name/nodes/:node/cordon` | GKE 노드 코돈 |
| `POST` | `/api/v1/gcp/kubernetes/clusters/:name/nodes/:node/uncordon` | GKE 노드 언코돈 |
| `GET` | `/api/v1/gcp/kubernetes/clusters/:name/nodes/:node/logs` | GKE 노드 로그 조회 |
| `GET` | `/api/v1/gcp/kubernetes/clusters/:name/nodes/:node/metrics` | GKE 노드 메트릭 조회 |
| `GET` | `/api/v1/gcp/kubernetes/clusters/:name/nodes/:node/ssh` | GKE SSH 접근 설정 |

### **GKE 특화 기능 (5개)**
| Method | URL | 설명 |
|--------|-----|------|
| `POST` | `/api/v1/gcp/kubernetes/clusters/:name/workload-identity` | Workload Identity 설정 |
| `GET` | `/api/v1/gcp/kubernetes/clusters/:name/workload-identity` | Workload Identity 조회 |
| `POST` | `/api/v1/gcp/kubernetes/clusters/:name/binary-authorization` | Binary Authorization 설정 |
| `GET` | `/api/v1/gcp/kubernetes/clusters/:name/binary-authorization` | Binary Authorization 조회 |
| `GET` | `/api/v1/gcp/kubernetes/clusters/:name/network-policy` | Network Policy 조회 |

---

## 📊 **API 사용 예시**

### **AWS EKS 클러스터 생성**
```bash
POST /api/v1/aws/kubernetes/clusters
Content-Type: application/json
Authorization: Bearer <token>

{
  "credential_id": "aws-credential-uuid",
  "name": "my-eks-cluster",
  "version": "1.28",
  "region": "us-west-2",
  "subnet_ids": ["subnet-12345", "subnet-67890"],
  "role_arn": "arn:aws:iam::123456789012:role/EKSClusterRole",
  "tags": {
    "Environment": "production",
    "Project": "skyclust"
  }
}
```

### **GCP GKE 클러스터 생성**
```bash
POST /api/v1/gcp/kubernetes/clusters
Content-Type: application/json
Authorization: Bearer <token>

{
  "credential_id": "gcp-credential-uuid",
  "name": "my-gke-cluster",
  "version": "1.28",
  "region": "asia-northeast3",
  "zone": "asia-northeast3-a",
  "project_id": "my-project",
  "vpc_id": "projects/my-project/global/networks/default",
  "subnet_ids": ["projects/my-project/regions/asia-northeast3/subnetworks/default"],
  "gcp_config": {
    "cluster_type": "standard",
    "network_config": {
      "private_nodes": false,
      "private_endpoint": false,
      "master_authorized_networks": ["0.0.0.0/0"]
    },
    "node_pool": {
      "name": "default-pool",
      "machine_type": "e2-medium",
      "node_count": 3,
      "auto_scaling": {
        "enabled": true,
        "min_node_count": 1,
        "max_node_count": 5
      }
    }
  }
}
```

### **노드풀 스케일링**
```bash
PUT /api/v1/gcp/kubernetes/clusters/my-gke-cluster/nodepools/default-pool/scale
Content-Type: application/json
Authorization: Bearer <token>

{
  "credential_id": "gcp-credential-uuid",
  "region": "asia-northeast3",
  "node_count": 5
}
```

### **Kubeconfig 생성**
```bash
GET /api/v1/aws/kubernetes/clusters/my-eks-cluster/kubeconfig?credential_id=aws-credential-uuid&region=us-west-2
Authorization: Bearer <token>
```

---

## 🔧 **공통 파라미터**

### **Path Parameters**
- `{provider}`: 클라우드 제공업체 (`aws`, `gcp`, `azure`, `ncp`)
- `:name`: 클러스터 이름
- `:nodepool`: 노드풀 이름
- `:nodegroup`: 노드그룹 이름
- `:node`: 노드 이름

### **Query Parameters**
- `credential_id`: 자격증명 ID (필수)
- `region`: 리전 (필수)
- `zone`: 존 (GCP 선택사항)

### **공통 응답 형식**
```json
{
  "success": true,
  "data": { ... },
  "message": "Operation completed successfully",
  "request_id": "uuid",
  "timestamp": "2025-01-01T00:00:00Z"
}
```

---

## 🎯 **주요 특징**

### **멀티 클라우드 지원**
- AWS EKS와 GCP GKE를 동일한 API 인터페이스로 관리
- 클라우드별 특화 기능 지원 (노드그룹 vs 노드풀)
- Provider별 Dispatch 패턴으로 확장성 확보

### **완전한 생명주기 관리**
- 클러스터 생성부터 삭제까지 전체 생명주기 관리
- 노드풀/노드그룹 자동 스케일링
- 클러스터 업그레이드 및 상태 모니터링

### **운영 편의성**
- Kubeconfig 자동 생성 및 다운로드
- SSH 접근 설정 자동화
- 노드 드레인, 코돈 등 운영 작업 지원

이러한 API를 통해 **멀티 클라우드 Kubernetes 환경을 통합 관리**할 수 있습니다!

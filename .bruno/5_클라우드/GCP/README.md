# GCP API 테스트 가이드

이 폴더는 Google Cloud Platform (GCP) 관련 API 테스트를 위한 Bruno 파일들을 포함합니다.

## 폴더 구조

```
GCP/
├── Kubernetes (GKE)/
│   ├── 클러스터/
│   │   ├── 클러스터 생성.bru
│   │   ├── 클러스터 목록 조회.bru
│   │   ├── 클러스터 상세 조회.bru
│   │   ├── 클러스터 삭제.bru
│   │   └── Kubeconfig 다운로드.bru
│   └── 노드 풀/
│       ├── 노드 풀 생성.bru
│       ├── 노드 풀 목록 조회.bru
│       ├── 노드 풀 상세 조회.bru
│       ├── 노드 풀 삭제.bru
│       └── 노드 풀 스케일링.bru
└── Network/
    ├── VPC/
    │   ├── VPC 생성.bru
    │   ├── VPC 목록 조회.bru
    │   ├── VPC 상세 조회.bru
    │   ├── VPC 수정.bru
    │   └── VPC 삭제.bru
    ├── 서브넷/
    │   ├── 서브넷 생성.bru
    │   └── 서브넷 목록 조회.bru
    └── 방화벽 규칙/
        ├── 방화벽 규칙 생성.bru
        ├── 방화벽 규칙 목록 조회.bru
        ├── 방화벽 규칙 상세 조회.bru
        ├── 방화벽 규칙 수정.bru
        └── 방화벽 규칙 삭제.bru
```

## 환경 변수 설정

GCP API 테스트를 위해 다음 환경 변수를 설정해야 합니다:

### 필수 환경 변수

- `baseUrl`: API 서버 기본 URL (예: `http://localhost:8080`)
- `apiVersion`: API 버전 (예: `v1`)
- `token`: JWT 인증 토큰
- `credentialsId`: GCP credential ID
- `gcp_project_id`: GCP 프로젝트 ID
- `gcpRegion`: GCP 리전 (예: `us-central1`)
- `gcpZone`: GCP 존 (예: `us-central1-a`)

### 선택적 환경 변수

- `clusterName`: 클러스터 이름 (동적으로 설정됨)
- `nodePoolName`: 노드 풀 이름 (동적으로 설정됨)
- `vpcId`: VPC ID (동적으로 설정됨)
- `subnetId`: 서브넷 ID (동적으로 설정됨)
- `firewallId`: 방화벽 규칙 ID (동적으로 설정됨)

## 사용 방법

1. **환경 변수 설정**: Bruno에서 환경 변수를 설정합니다.
2. **Credential 등록**: GCP 서비스 계정 키를 등록합니다.
3. **API 테스트**: 각 Bruno 파일을 실행하여 API를 테스트합니다.

## 주의사항

- GCP API는 실제 리소스를 생성/삭제하므로 테스트 시 주의하세요.
- 테스트 후 생성된 리소스는 정리하세요.
- 비용이 발생할 수 있으므로 테스트 환경에서만 사용하세요.

## 지원 기능

### Kubernetes (GKE)
- 클러스터 생성, 조회, 삭제
- 노드 풀 관리
- Kubeconfig 다운로드

### Network
- VPC 관리
- 서브넷 관리
- 방화벽 규칙 관리

## 문제 해결

### 일반적인 오류
- **401 Unauthorized**: JWT 토큰이 유효하지 않습니다.
- **403 Forbidden**: GCP credential이 올바르지 않습니다.
- **404 Not Found**: 리소스를 찾을 수 없습니다.
- **500 Internal Server Error**: 서버 내부 오류입니다.

### 디버깅 팁
- Bruno의 응답 로그를 확인하세요.
- GCP 콘솔에서 리소스 상태를 확인하세요.
- 서버 로그를 확인하세요.

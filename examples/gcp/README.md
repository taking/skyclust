# GCP Provider Example

이 예제는 Google Cloud Platform (GCP) Compute Engine을 위한 클라우드 프로바이더 플러그인을 만드는 방법을 보여줍니다.

## 기능

- Compute Engine 인스턴스 목록 조회
- Compute Engine 인스턴스 생성
- Compute Engine 인스턴스 삭제
- 인스턴스 상태 확인
- GCP 리전 목록 조회
- 비용 추정 (모의)

## 필요한 의존성

```go
require (
    cloud.google.com/go/compute v1.23.0
    google.golang.org/api v0.143.0
)
```

## 설정

GCP 자격 증명을 설정해야 합니다:

### 방법 1: gcloud CLI 설정
```bash
gcloud auth application-default login
```

### 방법 2: 서비스 계정 키 파일
```bash
export GOOGLE_APPLICATION_CREDENTIALS="path/to/service-account-key.json"
```

### 방법 3: 환경 변수
```bash
export GOOGLE_APPLICATION_CREDENTIALS="path/to/service-account-key.json"
export GOOGLE_CLOUD_PROJECT="your-project-id"
```

## 사용법

### 1. 플러그인 빌드
```bash
cd examples/gcp
go mod init gcp-example
go mod tidy
go build -buildmode=plugin -o ../../../plugins/gcp-example.so gcp-example.go
```

### 2. 서버 실행
```bash
cd ../..
make run
```

### 3. API 테스트
```bash
# 프로바이더 목록 조회
curl http://localhost:8080/api/v1/providers

# GCP 인스턴스 목록 조회
curl http://localhost:8080/api/v1/providers/gcp/instances

# 새 인스턴스 생성
curl -X POST http://localhost:8080/api/v1/providers/gcp/instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-instance",
    "type": "e2-micro",
    "region": "us-central1"
  }'
```

## 주요 구현 사항

### 1. Google Cloud Client Libraries 사용
최신 Google Cloud Client Libraries를 사용하여 Compute Engine API와 상호작용합니다.

### 2. 리전 및 존 관리
GCP의 리전-존 구조를 적절히 처리하여 인스턴스를 관리합니다.

### 3. 라벨 처리
GCP 인스턴스의 라벨을 태그로 변환하여 메타데이터를 관리합니다.

### 4. 네트워크 인터페이스 처리
인스턴스의 내부 IP와 외부 IP를 적절히 추출합니다.

## GCP 특화 기능

### 1. 존 기반 관리
GCP는 리전 내에 여러 존이 있으므로, 모든 존을 순회하여 인스턴스를 찾습니다.

### 2. 메타데이터 처리
GCP 인스턴스의 메타데이터를 사용하여 시작 스크립트 등을 설정할 수 있습니다.

### 3. 라벨 시스템
GCP의 라벨 시스템을 활용하여 인스턴스에 메타데이터를 추가할 수 있습니다.

## 확장 가능한 기능

- VPC 네트워크 관리
- 방화벽 규칙 관리
- 로드 밸런서 관리
- Cloud SQL 데이터베이스 관리
- Cloud Storage 버킷 관리
- IAM 정책 관리
- Kubernetes Engine 클러스터 관리

## 주의사항

- 실제 프로덕션 환경에서는 적절한 에러 처리와 로깅을 추가해야 합니다.
- GCP 자격 증명은 안전하게 관리해야 합니다.
- 비용 추정은 실제 GCP Pricing API를 사용하는 것이 좋습니다.
- 대용량 인스턴스 목록의 경우 페이지네이션을 구현해야 할 수 있습니다.
- GCP의 할당량 제한을 고려해야 합니다.

## GCP 프로젝트 설정

1. GCP 콘솔에서 새 프로젝트를 생성합니다.
2. Compute Engine API를 활성화합니다.
3. 서비스 계정을 생성하고 필요한 권한을 부여합니다.
4. 서비스 계정 키를 다운로드하여 설정합니다.

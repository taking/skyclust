# AWS Provider Example

이 예제는 AWS EC2를 위한 클라우드 프로바이더 플러그인을 만드는 방법을 보여줍니다.

## 기능

- EC2 인스턴스 목록 조회
- EC2 인스턴스 생성
- EC2 인스턴스 삭제
- 인스턴스 상태 확인
- AWS 리전 목록 조회
- 비용 추정 (모의)

## 필요한 의존성

```go
require (
    github.com/aws/aws-sdk-go-v2 v1.39.0
    github.com/aws/aws-sdk-go-v2/config v1.31.8
    github.com/aws/aws-sdk-go-v2/service/ec2 v1.251.2
    github.com/aws/aws-sdk-go-v2/service/iam v1.47.5
)
```

## 설정

AWS 자격 증명을 설정해야 합니다:

### 방법 1: AWS CLI 설정
```bash
aws configure
```

### 방법 2: 환경 변수
```bash
export AWS_ACCESS_KEY_ID=your-access-key
export AWS_SECRET_ACCESS_KEY=your-secret-key
export AWS_DEFAULT_REGION=us-east-1
```

### 방법 3: IAM 역할 (EC2에서 실행할 때)
EC2 인스턴스에 적절한 IAM 역할을 할당합니다.

## 사용법

### 1. 플러그인 빌드
```bash
cd examples/aws
go mod init aws-example
go mod tidy
go build -buildmode=plugin -o ../../../plugins/aws-example.so aws-example.go
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

# AWS 인스턴스 목록 조회
curl http://localhost:8080/api/v1/providers/aws/instances

# 새 인스턴스 생성
curl -X POST http://localhost:8080/api/v1/providers/aws/instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-instance",
    "type": "t3.micro",
    "image_id": "ami-0abcdef1234567890",
    "region": "us-east-1"
  }'
```

## 주요 구현 사항

### 1. AWS SDK v2 사용
최신 AWS SDK v2를 사용하여 EC2 서비스와 상호작용합니다.

### 2. 태그 처리
EC2 인스턴스의 태그를 적절히 처리하여 이름과 메타데이터를 추출합니다.

### 3. 에러 처리
AWS API 호출 시 발생하는 에러를 적절히 처리하고 사용자에게 의미있는 메시지를 제공합니다.

### 4. 비용 추정
실제 AWS Pricing API 대신 모의 데이터를 사용하여 비용을 추정합니다.

## 확장 가능한 기능

- VPC 및 서브넷 관리
- 보안 그룹 관리
- 로드 밸런서 관리
- RDS 데이터베이스 관리
- S3 버킷 관리
- IAM 사용자 및 역할 관리

## 주의사항

- 실제 프로덕션 환경에서는 적절한 에러 처리와 로깅을 추가해야 합니다.
- AWS 자격 증명은 안전하게 관리해야 합니다.
- 비용 추정은 실제 AWS Pricing API를 사용하는 것이 좋습니다.
- 대용량 인스턴스 목록의 경우 페이지네이션을 구현해야 할 수 있습니다.

# AWS EKS API 테스트 가이드

이 폴더는 AWS EKS 클러스터 관리와 관련된 API 테스트 파일들을 포함합니다.

## 📋 테스트 파일 목록

### 1. 리소스 조회 API
- **VPC 목록 조회.bru** - AWS VPC 목록 조회
- **서브넷 목록 조회.bru** - 특정 VPC의 서브넷 목록 조회
- **보안 그룹 목록 조회.bru** - 특정 VPC의 보안 그룹 목록 조회
- **IAM 역할 목록 조회.bru** - EKS 관련 IAM 역할 목록 조회

### 2. 클러스터 관리 API
- **클러스터 생성.bru** - 기본 EKS 클러스터 생성
- **클러스터 생성 (리소스 선택).bru** - 리소스 선택을 포함한 EKS 클러스터 생성
- **클러스터 목록 조회.bru** - EKS 클러스터 목록 조회
- **클러스터 상세 조회.bru** - 특정 EKS 클러스터 상세 정보 조회
- **클러스터 삭제.bru** - EKS 클러스터 삭제

### 3. Kubeconfig 관리 API
- **Kubeconfig 다운로드.bru** - EKS 클러스터 Kubeconfig 다운로드

## 🔧 환경 변수 설정

테스트를 실행하기 전에 다음 환경 변수들을 설정해야 합니다:

### 필수 변수
- `jwt_token`: 인증용 JWT 토큰
- `aws_credential_id`: AWS 자격 증명 ID (UUID)
- `vpc_id`: 사용할 VPC ID
- `subnet_id_1`, `subnet_id_2`: 사용할 서브넷 ID들
- `cluster_role_arn`: EKS 클러스터용 IAM 역할 ARN
- `node_role_arn`: EKS 노드 그룹용 IAM 역할 ARN

### 설정 방법
1. Bruno에서 "Development" 환경을 선택
2. 환경 변수 섹션에서 각 변수에 실제 값을 입력
3. 또는 `bruno.json` 파일에서 직접 수정

## 🚀 테스트 실행 순서

### 1단계: 리소스 조회
1. **VPC 목록 조회** - 사용 가능한 VPC 확인
2. **서브넷 목록 조회** - 선택한 VPC의 서브넷 확인
3. **보안 그룹 목록 조회** - 사용할 보안 그룹 확인
4. **IAM 역할 목록 조회** - EKS 관련 IAM 역할 확인

### 2단계: 클러스터 생성
1. **클러스터 생성 (리소스 선택)** - 조회한 리소스들을 사용하여 클러스터 생성

### 3단계: 클러스터 관리
1. **클러스터 목록 조회** - 생성된 클러스터 확인
2. **클러스터 상세 조회** - 클러스터 상세 정보 확인
3. **Kubeconfig 다운로드** - 클러스터 접속용 설정 파일 다운로드

### 4단계: 정리
1. **클러스터 삭제** - 테스트 완료 후 클러스터 삭제

## 📝 API 엔드포인트

### 리소스 조회
- `GET /api/v1/aws/kubernetes/vpcs` - VPC 목록
- `GET /api/v1/aws/kubernetes/subnets` - 서브넷 목록
- `GET /api/v1/aws/kubernetes/security-groups` - 보안 그룹 목록
- `GET /api/v1/aws/kubernetes/iam-roles` - IAM 역할 목록

### 클러스터 관리
- `POST /api/v1/aws/kubernetes/clusters` - 클러스터 생성
- `GET /api/v1/aws/kubernetes/clusters` - 클러스터 목록
- `GET /api/v1/aws/kubernetes/clusters/:name` - 클러스터 상세
- `DELETE /api/v1/aws/kubernetes/clusters/:name` - 클러스터 삭제
- `GET /api/v1/aws/kubernetes/clusters/:name/kubeconfig` - Kubeconfig 다운로드

## 🔍 응답 예시

### VPC 목록 조회 응답
```json
{
  "success": true,
  "data": {
    "vpcs": [
      {
        "id": "vpc-12345678",
        "name": "my-vpc",
        "cidr_block": "10.0.0.0/16",
        "state": "available",
        "is_default": false,
        "region": "us-west-2"
      }
    ]
  },
  "message": "VPCs retrieved successfully"
}
```

### 서브넷 목록 조회 응답
```json
{
  "success": true,
  "data": {
    "subnets": [
      {
        "id": "subnet-12345678",
        "name": "public-subnet-1",
        "vpc_id": "vpc-12345678",
        "cidr_block": "10.0.1.0/24",
        "availability_zone": "us-west-2a",
        "state": "available",
        "is_public": true,
        "region": "us-west-2"
      }
    ]
  },
  "message": "Subnets retrieved successfully"
}
```

## ⚠️ 주의사항

1. **비용**: EKS 클러스터 생성 시 AWS 비용이 발생합니다
2. **권한**: AWS 자격 증명에 EKS, EC2, IAM 권한이 필요합니다
3. **리전**: 모든 리소스는 동일한 AWS 리전에 있어야 합니다
4. **서브넷**: 최소 2개의 서브넷이 서로 다른 가용 영역에 있어야 합니다
5. **IAM 역할**: EKS 클러스터와 노드 그룹용 IAM 역할이 미리 생성되어 있어야 합니다

## 🆘 문제 해결

### 일반적인 오류
- **401 Unauthorized**: JWT 토큰이 유효하지 않음
- **404 Not Found**: 자격 증명 ID가 존재하지 않음
- **400 Bad Request**: 필수 파라미터 누락 또는 잘못된 형식
- **500 Internal Server Error**: AWS API 호출 실패

### 디버깅 팁
1. 서버 로그 확인
2. AWS 자격 증명 유효성 확인
3. IAM 권한 확인
4. 네트워크 연결 상태 확인

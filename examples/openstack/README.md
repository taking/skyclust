# OpenStack Provider Example

이 예제는 OpenStack Nova를 위한 클라우드 프로바이더 플러그인을 만드는 방법을 보여줍니다.

## 기능

- Nova 인스턴스 목록 조회
- Nova 인스턴스 생성
- Nova 인스턴스 삭제
- 인스턴스 상태 확인
- OpenStack 리전 목록 조회
- 비용 추정 (모의)

## 필요한 의존성

```go
require (
    github.com/gophercloud/gophercloud v1.8.0
)
```

## 설정

OpenStack 자격 증명을 설정해야 합니다:

### 방법 1: 환경 변수
```bash
export OS_AUTH_URL="http://your-openstack:5000/v3"
export OS_USERNAME="your-username"
export OS_PASSWORD="your-password"
export OS_DOMAIN_NAME="Default"
export OS_PROJECT_NAME="your-project"
export OS_REGION_NAME="RegionOne"
```

### 방법 2: 설정 파일
```yaml
providers:
  openstack:
    auth_url: "http://your-openstack:5000/v3"
    username: "your-username"
    password: "your-password"
    domain_name: "Default"
    project_name: "your-project"
    region: "RegionOne"
```

## 사용법

### 1. 플러그인 빌드
```bash
cd examples/openstack
go mod init openstack-example
go mod tidy
go build -buildmode=plugin -o ../../../plugins/openstack-example.so openstack-example.go
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

# OpenStack 인스턴스 목록 조회
curl http://localhost:8080/api/v1/providers/openstack/instances

# 새 인스턴스 생성
curl -X POST http://localhost:8080/api/v1/providers/openstack/instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-instance",
    "type": "m1.small",
    "image_id": "your-image-id",
    "region": "RegionOne"
  }'
```

## 주요 구현 사항

### 1. Gophercloud 라이브러리 사용
OpenStack의 공식 Go 클라이언트 라이브러리인 Gophercloud를 사용합니다.

### 2. Keystone 인증
OpenStack의 Keystone 서비스를 사용하여 인증을 처리합니다.

### 3. Nova API 사용
OpenStack의 Nova 서비스를 사용하여 인스턴스를 관리합니다.

### 4. 메타데이터 처리
OpenStack 인스턴스의 메타데이터를 태그로 변환합니다.

## OpenStack 특화 기능

### 1. 프로젝트 기반 관리
OpenStack은 프로젝트 기반으로 리소스를 관리하므로 프로젝트 ID를 추출합니다.

### 2. 플레이버 관리
OpenStack의 플레이버 시스템을 사용하여 인스턴스 타입을 관리합니다.

### 3. 네트워크 주소 처리
OpenStack의 네트워크 주소 구조를 처리하여 내부 IP와 외부 IP를 구분합니다.

### 4. 사용자 데이터 처리
OpenStack의 사용자 데이터 기능을 사용하여 인스턴스 초기화 스크립트를 설정할 수 있습니다.

## 확장 가능한 기능

- Neutron 네트워크 관리
- Cinder 볼륨 관리
- Glance 이미지 관리
- Heat 오케스트레이션 관리
- Keystone 사용자 및 역할 관리
- Cinder 스냅샷 관리
- Nova 키페어 관리

## 주의사항

- 실제 프로덕션 환경에서는 적절한 에러 처리와 로깅을 추가해야 합니다.
- OpenStack 자격 증명은 안전하게 관리해야 합니다.
- OpenStack의 할당량 제한을 고려해야 합니다.
- 대용량 인스턴스 목록의 경우 페이지네이션을 구현해야 할 수 있습니다.
- OpenStack 버전에 따라 API가 다를 수 있습니다.

## OpenStack 설정

1. OpenStack 환경에 접근할 수 있는 자격 증명을 준비합니다.
2. Keystone 서비스의 엔드포인트 URL을 확인합니다.
3. 프로젝트 이름과 도메인을 확인합니다.
4. 필요한 권한이 있는 사용자 계정을 사용합니다.

## 지원되는 OpenStack 서비스

- Nova (Compute)
- Neutron (Networking) - 확장 가능
- Cinder (Block Storage) - 확장 가능
- Glance (Image) - 확장 가능
- Keystone (Identity) - 확장 가능

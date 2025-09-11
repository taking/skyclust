# OpenStack Provider

OpenStack을 위한 클라우드 프로바이더 플러그인입니다. 이 플러그인은 OpenStack의 Nova (Compute), Neutron (Networking), Keystone (Identity) 서비스를 통합하여 VM 생성 및 관리 기능을 제공합니다.

## 주요 기능

### VM 관리
- 인스턴스 생성, 삭제, 상태 조회
- 인스턴스 목록 조회
- 리전 관리

### 네트워크 관리 (Neutron)
- VPC/Network 생성 및 관리
- Subnet 생성 및 관리
- Security Group 관리
- Security Group Rules 관리
- Key Pair 관리
- Load Balancer 관리

### IAM 관리 (Keystone)
- 사용자 관리
- 그룹 관리
- 역할 관리
- 정책 관리
- 권한 관리
- 액세스 키 관리

## 필요한 OpenStack 설정

### 1. OpenStack 환경 설정

OpenStack 환경에서 다음 정보가 필요합니다:

- **Auth URL**: Keystone 인증 엔드포인트
- **Username**: OpenStack 사용자명
- **Password**: OpenStack 비밀번호
- **Project ID**: 프로젝트 ID
- **Region**: 리전 이름 (기본값: RegionOne)

### 2. 권한 설정

사용자에게 다음 권한이 필요합니다:

- **Nova**: 인스턴스 관리 권한
- **Neutron**: 네트워크 관리 권한
- **Keystone**: IAM 관리 권한

## 설정 방법

### 1. 플러그인 빌드

```bash
cd plugins/private/openstack
go build -buildmode=plugin -o ../../../plugins/openstack.so openstack.go
```

### 2. config.yaml 설정

```yaml
providers:
  openstack:
    auth_url: "http://your-openstack-keystone:5000/v3"
    username: "your-username"
    password: "your-password"
    project_id: "your-project-id"
    region: "RegionOne"
```

### 3. 서버 실행

```bash
make run
```

## API 사용 예시

### OpenStack 인스턴스 목록 조회

```bash
curl http://localhost:8080/api/v1/providers/openstack/instances
```

### OpenStack VM 생성 (네트워크 및 키페어 포함)

```bash
curl -X POST http://localhost:8080/api/v1/providers/openstack/instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-openstack-vm",
    "type": "m1.small",
    "region": "RegionOne",
    "image_id": "ubuntu-20.04",
    "vpc_id": "openstack-network-001",
    "subnet_id": "openstack-subnet-001",
    "security_groups": ["openstack-sg-001"],
    "key_pair_name": "my-keypair",
    "public_ip": true,
    "root_volume_size": 20,
    "root_volume_type": "standard"
  }'
```

### 네트워크 관리

#### VPC 생성
```bash
curl -X POST http://localhost:8080/api/v1/providers/openstack/networks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-network",
    "cidr": "10.0.0.0/16",
    "region": "RegionOne"
  }'
```

#### Subnet 생성
```bash
curl -X POST http://localhost:8080/api/v1/providers/openstack/subnets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-subnet",
    "vpc_id": "openstack-network-001",
    "cidr": "10.0.1.0/24",
    "availability_zone": "nova"
  }'
```

#### Security Group 생성
```bash
curl -X POST http://localhost:8080/api/v1/providers/openstack/security-groups \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-security-group",
    "description": "My security group",
    "vpc_id": "openstack-network-001"
  }'
```

#### Key Pair 생성
```bash
curl -X POST http://localhost:8080/api/v1/providers/openstack/key-pairs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-keypair",
    "public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC..."
  }'
```

### IAM 관리

#### 사용자 생성
```bash
curl -X POST http://localhost:8080/api/v1/providers/openstack/iam/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "email": "newuser@example.com",
    "display_name": "New User"
  }'
```

#### 그룹 생성
```bash
curl -X POST http://localhost:8080/api/v1/providers/openstack/iam/groups \
  -H "Content-Type: application/json" \
  -d '{
    "name": "developers",
    "description": "Developer group"
  }'
```

## OpenStack 서비스 매핑

| 기능 | OpenStack 서비스 | 설명 |
|------|------------------|------|
| VM 관리 | Nova | 가상 머신 생성, 삭제, 관리 |
| 네트워크 관리 | Neutron | 네트워크, 서브넷, 보안 그룹 관리 |
| IAM 관리 | Keystone | 사용자, 그룹, 역할, 정책 관리 |
| 이미지 관리 | Glance | VM 이미지 관리 |
| 볼륨 관리 | Cinder | 스토리지 볼륨 관리 |

## 실제 OpenStack SDK 사용

현재는 모의 구현이지만, 실제 OpenStack SDK를 사용하려면:

1. OpenStack SDK 설치:
```bash
go get github.com/gophercloud/gophercloud
```

2. 인증 설정:
```go
import "github.com/gophercloud/gophercloud/openstack"

func (p *OpenStackProvider) Initialize(config map[string]interface{}) error {
    // OpenStack 인증 설정
    authOpts := gophercloud.AuthOptions{
        IdentityEndpoint: config["auth_url"].(string),
        Username:         config["username"].(string),
        Password:         config["password"].(string),
        TenantID:         config["project_id"].(string),
    }
    
    provider, err := openstack.AuthenticatedClient(authOpts)
    if err != nil {
        return err
    }
    
    // Nova 클라이언트 생성
    novaClient, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
        Region: config["region"].(string),
    })
    if err != nil {
        return err
    }
    
    p.novaClient = novaClient
    return nil
}
```

## 주의사항

- 실제 프로덕션 환경에서는 OpenStack SDK를 사용해야 합니다
- OpenStack 환경의 네트워크 설정을 확인하세요
- 적절한 권한이 있는 사용자를 사용하세요
- 보안 그룹 규칙을 신중하게 설정하세요

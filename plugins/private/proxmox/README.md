# Proxmox Provider

Proxmox VE를 위한 클라우드 프로바이더 플러그인입니다. 이 플러그인은 Proxmox VE의 API를 사용하여 VM 생성 및 관리 기능을 제공합니다.

## 주요 기능

### VM 관리
- 인스턴스 생성, 삭제, 상태 조회
- 인스턴스 목록 조회
- 노드 관리 (리전으로 표시)

## 필요한 Proxmox 설정

### 1. Proxmox VE 환경 설정

Proxmox VE 환경에서 다음 정보가 필요합니다:

- **Host**: Proxmox VE 서버 주소
- **Username**: Proxmox 사용자명
- **Password**: Proxmox 비밀번호
- **Realm**: 인증 영역 (기본값: pve)

### 2. 권한 설정

사용자에게 다음 권한이 필요합니다:

- **VM 관리**: VM 생성, 삭제, 시작, 중지 권한
- **노드 접근**: 노드 정보 조회 권한

## 설정 방법

### 1. 플러그인 빌드

```bash
cd plugins/private/proxmox
go build -buildmode=plugin -o ../../../plugins/private/proxmox.so proxmox.go
```

### 2. config.yaml 설정

```yaml
providers:
  proxmox:
    host: "your-proxmox-host"
    username: "your-username"
    password: "your-password"
    realm: "pve"
```

### 3. 서버 실행

```bash
make run
```

## API 사용 예시

### Proxmox VM 목록 조회

```bash
curl http://localhost:8080/api/v1/providers/private/proxmox/instances
```

### Proxmox VM 생성

```bash
curl -X POST http://localhost:8080/api/v1/providers/private/proxmox/instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-proxmox-vm",
    "type": "medium",
    "region": "pve-node-1",
    "image_id": "local:iso/ubuntu-20.04-server-amd64.iso"
  }'
```

### Proxmox 노드 목록 조회

```bash
curl http://localhost:8080/api/v1/providers/private/proxmox/regions
```

### 비용 계산

```bash
curl -X POST http://localhost:8080/api/v1/providers/private/proxmox/cost-estimate \
  -H "Content-Type: application/json" \
  -d '{
    "instance_type": "medium",
    "region": "pve-node-1",
    "duration": "1d"
  }'
```

## Proxmox VM 타입

| 타입 | vCPU | RAM | 설명 |
|------|------|-----|------|
| small | 1 | 1GB | 소규모 워크로드 |
| medium | 2 | 2GB | 중간 규모 워크로드 |
| large | 4 | 4GB | 대규모 워크로드 |

## 실제 Proxmox API 사용

이 플러그인은 `github.com/luthermonson/go-proxmox` 라이브러리를 사용합니다:

```go
import "github.com/luthermonson/go-proxmox"

func (p *ProxmoxProvider) Initialize(config map[string]interface{}) error {
    // Proxmox 클라이언트 생성
    p.client = proxmox.NewClient(fmt.Sprintf("https://%s:8006", p.host))
    
    // 인증
    _, err := p.client.Login(p.username, p.password, p.realm)
    if err != nil {
        return fmt.Errorf("failed to authenticate with Proxmox: %w", err)
    }
    
    return nil
}
```

## 주의사항

- Proxmox VE 서버가 네트워크에서 접근 가능해야 합니다
- 적절한 권한이 있는 사용자를 사용하세요
- VM 생성 시 충분한 리소스가 있는 노드를 선택하세요
- Proxmox VE의 API 제한을 고려하세요

## Proxmox VE 설정

### 1. API 활성화

Proxmox VE에서 API가 활성화되어 있는지 확인하세요:

```bash
# Proxmox VE 서버에서
systemctl status pveproxy
```

### 2. 사용자 권한 설정

Proxmox VE 웹 인터페이스에서 사용자 권한을 설정하세요:

1. **Datacenter** > **Permissions** > **Users**에서 사용자 생성
2. **Datacenter** > **Permissions** > **Roles**에서 역할 생성
3. **Datacenter** > **Permissions** > **User Permissions**에서 권한 할당

### 3. 방화벽 설정

Proxmox VE의 8006 포트가 열려있는지 확인하세요:

```bash
# 방화벽에서 8006 포트 열기
ufw allow 8006
```

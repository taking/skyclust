# Proxmox Provider Example

이 예제는 Proxmox VE를 위한 클라우드 프로바이더 플러그인을 만드는 방법을 보여줍니다.

## 기능

- Proxmox VM 목록 조회
- Proxmox VM 생성
- Proxmox VM 삭제
- VM 상태 확인
- Proxmox 노드 목록 조회
- 비용 추정 (모의)

## 필요한 의존성

```go
require (
    github.com/luthermonson/go-proxmox v0.2.3
)
```

## 설정

Proxmox 자격 증명을 설정해야 합니다:

### 방법 1: 환경 변수
```bash
export PROXMOX_HOST="your-proxmox-host"
export PROXMOX_USERNAME="your-username"
export PROXMOX_PASSWORD="your-password"
export PROXMOX_REALM="pve"
```

### 방법 2: 설정 파일
```yaml
providers:
  proxmox:
    host: "your-proxmox-host"
    username: "your-username"
    password: "your-password"
    realm: "pve"
```

## 사용법

### 1. 플러그인 빌드
```bash
cd examples/proxmox
go mod init proxmox-example
go mod tidy
go build -buildmode=plugin -o ../../../plugins/proxmox-example.so proxmox-example.go
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

# Proxmox VM 목록 조회
curl http://localhost:8080/api/v1/providers/proxmox/instances

# 새 VM 생성
curl -X POST http://localhost:8080/api/v1/providers/proxmox/instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-vm",
    "type": "qemu",
    "region": "node1"
  }'
```

## 주요 구현 사항

### 1. go-proxmox 라이브러리 사용
Proxmox의 공식 Go 클라이언트 라이브러리를 사용합니다.

### 2. 노드 기반 관리
Proxmox는 클러스터 환경에서 여러 노드를 관리하므로 노드별로 VM을 조회합니다.

### 3. VMID 관리
Proxmox는 VMID를 사용하여 VM을 식별하므로 VMID를 적절히 관리합니다.

### 4. 태그 처리
Proxmox VM의 태그를 메타데이터로 변환합니다.

## Proxmox 특화 기능

### 1. 클러스터 관리
Proxmox 클러스터의 모든 노드를 순회하여 VM을 관리합니다.

### 2. VMID 자동 할당
사용 가능한 VMID를 자동으로 찾아서 할당합니다.

### 3. 템플릿 지원
Proxmox의 템플릿 기능을 지원하여 VM과 템플릿을 구분합니다.

### 4. 노드별 리전 관리
각 Proxmox 노드를 별도의 리전으로 관리합니다.

## 확장 가능한 기능

- 스토리지 관리
- 네트워크 관리
- 백업 관리
- 스냅샷 관리
- 컨테이너 관리
- 클러스터 관리
- 사용자 및 권한 관리

## 주의사항

- 실제 프로덕션 환경에서는 적절한 에러 처리와 로깅을 추가해야 합니다.
- Proxmox 자격 증명은 안전하게 관리해야 합니다.
- Proxmox의 리소스 제한을 고려해야 합니다.
- 대용량 VM 목록의 경우 페이지네이션을 구현해야 할 수 있습니다.
- Proxmox 버전에 따라 API가 다를 수 있습니다.

## Proxmox 설정

1. Proxmox VE 환경에 접근할 수 있는 자격 증명을 준비합니다.
2. Proxmox 호스트의 IP 주소와 포트(기본: 8006)를 확인합니다.
3. 사용자 계정과 도메인을 확인합니다.
4. 필요한 권한이 있는 사용자 계정을 사용합니다.

## 지원되는 Proxmox 기능

- QEMU/KVM 가상머신
- LXC 컨테이너
- 스토리지 관리
- 네트워크 관리
- 백업 및 복원
- 스냅샷 관리
- 클러스터 관리

## Proxmox API 특징

- RESTful API 사용
- JSON 기반 통신
- 인증 토큰 사용
- 실시간 상태 모니터링
- 비동기 작업 처리

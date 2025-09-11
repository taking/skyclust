# Provider Examples

이 폴더에는 새로운 클라우드 프로바이더를 추가하는 방법을 보여주는 예제들이 포함되어 있습니다.

## 예제 목록

### 실제 구현된 프로바이더들
1. **AWS Provider** (`aws/`) - Amazon Web Services EC2
2. **GCP Provider** (`gcp/`) - Google Cloud Platform Compute Engine
3. **OpenStack Provider** (`openstack/`) - OpenStack Nova
4. **Proxmox Provider** (`proxmox/`) - Proxmox VE

### 기존 예제들
5. **Azure Provider** (`azure/`) - Microsoft Azure 클라우드 서비스
6. **DigitalOcean Provider** (`digitalocean/`) - DigitalOcean 클라우드 서비스
7. **Custom Provider** (`custom/`) - 커스텀 프로바이더 템플릿

## 새로운 프로바이더 추가 방법

### 1단계: 프로바이더 디렉토리 생성

```bash
mkdir plugins/your-provider
cd plugins/your-provider
```

### 2단계: go.mod 파일 생성

```go
module your-provider-plugin

go 1.21

require cmp v0.0.0

replace cmp => ../../
```

### 3단계: 프로바이더 구현

`CloudProvider` 인터페이스를 구현하는 Go 파일을 작성합니다:

```go
package main

import (
    "context"
    "cmp/pkg/interfaces"
)

type YourProvider struct {
    config map[string]interface{}
}

func New() interfaces.CloudProvider {
    return &YourProvider{}
}

// 필수 메서드들 구현...
```

### 4단계: 플러그인 빌드

```bash
go build -buildmode=plugin -o ../../plugins/your-provider.so your-provider.go
```

### 5단계: 설정 추가

`config.yaml`에 새 프로바이더 설정을 추가합니다:

```yaml
providers:
  your-provider:
    api_key: "your-api-key"
    region: "default-region"
```

### 6단계: 테스트

서버를 실행하고 새 프로바이더가 로드되는지 확인합니다:

```bash
make run
curl http://localhost:8080/api/v1/providers
```

## 빌드 스크립트

모든 예제를 한 번에 빌드할 수 있는 스크립트가 제공됩니다:

```bash
./build-examples.sh
```

이 스크립트는 다음 작업을 수행합니다:
- 각 예제 디렉토리에서 `go mod init` 실행
- 메인 프로젝트 인터페이스를 참조하도록 `replace` 지시문 추가
- `go mod tidy`로 의존성 정리
- `go build -buildmode=plugin`으로 플러그인 빌드
- 빌드 결과를 `../../plugins/` 디렉토리에 저장

## 각 예제별 상세 가이드

각 예제 폴더에는 해당 프로바이더의 구현 방법과 사용법이 자세히 설명되어 있습니다.

### 실제 구현된 프로바이더들
- [AWS 예제](aws/README.md) - Amazon Web Services EC2
- [GCP 예제](gcp/README.md) - Google Cloud Platform Compute Engine  
- [OpenStack 예제](openstack/README.md) - OpenStack Nova
- [Proxmox 예제](proxmox/README.md) - Proxmox VE

### 기존 예제들
- [Azure 예제](azure/README.md) - Microsoft Azure
- [DigitalOcean 예제](digitalocean/README.md) - DigitalOcean
- [커스텀 예제](custom/README.md) - 커스텀 프로바이더 템플릿

# Azure Provider Example

이 예제는 Microsoft Azure를 위한 클라우드 프로바이더 플러그인을 구현하는 방법을 보여줍니다.

## 구현 내용

- Azure Virtual Machines 관리
- Azure 리전 조회
- 비용 계산
- 인스턴스 생성/삭제/상태 조회

## 필요한 Azure 설정

### 1. Azure Active Directory 앱 등록

1. Azure Portal에서 "Azure Active Directory" > "앱 등록"으로 이동
2. "새 등록" 클릭
3. 앱 이름 입력 후 "등록"
4. "클라이언트 ID"와 "테넌트 ID" 기록

### 2. 클라이언트 시크릿 생성

1. 앱 등록 페이지에서 "인증서 및 비밀" 클릭
2. "새 클라이언트 비밀" 클릭
3. 설명 입력 후 "만료" 설정
4. "값" 복사하여 안전하게 보관

### 3. 구독 ID 확인

1. Azure Portal에서 "구독" 클릭
2. 구독 ID 복사

## 설정 방법

### 1. 플러그인 빌드

```bash
cd examples/azure
go build -buildmode=plugin -o ../../plugins/azure.so azure.go
```

### 2. config.yaml 설정

```yaml
providers:
  azure:
    subscription_id: "your-subscription-id"
    client_id: "your-client-id"
    client_secret: "your-client-secret"
    tenant_id: "your-tenant-id"
    location: "East US"
```

### 3. 서버 실행

```bash
make run
```

## API 사용 예시

### Azure 인스턴스 목록 조회

```bash
curl http://localhost:8080/api/v1/providers/azure/instances
```

### Azure 리전 목록 조회

```bash
curl http://localhost:8080/api/v1/providers/azure/regions
```

### Azure VM 생성

```bash
curl -X POST http://localhost:8080/api/v1/providers/azure/instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-azure-vm",
    "type": "Standard_B1s",
    "region": "eastus",
    "image_id": "Canonical:UbuntuServer:18.04-LTS:latest"
  }'
```

### 비용 계산

```bash
curl -X POST http://localhost:8080/api/v1/providers/azure/cost-estimate \
  -H "Content-Type: application/json" \
  -d '{
    "instance_type": "Standard_B1s",
    "region": "eastus",
    "duration": "1d"
  }'
```

## 실제 Azure SDK 사용

현재는 모의 구현이지만, 실제 Azure SDK를 사용하려면:

1. Azure SDK 설치:
```bash
go get github.com/Azure/azure-sdk-for-go/sdk/azidentity
go get github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute
```

2. 인증 설정:
```go
import (
    "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
    "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

func (p *AzureProvider) Initialize(config map[string]interface{}) error {
    // Azure 인증 설정
    cred, err := azidentity.NewClientSecretCredential(
        config["tenant_id"].(string),
        config["client_id"].(string),
        config["client_secret"].(string),
        nil,
    )
    if err != nil {
        return err
    }
    
    // Compute client 생성
    client, err := armcompute.NewVirtualMachinesClient(
        config["subscription_id"].(string),
        cred,
        nil,
    )
    if err != nil {
        return err
    }
    
    p.client = client
    return nil
}
```

## 주의사항

- 실제 프로덕션 환경에서는 Azure SDK를 사용해야 합니다
- 클라이언트 시크릿은 안전하게 보관하세요
- 적절한 권한이 있는 서비스 주체를 사용하세요

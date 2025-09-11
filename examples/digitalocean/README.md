# DigitalOcean Provider Example

이 예제는 DigitalOcean을 위한 클라우드 프로바이더 플러그인을 구현하는 방법을 보여줍니다.

## 구현 내용

- DigitalOcean Droplets 관리
- DigitalOcean 리전 조회
- 비용 계산
- 인스턴스 생성/삭제/상태 조회

## 필요한 DigitalOcean 설정

### 1. API 토큰 생성

1. DigitalOcean Control Panel에 로그인
2. "API" 메뉴 클릭
3. "Generate New Token" 클릭
4. 토큰 이름 입력 (예: "CMP Token")
5. "Generate Token" 클릭
6. 생성된 토큰을 안전하게 보관

### 2. 권한 설정

API 토큰에는 다음 권한이 필요합니다:
- Read (읽기)
- Write (쓰기)

## 설정 방법

### 1. 플러그인 빌드

```bash
cd examples/digitalocean
go build -buildmode=plugin -o ../../plugins/digitalocean.so digitalocean.go
```

### 2. config.yaml 설정

```yaml
providers:
  digitalocean:
    api_token: "your-digitalocean-api-token"
    region: "nyc1"
```

### 3. 서버 실행

```bash
make run
```

## API 사용 예시

### DigitalOcean Droplets 목록 조회

```bash
curl http://localhost:8080/api/v1/providers/digitalocean/instances
```

### DigitalOcean 리전 목록 조회

```bash
curl http://localhost:8080/api/v1/providers/digitalocean/regions
```

### DigitalOcean Droplet 생성

```bash
curl -X POST http://localhost:8080/api/v1/providers/digitalocean/instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-droplet",
    "type": "s-1vcpu-1gb",
    "region": "nyc1",
    "image_id": "ubuntu-20-04-x64"
  }'
```

### 비용 계산

```bash
curl -X POST http://localhost:8080/api/v1/providers/digitalocean/cost-estimate \
  -H "Content-Type: application/json" \
  -d '{
    "instance_type": "s-1vcpu-1gb",
    "region": "nyc1",
    "duration": "1d"
  }'
```

## 실제 DigitalOcean API 사용

현재는 모의 구현이지만, 실제 DigitalOcean API를 사용하려면:

1. HTTP 클라이언트 설정:
```go
import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type DigitalOceanClient struct {
    apiToken string
    baseURL  string
}

func NewDigitalOceanClient(apiToken string) *DigitalOceanClient {
    return &DigitalOceanClient{
        apiToken: apiToken,
        baseURL:  "https://api.digitalocean.com/v2",
    }
}

func (c *DigitalOceanClient) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
    var reqBody []byte
    var err error
    
    if body != nil {
        reqBody, err = json.Marshal(body)
        if err != nil {
            return nil, err
        }
    }
    
    req, err := http.NewRequest(method, c.baseURL+endpoint, bytes.NewBuffer(reqBody))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.apiToken)
    req.Header.Set("Content-Type", "application/json")
    
    return http.DefaultClient.Do(req)
}
```

2. Droplets 목록 조회:
```go
func (p *DigitalOceanProvider) ListInstances(ctx context.Context) ([]interfaces.Instance, error) {
    resp, err := p.client.makeRequest("GET", "/droplets", nil)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    // 응답 파싱 및 변환
    // ...
}
```

## DigitalOcean Droplet 타입

| 타입 | vCPU | RAM | 디스크 | 월 비용 |
|------|------|-----|--------|---------|
| s-1vcpu-1gb | 1 | 1GB | 25GB SSD | $5 |
| s-1vcpu-2gb | 1 | 2GB | 50GB SSD | $10 |
| s-2vcpu-2gb | 2 | 2GB | 50GB SSD | $20 |
| s-2vcpu-4gb | 2 | 4GB | 80GB SSD | $40 |
| s-4vcpu-8gb | 4 | 8GB | 160GB SSD | $80 |

## 주의사항

- 실제 프로덕션 환경에서는 DigitalOcean API를 사용해야 합니다
- API 토큰은 안전하게 보관하세요
- API 호출 제한을 고려하세요 (초당 1200회)
- 적절한 에러 처리를 구현하세요

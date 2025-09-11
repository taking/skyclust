# Custom Provider Template

이 예제는 새로운 클라우드 프로바이더를 만들기 위한 템플릿입니다. 이 템플릿을 기반으로 자신만의 클라우드 서비스 프로바이더를 구현할 수 있습니다.

## 템플릿 구조

```
custom/
├── custom.go          # 메인 프로바이더 구현
├── go.mod            # Go 모듈 정의
└── README.md         # 이 파일
```

## 구현 단계

### 1단계: 기본 구조 설정

1. 프로바이더 이름 변경:
```go
func (p *CustomProvider) GetName() string {
    return "Your Cloud Provider Name"
}
```

2. 버전 설정:
```go
func (p *CustomProvider) GetVersion() string {
    return "1.0.0"
}
```

### 2단계: 설정 검증 구현

`Initialize` 메서드에서 필요한 설정을 검증합니다:

```go
func (p *CustomProvider) Initialize(config map[string]interface{}) error {
    p.config = config
    
    // 필수 설정 검증
    if _, ok := config["api_endpoint"]; !ok {
        return fmt.Errorf("API endpoint is required")
    }
    if _, ok := config["api_key"]; !ok {
        return fmt.Errorf("API key is required")
    }
    
    // API 클라이언트 초기화
    p.client = NewAPIClient(config["api_endpoint"].(string), config["api_key"].(string))
    
    return nil
}
```

### 3단계: API 클라이언트 구현

실제 클라우드 서비스 API와 통신하는 클라이언트를 구현합니다:

```go
type APIClient struct {
    endpoint string
    apiKey   string
    httpClient *http.Client
}

func NewAPIClient(endpoint, apiKey string) *APIClient {
    return &APIClient{
        endpoint: endpoint,
        apiKey:   apiKey,
        httpClient: &http.Client{Timeout: 30 * time.Second},
    }
}

func (c *APIClient) makeRequest(method, path string, body interface{}) (*http.Response, error) {
    // HTTP 요청 구현
}
```

### 4단계: 각 메서드 구현

각 `CloudProvider` 인터페이스 메서드를 실제 API 호출로 구현합니다:

```go
func (p *CustomProvider) ListInstances(ctx context.Context) ([]interfaces.Instance, error) {
    // 실제 API 호출
    resp, err := p.client.GetInstances()
    if err != nil {
        return nil, err
    }
    
    // 응답을 interfaces.Instance 형태로 변환
    instances := make([]interfaces.Instance, len(resp.Instances))
    for i, inst := range resp.Instances {
        instances[i] = interfaces.Instance{
            ID:        inst.ID,
            Name:      inst.Name,
            Status:    inst.Status,
            Type:      inst.Type,
            Region:    inst.Region,
            CreatedAt: inst.CreatedAt,
            Tags:      inst.Tags,
            PublicIP:  inst.PublicIP,
            PrivateIP: inst.PrivateIP,
        }
    }
    
    return instances, nil
}
```

## 설정 예시

### config.yaml 설정

```yaml
providers:
  custom:
    api_endpoint: "https://api.yourcloud.com/v1"
    api_key: "your-api-key"
    region: "us-east-1"
    # 추가 설정...
```

## 빌드 및 테스트

### 1. 플러그인 빌드

```bash
cd examples/custom
go build -buildmode=plugin -o ../../plugins/custom.so custom.go
```

### 2. 서버 실행

```bash
make run
```

### 3. 테스트

```bash
# 프로바이더 목록 확인
curl http://localhost:8080/api/v1/providers

# 인스턴스 목록 조회
curl http://localhost:8080/api/v1/providers/custom/instances
```

## 실제 구현 예시

### HTTP 클라이언트 구현

```go
import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type APIClient struct {
    endpoint   string
    apiKey     string
    httpClient *http.Client
}

func NewAPIClient(endpoint, apiKey string) *APIClient {
    return &APIClient{
        endpoint: endpoint,
        apiKey:   apiKey,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (c *APIClient) makeRequest(method, path string, body interface{}) (*http.Response, error) {
    var reqBody []byte
    var err error
    
    if body != nil {
        reqBody, err = json.Marshal(body)
        if err != nil {
            return nil, err
        }
    }
    
    req, err := http.NewRequest(method, c.endpoint+path, bytes.NewBuffer(reqBody))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.apiKey)
    req.Header.Set("Content-Type", "application/json")
    
    return c.httpClient.Do(req)
}
```

### 에러 처리

```go
func (p *CustomProvider) ListInstances(ctx context.Context) ([]interfaces.Instance, error) {
    resp, err := p.client.makeRequest("GET", "/instances", nil)
    if err != nil {
        return nil, fmt.Errorf("failed to make API request: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
    }
    
    // 응답 처리...
}
```

## 필수 인터페이스 메서드

모든 프로바이더는 다음 메서드들을 구현해야 합니다:

```go
// 기본 정보
func (p *CustomProvider) GetName() string
func (p *CustomProvider) GetVersion() string
func (p *CustomProvider) Initialize(config map[string]interface{}) error

// 인스턴스 관리
func (p *CustomProvider) ListInstances(ctx context.Context) ([]interfaces.Instance, error)
func (p *CustomProvider) CreateInstance(ctx context.Context, req interfaces.CreateInstanceRequest) (*interfaces.Instance, error)
func (p *CustomProvider) DeleteInstance(ctx context.Context, instanceID string) error
func (p *CustomProvider) GetInstanceStatus(ctx context.Context, instanceID string) (string, error)

// 리전 관리
func (p *CustomProvider) ListRegions(ctx context.Context) ([]interfaces.Region, error)

// 비용 추정
func (p *CustomProvider) GetCostEstimate(ctx context.Context, req interfaces.CostEstimateRequest) (*interfaces.CostEstimate, error)

// 확장 기능 (선택사항)
func (p *CustomProvider) GetNetworkProvider() interfaces.NetworkProvider
func (p *CustomProvider) GetIAMProvider() interfaces.IAMProvider
```

## 주의사항

1. **에러 처리**: 모든 API 호출에 적절한 에러 처리를 구현하세요
2. **타임아웃**: HTTP 클라이언트에 적절한 타임아웃을 설정하세요
3. **인증**: API 키나 토큰을 안전하게 관리하세요
4. **로깅**: 디버깅을 위해 적절한 로깅을 추가하세요
5. **테스트**: 단위 테스트를 작성하여 구현을 검증하세요
6. **컨텍스트**: 모든 메서드에서 `context.Context`를 적절히 처리하세요
7. **인터페이스 준수**: `interfaces.CloudProvider` 인터페이스를 정확히 구현하세요

## 다음 단계

1. 이 템플릿을 복사하여 새로운 프로바이더 디렉토리 생성
2. 클라우드 서비스 API 문서 확인
3. 필요한 설정 파라미터 정의
4. API 클라이언트 구현
5. 각 메서드 구현
6. 테스트 및 검증

# SkyClust 플러그인 개발 가이드

이 문서는 SkyClust용 클라우드 프로바이더 플러그인을 개발하는 방법을 설명합니다.

## 플러그인 개요

SkyClust는 플러그인 기반 아키텍처를 사용하여 다양한 클라우드 프로바이더를 지원합니다. 각 플러그인은 표준화된 인터페이스를 구현하여 일관된 API를 제공합니다.

## 플러그인 구조

```
plugins/
├── public/                 # 공개 클라우드 프로바이더
│   ├── aws/               # AWS 플러그인
│   ├── gcp/               # GCP 플러그인
│   └── azure/             # Azure 플러그인
└── private/               # 프라이빗 클라우드 프로바이더
    ├── openstack/         # OpenStack 플러그인
    └── proxmox/           # Proxmox 플러그인
```

## 플러그인 인터페이스

### 기본 인터페이스

```go
// pkg/plugin/cloud_provider.go
package plugin

import (
    "context"
    "time"
)

// CloudProvider는 클라우드 프로바이더의 기본 인터페이스입니다
type CloudProvider interface {
    // 기본 정보
    GetName() string
    GetVersion() string
    GetDescription() string
    
    // 초기화
    Initialize(config map[string]interface{}) error
    ValidateConfig(config map[string]interface{}) error
    
    // VM 관리
    ListInstances(ctx context.Context) ([]*Instance, error)
    CreateInstance(ctx context.Context, req CreateInstanceRequest) (*Instance, error)
    GetInstance(ctx context.Context, id string) (*Instance, error)
    UpdateInstance(ctx context.Context, id string, req UpdateInstanceRequest) (*Instance, error)
    DeleteInstance(ctx context.Context, id string) error
    StartInstance(ctx context.Context, id string) error
    StopInstance(ctx context.Context, id string) error
    RestartInstance(ctx context.Context, id string) error
    
    // 리전 관리
    ListRegions(ctx context.Context) ([]*Region, error)
    GetRegion(ctx context.Context, id string) (*Region, error)
    
    // 이미지 관리
    ListImages(ctx context.Context, req ListImagesRequest) ([]*Image, error)
    GetImage(ctx context.Context, id string) (*Image, error)
    
    // 인스턴스 타입 관리
    ListInstanceTypes(ctx context.Context, req ListInstanceTypesRequest) ([]*InstanceType, error)
    GetInstanceType(ctx context.Context, id string) (*InstanceType, error)
    
    // 비용 관리
    GetCostEstimate(ctx context.Context, req CostEstimateRequest) (*CostEstimate, error)
    GetBillingInfo(ctx context.Context, req BillingInfoRequest) (*BillingInfo, error)
    
    // 네트워킹
    ListNetworks(ctx context.Context) ([]*Network, error)
    GetNetwork(ctx context.Context, id string) (*Network, error)
    CreateNetwork(ctx context.Context, req CreateNetworkRequest) (*Network, error)
    DeleteNetwork(ctx context.Context, id string) error
    
    // 스토리지
    ListVolumes(ctx context.Context) ([]*Volume, error)
    GetVolume(ctx context.Context, id string) (*Volume, error)
    CreateVolume(ctx context.Context, req CreateVolumeRequest) (*Volume, error)
    DeleteVolume(ctx context.Context, id string) error
    AttachVolume(ctx context.Context, volumeID, instanceID string) error
    DetachVolume(ctx context.Context, volumeID, instanceID string) error
    
    // 보안
    ListSecurityGroups(ctx context.Context) ([]*SecurityGroup, error)
    GetSecurityGroup(ctx context.Context, id string) (*SecurityGroup, error)
    CreateSecurityGroup(ctx context.Context, req CreateSecurityGroupRequest) (*SecurityGroup, error)
    DeleteSecurityGroup(ctx context.Context, id string) error
    
    // 키 페어 관리
    ListKeyPairs(ctx context.Context) ([]*KeyPair, error)
    GetKeyPair(ctx context.Context, id string) (*KeyPair, error)
    CreateKeyPair(ctx context.Context, req CreateKeyPairRequest) (*KeyPair, error)
    DeleteKeyPair(ctx context.Context, id string) error
    
    // 헬스 체크
    HealthCheck(ctx context.Context) error
    
    // 정리
    Cleanup() error
}
```

### 데이터 모델

```go
// Instance는 클라우드 인스턴스를 나타냅니다
type Instance struct {
    ID           string            `json:"id"`
    Name         string            `json:"name"`
    Status       string            `json:"status"`
    InstanceType string            `json:"instance_type"`
    ImageID      string            `json:"image_id"`
    Region       string            `json:"region"`
    Zone         string            `json:"zone"`
    PublicIP     string            `json:"public_ip"`
    PrivateIP    string            `json:"private_ip"`
    Tags         map[string]string `json:"tags"`
    CreatedAt    time.Time         `json:"created_at"`
    UpdatedAt    time.Time         `json:"updated_at"`
}

// Region은 클라우드 리전을 나타냅니다
type Region struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    DisplayName string `json:"display_name"`
    Endpoint    string `json:"endpoint"`
    IsActive    bool   `json:"is_active"`
}

// Image는 클라우드 이미지를 나타냅니다
type Image struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Description string            `json:"description"`
    OS          string            `json:"os"`
    Architecture string           `json:"architecture"`
    Size        int64             `json:"size"`
    Tags        map[string]string `json:"tags"`
    CreatedAt   time.Time         `json:"created_at"`
}

// InstanceType은 인스턴스 타입을 나타냅니다
type InstanceType struct {
    ID           string  `json:"id"`
    Name         string  `json:"name"`
    VCPUs        int     `json:"vcpus"`
    Memory       int     `json:"memory"`
    Storage      int     `json:"storage"`
    Network      int     `json:"network"`
    Price        float64 `json:"price"`
    Currency     string  `json:"currency"`
    IsAvailable  bool    `json:"is_available"`
}

// CostEstimate는 비용 추정을 나타냅니다
type CostEstimate struct {
    InstanceType string    `json:"instance_type"`
    Region       string    `json:"region"`
    Duration     string    `json:"duration"`
    Cost         float64   `json:"cost"`
    Currency     string    `json:"currency"`
    Breakdown    []CostItem `json:"breakdown"`
}

// CostItem은 비용 항목을 나타냅니다
type CostItem struct {
    Service string  `json:"service"`
    Cost    float64 `json:"cost"`
    Unit    string  `json:"unit"`
}
```

### 요청/응답 모델

```go
// CreateInstanceRequest는 인스턴스 생성 요청을 나타냅니다
type CreateInstanceRequest struct {
    Name         string            `json:"name"`
    InstanceType string            `json:"instance_type"`
    ImageID      string            `json:"image_id"`
    Region       string            `json:"region"`
    Zone         string            `json:"zone"`
    KeyPair      string            `json:"key_pair"`
    SecurityGroups []string        `json:"security_groups"`
    Tags         map[string]string `json:"tags"`
    UserData     string            `json:"user_data"`
}

// UpdateInstanceRequest는 인스턴스 업데이트 요청을 나타냅니다
type UpdateInstanceRequest struct {
    Name         string            `json:"name"`
    InstanceType string            `json:"instance_type"`
    Tags         map[string]string `json:"tags"`
}

// ListImagesRequest는 이미지 목록 조회 요청을 나타냅니다
type ListImagesRequest struct {
    OS           string `json:"os"`
    Architecture string `json:"architecture"`
    Limit        int    `json:"limit"`
    Offset       int    `json:"offset"`
}

// ListInstanceTypesRequest는 인스턴스 타입 목록 조회 요청을 나타냅니다
type ListInstanceTypesRequest struct {
    Region string `json:"region"`
    Limit  int    `json:"limit"`
    Offset int    `json:"offset"`
}

// CostEstimateRequest는 비용 추정 요청을 나타냅니다
type CostEstimateRequest struct {
    InstanceType string `json:"instance_type"`
    Region       string `json:"region"`
    Duration     string `json:"duration"`
}

// BillingInfoRequest는 청구 정보 조회 요청을 나타냅니다
type BillingInfoRequest struct {
    StartDate string `json:"start_date"`
    EndDate   string `json:"end_date"`
    Region    string `json:"region"`
}
```

## 플러그인 개발

### 1. 플러그인 프로젝트 생성

```bash
# 플러그인 디렉토리 생성
mkdir plugins/my-provider
cd plugins/my-provider

# Go 모듈 초기화
go mod init my-provider-plugin

# 의존성 추가
go get github.com/skyclust/skyclust/pkg/plugin
```

### 2. 플러그인 구현

```go
// main.go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/skyclust/skyclust/pkg/plugin"
)

// MyProvider는 MyProvider 클라우드 프로바이더를 나타냅니다
type MyProvider struct {
    config map[string]interface{}
    client *MyProviderClient
}

// New는 새로운 MyProvider 인스턴스를 생성합니다
func New() plugin.CloudProvider {
    return &MyProvider{}
}

// GetName은 프로바이더 이름을 반환합니다
func (p *MyProvider) GetName() string {
    return "MyProvider"
}

// GetVersion은 프로바이더 버전을 반환합니다
func (p *MyProvider) GetVersion() string {
    return "1.0.0"
}

// GetDescription은 프로바이더 설명을 반환합니다
func (p *MyProvider) GetDescription() string {
    return "MyProvider 클라우드 프로바이더 플러그인"
}

// Initialize는 프로바이더를 초기화합니다
func (p *MyProvider) Initialize(config map[string]interface{}) error {
    p.config = config
    
    // 클라이언트 초기화
    client, err := NewMyProviderClient(config)
    if err != nil {
        return fmt.Errorf("failed to initialize client: %w", err)
    }
    
    p.client = client
    return nil
}

// ValidateConfig는 설정을 검증합니다
func (p *MyProvider) ValidateConfig(config map[string]interface{}) error {
    required := []string{"api_key", "api_secret", "region"}
    
    for _, field := range required {
        if _, exists := config[field]; !exists {
            return fmt.Errorf("required field %s is missing", field)
        }
    }
    
    return nil
}

// ListInstances는 인스턴스 목록을 조회합니다
func (p *MyProvider) ListInstances(ctx context.Context) ([]*plugin.Instance, error) {
    instances, err := p.client.ListInstances(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to list instances: %w", err)
    }
    
    return instances, nil
}

// CreateInstance는 인스턴스를 생성합니다
func (p *MyProvider) CreateInstance(ctx context.Context, req plugin.CreateInstanceRequest) (*plugin.Instance, error) {
    instance, err := p.client.CreateInstance(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to create instance: %w", err)
    }
    
    return instance, nil
}

// GetInstance는 인스턴스를 조회합니다
func (p *MyProvider) GetInstance(ctx context.Context, id string) (*plugin.Instance, error) {
    instance, err := p.client.GetInstance(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get instance: %w", err)
    }
    
    return instance, nil
}

// UpdateInstance는 인스턴스를 업데이트합니다
func (p *MyProvider) UpdateInstance(ctx context.Context, id string, req plugin.UpdateInstanceRequest) (*plugin.Instance, error) {
    instance, err := p.client.UpdateInstance(ctx, id, req)
    if err != nil {
        return nil, fmt.Errorf("failed to update instance: %w", err)
    }
    
    return instance, nil
}

// DeleteInstance는 인스턴스를 삭제합니다
func (p *MyProvider) DeleteInstance(ctx context.Context, id string) error {
    err := p.client.DeleteInstance(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to delete instance: %w", err)
    }
    
    return nil
}

// StartInstance는 인스턴스를 시작합니다
func (p *MyProvider) StartInstance(ctx context.Context, id string) error {
    err := p.client.StartInstance(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to start instance: %w", err)
    }
    
    return nil
}

// StopInstance는 인스턴스를 중지합니다
func (p *MyProvider) StopInstance(ctx context.Context, id string) error {
    err := p.client.StopInstance(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to stop instance: %w", err)
    }
    
    return nil
}

// RestartInstance는 인스턴스를 재시작합니다
func (p *MyProvider) RestartInstance(ctx context.Context, id string) error {
    err := p.client.RestartInstance(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to restart instance: %w", err)
    }
    
    return nil
}

// ListRegions은 리전 목록을 조회합니다
func (p *MyProvider) ListRegions(ctx context.Context) ([]*plugin.Region, error) {
    regions, err := p.client.ListRegions(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to list regions: %w", err)
    }
    
    return regions, nil
}

// GetRegion은 리전을 조회합니다
func (p *MyProvider) GetRegion(ctx context.Context, id string) (*plugin.Region, error) {
    region, err := p.client.GetRegion(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get region: %w", err)
    }
    
    return region, nil
}

// ListImages는 이미지 목록을 조회합니다
func (p *MyProvider) ListImages(ctx context.Context, req plugin.ListImagesRequest) ([]*plugin.Image, error) {
    images, err := p.client.ListImages(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to list images: %w", err)
    }
    
    return images, nil
}

// GetImage는 이미지를 조회합니다
func (p *MyProvider) GetImage(ctx context.Context, id string) (*plugin.Image, error) {
    image, err := p.client.GetImage(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get image: %w", err)
    }
    
    return image, nil
}

// ListInstanceTypes는 인스턴스 타입 목록을 조회합니다
func (p *MyProvider) ListInstanceTypes(ctx context.Context, req plugin.ListInstanceTypesRequest) ([]*plugin.InstanceType, error) {
    instanceTypes, err := p.client.ListInstanceTypes(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to list instance types: %w", err)
    }
    
    return instanceTypes, nil
}

// GetInstanceType은 인스턴스 타입을 조회합니다
func (p *MyProvider) GetInstanceType(ctx context.Context, id string) (*plugin.InstanceType, error) {
    instanceType, err := p.client.GetInstanceType(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get instance type: %w", err)
    }
    
    return instanceType, nil
}

// GetCostEstimate는 비용을 추정합니다
func (p *MyProvider) GetCostEstimate(ctx context.Context, req plugin.CostEstimateRequest) (*plugin.CostEstimate, error) {
    cost, err := p.client.GetCostEstimate(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to get cost estimate: %w", err)
    }
    
    return cost, nil
}

// GetBillingInfo는 청구 정보를 조회합니다
func (p *MyProvider) GetBillingInfo(ctx context.Context, req plugin.BillingInfoRequest) (*plugin.BillingInfo, error) {
    billing, err := p.client.GetBillingInfo(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to get billing info: %w", err)
    }
    
    return billing, nil
}

// ListNetworks는 네트워크 목록을 조회합니다
func (p *MyProvider) ListNetworks(ctx context.Context) ([]*plugin.Network, error) {
    networks, err := p.client.ListNetworks(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to list networks: %w", err)
    }
    
    return networks, nil
}

// GetNetwork은 네트워크를 조회합니다
func (p *MyProvider) GetNetwork(ctx context.Context, id string) (*plugin.Network, error) {
    network, err := p.client.GetNetwork(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get network: %w", err)
    }
    
    return network, nil
}

// CreateNetwork은 네트워크를 생성합니다
func (p *MyProvider) CreateNetwork(ctx context.Context, req plugin.CreateNetworkRequest) (*plugin.Network, error) {
    network, err := p.client.CreateNetwork(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to create network: %w", err)
    }
    
    return network, nil
}

// DeleteNetwork은 네트워크를 삭제합니다
func (p *MyProvider) DeleteNetwork(ctx context.Context, id string) error {
    err := p.client.DeleteNetwork(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to delete network: %w", err)
    }
    
    return nil
}

// ListVolumes은 볼륨 목록을 조회합니다
func (p *MyProvider) ListVolumes(ctx context.Context) ([]*plugin.Volume, error) {
    volumes, err := p.client.ListVolumes(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to list volumes: %w", err)
    }
    
    return volumes, nil
}

// GetVolume은 볼륨을 조회합니다
func (p *MyProvider) GetVolume(ctx context.Context, id string) (*plugin.Volume, error) {
    volume, err := p.client.GetVolume(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get volume: %w", err)
    }
    
    return volume, nil
}

// CreateVolume은 볼륨을 생성합니다
func (p *MyProvider) CreateVolume(ctx context.Context, req plugin.CreateVolumeRequest) (*plugin.Volume, error) {
    volume, err := p.client.CreateVolume(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to create volume: %w", err)
    }
    
    return volume, nil
}

// DeleteVolume은 볼륨을 삭제합니다
func (p *MyProvider) DeleteVolume(ctx context.Context, id string) error {
    err := p.client.DeleteVolume(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to delete volume: %w", err)
    }
    
    return nil
}

// AttachVolume은 볼륨을 연결합니다
func (p *MyProvider) AttachVolume(ctx context.Context, volumeID, instanceID string) error {
    err := p.client.AttachVolume(ctx, volumeID, instanceID)
    if err != nil {
        return fmt.Errorf("failed to attach volume: %w", err)
    }
    
    return nil
}

// DetachVolume은 볼륨을 분리합니다
func (p *MyProvider) DetachVolume(ctx context.Context, volumeID, instanceID string) error {
    err := p.client.DetachVolume(ctx, volumeID, instanceID)
    if err != nil {
        return fmt.Errorf("failed to detach volume: %w", err)
    }
    
    return nil
}

// ListSecurityGroups은 보안 그룹 목록을 조회합니다
func (p *MyProvider) ListSecurityGroups(ctx context.Context) ([]*plugin.SecurityGroup, error) {
    securityGroups, err := p.client.ListSecurityGroups(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to list security groups: %w", err)
    }
    
    return securityGroups, nil
}

// GetSecurityGroup은 보안 그룹을 조회합니다
func (p *MyProvider) GetSecurityGroup(ctx context.Context, id string) (*plugin.SecurityGroup, error) {
    securityGroup, err := p.client.GetSecurityGroup(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get security group: %w", err)
    }
    
    return securityGroup, nil
}

// CreateSecurityGroup은 보안 그룹을 생성합니다
func (p *MyProvider) CreateSecurityGroup(ctx context.Context, req plugin.CreateSecurityGroupRequest) (*plugin.SecurityGroup, error) {
    securityGroup, err := p.client.CreateSecurityGroup(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to create security group: %w", err)
    }
    
    return securityGroup, nil
}

// DeleteSecurityGroup은 보안 그룹을 삭제합니다
func (p *MyProvider) DeleteSecurityGroup(ctx context.Context, id string) error {
    err := p.client.DeleteSecurityGroup(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to delete security group: %w", err)
    }
    
    return nil
}

// ListKeyPairs은 키 페어 목록을 조회합니다
func (p *MyProvider) ListKeyPairs(ctx context.Context) ([]*plugin.KeyPair, error) {
    keyPairs, err := p.client.ListKeyPairs(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to list key pairs: %w", err)
    }
    
    return keyPairs, nil
}

// GetKeyPair은 키 페어를 조회합니다
func (p *MyProvider) GetKeyPair(ctx context.Context, id string) (*plugin.KeyPair, error) {
    keyPair, err := p.client.GetKeyPair(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get key pair: %w", err)
    }
    
    return keyPair, nil
}

// CreateKeyPair은 키 페어를 생성합니다
func (p *MyProvider) CreateKeyPair(ctx context.Context, req plugin.CreateKeyPairRequest) (*plugin.KeyPair, error) {
    keyPair, err := p.client.CreateKeyPair(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to create key pair: %w", err)
    }
    
    return keyPair, nil
}

// DeleteKeyPair은 키 페어를 삭제합니다
func (p *MyProvider) DeleteKeyPair(ctx context.Context, id string) error {
    err := p.client.DeleteKeyPair(ctx, id)
    if err != nil {
        return fmt.Errorf("failed to delete key pair: %w", err)
    }
    
    return nil
}

// HealthCheck은 헬스 체크를 수행합니다
func (p *MyProvider) HealthCheck(ctx context.Context) error {
    err := p.client.HealthCheck(ctx)
    if err != nil {
        return fmt.Errorf("health check failed: %w", err)
    }
    
    return nil
}

// Cleanup은 리소스를 정리합니다
func (p *MyProvider) Cleanup() error {
    if p.client != nil {
        return p.client.Close()
    }
    
    return nil
}

// main 함수는 플러그인 진입점입니다
func main() {
    // 플러그인은 main 함수를 구현하지 않아도 됩니다
    // SkyClust가 플러그인을 동적으로 로드합니다
}
```

### 3. 클라이언트 구현

```go
// client.go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/skyclust/skyclust/pkg/plugin"
)

// MyProviderClient는 MyProvider API 클라이언트입니다
type MyProviderClient struct {
    apiKey    string
    apiSecret string
    region    string
    endpoint  string
}

// NewMyProviderClient는 새로운 클라이언트를 생성합니다
func NewMyProviderClient(config map[string]interface{}) (*MyProviderClient, error) {
    apiKey, ok := config["api_key"].(string)
    if !ok {
        return nil, fmt.Errorf("api_key is required")
    }
    
    apiSecret, ok := config["api_secret"].(string)
    if !ok {
        return nil, fmt.Errorf("api_secret is required")
    }
    
    region, ok := config["region"].(string)
    if !ok {
        return nil, fmt.Errorf("region is required")
    }
    
    endpoint, ok := config["endpoint"].(string)
    if !ok {
        endpoint = fmt.Sprintf("https://api.myprovider.com/%s", region)
    }
    
    return &MyProviderClient{
        apiKey:    apiKey,
        apiSecret: apiSecret,
        region:    region,
        endpoint:  endpoint,
    }, nil
}

// ListInstances는 인스턴스 목록을 조회합니다
func (c *MyProviderClient) ListInstances(ctx context.Context) ([]*plugin.Instance, error) {
    // API 호출 구현
    // ...
    
    return []*plugin.Instance{
        {
            ID:           "i-1234567890abcdef0",
            Name:         "my-instance",
            Status:       "running",
            InstanceType: "t3.micro",
            ImageID:      "ami-12345678",
            Region:       c.region,
            Zone:         "us-east-1a",
            PublicIP:     "1.2.3.4",
            PrivateIP:    "10.0.0.1",
            Tags: map[string]string{
                "Name": "my-instance",
                "Env":  "production",
            },
            CreatedAt: time.Now().Add(-24 * time.Hour),
            UpdatedAt: time.Now(),
        },
    }, nil
}

// CreateInstance는 인스턴스를 생성합니다
func (c *MyProviderClient) CreateInstance(ctx context.Context, req plugin.CreateInstanceRequest) (*plugin.Instance, error) {
    // API 호출 구현
    // ...
    
    return &plugin.Instance{
        ID:           "i-1234567890abcdef0",
        Name:         req.Name,
        Status:       "pending",
        InstanceType: req.InstanceType,
        ImageID:      req.ImageID,
        Region:       req.Region,
        Zone:         req.Zone,
        Tags:         req.Tags,
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }, nil
}

// GetInstance는 인스턴스를 조회합니다
func (c *MyProviderClient) GetInstance(ctx context.Context, id string) (*plugin.Instance, error) {
    // API 호출 구현
    // ...
    
    return &plugin.Instance{
        ID:           id,
        Name:         "my-instance",
        Status:       "running",
        InstanceType: "t3.micro",
        ImageID:      "ami-12345678",
        Region:       c.region,
        Zone:         "us-east-1a",
        PublicIP:     "1.2.3.4",
        PrivateIP:    "10.0.0.1",
        Tags: map[string]string{
            "Name": "my-instance",
            "Env":  "production",
        },
        CreatedAt: time.Now().Add(-24 * time.Hour),
        UpdatedAt: time.Now(),
    }, nil
}

// UpdateInstance는 인스턴스를 업데이트합니다
func (c *MyProviderClient) UpdateInstance(ctx context.Context, id string, req plugin.UpdateInstanceRequest) (*plugin.Instance, error) {
    // API 호출 구현
    // ...
    
    return &plugin.Instance{
        ID:           id,
        Name:         req.Name,
        Status:       "running",
        InstanceType: req.InstanceType,
        ImageID:      "ami-12345678",
        Region:       c.region,
        Zone:         "us-east-1a",
        Tags:         req.Tags,
        CreatedAt:    time.Now().Add(-24 * time.Hour),
        UpdatedAt:    time.Now(),
    }, nil
}

// DeleteInstance는 인스턴스를 삭제합니다
func (c *MyProviderClient) DeleteInstance(ctx context.Context, id string) error {
    // API 호출 구현
    // ...
    
    return nil
}

// StartInstance는 인스턴스를 시작합니다
func (c *MyProviderClient) StartInstance(ctx context.Context, id string) error {
    // API 호출 구현
    // ...
    
    return nil
}

// StopInstance는 인스턴스를 중지합니다
func (c *MyProviderClient) StopInstance(ctx context.Context, id string) error {
    // API 호출 구현
    // ...
    
    return nil
}

// RestartInstance는 인스턴스를 재시작합니다
func (c *MyProviderClient) RestartInstance(ctx context.Context, id string) error {
    // API 호출 구현
    // ...
    
    return nil
}

// ListRegions은 리전 목록을 조회합니다
func (c *MyProviderClient) ListRegions(ctx context.Context) ([]*plugin.Region, error) {
    // API 호출 구현
    // ...
    
    return []*plugin.Region{
        {
            ID:          "us-east-1",
            Name:        "US East (N. Virginia)",
            DisplayName: "US East (N. Virginia)",
            Endpoint:    "https://api.myprovider.com/us-east-1",
            IsActive:    true,
        },
        {
            ID:          "us-west-2",
            Name:        "US West (Oregon)",
            DisplayName: "US West (Oregon)",
            Endpoint:    "https://api.myprovider.com/us-west-2",
            IsActive:    true,
        },
    }, nil
}

// GetRegion은 리전을 조회합니다
func (c *MyProviderClient) GetRegion(ctx context.Context, id string) (*plugin.Region, error) {
    // API 호출 구현
    // ...
    
    return &plugin.Region{
        ID:          id,
        Name:        "US East (N. Virginia)",
        DisplayName: "US East (N. Virginia)",
        Endpoint:    "https://api.myprovider.com/us-east-1",
        IsActive:    true,
    }, nil
}

// ListImages는 이미지 목록을 조회합니다
func (c *MyProviderClient) ListImages(ctx context.Context, req plugin.ListImagesRequest) ([]*plugin.Image, error) {
    // API 호출 구현
    // ...
    
    return []*plugin.Image{
        {
            ID:           "ami-12345678",
            Name:         "Ubuntu 20.04 LTS",
            Description:  "Ubuntu 20.04 LTS Server",
            OS:           "ubuntu",
            Architecture: "x86_64",
            Size:         8589934592, // 8GB
            Tags: map[string]string{
                "OS":           "ubuntu",
                "Version":      "20.04",
                "Architecture": "x86_64",
            },
            CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
        },
    }, nil
}

// GetImage는 이미지를 조회합니다
func (c *MyProviderClient) GetImage(ctx context.Context, id string) (*plugin.Image, error) {
    // API 호출 구현
    // ...
    
    return &plugin.Image{
        ID:           id,
        Name:         "Ubuntu 20.04 LTS",
        Description:  "Ubuntu 20.04 LTS Server",
        OS:           "ubuntu",
        Architecture: "x86_64",
        Size:         8589934592, // 8GB
        Tags: map[string]string{
            "OS":           "ubuntu",
            "Version":      "20.04",
            "Architecture": "x86_64",
        },
        CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
    }, nil
}

// ListInstanceTypes는 인스턴스 타입 목록을 조회합니다
func (c *MyProviderClient) ListInstanceTypes(ctx context.Context, req plugin.ListInstanceTypesRequest) ([]*plugin.InstanceType, error) {
    // API 호출 구현
    // ...
    
    return []*plugin.InstanceType{
        {
            ID:          "t3.micro",
            Name:        "t3.micro",
            VCPUs:       2,
            Memory:      1024, // 1GB
            Storage:     0,
            Network:     1000, // 1Gbps
            Price:       0.0104,
            Currency:    "USD",
            IsAvailable: true,
        },
        {
            ID:          "t3.small",
            Name:        "t3.small",
            VCPUs:       2,
            Memory:      2048, // 2GB
            Storage:     0,
            Network:     1000, // 1Gbps
            Price:       0.0208,
            Currency:    "USD",
            IsAvailable: true,
        },
    }, nil
}

// GetInstanceType은 인스턴스 타입을 조회합니다
func (c *MyProviderClient) GetInstanceType(ctx context.Context, id string) (*plugin.InstanceType, error) {
    // API 호출 구현
    // ...
    
    return &plugin.InstanceType{
        ID:          id,
        Name:        id,
        VCPUs:       2,
        Memory:      1024, // 1GB
        Storage:     0,
        Network:     1000, // 1Gbps
        Price:       0.0104,
        Currency:    "USD",
        IsAvailable: true,
    }, nil
}

// GetCostEstimate는 비용을 추정합니다
func (c *MyProviderClient) GetCostEstimate(ctx context.Context, req plugin.CostEstimateRequest) (*plugin.CostEstimate, error) {
    // API 호출 구현
    // ...
    
    return &plugin.CostEstimate{
        InstanceType: req.InstanceType,
        Region:       req.Region,
        Duration:     req.Duration,
        Cost:         0.0104, // 시간당 비용
        Currency:     "USD",
        Breakdown: []plugin.CostItem{
            {
                Service: "Compute",
                Cost:    0.0104,
                Unit:    "per hour",
            },
        },
    }, nil
}

// GetBillingInfo는 청구 정보를 조회합니다
func (c *MyProviderClient) GetBillingInfo(ctx context.Context, req plugin.BillingInfoRequest) (*plugin.BillingInfo, error) {
    // API 호출 구현
    // ...
    
    return &plugin.BillingInfo{
        TotalCost: 1250.50,
        Currency:  "USD",
        Period:    "30d",
        Breakdown: []plugin.CostItem{
            {
                Service: "Compute",
                Cost:    1000.00,
                Unit:    "total",
            },
            {
                Service: "Storage",
                Cost:    250.50,
                Unit:    "total",
            },
        },
    }, nil
}

// ListNetworks는 네트워크 목록을 조회합니다
func (c *MyProviderClient) ListNetworks(ctx context.Context) ([]*plugin.Network, error) {
    // API 호출 구현
    // ...
    
    return []*plugin.Network{
        {
            ID:          "vpc-12345678",
            Name:        "default-vpc",
            CIDR:        "10.0.0.0/16",
            Region:      c.region,
            IsDefault:   true,
            IsActive:    true,
            CreatedAt:   time.Now().Add(-365 * 24 * time.Hour),
        },
    }, nil
}

// GetNetwork은 네트워크를 조회합니다
func (c *MyProviderClient) GetNetwork(ctx context.Context, id string) (*plugin.Network, error) {
    // API 호출 구현
    // ...
    
    return &plugin.Network{
        ID:        id,
        Name:      "default-vpc",
        CIDR:      "10.0.0.0/16",
        Region:    c.region,
        IsDefault: true,
        IsActive:  true,
        CreatedAt: time.Now().Add(-365 * 24 * time.Hour),
    }, nil
}

// CreateNetwork은 네트워크를 생성합니다
func (c *MyProviderClient) CreateNetwork(ctx context.Context, req plugin.CreateNetworkRequest) (*plugin.Network, error) {
    // API 호출 구현
    // ...
    
    return &plugin.Network{
        ID:        "vpc-12345678",
        Name:      req.Name,
        CIDR:      req.CIDR,
        Region:    c.region,
        IsDefault: false,
        IsActive:  true,
        CreatedAt: time.Now(),
    }, nil
}

// DeleteNetwork은 네트워크를 삭제합니다
func (c *MyProviderClient) DeleteNetwork(ctx context.Context, id string) error {
    // API 호출 구현
    // ...
    
    return nil
}

// ListVolumes은 볼륨 목록을 조회합니다
func (c *MyProviderClient) ListVolumes(ctx context.Context) ([]*plugin.Volume, error) {
    // API 호출 구현
    // ...
    
    return []*plugin.Volume{
        {
            ID:         "vol-12345678",
            Name:       "my-volume",
            Size:       20, // 20GB
            Type:       "gp2",
            Status:     "available",
            Region:     c.region,
            Zone:       "us-east-1a",
            IsEncrypted: false,
            CreatedAt:  time.Now().Add(-7 * 24 * time.Hour),
        },
    }, nil
}

// GetVolume은 볼륨을 조회합니다
func (c *MyProviderClient) GetVolume(ctx context.Context, id string) (*plugin.Volume, error) {
    // API 호출 구현
    // ...
    
    return &plugin.Volume{
        ID:         id,
        Name:       "my-volume",
        Size:       20, // 20GB
        Type:       "gp2",
        Status:     "available",
        Region:     c.region,
        Zone:       "us-east-1a",
        IsEncrypted: false,
        CreatedAt:  time.Now().Add(-7 * 24 * time.Hour),
    }, nil
}

// CreateVolume은 볼륨을 생성합니다
func (c *MyProviderClient) CreateVolume(ctx context.Context, req plugin.CreateVolumeRequest) (*plugin.Volume, error) {
    // API 호출 구현
    // ...
    
    return &plugin.Volume{
        ID:         "vol-12345678",
        Name:       req.Name,
        Size:       req.Size,
        Type:       req.Type,
        Status:     "creating",
        Region:     c.region,
        Zone:       req.Zone,
        IsEncrypted: req.IsEncrypted,
        CreatedAt:  time.Now(),
    }, nil
}

// DeleteVolume은 볼륨을 삭제합니다
func (c *MyProviderClient) DeleteVolume(ctx context.Context, id string) error {
    // API 호출 구현
    // ...
    
    return nil
}

// AttachVolume은 볼륨을 연결합니다
func (c *MyProviderClient) AttachVolume(ctx context.Context, volumeID, instanceID string) error {
    // API 호출 구현
    // ...
    
    return nil
}

// DetachVolume은 볼륨을 분리합니다
func (c *MyProviderClient) DetachVolume(ctx context.Context, volumeID, instanceID string) error {
    // API 호출 구현
    // ...
    
    return nil
}

// ListSecurityGroups은 보안 그룹 목록을 조회합니다
func (c *MyProviderClient) ListSecurityGroups(ctx context.Context) ([]*plugin.SecurityGroup, error) {
    // API 호출 구현
    // ...
    
    return []*plugin.SecurityGroup{
        {
            ID:          "sg-12345678",
            Name:        "default",
            Description: "Default security group",
            Region:      c.region,
            IsDefault:   true,
            IsActive:    true,
            CreatedAt:   time.Now().Add(-365 * 24 * time.Hour),
        },
    }, nil
}

// GetSecurityGroup은 보안 그룹을 조회합니다
func (c *MyProviderClient) GetSecurityGroup(ctx context.Context, id string) (*plugin.SecurityGroup, error) {
    // API 호출 구현
    // ...
    
    return &plugin.SecurityGroup{
        ID:          id,
        Name:        "default",
        Description: "Default security group",
        Region:      c.region,
        IsDefault:   true,
        IsActive:    true,
        CreatedAt:   time.Now().Add(-365 * 24 * time.Hour),
    }, nil
}

// CreateSecurityGroup은 보안 그룹을 생성합니다
func (c *MyProviderClient) CreateSecurityGroup(ctx context.Context, req plugin.CreateSecurityGroupRequest) (*plugin.SecurityGroup, error) {
    // API 호출 구현
    // ...
    
    return &plugin.SecurityGroup{
        ID:          "sg-12345678",
        Name:        req.Name,
        Description: req.Description,
        Region:      c.region,
        IsDefault:   false,
        IsActive:    true,
        CreatedAt:   time.Now(),
    }, nil
}

// DeleteSecurityGroup은 보안 그룹을 삭제합니다
func (c *MyProviderClient) DeleteSecurityGroup(ctx context.Context, id string) error {
    // API 호출 구현
    // ...
    
    return nil
}

// ListKeyPairs은 키 페어 목록을 조회합니다
func (c *MyProviderClient) ListKeyPairs(ctx context.Context) ([]*plugin.KeyPair, error) {
    // API 호출 구현
    // ...
    
    return []*plugin.KeyPair{
        {
            ID:        "kp-12345678",
            Name:      "my-key-pair",
            Fingerprint: "12:34:56:78:90:ab:cd:ef:12:34:56:78:90:ab:cd:ef",
            Region:    c.region,
            IsActive:  true,
            CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
        },
    }, nil
}

// GetKeyPair은 키 페어를 조회합니다
func (c *MyProviderClient) GetKeyPair(ctx context.Context, id string) (*plugin.KeyPair, error) {
    // API 호출 구현
    // ...
    
    return &plugin.KeyPair{
        ID:          id,
        Name:        "my-key-pair",
        Fingerprint: "12:34:56:78:90:ab:cd:ef:12:34:56:78:90:ab:cd:ef",
        Region:      c.region,
        IsActive:    true,
        CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
    }, nil
}

// CreateKeyPair은 키 페어를 생성합니다
func (c *MyProviderClient) CreateKeyPair(ctx context.Context, req plugin.CreateKeyPairRequest) (*plugin.KeyPair, error) {
    // API 호출 구현
    // ...
    
    return &plugin.KeyPair{
        ID:          "kp-12345678",
        Name:        req.Name,
        Fingerprint: "12:34:56:78:90:ab:cd:ef:12:34:56:78:90:ab:cd:ef",
        Region:      c.region,
        IsActive:    true,
        CreatedAt:   time.Now(),
    }, nil
}

// DeleteKeyPair은 키 페어를 삭제합니다
func (c *MyProviderClient) DeleteKeyPair(ctx context.Context, id string) error {
    // API 호출 구현
    // ...
    
    return nil
}

// HealthCheck은 헬스 체크를 수행합니다
func (c *MyProviderClient) HealthCheck(ctx context.Context) error {
    // API 호출 구현
    // ...
    
    return nil
}

// Close는 클라이언트를 닫습니다
func (c *MyProviderClient) Close() error {
    // 리소스 정리
    // ...
    
    return nil
}
```

### 4. 플러그인 빌드

```bash
# 플러그인 빌드
go build -buildmode=plugin -o my-provider.so main.go client.go

# 플러그인 테스트
go test ./...

# 플러그인 검증
go run cmd/server/main.go --validate-plugin my-provider.so
```

### 5. 플러그인 등록

```yaml
# config.yaml
plugins:
  directory: "plugins"
  auto_load: true
  timeout: 30s
  providers:
    my-provider:
      enabled: true
      config:
        api_key: "your-api-key"
        api_secret: "your-api-secret"
        region: "us-east-1"
        endpoint: "https://api.myprovider.com"
```

## 플러그인 테스트

### 1. 단위 테스트

```go
// main_test.go
package main

import (
    "context"
    "testing"
    
    "github.com/skyclust/skyclust/pkg/plugin"
    "github.com/stretchr/testify/assert"
)

func TestMyProvider_Initialize(t *testing.T) {
    provider := New()
    
    config := map[string]interface{}{
        "api_key":    "test-key",
        "api_secret": "test-secret",
        "region":     "us-east-1",
    }
    
    err := provider.Initialize(config)
    assert.NoError(t, err)
}

func TestMyProvider_ValidateConfig(t *testing.T) {
    provider := New()
    
    config := map[string]interface{}{
        "api_key":    "test-key",
        "api_secret": "test-secret",
        "region":     "us-east-1",
    }
    
    err := provider.ValidateConfig(config)
    assert.NoError(t, err)
}

func TestMyProvider_ListInstances(t *testing.T) {
    provider := New()
    
    config := map[string]interface{}{
        "api_key":    "test-key",
        "api_secret": "test-secret",
        "region":     "us-east-1",
    }
    
    err := provider.Initialize(config)
    assert.NoError(t, err)
    
    instances, err := provider.ListInstances(context.Background())
    assert.NoError(t, err)
    assert.NotNil(t, instances)
}
```

### 2. 통합 테스트

```go
// integration_test.go
package main

import (
    "context"
    "testing"
    
    "github.com/skyclust/skyclust/pkg/plugin"
    "github.com/stretchr/testify/assert"
)

func TestMyProvider_Integration(t *testing.T) {
    provider := New()
    
    config := map[string]interface{}{
        "api_key":    "test-key",
        "api_secret": "test-secret",
        "region":     "us-east-1",
    }
    
    err := provider.Initialize(config)
    assert.NoError(t, err)
    
    // 인스턴스 생성 테스트
    req := plugin.CreateInstanceRequest{
        Name:         "test-instance",
        InstanceType: "t3.micro",
        ImageID:      "ami-12345678",
        Region:       "us-east-1",
    }
    
    instance, err := provider.CreateInstance(context.Background(), req)
    assert.NoError(t, err)
    assert.NotNil(t, instance)
    assert.Equal(t, "test-instance", instance.Name)
    
    // 인스턴스 조회 테스트
    retrieved, err := provider.GetInstance(context.Background(), instance.ID)
    assert.NoError(t, err)
    assert.NotNil(t, retrieved)
    assert.Equal(t, instance.ID, retrieved.ID)
    
    // 인스턴스 삭제 테스트
    err = provider.DeleteInstance(context.Background(), instance.ID)
    assert.NoError(t, err)
}
```

### 3. 성능 테스트

```go
// benchmark_test.go
package main

import (
    "context"
    "testing"
    
    "github.com/skyclust/skyclust/pkg/plugin"
)

func BenchmarkMyProvider_ListInstances(b *testing.B) {
    provider := New()
    
    config := map[string]interface{}{
        "api_key":    "test-key",
        "api_secret": "test-secret",
        "region":     "us-east-1",
    }
    
    err := provider.Initialize(config)
    if err != nil {
        b.Fatal(err)
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _, err := provider.ListInstances(context.Background())
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## 플러그인 배포

### 1. 플러그인 패키징

```bash
# 플러그인 패키지 생성
tar -czf my-provider-plugin.tar.gz \
  my-provider.so \
  config.yaml \
  README.md

# 플러그인 서명
gpg --sign --armor my-provider-plugin.tar.gz
```

### 2. 플러그인 설치

```bash
# 플러그인 디렉토리에 복사
cp my-provider.so plugins/my-provider/

# 설정 파일 업데이트
echo "plugins:
  my-provider:
    enabled: true
    config:
      api_key: \"your-api-key\"
      api_secret: \"your-api-secret\"
      region: \"us-east-1\"" >> config.yaml
```

### 3. 플러그인 활성화

```bash
# 플러그인 활성화
go run cmd/server/main.go --enable-plugin my-provider

# 플러그인 상태 확인
go run cmd/server/main.go --list-plugins
```

## 플러그인 모범 사례

### 1. 에러 처리

```go
// 에러 처리 예시
func (p *MyProvider) ListInstances(ctx context.Context) ([]*plugin.Instance, error) {
    instances, err := p.client.ListInstances(ctx)
    if err != nil {
        // 구체적인 에러 메시지 제공
        return nil, fmt.Errorf("failed to list instances from MyProvider API: %w", err)
    }
    
    return instances, nil
}
```

### 2. 로깅

```go
// 로깅 예시
func (p *MyProvider) CreateInstance(ctx context.Context, req plugin.CreateInstanceRequest) (*plugin.Instance, error) {
    log.Info("Creating instance",
        zap.String("provider", p.GetName()),
        zap.String("name", req.Name),
        zap.String("instance_type", req.InstanceType),
        zap.String("region", req.Region),
    )
    
    instance, err := p.client.CreateInstance(ctx, req)
    if err != nil {
        log.Error("Failed to create instance",
            zap.String("provider", p.GetName()),
            zap.String("name", req.Name),
            zap.Error(err),
        )
        return nil, err
    }
    
    log.Info("Instance created successfully",
        zap.String("provider", p.GetName()),
        zap.String("instance_id", instance.ID),
        zap.String("name", instance.Name),
    )
    
    return instance, nil
}
```

### 3. 설정 검증

```go
// 설정 검증 예시
func (p *MyProvider) ValidateConfig(config map[string]interface{}) error {
    required := []string{"api_key", "api_secret", "region"}
    
    for _, field := range required {
        if _, exists := config[field]; !exists {
            return fmt.Errorf("required field %s is missing", field)
        }
    }
    
    // API 키 형식 검증
    apiKey, ok := config["api_key"].(string)
    if !ok || len(apiKey) < 10 {
        return fmt.Errorf("invalid api_key format")
    }
    
    // 리전 검증
    region, ok := config["region"].(string)
    if !ok {
        return fmt.Errorf("region must be a string")
    }
    
    validRegions := []string{"us-east-1", "us-west-2", "eu-west-1"}
    if !contains(validRegions, region) {
        return fmt.Errorf("invalid region: %s", region)
    }
    
    return nil
}
```

### 4. 리소스 정리

```go
// 리소스 정리 예시
func (p *MyProvider) Cleanup() error {
    if p.client != nil {
        err := p.client.Close()
        if err != nil {
            log.Error("Failed to close client",
                zap.String("provider", p.GetName()),
                zap.Error(err),
            )
            return err
        }
    }
    
    log.Info("Provider cleanup completed",
        zap.String("provider", p.GetName()),
    )
    
    return nil
}
```

## 문제 해결

### 1. 일반적인 문제

#### 플러그인 로드 실패
```bash
# 플러그인 로그 확인
tail -f /var/log/skyclust/plugin.log

# 플러그인 검증
go run cmd/server/main.go --validate-plugin my-provider.so
```

#### API 연결 실패
```bash
# 네트워크 연결 확인
curl -v https://api.myprovider.com/health

# 인증 정보 확인
echo $MYPROVIDER_API_KEY
echo $MYPROVIDER_API_SECRET
```

#### 설정 오류
```bash
# 설정 파일 검증
go run cmd/server/main.go --validate-config config.yaml

# 플러그인 설정 확인
go run cmd/server/main.go --list-plugins
```

### 2. 디버깅

```go
// 디버깅 로그 추가
func (p *MyProvider) ListInstances(ctx context.Context) ([]*plugin.Instance, error) {
    log.Debug("Listing instances",
        zap.String("provider", p.GetName()),
        zap.String("region", p.client.region),
    )
    
    instances, err := p.client.ListInstances(ctx)
    if err != nil {
        log.Error("Failed to list instances",
            zap.String("provider", p.GetName()),
            zap.Error(err),
        )
        return nil, err
    }
    
    log.Debug("Instances listed successfully",
        zap.String("provider", p.GetName()),
        zap.Int("count", len(instances)),
    )
    
    return instances, nil
}
```

### 3. 성능 최적화

```go
// 연결 풀 사용
func (c *MyProviderClient) ListInstances(ctx context.Context) ([]*plugin.Instance, error) {
    // HTTP 클라이언트 재사용
    if c.httpClient == nil {
        c.httpClient = &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 10,
                IdleConnTimeout:     90 * time.Second,
            },
        }
    }
    
    // API 호출 구현
    // ...
}
```

## 참고 자료

- [Go 플러그인 문서](https://golang.org/pkg/plugin/)
- [SkyClust 플러그인 인터페이스](https://github.com/skyclust/skyclust/tree/main/pkg/plugin)
- [AWS SDK for Go](https://github.com/aws/aws-sdk-go)
- [Google Cloud Go SDK](https://github.com/googleapis/google-cloud-go)
- [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go)

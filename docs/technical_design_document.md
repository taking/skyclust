# SkyClust 기술 설계 문서

## 📋 **목차**
1. [시스템 개요](#시스템-개요)
2. [아키텍처 설계](#아키텍처-설계)
3. [데이터 모델](#데이터-모델)
4. [API 설계](#api-설계)
5. [보안 설계](#보안-설계)
6. [성능 최적화](#성능-최적화)
7. [모니터링 및 로깅](#모니터링-및-로깅)
8. [배포 전략](#배포-전략)

---

## 🎯 **시스템 개요**

### **프로젝트 목표**
멀티 클라우드 환경에서 Kubernetes 클러스터와 네트워크 인프라를 통합 관리하는 플랫폼 구축

### **핵심 요구사항**
- **멀티 클라우드 지원**: AWS EKS, GCP GKE 통합 관리
- **네트워크 자동화**: VPC, 서브넷, 보안 그룹 자동 생성/관리
- **보안**: 암호화된 자격증명 관리 및 RBAC
- **확장성**: 새로운 클라우드 제공업체 쉽게 추가
- **일관성**: 클라우드별 차이점 추상화

---

## 🏗️ **아키텍처 설계**

### **Clean Architecture 적용**

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                       │
│  ┌──────────────────┐  ┌──────────────────┐                  │
│  │   HTTP Handlers  │  │   gRPC Handlers  │                  │
│  │   (Gin Router)   │  │   (Provider)     │                  │
│  └──────────────────┘  └──────────────────┘                  │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                        │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │ Kubernetes      │  │ Network         │  │ Credential  │  │
│  │ Service         │  │ Service         │  │ Service     │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                      Domain Layer                           │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │ Cluster         │  │ Network         │  │ User        │  │
│  │ Entity          │  │ Entity          │  │ Entity      │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                   Infrastructure Layer                      │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────┐  │
│  │ Cloud SDK       │  │ Database        │  │ External    │  │
│  │ Integration     │  │ Repository      │  │ Services    │  │
│  └─────────────────┘  └─────────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### **서비스 레이어 설계**

#### **Kubernetes Service**
```go
type KubernetesService struct {
    credentialService domain.CredentialService
    logger            *zap.Logger
}

// Provider별 Dispatch 패턴
func (s *KubernetesService) CreateCluster(ctx context.Context, credential *domain.Credential, req dto.CreateClusterRequest) (*dto.CreateClusterResponse, error) {
    switch credential.Provider {
    case "aws":
        return s.createAWSEKSCluster(ctx, credential, req)
    case "gcp":
        return s.createGCPGKECluster(ctx, credential, req)
    case "azure":
        return s.createAzureAKSCluster(ctx, credential, req)
    default:
        return nil, fmt.Errorf("unsupported provider: %s", credential.Provider)
    }
}
```

#### **Network Service**
```go
type NetworkService struct {
    credentialService domain.CredentialService
    logger            *zap.Logger
}

// 통합 네트워크 관리
func (s *NetworkService) CreateVPC(ctx context.Context, credential *domain.Credential, req dto.CreateVPCRequest) (*dto.VPCInfo, error) {
    switch credential.Provider {
    case "aws":
        return s.createAWSVPC(ctx, credential, req)
    case "gcp":
        return s.createGCPVPC(ctx, credential, req)
    default:
        return nil, fmt.Errorf("unsupported provider: %s", credential.Provider)
    }
}
```

---

## 📊 **데이터 모델**

### **핵심 엔티티**

#### **User Entity**
```go
type User struct {
    ID        uuid.UUID `gorm:"type:uuid;primary_key"`
    Email     string    `gorm:"unique;not null"`
    Name      string    `gorm:"not null"`
    Role      string    `gorm:"not null"` // admin, user, viewer
    WorkspaceID uuid.UUID
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

#### **Credential Entity (Workspace 기반)**
```go
type Credential struct {
    ID            uuid.UUID  `gorm:"type:uuid;primary_key"`
    WorkspaceID    uuid.UUID  `gorm:"type:uuid;not null"`
    CreatedBy      uuid.UUID  `gorm:"type:uuid;not null"`
    Provider       string     `gorm:"not null"` // aws, gcp, azure, ncp
    Name           string     `gorm:"not null"`
    EncryptedData  []byte     `gorm:"type:bytea;not null"`
    IsActive       bool       `gorm:"default:true"`
    CreatedAt      time.Time
    UpdatedAt      time.Time
    DeletedAt      *time.Time `gorm:"index"`
}
```

#### **Cluster Entity**
```go
type Cluster struct {
    ID           uuid.UUID `gorm:"type:uuid;primary_key"`
    UserID       uuid.UUID `gorm:"not null"`
    CredentialID uuid.UUID `gorm:"not null"`
    Name         string    `gorm:"not null"`
    Provider     string    `gorm:"not null"`
    Region       string    `gorm:"not null"`
    Status       string    `gorm:"not null"`
    Config       []byte    `gorm:"type:jsonb"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### **DTO 설계**

#### **통합 클러스터 DTO**
```go
type CreateClusterRequest struct {
    CredentialID string            `json:"credential_id" validate:"required,uuid"`
    Name         string            `json:"name" validate:"required,min=1,max=255"`
    Version      string            `json:"version" validate:"required"`
    Region       string            `json:"region" validate:"required"`
    Tags         map[string]string `json:"tags,omitempty"`
    
    // Provider-specific configurations
    AWSConfig *AWSClusterConfig `json:"aws_config,omitempty"`
    GCPConfig *GCPClusterConfig `json:"gcp_config,omitempty"`
}

type ClusterInfo struct {
    ID        string            `json:"id"`
    Name      string            `json:"name"`
    Version   string            `json:"version"`
    Status    string            `json:"status"`
    Region    string            `json:"region"`
    Provider  string            `json:"provider"`
    Endpoint  string            `json:"endpoint,omitempty"`
    CreatedAt string            `json:"created_at,omitempty"`
    Tags      map[string]string `json:"tags,omitempty"`
}
```

#### **네트워크 DTO**
```go
type VPCInfo struct {
    ID                string            `json:"id"`
    Name              string            `json:"name"`
    CIDRBlock         string            `json:"cidr_block,omitempty"`
    State             string            `json:"state"`
    IsDefault         bool              `json:"is_default"`
    Region            string            `json:"region"`
    Tags              map[string]string `json:"tags,omitempty"`
    
    // Provider-specific fields
    NetworkMode       string            `json:"network_mode,omitempty"`       // GCP
    RoutingMode       string            `json:"routing_mode,omitempty"`       // GCP
    MTU               int64             `json:"mtu,omitempty"`                // GCP
    AutoSubnets       bool              `json:"auto_subnets,omitempty"`       // GCP
    FirewallRuleCount int32             `json:"firewall_rule_count,omitempty"` // GCP
}

type SubnetInfo struct {
    ID               string            `json:"id"`
    Name             string            `json:"name"`
    VPCID            string            `json:"vpc_id"`
    CIDRBlock        string            `json:"cidr_block"`
    AvailabilityZone string            `json:"availability_zone"`
    State            string            `json:"state"`
    IsPublic         bool              `json:"is_public"`
    Region           string            `json:"region"`
    Tags             map[string]string `json:"tags,omitempty"`
}
```

---

## 🔌 **API 설계**

### **RESTful API 구조**

#### **리소스 기반 URL 설계**
```
/api/v1/
├── auth/                           # 인증 관리
│   ├── login
│   ├── logout
│   └── refresh
├── credentials/                    # 자격증명 관리
│   ├── GET    /                   # 목록 조회
│   ├── POST   /                   # 생성
│   ├── GET    /:id                # 상세 조회
│   ├── PUT    /:id                # 수정
│   └── DELETE /:id                # 삭제
├── {provider}/                     # 클라우드별 리소스
│   ├── kubernetes/                # Kubernetes 관리
│   │   ├── clusters/              # 클러스터 관리
│   │   └── node-groups/           # 노드 그룹 관리
│   └── network/                   # 네트워크 관리
│       ├── vpcs/                  # VPC 관리
│       ├── subnets/              # 서브넷 관리
│       └── security-groups/       # 보안 그룹 관리
└── monitoring/                     # 모니터링
    ├── metrics/
    └── alerts/
```

#### **API 응답 표준화**
```go
type APIResponse struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data,omitempty"`
    Error     *APIError   `json:"error,omitempty"`
    Message   string      `json:"message,omitempty"`
    RequestID string      `json:"request_id"`
    Timestamp string      `json:"timestamp"`
}

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}
```

### **Provider별 Dispatch 패턴**

#### **Handler 레벨**
```go
func (h *KubernetesHandler) CreateCluster(c *gin.Context) {
    // Provider 추출
    provider := c.Param("provider")
    
    // Provider별 핸들러 호출
    switch provider {
    case "aws":
        h.CreateEKSCluster(c)
    case "gcp":
        h.CreateGKECluster(c)
    case "azure":
        h.CreateAKSCluster(c)
    default:
        responses.BadRequest(c, "Unsupported provider")
    }
}
```

#### **Service 레벨**
```go
func (s *KubernetesService) CreateCluster(ctx context.Context, credential *domain.Credential, req dto.CreateClusterRequest) (*dto.CreateClusterResponse, error) {
    // Provider별 구현체 호출
    switch credential.Provider {
    case "aws":
        return s.createAWSEKSCluster(ctx, credential, req)
    case "gcp":
        return s.createGCPGKECluster(ctx, credential, req)
    case "azure":
        return s.createAzureAKSCluster(ctx, credential, req)
    default:
        return nil, domain.NewDomainError(
            domain.ErrCodeNotSupported,
            fmt.Sprintf("Unsupported provider: %s", credential.Provider),
            400,
        )
    }
}
```

---

## 🔒 **보안 설계**

### **인증 및 인가**

#### **JWT 기반 인증**
```go
type Claims struct {
    UserID      uuid.UUID `json:"user_id"`
    Email       string    `json:"email"`
    Role        string    `json:"role"`
    WorkspaceID uuid.UUID `json:"workspace_id"`
    jwt.RegisteredClaims
}

func GenerateToken(user *domain.User) (string, error) {
    claims := Claims{
        UserID:      user.ID,
        Email:       user.Email,
        Role:        user.Role,
        WorkspaceID: user.WorkspaceID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
```

#### **RBAC 미들웨어**
```go
func RBACMiddleware(requiredRole string) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            responses.Unauthorized(c, "Authorization header required")
            c.Abort()
            return
        }
        
        claims, err := validateToken(token)
        if err != nil {
            responses.Unauthorized(c, "Invalid token")
            c.Abort()
            return
        }
        
        if !hasPermission(claims.Role, requiredRole) {
            responses.Forbidden(c, "Insufficient permissions")
            c.Abort()
            return
        }
        
        c.Set("user_id", claims.UserID)
        c.Set("workspace_id", claims.WorkspaceID)
        c.Next()
    }
}
```

### **자격증명 암호화**

#### **AES 암호화**
```go
type CredentialEncryption struct {
    key []byte
}

func (ce *CredentialEncryption) Encrypt(data []byte) ([]byte, error) {
    block, err := aes.NewCipher(ce.key)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }
    
    ciphertext := gcm.Seal(nonce, nonce, data, nil)
    return ciphertext, nil
}

func (ce *CredentialEncryption) Decrypt(data []byte) ([]byte, error) {
    block, err := aes.NewCipher(ce.key)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return nil, fmt.Errorf("ciphertext too short")
    }
    
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, err
    }
    
    return plaintext, nil
}
```

---

## ⚡ **성능 최적화**

### **데이터베이스 최적화**

#### **인덱스 설계**
```sql
-- 자주 조회되는 필드에 인덱스 생성
CREATE INDEX idx_credentials_user_provider ON credentials(user_id, provider);
CREATE INDEX idx_clusters_user_region ON clusters(user_id, region);
CREATE INDEX idx_audit_logs_user_time ON audit_logs(user_id, created_at);

-- 복합 인덱스
CREATE INDEX idx_clusters_user_provider_status ON clusters(user_id, provider, status);
```

#### **쿼리 최적화**
```go
// N+1 문제 해결을 위한 Preload 사용
func (r *ClusterRepository) GetClustersWithDetails(userID uuid.UUID) ([]*domain.Cluster, error) {
    var clusters []*domain.Cluster
    err := r.db.Preload("Credential").
        Preload("NodeGroups").
        Where("user_id = ?", userID).
        Find(&clusters).Error
    return clusters, err
}
```

### **캐싱 전략**

#### **Redis 캐싱**
```go
type CacheService struct {
    redis *redis.Client
}

func (cs *CacheService) GetClusterInfo(clusterID string) (*dto.ClusterInfo, error) {
    key := fmt.Sprintf("cluster:%s", clusterID)
    
    // 캐시에서 조회
    cached, err := cs.redis.Get(context.Background(), key).Result()
    if err == nil {
        var clusterInfo dto.ClusterInfo
        json.Unmarshal([]byte(cached), &clusterInfo)
        return &clusterInfo, nil
    }
    
    // 캐시 미스 시 DB에서 조회 후 캐시 저장
    clusterInfo, err := cs.fetchFromDatabase(clusterID)
    if err != nil {
        return nil, err
    }
    
    data, _ := json.Marshal(clusterInfo)
    cs.redis.Set(context.Background(), key, data, 5*time.Minute)
    
    return clusterInfo, nil
}
```

### **병렬 처리**

#### **Goroutine을 활용한 병렬 처리**
```go
func (s *KubernetesService) CreateClusterWithNetworking(ctx context.Context, credential *domain.Credential, req dto.CreateClusterRequest) (*dto.CreateClusterResponse, error) {
    // 병렬로 네트워크 리소스 생성
    var wg sync.WaitGroup
    var vpcErr, subnetErr error
    var vpcInfo *dto.VPCInfo
    var subnetInfo *dto.SubnetInfo
    
    wg.Add(2)
    
    // VPC 생성
    go func() {
        defer wg.Done()
        vpcReq := dto.CreateVPCRequest{
            CredentialID: req.CredentialID,
            Name:         req.Name + "-vpc",
            CIDRBlock:    "10.0.0.0/16",
        }
        vpcInfo, vpcErr = s.networkService.CreateVPC(ctx, credential, vpcReq)
    }()
    
    // 서브넷 생성
    go func() {
        defer wg.Done()
        subnetReq := dto.CreateSubnetRequest{
            CredentialID: req.CredentialID,
            Name:         req.Name + "-subnet",
            CIDRBlock:    "10.0.1.0/24",
        }
        subnetInfo, subnetErr = s.networkService.CreateSubnet(ctx, credential, subnetReq)
    }()
    
    wg.Wait()
    
    if vpcErr != nil || subnetErr != nil {
        return nil, fmt.Errorf("network creation failed: vpc=%v, subnet=%v", vpcErr, subnetErr)
    }
    
    // 클러스터 생성
    return s.CreateCluster(ctx, credential, req)
}
```

---

## 💰 **비용 분석 시스템**

### **비용 분석 아키텍처**

#### **CostAnalysisService 구조**
```go
type CostAnalysisService struct {
    vmRepo            domain.VMRepository
    credentialRepo    domain.CredentialRepository
    workspaceRepo     domain.WorkspaceRepository
    auditLogRepo      domain.AuditLogRepository
    credentialService domain.CredentialService
    kubernetesService *KubernetesService
}
```

### **지원 기능**

#### **비용 데이터 소스**
1. **AWS Cost Explorer API**: 실제 AWS 비용 데이터
2. **GCP Cloud Billing API**: 실제 GCP 비용 데이터
3. **추정 비용**: API 접근 불가 시 VM 사양 기반 추정

#### **리소스 타입 지원**
- **VM 비용**: EC2, Compute Engine 등
- **Kubernetes 클러스터 비용**: EKS, GKE
- **노드 그룹/풀 비용**: EKS 노드 그룹, GKE 노드 풀

#### **비용 분석 기능**
- **비용 요약**: 기간별 총 비용 및 프로바이더별 분석
- **비용 예측**: 선형 회귀를 사용한 미래 비용 예측
- **비용 트렌드**: 전반기 대비 후반기 변화율 분석
- **비용 세부 분석**: 서비스, 프로바이더, 리전별 세부 분석
- **비용 비교**: 현재 기간과 이전 기간 비교
- **예산 알림**: 예산 초과 및 경고 알림

#### **리소스 타입 필터링**
```go
// resource_types 쿼리 파라미터
- "all": 모든 리소스 (기본값)
- "vm": VM만
- "cluster": Kubernetes 클러스터만
- "vm,cluster": VM과 클러스터 함께
```

### **비용 계산 흐름**

#### **VM 비용 계산**
```go
func (s *CostAnalysisService) calculateVMCosts(ctx context.Context, vm *domain.VM, startDate, endDate time.Time) ([]CostData, error) {
    // 1. 워크스페이스의 자격증명 조회
    credentials, err := s.credentialRepo.GetByWorkspaceIDAndProvider(workspaceUUID, vm.Provider)
    
    // 2. 프로바이더별 API 호출
    switch vm.Provider {
    case "aws":
        return s.getAWSCosts(ctx, credential, vm, startDate, endDate)
    case "gcp":
        return s.getGCPCosts(ctx, credential, vm, startDate, endDate)
    default:
        // 3. API 실패 시 추정 비용 사용
        return s.calculateEstimatedCosts(vm, startDate, endDate)
    }
}
```

#### **Kubernetes 비용 계산**
```go
func (s *CostAnalysisService) calculateKubernetesCosts(ctx context.Context, workspaceID string, startDate, endDate time.Time, includeNodeGroups bool) ([]CostData, []CostWarning, error) {
    // 1. 워크스페이스의 모든 자격증명 조회
    allCredentials, err := s.credentialRepo.GetByWorkspaceID(workspaceUUID)
    
    // 2. 프로바이더별로 그룹화
    // 3. AWS: EKS 비용 (Cost Explorer API)
    // 4. GCP: GKE 비용 (Cloud Billing API, BigQuery Export 권장)
    // 5. 경고 정보 반환
}
```

### **경고 시스템**

#### **CostWarning 구조**
```go
type CostWarning struct {
    Code         string `json:"code"`    // API_PERMISSION_DENIED, API_NOT_ENABLED 등
    Message      string `json:"message"` // 사용자 친화적 메시지
    Provider     string `json:"provider,omitempty"`
    ResourceType string `json:"resource_type,omitempty"`
}
```

#### **주요 경고 코드**
- `API_PERMISSION_DENIED`: 클라우드 API 권한 부족
- `API_NOT_ENABLED`: 클라우드 API 미활성화
- `VM_COST_CALCULATION_FAILED`: VM 비용 계산 실패
- `KUBERNETES_COST_CALCULATION_FAILED`: Kubernetes 비용 계산 실패
- `CREDENTIAL_ERROR`: 자격증명 오류
- `GKE_COST_NOT_IMPLEMENTED`: GKE 비용 계산 미구현 (BigQuery Export 필요)

### **API 통합**

#### **AWS Cost Explorer**
```go
func (s *CostAnalysisService) getAWSCosts(ctx context.Context, credential *domain.Credential, vm *domain.VM, startDate, endDate time.Time) ([]CostData, error) {
    ceClient := costexplorer.NewFromConfig(cfg)
    
    input := &costexplorer.GetCostAndUsageInput{
        TimePeriod: &types.DateInterval{
            Start: aws.String(startDate.Format("2006-01-02")),
            End:   aws.String(endDate.Format("2006-01-02")),
        },
        Granularity: types.GranularityDaily,
        Metrics:     []string{"BlendedCost"},
        GroupBy:     []types.GroupDefinition{...},
        Filter:      &types.Expression{...},
    }
    
    result, err := ceClient.GetCostAndUsage(ctx, input)
    // 결과 파싱 및 반환
}
```

#### **GCP Cloud Billing**
```go
func (s *CostAnalysisService) getGCPCosts(ctx context.Context, credential *domain.Credential, vm *domain.VM, startDate, endDate time.Time) ([]CostData, error) {
    billingClient, err := billingv1.NewCloudBillingClient(ctx, option.WithCredentialsJSON(keyBytes))
    
    projectInfo, err := billingClient.GetProjectBillingInfo(ctx, &billingpb.GetProjectBillingInfoRequest{
        Name: fmt.Sprintf("projects/%s", projectID),
    })
    
    // GCP는 BigQuery Export를 통한 상세 비용 조회 권장
}
```

### **예산 알림**

#### **예산 알림 로직**
```go
func (s *CostAnalysisService) CheckBudgetAlerts(ctx context.Context, workspaceID string, budgetLimit float64) ([]BudgetAlert, error) {
    summary, err := s.GetCostSummary(ctx, workspaceID, "1m", "all")
    
    percentage := (summary.TotalCost / budgetLimit) * 100
    
    if percentage >= 100 {
        // Critical: 예산 초과
    } else if percentage >= 80 {
        // Warning: 예산 80% 이상
    }
}
```

---

## 📊 **모니터링 및 로깅**

### **구조화된 로깅**

#### **Zap 로거 설정**
```go
func NewLogger() *zap.Logger {
    config := zap.NewProductionConfig()
    config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    
    logger, _ := config.Build()
    return logger
}

// 서비스에서 로깅 사용
func (s *KubernetesService) CreateCluster(ctx context.Context, credential *domain.Credential, req dto.CreateClusterRequest) (*dto.CreateClusterResponse, error) {
    s.logger.Info("Creating cluster",
        zap.String("provider", credential.Provider),
        zap.String("cluster_name", req.Name),
        zap.String("region", req.Region),
        zap.String("user_id", credential.UserID.String()))
    
    // 클러스터 생성 로직...
    
    s.logger.Info("Cluster created successfully",
        zap.String("cluster_id", response.ID),
        zap.String("provider", credential.Provider),
        zap.Duration("duration", time.Since(start)))
    
    return response, nil
}
```

### **메트릭 수집**

#### **Prometheus 메트릭**
```go
var (
    clusterCreationDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "cluster_creation_duration_seconds",
            Help: "Time taken to create a cluster",
        },
        []string{"provider", "region"},
    )
    
    activeClusters = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "active_clusters_total",
            Help: "Number of active clusters",
        },
        []string{"provider", "region"},
    )
)

func init() {
    prometheus.MustRegister(clusterCreationDuration)
    prometheus.MustRegister(activeClusters)
}
```

### **감사 로깅**

#### **감사 로그 구조**
```go
type AuditLog struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key"`
    UserID      uuid.UUID `gorm:"not null"`
    Action      string    `gorm:"not null"` // CREATE, UPDATE, DELETE
    Resource    string    `gorm:"not null"` // cluster, vpc, subnet
    ResourceID  string    `gorm:"not null"`
    Provider    string    `gorm:"not null"`
    Details     []byte    `gorm:"type:jsonb"`
    IPAddress   string
    UserAgent   string
    CreatedAt   time.Time
}

func (s *AuditService) LogAction(ctx context.Context, userID uuid.UUID, action, resource, resourceID, provider string, details interface{}) error {
    log := AuditLog{
        ID:         uuid.New(),
        UserID:     userID,
        Action:     action,
        Resource:   resource,
        ResourceID: resourceID,
        Provider:   provider,
        IPAddress:  getClientIP(ctx),
        UserAgent:  getUserAgent(ctx),
        CreatedAt:  time.Now(),
    }
    
    if details != nil {
        detailsJSON, _ := json.Marshal(details)
        log.Details = detailsJSON
    }
    
    return s.db.Create(&log).Error
}
```

---

## 🚀 **배포 전략**

### **컨테이너화**

#### **Dockerfile**
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

EXPOSE 8080
CMD ["./main"]
```

### **Kubernetes 배포**

#### **Deployment YAML**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: skyclust-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: skyclust-api
  template:
    metadata:
      labels:
        app: skyclust-api
    spec:
      containers:
      - name: skyclust-api
        image: skyclust/api:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: skyclust-secrets
              key: database-url
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: skyclust-secrets
              key: jwt-secret
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### **CI/CD 파이프라인**

#### **GitHub Actions**
```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.21
    - run: go test ./...
    - run: go vet ./...
    - run: golangci-lint run

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Build Docker image
      run: docker build -t skyclust/api:${{ github.sha }} .
    - name: Push to registry
      run: docker push skyclust/api:${{ github.sha }}

  deploy:
    needs: build
    runs-on: ubuntu-latest
    steps:
    - name: Deploy to Kubernetes
      run: |
        kubectl set image deployment/skyclust-api skyclust-api=skyclust/api:${{ github.sha }}
        kubectl rollout status deployment/skyclust-api
```

---

## 📈 **성능 벤치마크**

### **예상 성능 지표**

| 메트릭 | 목표값 | 측정 방법 |
|--------|--------|-----------|
| **API 응답 시간** | < 200ms | 95th percentile |
| **클러스터 생성 시간** | < 5분 | AWS EKS, GCP GKE 평균 |
| **동시 사용자** | 1000+ | 부하 테스트 |
| **데이터베이스 쿼리** | < 50ms | 복잡한 조인 쿼리 |
| **메모리 사용량** | < 512MB | 컨테이너당 |

### **부하 테스트**

#### **K6 부하 테스트 스크립트**
```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '2m', target: 100 },
    { duration: '5m', target: 100 },
    { duration: '2m', target: 200 },
    { duration: '5m', target: 200 },
    { duration: '2m', target: 0 },
  ],
};

export default function() {
  let response = http.get('http://localhost:8080/api/v1/aws/kubernetes/clusters', {
    headers: {
      'Authorization': 'Bearer ' + __ENV.JWT_TOKEN,
    },
  });
  
  check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 200ms': (r) => r.timings.duration < 200,
  });
  
  sleep(1);
}
```

---

## 🔮 **향후 확장 계획**

### **단기 계획 (3개월)**
- ✅ Azure AKS 지원 추가
- ✅ 실시간 모니터링 대시보드
- ✅ 비용 최적화 추천 엔진
- ✅ GitOps 통합

### **중기 계획 (6개월)**
- 🔄 Kubernetes 네이티브 오퍼레이터
- 🔄 멀티 리전 클러스터 관리
- 🔄 자동 스케일링 정책
- 🔄 보안 스캐닝 통합

### **장기 계획 (1년)**
- 🔄 AI 기반 자동 스케일링
- 🔄 하이브리드 클라우드 네이티브 애플리케이션
- 🔄 글로벌 멀티 클라우드 오케스트레이션
- 🔄 Edge Computing 지원

---

## 📝 **결론**

SkyClust는 **Clean Architecture**와 **멀티 클라우드 통합**을 통해 다음과 같은 가치를 제공합니다:

### **기술적 가치**
- **확장성**: 새로운 클라우드 제공업체 쉽게 추가
- **유지보수성**: Clean Architecture로 코드 품질 향상
- **성능**: 최적화된 데이터베이스 쿼리 및 캐싱
- **보안**: 엔터프라이즈급 보안 및 감사 기능

### **비즈니스 가치**
- **비용 절감**: 멀티 클라우드 비용 최적화
- **운영 효율성**: 통합된 관리 인터페이스
- **개발 생산성**: 자동화된 인프라 관리
- **위험 감소**: 클라우드 벤더 락인 방지

이를 통해 개발팀은 **클라우드 복잡성에서 해방**되어 **애플리케이션 개발에 집중**할 수 있습니다.

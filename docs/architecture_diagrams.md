# SkyClust 아키텍처 다이어그램

## 전체 시스템 아키텍처

```mermaid
graph TB
    subgraph "Frontend Layer"
        UI[Next.js Dashboard]
        API_CLIENT[API Client]
    end
    
    subgraph "API Gateway"
        GIN[Gin HTTP Server]
        MIDDLEWARE[Auth & RBAC Middleware]
    end
    
    subgraph "Service Layer"
        K8S_SVC[Kubernetes Service]
        NET_SVC[Network Service]
        CRED_SVC[Credential Service]
        COST_SVC[Cost Analysis Service]
        USER_SVC[User Service]
        WORKSPACE_SVC[Workspace Service]
        NOTIFICATION_SVC[Notification Service]
        OIDC_SVC[OIDC Service]
    end
    
    subgraph "Cloud Providers"
        subgraph "AWS"
            EKS[AWS EKS]
            VPC[AWS VPC]
            EC2[AWS EC2]
            COST_EXPLORER[AWS Cost Explorer]
        end
        
        subgraph "GCP"
            GKE[GCP GKE]
            GCP_VPC[GCP VPC]
            COMPUTE[GCP Compute]
            BILLING[GCP Cloud Billing]
        end
        
        subgraph "Azure"
            AKS[Azure AKS]
            VNET[Azure VNet]
        end
    end
    
    subgraph "Data Layer"
        POSTGRES[(PostgreSQL)]
        REDIS[(Redis Cache)]
    end
    
    UI --> API_CLIENT
    API_CLIENT --> GIN
    GIN --> MIDDLEWARE
    MIDDLEWARE --> K8S_SVC
    MIDDLEWARE --> NET_SVC
    MIDDLEWARE --> CRED_SVC
    MIDDLEWARE --> COST_SVC
    MIDDLEWARE --> USER_SVC
    MIDDLEWARE --> WORKSPACE_SVC
    MIDDLEWARE --> NOTIFICATION_SVC
    MIDDLEWARE --> OIDC_SVC
    
    K8S_SVC --> EKS
    K8S_SVC --> GKE
    K8S_SVC --> AKS
    
    NET_SVC --> VPC
    NET_SVC --> GCP_VPC
    NET_SVC --> VNET
    
    COST_SVC --> COST_EXPLORER
    COST_SVC --> BILLING
    COST_SVC --> EC2
    COST_SVC --> COMPUTE
    
    CRED_SVC --> POSTGRES
    USER_SVC --> POSTGRES
    WORKSPACE_SVC --> POSTGRES
    K8S_SVC --> REDIS
    NET_SVC --> REDIS
```

## Kubernetes 클러스터 생성 워크플로우

```mermaid
sequenceDiagram
    participant User
    participant API
    participant K8S_SVC
    participant NET_SVC
    participant AWS
    participant GCP
    
    User->>API: POST /clusters
    API->>K8S_SVC: CreateCluster()
    
    alt AWS EKS
        K8S_SVC->>NET_SVC: CreateVPC()
        NET_SVC->>AWS: Create VPC
        AWS-->>NET_SVC: VPC Created
        K8S_SVC->>NET_SVC: CreateSubnet()
        NET_SVC->>AWS: Create Subnet
        AWS-->>NET_SVC: Subnet Created
        K8S_SVC->>AWS: Create EKS Cluster
        AWS-->>K8S_SVC: Cluster Created
        K8S_SVC->>AWS: Create Node Group
        AWS-->>K8S_SVC: Node Group Created
    else GCP GKE
        K8S_SVC->>NET_SVC: CreateVPC()
        NET_SVC->>GCP: Create VPC
        GCP-->>NET_SVC: VPC Created
        K8S_SVC->>NET_SVC: CreateSubnet()
        NET_SVC->>GCP: Create Subnet
        GCP-->>NET_SVC: Subnet Created
        K8S_SVC->>GCP: Create GKE Cluster
        GCP-->>K8S_SVC: Cluster Created
        K8S_SVC->>GCP: Create Node Pool
        GCP-->>K8S_SVC: Node Pool Created
    end
    
    K8S_SVC-->>API: Cluster Created
    API-->>User: Success Response
```

## 비용 분석 아키텍처

```mermaid
graph LR
    subgraph "Cost Analysis Service"
        COST_SVC[Cost Analysis Service]
        VM_COST[VM Cost Calculator]
        K8S_COST[Kubernetes Cost Calculator]
        PREDICT[Cost Predictor]
    end
    
    subgraph "Cloud APIs"
        AWS_COST[AWS Cost Explorer API]
        GCP_BILLING[GCP Cloud Billing API]
    end
    
    subgraph "Data Sources"
        VM_REPO[VM Repository]
        K8S_SVC[Kubernetes Service]
        CRED_SVC[Credential Service]
    end
    
    COST_SVC --> VM_COST
    COST_SVC --> K8S_COST
    COST_SVC --> PREDICT
    
    VM_COST --> AWS_COST
    VM_COST --> GCP_BILLING
    K8S_COST --> AWS_COST
    K8S_COST --> GCP_BILLING
    
    VM_COST --> VM_REPO
    K8S_COST --> K8S_SVC
    VM_COST --> CRED_SVC
    K8S_COST --> CRED_SVC
```

## 워크스페이스 기반 멀티 테넌트 아키텍처

```mermaid
graph TB
    subgraph "Users"
        USER1[User 1]
        USER2[User 2]
        USER3[User 3]
    end
    
    subgraph "Workspaces"
        WS1[Workspace 1]
        WS2[Workspace 2]
        WS3[Workspace 3]
    end
    
    subgraph "Resources"
        subgraph "Workspace 1 Resources"
            CRED1[Credentials 1]
            VM1[VMs 1]
            K8S1[K8s Clusters 1]
        end
        
        subgraph "Workspace 2 Resources"
            CRED2[Credentials 2]
            VM2[VMs 2]
            K8S2[K8s Clusters 2]
        end
        
        subgraph "Workspace 3 Resources"
            CRED3[Credentials 3]
            VM3[VMs 3]
            K8S3[K8s Clusters 3]
        end
    end
    
    USER1 --> WS1
    USER2 --> WS2
    USER3 --> WS3
    
    WS1 --> CRED1
    WS1 --> VM1
    WS1 --> K8S1
    
    WS2 --> CRED2
    WS2 --> VM2
    WS2 --> K8S2
    
    WS3 --> CRED3
    WS3 --> VM3
    WS3 --> K8S3
```

## 멀티 클라우드 네트워크 관리

```mermaid
graph LR
    subgraph "SkyClust Network Service"
        NS[Network Service]
        VPC_MGR[VPC Manager]
        SUBNET_MGR[Subnet Manager]
        SEC_MGR[Security Manager]
    end
    
    subgraph "AWS Network"
        AWS_VPC[AWS VPC]
        AWS_SUBNET[AWS Subnet]
        AWS_SG[AWS Security Group]
    end
    
    subgraph "GCP Network"
        GCP_VPC[GCP VPC]
        GCP_SUBNET[GCP Subnet]
        GCP_FW[GCP Firewall]
    end
    
    NS --> VPC_MGR
    NS --> SUBNET_MGR
    NS --> SEC_MGR
    
    VPC_MGR --> AWS_VPC
    VPC_MGR --> GCP_VPC
    
    SUBNET_MGR --> AWS_SUBNET
    SUBNET_MGR --> GCP_SUBNET
    
    SEC_MGR --> AWS_SG
    SEC_MGR --> GCP_FW
```

## 데이터 플로우 아키텍처

```mermaid
graph TD
    subgraph "Request Flow"
        REQ[HTTP Request]
        AUTH[Authentication]
        RBAC[Authorization]
        VALIDATE[Validation]
    end
    
    subgraph "Processing Layer"
        SERVICE[Service Layer]
        PROVIDER[Provider Dispatch]
        CLOUD[Cloud SDK]
    end
    
    subgraph "Response Flow"
        RESPONSE[Response]
        LOG[Audit Log]
        CACHE[Cache Update]
    end
    
    REQ --> AUTH
    AUTH --> RBAC
    RBAC --> VALIDATE
    VALIDATE --> SERVICE
    SERVICE --> PROVIDER
    PROVIDER --> CLOUD
    CLOUD --> RESPONSE
    RESPONSE --> LOG
    RESPONSE --> CACHE
```

## 보안 아키텍처

```mermaid
graph TB
    subgraph "Security Layers"
        JWT[JWT Authentication]
        RBAC[Role-Based Access Control]
        ENCRYPT[Data Encryption]
        AUDIT[Audit Logging]
    end
    
    subgraph "Credential Management"
        STORE[Encrypted Storage<br/>Workspace-based]
        ROTATE[Credential Rotation]
        MASK[Masked Display]
    end
    
    subgraph "Network Security"
        TLS[TLS Encryption]
        FIREWALL[Firewall Rules]
        VPN[VPN Access]
    end
    
    JWT --> RBAC
    RBAC --> ENCRYPT
    ENCRYPT --> AUDIT
    
    STORE --> ROTATE
    ROTATE --> MASK
    
    TLS --> FIREWALL
    FIREWALL --> VPN
```

## Clean Architecture 레이어 구조

```mermaid
graph TB
    subgraph "Presentation Layer"
        HTTP[HTTP Handlers]
        SSE[SSE Handlers]
    end
    
    subgraph "Application Layer"
        SERVICES[Services<br/>Business Logic]
        DTO[DTOs]
    end
    
    subgraph "Domain Layer"
        ENTITIES[Domain Entities]
        INTERFACES[Interfaces]
        EVENTS[Domain Events]
    end
    
    subgraph "Infrastructure Layer"
        REPO[Repositories]
        EXTERNAL[External APIs]
        DB[(PostgreSQL)]
        CACHE[(Redis)]
    end
    
    HTTP --> SERVICES
    SSE --> SERVICES
    SERVICES --> ENTITIES
    SERVICES --> INTERFACES
    SERVICES --> DTO
    INTERFACES --> REPO
    REPO --> DB
    REPO --> CACHE
    SERVICES --> EXTERNAL
```

## 확장성 아키텍처

```mermaid
graph TB
    subgraph "Load Balancer"
        LB[Load Balancer]
    end
    
    subgraph "Application Instances"
        APP1[App Instance 1]
        APP2[App Instance 2]
        APP3[App Instance 3]
    end
    
    subgraph "Database Cluster"
        MASTER[(Master DB)]
        SLAVE1[(Slave DB 1)]
        SLAVE2[(Slave DB 2)]
    end
    
    subgraph "Cache Cluster"
        REDIS1[(Redis 1)]
        REDIS2[(Redis 2)]
        REDIS3[(Redis 3)]
    end
    
    LB --> APP1
    LB --> APP2
    LB --> APP3
    
    APP1 --> MASTER
    APP2 --> MASTER
    APP3 --> MASTER
    
    MASTER --> SLAVE1
    MASTER --> SLAVE2
    
    APP1 --> REDIS1
    APP2 --> REDIS2
    APP3 --> REDIS3
```

## 실시간 모니터링 아키텍처

```mermaid
graph LR
    subgraph "Clients"
        CLIENT1[Client 1]
        CLIENT2[Client 2]
        CLIENT3[Client 3]
    end
    
    subgraph "SSE Service"
        SSE_SVC[SSE Service]
        MONITORING[Monitoring Stream]
        NOTIFICATIONS[Notifications Stream]
    end
    
    subgraph "Event Sources"
        K8S_EVENTS[K8s Events]
        VM_EVENTS[VM Events]
        COST_EVENTS[Cost Events]
        SYSTEM_EVENTS[System Events]
    end
    
    CLIENT1 --> SSE_SVC
    CLIENT2 --> SSE_SVC
    CLIENT3 --> SSE_SVC
    
    SSE_SVC --> MONITORING
    SSE_SVC --> NOTIFICATIONS
    
    K8S_EVENTS --> SSE_SVC
    VM_EVENTS --> SSE_SVC
    COST_EVENTS --> SSE_SVC
    SYSTEM_EVENTS --> SSE_SVC
```

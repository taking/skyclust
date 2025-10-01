# Cloud Management Portal (CMP)

Cloud Management Portal (CMP)로, 다중 클라우드 환경을 통합 관리하는 플러그인 기반 플랫폼

## Features

- 다중 클라우드 지원: AWS, GCP, OpenStack, Proxmox
- 플러그인 아키텍처: 동적 로딩 가능한 클라우드 프로바이더
- 워크스페이스 기반 멀티테넌시: 사용자별 격리된 환경
- VM 관리: 인스턴스 생성, 삭제, 상태 관리
- IaC 지원: OpenTofu를 통한 인프라 관리
- 실시간 통신: WebSocket/SSE를 통한 실시간 업데이트
- Kubernetes 관리: 클러스터 및 리소스 관리
- 자격증명 관리: 암호화된 클라우드 자격증명 저장

## Architecture

백엔드 (Go)
- 프레임워크: Gin (HTTP), GORM (ORM)
- 데이터베이스: PostgreSQL
- 메시징: NATS (이벤트 버스)
- 인증: JWT 기반
- 암호화: AES 암호화 서비스

프론트엔드 (React + TypeScript)
- UI 라이브러리: Mantine
- 상태 관리: React Query
- HTTP 클라이언트: Axios
- 빌드 도구: Vite

인프라
- 컨테이너화: Docker + Docker Compose
- 데이터베이스: PostgreSQL 15
- 메시징: NATS 2.10

## Project Structure

```
cmp/
├── cmd/server/          # Main application
├── pkg/
│   ├── interfaces/      # Cloud provider interface definitions
│   │   ├── cloud_provider.go  # Main cloud provider interface
│   │   ├── iam.go            # IAM management interface
│   │   └── network.go        # Network management interface
│   └── plugin/          # Plugin management system
├── plugins/             # Plugin directory
│   ├── public/         # Public cloud providers
│   │   ├── aws/        # AWS plugin (aws-sdk-go-v2)
│   │   └── gcp/        # GCP plugin (google-cloud-go)
│   └── private/        # Private cloud providers
│       ├── openstack/  # OpenStack plugin (gophercloud)
│       └── proxmox/    # Proxmox plugin (go-proxmox)
├── examples/           # Provider examples
└── config.yaml         # Configuration file
```

## Quick Start

### Prerequisites

- Go 1.21 or later
- Make (optional, for using Makefile)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd cmp
```

2. Install dependencies:
```bash
make deps
# or
go mod tidy
```

3. Build the application and plugins:
```bash
make build-all
# or
make build && make build-plugins
```

4. Configure your cloud providers in `config.yaml`:
```yaml
providers:
  aws:
    access_key: "your-aws-access-key"
    secret_key: "your-aws-secret-key"
    region: "us-east-1"
  
  gcp:
    project_id: "your-gcp-project-id"
    credentials_file: "/path/to/credentials.json"
    region: "us-central1"
  
  openstack:
    auth_url: "http://your-openstack-keystone:5000/v3"
    username: "your-username"
    password: "your-password"
    project_id: "your-project-id"
    region: "RegionOne"
  
  proxmox:
    host: "your-proxmox-host"
    username: "your-username"
    password: "your-password"
    realm: "pve"
```

5. Run the server:

**Option A: Using config.yaml (Recommended)**
```bash
go run cmd/server/main.go --config config.yaml
```

**Option B: Using environment variables**
```bash
export CMP_DB_HOST=localhost
export CMP_DB_USER=cmp_user
export CMP_DB_PASSWORD=cmp_password
export CMP_DB_NAME=cmp
go run cmd/server/main.go --config config.yaml
```

**Option C: Using Makefile**
```bash
make run
# or
./bin/cmp-server --plugins plugins
```

The server will start on `http://localhost:8080` (or the port specified in your configuration)

## API Endpoints

### Health Check
```bash
GET /health
```

### Provider Management
```bash
# List all providers
GET /api/v1/providers

# Get provider info
GET /api/v1/providers/{name}

# Initialize provider
POST /api/v1/providers/{name}/initialize
```

### Instance Management
```bash
# List instances
GET /api/v1/providers/{name}/instances

# Create instance
POST /api/v1/providers/{name}/instances
{
  "name": "my-instance",
  "type": "t3.micro",
  "region": "us-east-1",
  "image_id": "ami-12345678"
}

# Get instance status
GET /api/v1/providers/{name}/instances/{id}

# Delete instance
DELETE /api/v1/providers/{name}/instances/{id}
```

### Region Management
```bash
# List regions
GET /api/v1/providers/{name}/regions
```

### Cost Estimation
```bash
# Get cost estimate
POST /api/v1/providers/{name}/cost-estimate
{
  "instance_type": "t3.micro",
  "region": "us-east-1",
  "duration": "1d"
}
```

## Plugin Development

### Creating a New Plugin

1. Create a new directory under `plugins/`:
```bash
mkdir plugins/azure
cd plugins/azure
```

2. Create `go.mod`:
```go
module azure-plugin

go 1.21

require cmp v0.0.0

replace cmp => ../../
```

3. Implement the plugin:
```go
package main

import (
    "context"
    "cmp/pkg/interfaces"
)

type AzureProvider struct {
    config map[string]interface{}
}

func New() interfaces.CloudProvider {
    return &AzureProvider{}
}

func (p *AzureProvider) GetName() string {
    return "Azure"
}

func (p *AzureProvider) GetVersion() string {
    return "1.0.0"
}

func (p *AzureProvider) Initialize(config map[string]interface{}) error {
    p.config = config
    return nil
}

// Implement other required methods...
```

4. Build the plugin:
```bash
go build -buildmode=plugin -o ../../plugins/azure.so azure.go
```

### Plugin Interface

All plugins must implement the `CloudProvider` interface:

```go
type CloudProvider interface {
    GetName() string
    GetVersion() string
    Initialize(config map[string]interface{}) error
    ListInstances(ctx context.Context) ([]Instance, error)
    CreateInstance(ctx context.Context, req CreateInstanceRequest) (*Instance, error)
    DeleteInstance(ctx context.Context, instanceID string) error
    GetInstanceStatus(ctx context.Context, instanceID string) (string, error)
    ListRegions(ctx context.Context) ([]Region, error)
    GetCostEstimate(ctx context.Context, req CostEstimateRequest) (*CostEstimate, error)
}
```

## Configuration

The application supports multiple configuration methods with a clear priority order. You can use YAML configuration files, environment variables, or command-line flags.

### Configuration Priority Order

The configuration system follows this priority order (highest to lowest):

1. **Command-line flags** (e.g., `--port 8080`)
2. **Environment variables** (e.g., `CMP_PORT=8080`)
3. **Configuration file** (e.g., `config.yaml`)
4. **Default values**

### Configuration Methods

#### 1. Configuration File (config.yaml)

Create a `config.yaml` file in your project root:

```yaml
server:
  port: "8080"
  host: "0.0.0.0"

database:
  host: "localhost"
  port: 5432
  user: "cmp_user"
  password: "cmp_password"
  name: "cmp"
  sslmode: "disable"

jwt:
  secret: "your-jwt-secret-here"
  expiry: "24h"

encryption:
  key: "your-32-byte-encryption-key-here"

nats:
  url: "nats://localhost:4222"
  cluster: "cmp-cluster"

plugins:
  directory: "plugins"

providers:
  aws:
    access_key: "your-aws-access-key"
    secret_key: "your-aws-secret-key"
    region: "us-east-1"
  
  gcp:
    project_id: "your-gcp-project-id"
    credentials_file: "/path/to/credentials.json"
    region: "us-central1"
  
  openstack:
    auth_url: "http://your-openstack-keystone:5000/v3"
    username: "your-username"
    password: "your-password"
    project_id: "your-project-id"
    region: "RegionOne"
  
  proxmox:
    host: "your-proxmox-host"
    username: "your-username"
    password: "your-password"
    realm: "pve"
```

#### 2. Environment Variables

Set environment variables to override configuration file values:

```bash
# Server configuration
export CMP_PORT=8080
export CMP_HOST=0.0.0.0

# Database configuration
export CMP_DB_HOST=localhost
export CMP_DB_PORT=5432
export CMP_DB_USER=cmp_user
export CMP_DB_PASSWORD=cmp_password
export CMP_DB_NAME=cmp
export CMP_DB_SSLMODE=disable

# Security configuration
export CMP_JWT_SECRET=your-jwt-secret-here
export CMP_ENCRYPTION_KEY=your-32-byte-encryption-key-here

# NATS configuration
export CMP_NATS_URL=nats://localhost:4222
export CMP_NATS_CLUSTER=cmp-cluster

# Plugin configuration
export CMP_PLUGINS_DIR=plugins

# Cloud provider configuration
export CMP_AWS_ACCESS_KEY=your-aws-access-key
export CMP_AWS_SECRET_KEY=your-aws-secret-key
export CMP_AWS_REGION=us-east-1

export CMP_GCP_PROJECT_ID=your-gcp-project-id
export CMP_GCP_CREDENTIALS_FILE=/path/to/credentials.json
export CMP_GCP_REGION=us-central1

export CMP_OPENSTACK_AUTH_URL=http://your-openstack-keystone:5000/v3
export CMP_OPENSTACK_USERNAME=your-username
export CMP_OPENSTACK_PASSWORD=your-password
export CMP_OPENSTACK_PROJECT_ID=your-project-id
export CMP_OPENSTACK_REGION=RegionOne

export CMP_PROXMOX_HOST=your-proxmox-host
export CMP_PROXMOX_USERNAME=your-username
export CMP_PROXMOX_PASSWORD=your-password
export CMP_PROXMOX_REALM=pve
```

#### 3. Command-line Flags

Use command-line flags for quick overrides:

```bash
# Basic usage
go run cmd/server/main.go --config config.yaml

# Override port
go run cmd/server/main.go --config config.yaml --port 8082

# Override plugin directory
go run cmd/server/main.go --config config.yaml --plugins /path/to/plugins

# Use different config file
go run cmd/server/main.go --config /path/to/production.yaml
```

### Configuration Examples

#### Development Environment

```bash
# Use config.yaml with minimal setup
go run cmd/server/main.go --config config.yaml
```

#### Production Environment

```bash
# Use environment variables for sensitive data
export CMP_DB_PASSWORD=secure-production-password
export CMP_JWT_SECRET=secure-jwt-secret-for-production
export CMP_ENCRYPTION_KEY=secure-32-byte-encryption-key
go run cmd/server/main.go --config config.yaml
```

#### Docker Environment

```bash
# Use environment variables in Docker
docker run -e CMP_PORT=8080 -e CMP_DB_HOST=postgres cmp-server
```

#### Testing Environment

```bash
# Override specific settings for testing
export CMP_PORT=8081
export CMP_DB_NAME=cmp_test
go run cmd/server/main.go --config config.yaml
```

### Configuration Options

| Option | Environment Variable | Default | Description |
|--------|-------------------|---------|-------------|
| `server.port` | `CMP_PORT` | `8080` | Server port |
| `server.host` | `CMP_HOST` | `0.0.0.0` | Server host |
| `database.host` | `CMP_DB_HOST` | `localhost` | Database host |
| `database.port` | `CMP_DB_PORT` | `5432` | Database port |
| `database.user` | `CMP_DB_USER` | `cmp_user` | Database user |
| `database.password` | `CMP_DB_PASSWORD` | `cmp_password` | Database password |
| `database.name` | `CMP_DB_NAME` | `cmp` | Database name |
| `database.sslmode` | `CMP_DB_SSLMODE` | `disable` | Database SSL mode |
| `jwt.secret` | `CMP_JWT_SECRET` | Generated | JWT secret key |
| `encryption.key` | `CMP_ENCRYPTION_KEY` | Generated | Encryption key |
| `nats.url` | `CMP_NATS_URL` | `nats://localhost:4222` | NATS server URL |
| `plugins.directory` | `CMP_PLUGINS_DIR` | `plugins` | Plugin directory path |

### Security Considerations

- **Production Environment**: Always use environment variables for sensitive data like passwords and secrets
- **JWT Secret**: Must be at least 32 characters long
- **Encryption Key**: Must be exactly 32 bytes (64 hex characters)
- **Database SSL**: Enable SSL in production (`sslmode=require` or `sslmode=verify-full`)
- **Default Values**: Never use default values in production

## Development

### Running in Development Mode

**Option A: Using config.yaml (Recommended)**
```bash
go run cmd/server/main.go --config config.yaml
```

**Option B: Using Makefile**
```bash
make dev
```

This will:
- Build the application and plugins
- Start the server with hot reload
- Display helpful URLs

**Option C: With environment variables**
```bash
export CMP_PORT=8080
export CMP_DB_HOST=localhost
export CMP_DB_USER=cmp_user
export CMP_DB_PASSWORD=cmp_password
export CMP_DB_NAME=cmp
go run cmd/server/main.go --config config.yaml
```

### Testing

```bash
make test
```

### Code Formatting

```bash
make fmt
```

### Linting

```bash
make lint
```

## Examples

### List all instances from AWS

```bash
curl http://localhost:8080/api/v1/providers/aws/instances
```

### Create a new GCP instance

```bash
curl -X POST http://localhost:8080/api/v1/providers/gcp/instances \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-gcp-instance",
    "type": "e2-micro",
    "region": "us-central1",
    "image_id": "debian-cloud/debian-11"
  }'
```

### Get cost estimate

```bash
curl -X POST http://localhost:8080/api/v1/providers/aws/cost-estimate \
  -H "Content-Type: application/json" \
  -d '{
    "instance_type": "t3.micro",
    "region": "us-east-1",
    "duration": "1m"
  }'
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Supported Cloud Providers

### Public Clouds
- **AWS** - Using [aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2)
- **GCP** - Using [google-cloud-go](https://github.com/googleapis/google-cloud-go)

### Private Clouds
- **OpenStack** - Using [gophercloud](https://github.com/gophercloud/gophercloud)
- **Proxmox VE** - Using [go-proxmox](https://github.com/luthermonson/go-proxmox)

## SDK Integration

This project uses official and community-maintained Go SDKs for each cloud provider:

- **AWS**: Official AWS SDK v2 with support for EC2, IAM, and other services
- **GCP**: Official Google Cloud Go SDK with Compute Engine support
- **OpenStack**: Gophercloud library for comprehensive OpenStack API support
- **Proxmox**: Community-maintained go-proxmox library for Proxmox VE API

## Roadmap

- [ ] Add more cloud providers (Azure, DigitalOcean, etc.)
- [ ] Implement authentication and authorization
- [ ] Add web UI dashboard
- [ ] Support for more resource types (databases, storage, etc.)
- [ ] Plugin marketplace
- [ ] Metrics and monitoring
- [ ] Multi-tenant support

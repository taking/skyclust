# Cloud Management Portal (CMP)

A plugin-based cloud management portal built with Go that supports multiple cloud providers through a unified interface.

## Features

- 🔌 **Plugin Architecture**: Easily add new cloud providers as plugins
- 🌐 **Multi-Cloud Support**: Supports AWS, GCP, OpenStack, and Proxmox
- 🏢 **Public/Private Cloud**: Separate plugin directories for public and private clouds
- 🚀 **RESTful API**: Clean HTTP API for all operations
- ⚡ **Dynamic Loading**: Load plugins at runtime without recompilation
- 🔧 **Configurable**: YAML-based configuration
- 📊 **Cost Estimation**: Built-in cost calculation for resources
- 🔐 **IAM Management**: User, group, role, and policy management
- 🌐 **Network Management**: VPC, subnet, security group, and key pair management
- 🔑 **Key Pair Management**: SSH key pair creation and management

## Architecture

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
```bash
make run
# or
./bin/cmp-server --plugins plugins
```

The server will start on `http://localhost:8080`

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

The application uses YAML configuration. You can specify the config file with the `--config` flag:

```bash
./bin/cmp-server --config /path/to/config.yaml
```

### Configuration Options

- `server.port`: Server port (default: 8080)
- `plugins.directory`: Plugin directory path (default: plugins)
- `providers.{name}.*`: Provider-specific configuration

## Development

### Running in Development Mode

```bash
make dev
```

This will:
- Build the application and plugins
- Start the server with hot reload
- Display helpful URLs

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

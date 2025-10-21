# SkyClust gRPC Protocol Definitions

This directory contains Protocol Buffer (proto) definitions for SkyClust cloud provider services.

## 📁 Structure

```
api/proto/v1/
├── cloud_provider.proto  # Core cloud provider operations
├── iam.proto            # Identity and Access Management
├── network.proto        # Network management (VPC, Subnet, Security Groups)
├── kubernetes.proto     # Kubernetes cluster management
└── README.md           # This file
```

## 🚀 Quick Start

### Prerequisites

Install buf (gRPC build tool):
```bash
go install github.com/bufbuild/buf/cmd/buf@latest
```

### Generate Code

```bash
make proto-gen
```

This will generate Go code in `api/gen/` directory.

### Lint Proto Files

```bash
make proto-lint
```

### Check Breaking Changes

```bash
make proto-breaking
```

## 📚 Services Overview

### 1. CloudProviderService

Core cloud provider operations:
- Instance management (Create, List, Delete, Start, Stop, Restart)
- Region listing
- Cost estimation
- Health checks

### 2. IAMService

Identity and Access Management:
- User management
- Group management
- Role management
- Policy management
- Access key management

### 3. NetworkService

Network infrastructure:
- VPC management
- Subnet management
- Security Group management
- Load Balancer management
- Firewall rules

### 4. KubernetesService

Kubernetes operations:
- Cluster management
- Node pool management
- Addon management

## 🔧 Development

### Adding New RPCs

1. Edit the appropriate `.proto` file
2. Run `make proto-gen` to generate code
3. Implement the RPC in your provider service
4. Test the implementation

### Versioning

- Use semantic versioning for proto packages
- Check breaking changes before releases: `make proto-breaking`

## 📖 Documentation

For detailed API documentation, see the generated documentation after running:

```bash
buf generate --template buf.gen.yaml
```

## 🧪 Testing

Test your proto definitions:

```bash
# Lint
make proto-lint

# Breaking changes
make proto-breaking

# Generate and test
make proto-gen
go test ./api/gen/...
```


# SkyClust Cloud Provider Services

This directory contains gRPC-based cloud provider services that implement the SkyClust provider protocols.

## 📁 Structure

```
providers/
├── aws/          # AWS Provider Service
├── gcp/          # Google Cloud Provider Service
├── openstack/    # OpenStack Provider Service
├── proxmox/      # Proxmox Provider Service
└── README.md     # This file
```

## 🚀 Quick Start

### Build All Providers

```bash
make build-providers
```

### Run Individual Provider

```bash
# AWS
cd providers/aws && go run . --port 50051

# GCP  
cd providers/gcp && go run . --port 50052

# OpenStack
cd providers/openstack && go run . --port 50053

# Proxmox
cd providers/proxmox && go run . --port 50054
```

### Docker

```bash
# Build
docker build -t skyclust-aws-provider ./providers/aws
docker build -t skyclust-gcp-provider ./providers/gcp
docker build -t skyclust-openstack-provider ./providers/openstack
docker build -t skyclust-proxmox-provider ./providers/proxmox

# Run
docker run -p 50051:50051 skyclust-aws-provider
docker run -p 50052:50052 skyclust-gcp-provider
docker run -p 50053:50053 skyclust-openstack-provider
docker run -p 50054:50054 skyclust-proxmox-provider
```

### Docker Compose

```bash
docker-compose up -d
```

## 🔧 Development

### Adding a New Provider

1. Create a new directory under `providers/`
2. Implement the gRPC server based on the protocol
3. Add Dockerfile
4. Update docker-compose.yml
5. Add to Makefile

### Testing

```bash
# Test with grpcurl
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext localhost:50051 skyclust.provider.v1.CloudProviderService/GetProviderInfo
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check
```

## 📖 Provider Details

### AWS Provider

**Port**: 50051  
**Features**:
- EC2 instance management
- Region listing
- Cost estimation
- Health checks

**Environment Variables**:
- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `AWS_REGION` (default: us-east-1)

### GCP Provider

**Port**: 50052  
**Features**:
- Compute Engine management
- Region/Zone listing
- Cost estimation
- Health checks

**Environment Variables**:
- `GOOGLE_APPLICATION_CREDENTIALS`
- `GCP_PROJECT_ID`
- `GCP_REGION` (default: us-central1)

### OpenStack Provider

**Port**: 50053  
**Features**:
- Nova instance management
- Region listing
- Health checks

**Environment Variables**:
- `OS_AUTH_URL`
- `OS_USERNAME`
- `OS_PASSWORD`
- `OS_PROJECT_ID`
- `OS_REGION_NAME`

### Proxmox Provider

**Port**: 50054  
**Features**:
- VM management
- Node listing
- Health checks

**Environment Variables**:
- `PROXMOX_HOST`
- `PROXMOX_USER`
- `PROXMOX_PASSWORD`
- `PROXMOX_REALM` (default: pve)

## 🧪 Testing

### Health Check

```bash
# AWS
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check

# GCP
grpcurl -plaintext localhost:50052 grpc.health.v1.Health/Check

# OpenStack
grpcurl -plaintext localhost:50053 grpc.health.v1.Health/Check

# Proxmox
grpcurl -plaintext localhost:50054 grpc.health.v1.Health/Check
```

### List Instances

```bash
grpcurl -plaintext -d '{"region":"us-east-1"}' \
  localhost:50051 skyclust.provider.v1.CloudProviderService/ListInstances
```

## 📚 Documentation

For detailed protocol documentation, see `api/proto/v1/README.md`.

## 🔐 Security

- Use TLS for production deployments
- Implement mTLS for service-to-service communication
- Store credentials securely (Vault, Secrets Manager)
- Follow least-privilege principle

## 🐛 Troubleshooting

### Connection Refused

- Check if the service is running: `docker ps` or `ps aux | grep provider`
- Verify the port is correct
- Check firewall rules

### Authentication Failed

- Verify environment variables are set
- Check credential validity
- Review provider-specific auth requirements

### Instance Not Found

- Verify the instance exists in the provider
- Check region/zone settings
- Confirm permissions

## 📝 Notes

- All providers implement the CloudProviderService protocol
- Health checks are available on all providers
- Reflection is enabled for debugging (disable in production)
- Graceful shutdown is handled via SIGTERM/SIGINT


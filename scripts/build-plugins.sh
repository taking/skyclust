#!/bin/bash

# Build all plugins script
echo "Building all plugins..."

# Get the script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Project root: $PROJECT_ROOT"

# Build public plugins
echo "Building public plugins..."
cd "$PROJECT_ROOT/plugins/public/aws"
go mod tidy
go build -buildmode=plugin -o aws.so aws.go
echo "AWS plugin built"

cd "$PROJECT_ROOT/plugins/public/gcp"
go mod tidy
go build -buildmode=plugin -o gcp.so gcp.go
echo "GCP plugin built"

# Build private plugins
echo "Building private plugins..."
cd "$PROJECT_ROOT/plugins/private/openstack"
go mod tidy
go build -buildmode=plugin -o openstack.so openstack.go
echo "OpenStack plugin built"

cd "$PROJECT_ROOT/plugins/private/proxmox"
go mod tidy
go build -buildmode=plugin -o proxmox.so proxmox.go
echo "Proxmox plugin built"

echo "All plugins built successfully!"

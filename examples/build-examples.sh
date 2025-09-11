#!/bin/bash

# Build all example plugins
# This script builds all example providers into the main plugins directory

echo "Building example plugins..."

# Create plugins directory if it doesn't exist
mkdir -p ../../plugins

# Build AWS plugin
echo "Building AWS plugin..."
cd aws
go mod init aws-example 2>/dev/null || true
echo "replace cmp => ../../" >> go.mod
go mod tidy
go build -buildmode=plugin -o ../../plugins/aws-example.so aws-example.go
if [ $? -eq 0 ]; then
    echo "✓ AWS plugin built successfully"
else
    echo "✗ Failed to build AWS plugin"
fi
cd ..

# Build GCP plugin
echo "Building GCP plugin..."
cd gcp
go mod init gcp-example 2>/dev/null || true
echo "replace cmp => ../../" >> go.mod
go mod tidy
go build -buildmode=plugin -o ../../plugins/gcp-example.so gcp-example.go
if [ $? -eq 0 ]; then
    echo "✓ GCP plugin built successfully"
else
    echo "✗ Failed to build GCP plugin"
fi
cd ..

# Build OpenStack plugin
echo "Building OpenStack plugin..."
cd openstack
go mod init openstack-example 2>/dev/null || true
echo "replace cmp => ../../" >> go.mod
go mod tidy
go build -buildmode=plugin -o ../../plugins/openstack-example.so openstack-example.go
if [ $? -eq 0 ]; then
    echo "✓ OpenStack plugin built successfully"
else
    echo "✗ Failed to build OpenStack plugin"
fi
cd ..

# Build Proxmox plugin
echo "Building Proxmox plugin..."
cd proxmox
go mod init proxmox-example 2>/dev/null || true
echo "replace cmp => ../../" >> go.mod
go mod tidy
go build -buildmode=plugin -o ../../plugins/proxmox-example.so proxmox-example.go
if [ $? -eq 0 ]; then
    echo "✓ Proxmox plugin built successfully"
else
    echo "✗ Failed to build Proxmox plugin"
fi
cd ..

# Build Custom plugin
echo "Building Custom plugin..."
cd custom
go mod init custom-example 2>/dev/null || true
echo "replace cmp => ../../" >> go.mod
go mod tidy
go build -buildmode=plugin -o ../../plugins/custom-example.so custom.go
if [ $? -eq 0 ]; then
    echo "✓ Custom plugin built successfully"
else
    echo "✗ Failed to build Custom plugin"
fi
cd ..

# Build Azure plugin (legacy)
echo "Building Azure plugin..."
cd azure
go mod init azure-example 2>/dev/null || true
echo "replace cmp => ../../" >> go.mod
go mod tidy
go build -buildmode=plugin -o ../../plugins/azure-example.so azure.go
if [ $? -eq 0 ]; then
    echo "✓ Azure plugin built successfully"
else
    echo "✗ Failed to build Azure plugin"
fi
cd ..

# Build DigitalOcean plugin (legacy)
echo "Building DigitalOcean plugin..."
cd digitalocean
go mod init digitalocean-example 2>/dev/null || true
echo "replace cmp => ../../" >> go.mod
go mod tidy
go build -buildmode=plugin -o ../../plugins/digitalocean-example.so digitalocean.go
if [ $? -eq 0 ]; then
    echo "✓ DigitalOcean plugin built successfully"
else
    echo "✗ Failed to build DigitalOcean plugin"
fi
cd ..

echo ""
echo "All example plugins built!"
echo "Plugins are available in: ../../plugins/"
echo ""
echo "To test the plugins, run:"
echo "  make run"
echo ""
echo "Then check available providers:"
echo "  curl http://localhost:8080/api/v1/providers"

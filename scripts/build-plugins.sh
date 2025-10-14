#!/bin/bash

# Build all plugins script
echo "Building all plugins..."

# Build public plugins
echo "Building public plugins..."
cd /home/taking/skyclust/plugins/public/aws
go mod tidy
go build -buildmode=plugin -o aws.so aws.go
echo "AWS plugin built"

cd /home/taking/skyclust/plugins/public/gcp
go mod tidy
go build -buildmode=plugin -o gcp.so gcp.go
echo "GCP plugin built"

# Build private plugins
echo "Building private plugins..."
cd /home/taking/skyclust/plugins/private/openstack
go mod tidy
go build -buildmode=plugin -o openstack.so openstack.go
echo "OpenStack plugin built"

cd /home/taking/skyclust/plugins/private/proxmox
go mod tidy
go build -buildmode=plugin -o proxmox.so proxmox.go
echo "Proxmox plugin built"

echo "All plugins built successfully!"

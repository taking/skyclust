.PHONY: build build-plugins clean run test

# Build the main application
build:
	go build -o bin/cmp-server cmd/server/main.go

# Build all plugins
build-plugins:
	@echo "Building public plugins..."
	@echo "Building AWS plugin..."
	cd plugins/public/aws && go build -buildmode=plugin -o ../../../plugins/public/aws.so aws.go
	
	@echo "Building GCP plugin..."
	cd plugins/public/gcp && go build -buildmode=plugin -o ../../../plugins/public/gcp.so gcp.go
	
	@echo "Building private plugins..."
	@echo "Building OpenStack plugin..."
	cd plugins/private/openstack && go build -buildmode=plugin -o ../../../plugins/private/openstack.so openstack.go
	
	@echo "Building Proxmox plugin..."
	cd plugins/private/proxmox && go build -buildmode=plugin -o ../../../plugins/private/proxmox.so proxmox.go

# Build everything
build-all: build build-plugins

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f plugins/public/*.so
	rm -f plugins/private/*.so

# Run the server
run: build-all
	./bin/cmp-server --plugins plugins

# Run with specific config
run-config: build-all
	./bin/cmp-server --config config.yaml --plugins plugins --port 8080

# Test the application
test:
	go test ./...

# Install dependencies
deps:
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Create plugin directory
setup:
	mkdir -p plugins/public
	mkdir -p plugins/private
	mkdir -p bin

# Development mode (with hot reload)
dev: build-all
	@echo "Starting development server..."
	@echo "API will be available at http://localhost:8080"
	@echo "Health check: http://localhost:8080/health"
	@echo "Providers: http://localhost:8080/api/v1/providers"
	./bin/cmp-server --plugins plugins --port 8080

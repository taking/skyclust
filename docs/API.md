# CMP API Documentation

## Overview
Cloud Management Portal (CMP) provides a unified interface for managing multiple cloud providers through a plugin-based architecture.

## Base URL
```
http://localhost:8081
```

## Authentication
Currently, the simple server does not require authentication. For production, JWT tokens are required.

## Endpoints

### Health Check
```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-10-01T11:25:05Z",
  "version": "1.0.0"
}
```

### Providers

#### List All Providers
```http
GET /api/v1/providers
```

**Response:**
```json
{
  "providers": ["public/aws"],
  "count": 1
}
```

#### Get Provider Information
```http
GET /api/v1/providers/{name}
```

**Example:**
```http
GET /api/v1/providers/aws
```

**Response:**
```json
{
  "name": "AWS",
  "version": "1.0.0"
}
```

#### List Provider Instances
```http
GET /api/v1/providers/{name}/instances?region={region}
```

**Parameters:**
- `region` (optional): Filter instances by region (e.g., `us-east-1`)

**Example:**
```http
GET /api/v1/providers/aws/instances?region=us-east-1
```

**Response:**
```json
{
  "instances": [
    {
      "id": "i-1234567890abcdef0",
      "name": "web-server-01",
      "type": "t3.micro",
      "status": "running",
      "region": "us-east-1"
    }
  ],
  "provider": "aws",
  "region": "us-east-1",
  "count": 1
}
```

#### List Provider Regions
```http
GET /api/v1/providers/{name}/regions
```

**Example:**
```http
GET /api/v1/providers/aws/regions
```

**Response:**
```json
{
  "regions": [
    {
      "name": "us-east-1",
      "display_name": "US East (N. Virginia)"
    },
    {
      "name": "us-west-2",
      "display_name": "US West (Oregon)"
    }
  ],
  "provider": "aws",
  "count": 2
}
```

#### Get Cost Estimates
```http
GET /api/v1/providers/{name}/cost-estimates
```

**Response:**
```json
{
  "estimates": [],
  "count": 0
}
```

#### Create Cost Estimate
```http
POST /api/v1/providers/{name}/cost-estimates
```

**Request Body:**
```json
{
  "instance_type": "t3.micro",
  "region": "us-east-1",
  "duration": "1h"
}
```

**Response:**
```json
{
  "instance_type": "t3.micro",
  "region": "us-east-1",
  "duration": "1h",
  "cost": 0.0104,
  "currency": "USD"
}
```

## Error Responses

### Provider Not Found
```json
{
  "error": "provider not found"
}
```

### Provider Not Initialized
```json
{
  "error": "AWS provider not initialized. Please configure AWS credentials"
}
```

### Internal Server Error
```json
{
  "error": "Failed to list instances"
}
```

## CORS Support
The API supports Cross-Origin Resource Sharing (CORS) with the following headers:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Origin, Content-Type, Accept, Authorization`

## Rate Limiting
Currently, no rate limiting is implemented in the simple server. Production servers should implement rate limiting.

## Examples

### Get AWS Instances in US East
```bash
curl "http://localhost:8081/api/v1/providers/aws/instances?region=us-east-1"
```

### Get All AWS Regions
```bash
curl "http://localhost:8081/api/v1/providers/aws/regions"
```

### Health Check
```bash
curl "http://localhost:8081/health"
```

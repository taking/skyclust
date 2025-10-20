# SkyClust API 문서

이 문서는 SkyClust의 REST API에 대한 포괄적인 API 문서를 제공합니다.

## 기본 정보

- **Base URL**: `http://localhost:8080/api/v1`
- **인증**: JWT Bearer Token
- **응답 형식**: JSON
- **문자 인코딩**: UTF-8

## 인증 및 권한 관리

모든 API 엔드포인트는 로그인 및 등록을 제외하고 인증이 필요합니다.

### 인증 방법

1. **JWT 토큰** (권장)
2. **API 키** (서비스 간 통신용)

### JWT 인증

Authorization 헤더에 JWT 토큰을 포함:

```http
Authorization: Bearer <jwt-token>
```

### API 키 인증

X-API-Key 헤더에 API 키를 포함:

```http
X-API-Key: <api-key>
```

### 역할 기반 접근 제어 (RBAC)

SkyClust는 역할 기반 접근 제어 시스템을 사용합니다.

#### 사용자 역할
- **admin**: 시스템 관리자 권한
- **user**: 일반 사용자 권한  
- **viewer**: 읽기 전용 권한

#### 권한 시스템
- **user:create**: 사용자 생성
- **user:read**: 사용자 조회
- **user:update**: 사용자 수정
- **user:delete**: 사용자 삭제
- **system:read**: 시스템 정보 조회
- **system:manage**: 시스템 관리
- **audit:read**: 감사 로그 조회
- **workspace:manage**: 워크스페이스 관리

#### 특별 사항
- 첫 번째 등록 사용자는 자동으로 `admin` 역할을 받습니다
- 이후 등록 사용자는 `user` 역할을 받습니다

## 공통 응답 형식

### 성공 응답

```json
{
  "success": true,
  "message": "작업이 성공적으로 완료되었습니다",
  "data": {
    // 응답 데이터
  },
  "request_id": "req_123456789",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### 에러 응답

```json
{
  "success": false,
  "error": "에러 메시지",
  "code": "ERROR_CODE",
  "data": {
    // 추가 에러 세부사항
  },
  "request_id": "req_123456789",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## 엔드포인트

### 인증

#### 사용자 등록

```http
POST /auth/register
```

**요청 본문:**
```json
{
  "username": "string",
  "email": "string",
  "password": "string"
}
```

**응답:**
```json
{
  "success": true,
  "message": "사용자가 성공적으로 등록되었습니다",
  "data": {
    "token": "jwt-token",
    "user": {
      "id": "uuid",
      "username": "string",
      "email": "string",
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

#### 로그인

```http
POST /auth/login
```

**요청 본문:**
```json
{
  "email": "string",
  "password": "string"
}
```

**응답:**
```json
{
  "success": true,
  "message": "로그인 성공",
  "data": {
    "token": "jwt-token",
    "user": {
      "id": "uuid",
      "username": "string",
      "email": "string"
    }
  }
}
```

#### 로그아웃

```http
POST /auth/logout
```

**응답:**
```json
{
  "success": true,
  "message": "로그아웃 성공"
}
```

### 워크스페이스 관리

#### 워크스페이스 목록 조회

```http
GET /workspaces
```

**쿼리 매개변수:**
- `limit` (int): 결과 수 (기본값: 10, 최대: 100)
- `offset` (int): 건너뛸 결과 수 (기본값: 0)
- `sort` (string): 정렬 필드와 방향 (예: "name:asc")

**응답:**
```json
{
  "success": true,
  "data": {
    "workspaces": [
      {
        "id": "uuid",
        "name": "string",
        "description": "string",
        "owner_id": "uuid",
        "is_active": true,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "limit": 10,
      "offset": 0,
      "total": 25
    }
  }
}
```

#### 워크스페이스 생성

```http
POST /workspaces
```

**요청 본문:**
```json
{
  "name": "string",
  "description": "string"
}
```

**응답:**
```json
{
  "success": true,
  "message": "워크스페이스가 성공적으로 생성되었습니다",
  "data": {
    "id": "uuid",
    "name": "string",
    "description": "string",
    "owner_id": "uuid",
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 워크스페이스 조회

```http
GET /workspaces/{id}
```

**응답:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "string",
    "description": "string",
    "owner_id": "uuid",
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 워크스페이스 수정

```http
PUT /workspaces/{id}
```

**요청 본문:**
```json
{
  "name": "string",
  "description": "string"
}
```

**응답:**
```json
{
  "success": true,
  "message": "워크스페이스가 성공적으로 수정되었습니다",
  "data": {
    "id": "uuid",
    "name": "string",
    "description": "string",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 워크스페이스 삭제

```http
DELETE /workspaces/{id}
```

**응답:**
```json
{
  "success": true,
  "message": "워크스페이스가 성공적으로 삭제되었습니다"
}
```

### VM 관리

#### VM 목록 조회

```http
GET /vms
```

**쿼리 매개변수:**
- `workspace_id` (string): 워크스페이스별 필터
- `provider` (string): 클라우드 프로바이더별 필터
- `status` (string): VM 상태별 필터
- `limit` (int): 결과 수 (기본값: 20)
- `offset` (int): 건너뛸 결과 수 (기본값: 0)

**응답:**
```json
{
  "success": true,
  "data": {
    "vms": [
      {
        "id": "uuid",
        "name": "string",
        "provider": "aws",
        "instance_type": "t3.micro",
        "status": "running",
        "region": "us-east-1",
        "workspace_id": "uuid",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "limit": 20,
      "offset": 0,
      "total": 50
    }
  }
}
```

#### VM 생성

```http
POST /vms
```

**요청 본문:**
```json
{
  "name": "string",
  "provider": "aws",
  "instance_type": "t3.micro",
  "region": "us-east-1",
  "workspace_id": "uuid",
  "image_id": "ami-12345678",
  "key_pair": "my-key-pair"
}
```

**응답:**
```json
{
  "success": true,
  "message": "VM 생성이 시작되었습니다",
  "data": {
    "id": "uuid",
    "name": "string",
    "provider": "aws",
    "status": "pending",
    "workspace_id": "uuid"
  }
}
```

#### VM 조회

```http
GET /vms/{id}
```

**응답:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "string",
    "provider": "aws",
    "instance_type": "t3.micro",
    "status": "running",
    "region": "us-east-1",
    "public_ip": "1.2.3.4",
    "private_ip": "10.0.0.1",
    "workspace_id": "uuid",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### VM 삭제

```http
DELETE /vms/{id}
```

**응답:**
```json
{
  "success": true,
  "message": "VM 삭제가 시작되었습니다"
}
```

### 자격증명 관리

#### 자격증명 목록 조회

```http
GET /credentials
```

**쿼리 매개변수:**
- `provider` (string): 클라우드 프로바이더별 필터
- `limit` (int): 결과 수 (기본값: 20)
- `offset` (int): 건너뛸 결과 수 (기본값: 0)

**응답:**
```json
{
  "success": true,
  "data": {
    "credentials": [
      {
        "id": "uuid",
        "name": "string",
        "provider": "aws",
        "is_active": true,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "limit": 20,
      "offset": 0,
      "total": 10
    }
  }
}
```

#### 자격증명 생성

```http
POST /credentials
```

**요청 본문:**
```json
{
  "name": "string",
  "provider": "aws",
  "data": {
    "access_key": "string",
    "secret_key": "string",
    "region": "us-east-1"
  }
}
```

**응답:**
```json
{
  "success": true,
  "message": "자격증명이 성공적으로 생성되었습니다",
  "data": {
    "id": "uuid",
    "name": "string",
    "provider": "aws",
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 자격증명 수정

```http
PUT /credentials/{id}
```

**요청 본문:**
```json
{
  "name": "string",
  "data": {
    "access_key": "string",
    "secret_key": "string",
    "region": "us-east-1"
  }
}
```

**응답:**
```json
{
  "success": true,
  "message": "자격증명이 성공적으로 수정되었습니다",
  "data": {
    "id": "uuid",
    "name": "string",
    "provider": "aws",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 자격증명 삭제

```http
DELETE /credentials/{id}
```

**응답:**
```json
{
  "success": true,
  "message": "자격증명이 성공적으로 삭제되었습니다"
}
```

### 비용 분석

#### 비용 분석 조회

```http
GET /cost-analysis
```

**쿼리 매개변수:**
- `period` (string): 기간 (1d, 7d, 30d, 90d)
- `workspace_id` (string): 워크스페이스별 필터
- `provider` (string): 클라우드 프로바이더별 필터

**응답:**
```json
{
  "success": true,
  "data": {
    "total_cost": 1250.50,
    "currency": "USD",
    "period": "30d",
    "breakdown": [
      {
        "provider": "aws",
        "cost": 800.25,
        "percentage": 64.0
      },
      {
        "provider": "gcp",
        "cost": 450.25,
        "percentage": 36.0
      }
    ],
    "trend": [
      {
        "date": "2024-01-01",
        "cost": 45.50
      }
    ]
  }
}
```

#### 비용 세부 분석

```http
GET /cost-analysis/breakdown
```

**쿼리 매개변수:**
- `provider` (string): 클라우드 프로바이더
- `service` (string): 서비스 유형
- `period` (string): 기간

**응답:**
```json
{
  "success": true,
  "data": {
    "provider": "aws",
    "total_cost": 800.25,
    "services": [
      {
        "service": "EC2",
        "cost": 600.00,
        "percentage": 75.0
      },
      {
        "service": "S3",
        "cost": 200.25,
        "percentage": 25.0
      }
    ]
  }
}
```

### 알림 관리

#### 알림 목록 조회

```http
GET /notifications
```

**쿼리 매개변수:**
- `unread_only` (boolean): 읽지 않은 알림만 필터
- `type` (string): 알림 유형별 필터
- `priority` (string): 우선순위별 필터
- `limit` (int): 결과 수 (기본값: 20)
- `offset` (int): 건너뛸 결과 수 (기본값: 0)

**응답:**
```json
{
  "success": true,
  "data": {
    "notifications": [
      {
        "id": "uuid",
        "title": "string",
        "message": "string",
        "type": "info",
        "priority": "medium",
        "is_read": false,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "limit": 20,
      "offset": 0,
      "total": 15
    }
  }
}
```

#### 알림 읽음 처리

```http
PUT /notifications/{id}/read
```

**응답:**
```json
{
  "success": true,
  "message": "알림이 읽음 처리되었습니다"
}
```

#### 모든 알림 읽음 처리

```http
PUT /notifications/read-all
```

**응답:**
```json
{
  "success": true,
  "message": "모든 알림이 읽음 처리되었습니다"
}
```

### 감사 로그

#### 감사 로그 조회

```http
GET /audit-logs
```

**쿼리 매개변수:**
- `user_id` (string): 사용자별 필터
- `action` (string): 액션별 필터
- `resource` (string): 리소스별 필터
- `start_time` (string): 시작 시간 (ISO 8601)
- `end_time` (string): 종료 시간 (ISO 8601)
- `limit` (int): 결과 수 (기본값: 50)
- `offset` (int): 건너뛸 결과 수 (기본값: 0)

**응답:**
```json
{
  "success": true,
  "data": {
    "audit_logs": [
      {
        "id": "uuid",
        "user_id": "uuid",
        "action": "create",
        "resource": "vm",
        "resource_id": "uuid",
        "success": true,
        "ip_address": "192.168.1.1",
        "user_agent": "Mozilla/5.0...",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "limit": 50,
      "offset": 0,
      "total": 100
    }
  }
}
```

### 시스템 상태

#### 헬스 체크

```http
GET /health
```

**응답:**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": "2024-01-01T00:00:00Z",
    "uptime": "72h30m15s",
    "version": "1.0.0",
    "environment": "production",
    "metrics": {
      "memory": {
        "alloc_mb": 45.2,
        "total_alloc_mb": 120.5,
        "sys_mb": 256.0,
        "num_gc": 15
      },
      "goroutines": 25
    }
  }
}
```

#### 데이터베이스 헬스 체크

```http
GET /health/db
```

**응답:**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "connection_count": 5,
    "max_connections": 100,
    "response_time": "2ms"
  }
}
```

#### Redis 헬스 체크

```http
GET /health/redis
```

**응답:**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "connected_clients": 3,
    "used_memory": "2.5MB",
    "response_time": "1ms"
  }
}
```

## 에러 코드

| 코드 | HTTP 상태 | 설명 |
|------|-----------|------|
| `VALIDATION_ERROR` | 400 | 요청 검증 실패 |
| `UNAUTHORIZED` | 401 | 인증 필요 |
| `FORBIDDEN` | 403 | 권한 부족 |
| `NOT_FOUND` | 404 | 리소스 없음 |
| `CONFLICT` | 409 | 리소스 중복 |
| `RATE_LIMITED` | 429 | 요청 제한 초과 |
| `INTERNAL_ERROR` | 500 | 내부 서버 오류 |
| `SERVICE_UNAVAILABLE` | 503 | 서비스 일시 중단 |

## Rate Limiting

API 요청은 남용 방지를 위해 제한됩니다:

- **인증된 사용자**: 시간당 1000 요청
- **인증되지 않은 사용자**: 시간당 100 요청
- **API 키**: 시간당 10000 요청

Rate limit 헤더가 응답에 포함됩니다:

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## 페이지네이션

목록 엔드포인트는 페이지네이션을 지원합니다:

- `limit`: 결과 수 (기본값: 20, 최대: 100)
- `offset`: 건너뛸 결과 수 (기본값: 0)

응답에 페이지네이션 메타데이터가 포함됩니다:

```json
{
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 150,
    "has_next": true,
    "has_prev": false
  }
}
```

## 정렬

목록 엔드포인트는 정렬을 지원합니다:

- `sort`: 정렬 필드와 방향 (예: "name:asc", "created_at:desc")
- 사용 가능한 필드는 엔드포인트마다 다름

## 필터링

목록 엔드포인트는 필터링을 지원합니다:

- 쿼리 매개변수로 필터링
- 여러 값 지원 (예: `status=running&status=pending`)
- 날짜 범위 지원 (예: `start_date=2024-01-01&end_date=2024-01-31`)

## WebSocket 이벤트

실시간 업데이트는 WebSocket을 통해 제공됩니다:

### 연결

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
```

### 인증

```javascript
ws.send(JSON.stringify({
  type: 'auth',
  token: 'jwt-token'
}));
```

### 이벤트 유형

- `vm_status_changed`: VM 상태 업데이트
- `cost_updated`: 비용 분석 업데이트
- `notification_received`: 새 알림
- `audit_log_created`: 새 감사 로그 항목

### 예시 이벤트

```json
{
  "type": "vm_status_changed",
  "data": {
    "vm_id": "uuid",
    "status": "running",
    "timestamp": "2024-01-01T00:00:00Z"
  }
}
```

## SDK 예제

### JavaScript/TypeScript

```typescript
import { SkyClustClient } from '@skyclust/sdk';

const client = new SkyClustClient({
  baseURL: 'http://localhost:8080/api/v1',
  token: 'your-jwt-token'
});

// 워크스페이스 목록 조회
const workspaces = await client.workspaces.list();

// VM 생성
const vm = await client.vms.create({
  name: 'my-vm',
  provider: 'aws',
  instance_type: 't3.micro',
  region: 'us-east-1'
});
```

### Python

```python
from skyclust import SkyClustClient

client = SkyClustClient(
    base_url='http://localhost:8080/api/v1',
    token='your-jwt-token'
)

# 워크스페이스 목록 조회
workspaces = client.workspaces.list()

# VM 생성
vm = client.vms.create({
    'name': 'my-vm',
    'provider': 'aws',
    'instance_type': 't3.micro',
    'region': 'us-east-1'
})
```

### Go

```go
package main

import (
    "github.com/skyclust/sdk-go"
)

client := skyclust.NewClient("http://localhost:8080/api/v1", "your-jwt-token")

// 워크스페이스 목록 조회
workspaces, err := client.Workspaces.List()

// VM 생성
vm, err := client.VMs.Create(skyclust.CreateVMRequest{
    Name: "my-vm",
    Provider: "aws",
    InstanceType: "t3.micro",
    Region: "us-east-1",
})
```
# RBAC API 문서

## 개요

이 문서는 SkyClust의 Role-Based Access Control (RBAC) 시스템 API에 대해 설명합니다.

## 인증 및 권한

### JWT 토큰
모든 보호된 엔드포인트는 JWT 토큰을 `Authorization: Bearer <token>` 헤더로 전달해야 합니다.

### 역할 (Roles)
- **admin**: 시스템 관리자 권한
- **user**: 일반 사용자 권한  
- **viewer**: 읽기 전용 권한

### 권한 (Permissions)
- **user:create**: 사용자 생성
- **user:read**: 사용자 조회
- **user:update**: 사용자 수정
- **user:delete**: 사용자 삭제
- **user:manage**: 사용자 관리
- **system:read**: 시스템 정보 조회
- **system:update**: 시스템 설정 수정
- **system:manage**: 시스템 관리
- **audit:read**: 감사 로그 조회
- **audit:export**: 감사 로그 내보내기
- **audit:manage**: 감사 로그 관리
- **workspace:create**: 워크스페이스 생성
- **workspace:read**: 워크스페이스 조회
- **workspace:update**: 워크스페이스 수정
- **workspace:delete**: 워크스페이스 삭제
- **workspace:manage**: 워크스페이스 관리
- **provider:read**: 프로바이더 조회
- **provider:manage**: 프로바이더 관리

## 사용자 관리 API

### 사용자 등록
```http
POST /api/v1/auth/register
Content-Type: application/json

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
  "data": {
    "token": "string",
    "user": {
      "id": "uuid",
      "username": "string",
      "email": "string",
      "is_active": true,
      "created_at": "datetime",
      "updated_at": "datetime"
    }
  },
  "message": "User registered successfully"
}
```

**특별 사항:**
- 첫 번째 사용자는 자동으로 `admin` 역할을 받습니다
- 이후 사용자는 `user` 역할을 받습니다

### 사용자 로그인
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "string",
  "password": "string"
}
```

**응답:**
```json
{
  "success": true,
  "data": {
    "token": "string",
    "user": {
      "id": "uuid",
      "username": "string",
      "email": "string",
      "is_active": true,
      "created_at": "datetime",
      "updated_at": "datetime"
    }
  },
  "message": "Login successful"
}
```

### 사용자 목록 조회 (관리자만)
```http
GET /api/v1/users
Authorization: Bearer <token>
```

**쿼리 파라미터:**
- `limit`: 페이지 크기 (기본값: 10)
- `offset`: 오프셋 (기본값: 0)
- `search`: 검색어
- `role`: 역할 필터

**응답:**
```json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": "uuid",
        "username": "string",
        "email": "string",
        "is_active": true,
        "created_at": "datetime",
        "updated_at": "datetime"
      }
    ],
    "pagination": {
      "total": 100,
      "limit": 10,
      "offset": 0,
      "current_page": 1,
      "total_pages": 10
    }
  },
  "message": "Users retrieved successfully"
}
```

**권한 요구사항:** `admin` 역할

### 현재 사용자 정보 조회
```http
GET /api/v1/auth/me
Authorization: Bearer <token>
```

**응답:**
```json
{
  "success": true,
  "data": {
    "id": "uuid",
    "username": "string",
    "email": "string",
    "is_active": true,
    "created_at": "datetime",
    "updated_at": "datetime"
  },
  "message": "User information retrieved successfully"
}
```

## 관리자 API

### 사용자 역할 할당
```http
POST /api/v1/admin/users/{user_id}/roles
Authorization: Bearer <token>
Content-Type: application/json

{
  "role": "admin|user|viewer"
}
```

**응답:**
```json
{
  "success": true,
  "data": {
    "message": "Role assigned successfully"
  },
  "message": "Role assigned successfully"
}
```

**권한 요구사항:** `admin` 역할

### 사용자 역할 제거
```http
DELETE /api/v1/admin/users/{user_id}/roles
Authorization: Bearer <token>
Content-Type: application/json

{
  "role": "admin|user|viewer"
}
```

**응답:**
```json
{
  "success": true,
  "data": {
    "message": "Role removed successfully"
  },
  "message": "Role removed successfully"
}
```

**권한 요구사항:** `admin` 역할

### 사용자 통계 조회
```http
GET /api/v1/admin/users/stats
Authorization: Bearer <token>
```

**응답:**
```json
{
  "success": true,
  "data": {
    "total_users": 100,
    "active_users": 95,
    "inactive_users": 5,
    "new_users_today": 3,
    "role_distribution": {
      "admin": 2,
      "user": 93,
      "viewer": 5
    },
    "last_updated": "2025-10-20T10:28:15Z"
  },
  "message": "User statistics retrieved successfully"
}
```

**권한 요구사항:** `admin` 역할

## 시스템 관리 API

### 시스템 상태 조회
```http
GET /api/v1/admin/system/status
Authorization: Bearer <token>
```

**응답:**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": "2025-10-20T10:28:15Z",
    "uptime": "72h30m15s",
    "version": "1.0.0",
    "environment": "production",
    "metrics": {
      "memory": {
        "alloc_mb": 25.5,
        "total_alloc_mb": 1024.0,
        "sys_mb": 512.0,
        "num_gc": 150
      },
      "goroutines": 45
    }
  },
  "message": "System status retrieved successfully"
}
```

**권한 요구사항:** `admin` 역할

### 시스템 설정 조회
```http
GET /api/v1/admin/system/config
Authorization: Bearer <token>
```

**응답:**
```json
{
  "success": true,
  "data": {
    "server": {
      "port": 8081,
      "host": "0.0.0.0"
    },
    "database": {
      "host": "localhost",
      "port": 5432,
      "name": "skyclust"
    },
    "redis": {
      "host": "localhost",
      "port": 6379
    }
  },
  "message": "System configuration retrieved successfully"
}
```

**권한 요구사항:** `admin` 역할

### 시스템 설정 업데이트
```http
PUT /api/v1/admin/system/config
Authorization: Bearer <token>
Content-Type: application/json

{
  "server": {
    "port": 8082
  }
}
```

**응답:**
```json
{
  "success": true,
  "data": {
    "message": "System configuration updated successfully",
    "config": {
      "server": {
        "port": 8082
      }
    }
  },
  "message": "System configuration updated successfully"
}
```

**권한 요구사항:** `admin` 역할

## 감사 로그 API

### 감사 로그 조회
```http
GET /api/v1/admin/audit/
Authorization: Bearer <token>
```

**쿼리 파라미터:**
- `limit`: 페이지 크기 (기본값: 50)
- `offset`: 오프셋 (기본값: 0)
- `action`: 액션 필터
- `user_id`: 사용자 ID 필터
- `start_date`: 시작 날짜
- `end_date`: 종료 날짜

**응답:**
```json
{
  "success": true,
  "data": {
    "logs": [
      {
        "id": "uuid",
        "user_id": "uuid",
        "action": "user_login",
        "resource": "auth",
        "ip_address": "192.168.1.100",
        "user_agent": "Mozilla/5.0...",
        "details": {
          "username": "admin"
        },
        "created_at": "2025-10-20T10:28:15Z"
      }
    ],
    "pagination": {
      "total": 1000,
      "limit": 50,
      "offset": 0,
      "current_page": 1,
      "total_pages": 20
    }
  },
  "message": "Audit logs retrieved successfully"
}
```

**권한 요구사항:** `admin` 역할

## 에러 응답

### 권한 부족 (403 Forbidden)
```json
{
  "success": false,
  "data": {},
  "error": "Insufficient permissions - admin role required",
  "code": "FORBIDDEN",
  "request_id": "uuid",
  "timestamp": "2025-10-20T10:28:15Z"
}
```

### 인증 실패 (401 Unauthorized)
```json
{
  "success": false,
  "data": {},
  "error": "Authorization header required",
  "code": "UNAUTHORIZED",
  "request_id": "uuid",
  "timestamp": "2025-10-20T10:28:15Z"
}
```

### 잘못된 토큰 (401 Unauthorized)
```json
{
  "success": false,
  "data": {},
  "error": "Invalid token",
  "code": "UNAUTHORIZED",
  "request_id": "uuid",
  "timestamp": "2025-10-20T10:28:15Z"
}
```

## 사용 예시

### 1. 첫 번째 사용자 등록 (관리자)
```bash
curl -X POST http://localhost:8081/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "securepassword123"
  }'
```

### 2. 일반 사용자 등록
```bash
curl -X POST http://localhost:8081/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user1",
    "email": "user1@example.com",
    "password": "password123"
  }'
```

### 3. 관리자로 사용자 목록 조회
```bash
curl -X GET http://localhost:8081/api/v1/users \
  -H "Authorization: Bearer <admin_token>"
```

### 4. 사용자에게 역할 할당
```bash
curl -X POST http://localhost:8081/api/v1/admin/users/{user_id}/roles \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"role": "admin"}'
```

### 5. 시스템 상태 확인
```bash
curl -X GET http://localhost:8081/api/v1/admin/system/status \
  -H "Authorization: Bearer <admin_token>"
```

## 보안 고려사항

1. **JWT 토큰 보안**: 토큰은 안전하게 저장하고 전송해야 합니다
2. **HTTPS 사용**: 프로덕션 환경에서는 반드시 HTTPS를 사용하세요
3. **토큰 만료**: 토큰은 적절한 시간 후 만료됩니다
4. **권한 최소화**: 사용자에게 필요한 최소한의 권한만 부여하세요
5. **감사 로깅**: 모든 관리 작업은 감사 로그에 기록됩니다

## 마이그레이션

기존 시스템에서 RBAC로 마이그레이션하려면 `scripts/migrate_to_rbac.sql` 스크립트를 실행하세요:

```bash
psql -d skyclust -f scripts/migrate_to_rbac.sql
```

이 스크립트는:
- `user_roles` 테이블 생성
- `role_permissions` 테이블 생성
- 기본 권한 데이터 삽입
- 기존 사용자를 RBAC 시스템으로 마이그레이션
- 첫 번째 사용자를 관리자로 설정

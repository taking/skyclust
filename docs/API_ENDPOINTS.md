# SkyClust API 엔드포인트 목록

이 문서는 SkyClust 프로젝트의 모든 API 엔드포인트를 정리한 것입니다.

## 📋 목차

- [인증 API](#인증-api)
- [사용자 API](#사용자-api)
- [워크스페이스 API](#워크스페이스-api)
- [자격증명 API](#자격증명-api)
- [VM API](#vm-api)
- [프로바이더 API](#프로바이더-api)
- [IaC API](#iac-api)
- [OIDC API](#oidc-api)
- [시스템 API](#시스템-api)

---

## 인증 API

### 회원가입
- **POST** `/api/v1/auth/register`
- **설명**: 새 사용자 계정을 생성합니다.
- **요청 본문**:
  ```json
  {
    "email": "user@example.com",
    "password": "password123",
    "name": "사용자 이름"
  }
  ```
- **응답**: `201 Created`
  ```json
  {
    "success": true,
    "data": {
      "id": "user-uuid",
      "email": "user@example.com",
      "name": "사용자 이름",
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
  ```

### 로그인
- **POST** `/api/v1/auth/login`
- **설명**: 사용자 인증을 수행하고 JWT 토큰을 반환합니다.
- **요청 본문**:
  ```json
  {
    "email": "user@example.com",
    "password": "password123"
  }
  ```
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": {
      "token": "jwt-token",
      "expires_at": "2024-01-02T00:00:00Z",
      "user": {
        "id": "user-uuid",
        "email": "user@example.com",
        "name": "사용자 이름"
      }
    }
  }
  ```

### 로그아웃
- **POST** `/api/v1/auth/logout`
- **설명**: 사용자 로그아웃을 수행합니다.
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "message": "로그아웃되었습니다."
  }
  ```

---

## 사용자 API

### 사용자 조회
- **GET** `/api/v1/users/:id`
- **설명**: 특정 사용자 정보를 조회합니다.
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": {
      "id": "user-uuid",
      "email": "user@example.com",
      "name": "사용자 이름",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  }
  ```

### 사용자 정보 수정
- **PUT** `/api/v1/users/:id`
- **설명**: 사용자 정보를 수정합니다.
- **헤더**: `Authorization: Bearer <token>`
- **요청 본문**:
  ```json
  {
    "name": "새로운 이름",
    "email": "new@example.com"
  }
  ```
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": {
      "id": "user-uuid",
      "email": "new@example.com",
      "name": "새로운 이름",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  }
  ```

### 사용자 삭제
- **DELETE** `/api/v1/users/:id`
- **설명**: 사용자 계정을 삭제합니다.
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "message": "사용자가 삭제되었습니다."
  }
  ```

---

## 워크스페이스 API

### 워크스페이스 생성
- **POST** `/api/v1/workspaces`
- **설명**: 새 워크스페이스를 생성합니다.
- **헤더**: `Authorization: Bearer <token>`
- **요청 본문**:
  ```json
  {
    "name": "워크스페이스 이름",
    "description": "워크스페이스 설명"
  }
  ```
- **응답**: `201 Created`
  ```json
  {
    "success": true,
    "data": {
      "id": "workspace-uuid",
      "name": "워크스페이스 이름",
      "description": "워크스페이스 설명",
      "owner_id": "user-uuid",
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
  ```

### 워크스페이스 조회
- **GET** `/api/v1/workspaces/:id`
- **설명**: 특정 워크스페이스 정보를 조회합니다.
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": {
      "id": "workspace-uuid",
      "name": "워크스페이스 이름",
      "description": "워크스페이스 설명",
      "owner_id": "user-uuid",
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
  ```

### 워크스페이스 목록 조회
- **GET** `/api/v1/workspaces`
- **설명**: 사용자의 워크스페이스 목록을 조회합니다.
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": [
      {
        "id": "workspace-uuid",
        "name": "워크스페이스 이름",
        "description": "워크스페이스 설명",
        "owner_id": "user-uuid",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
  ```

### 워크스페이스 수정
- **PUT** `/api/v1/workspaces/:id`
- **설명**: 워크스페이스 정보를 수정합니다.
- **헤더**: `Authorization: Bearer <token>`
- **요청 본문**:
  ```json
  {
    "name": "수정된 이름",
    "description": "수정된 설명"
  }
  ```
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": {
      "id": "workspace-uuid",
      "name": "수정된 이름",
      "description": "수정된 설명",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  }
  ```

### 워크스페이스 삭제
- **DELETE** `/api/v1/workspaces/:id`
- **설명**: 워크스페이스를 삭제합니다.
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "message": "워크스페이스가 삭제되었습니다."
  }
  ```

### 멤버 추가
- **POST** `/api/v1/workspaces/:id/members`
- **설명**: 워크스페이스에 멤버를 추가합니다.
- **헤더**: `Authorization: Bearer <token>`
- **요청 본문**:
  ```json
  {
    "user_id": "user-uuid",
    "role": "member"
  }
  ```
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "message": "멤버가 추가되었습니다."
  }
  ```

### 멤버 제거
- **DELETE** `/api/v1/workspaces/:id/members/:userID`
- **설명**: 워크스페이스에서 멤버를 제거합니다.
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "message": "멤버가 제거되었습니다."
  }
  ```

---

## 자격증명 API

### 자격증명 생성
- **POST** `/api/v1/credentials`
- **설명**: 새 클라우드 자격증명을 생성합니다.
- **헤더**: `Authorization: Bearer <token>`
- **요청 본문**:
  ```json
  {
    "workspace_id": "workspace-uuid",
    "provider": "aws",
    "credentials": {
      "access_key": "AKIA...",
      "secret_key": "...",
      "region": "us-east-1"
    }
  }
  ```
- **응답**: `201 Created`
  ```json
  {
    "success": true,
    "data": {
      "id": "credential-uuid",
      "workspace_id": "workspace-uuid",
      "provider": "aws",
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
  ```

### 자격증명 목록 조회
- **GET** `/api/v1/credentials`
- **설명**: 워크스페이스의 자격증명 목록을 조회합니다.
- **헤더**: `Authorization: Bearer <token>`
- **쿼리 파라미터**: `workspace_id`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": [
      {
        "id": "credential-uuid",
        "workspace_id": "workspace-uuid",
        "provider": "aws",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
  ```

### 자격증명 조회
- **GET** `/api/v1/credentials/:id`
- **설명**: 특정 자격증명을 조회합니다.
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": {
      "id": "credential-uuid",
      "workspace_id": "workspace-uuid",
      "provider": "aws",
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
  ```

### 자격증명 수정
- **PUT** `/api/v1/credentials/:id`
- **설명**: 자격증명을 수정합니다.
- **헤더**: `Authorization: Bearer <token>`
- **요청 본문**:
  ```json
  {
    "credentials": {
      "access_key": "NEW_AKIA...",
      "secret_key": "new_secret...",
      "region": "us-west-2"
    }
  }
  ```
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": {
      "id": "credential-uuid",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  }
  ```

### 자격증명 삭제
- **DELETE** `/api/v1/credentials/:id`
- **설명**: 자격증명을 삭제합니다.
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "message": "자격증명이 삭제되었습니다."
  }
  ```

---

## VM API

### VM 생성
- **POST** `/api/v1/vms`
- **설명**: 새 가상머신을 생성합니다.
- **헤더**: `Authorization: Bearer <token>`
- **요청 본문**:
  ```json
  {
    "workspace_id": "workspace-uuid",
    "name": "vm-name",
    "provider": "aws",
    "instance_type": "t3.micro",
    "region": "us-east-1",
    "image_id": "ami-12345678"
  }
  ```
- **응답**: `201 Created`
  ```json
  {
    "success": true,
    "data": {
      "id": "vm-uuid",
      "workspace_id": "workspace-uuid",
      "name": "vm-name",
      "provider": "aws",
      "instance_id": "i-1234567890abcdef0",
      "status": "running",
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
  ```

### VM 목록 조회
- **GET** `/api/v1/vms`
- **설명**: 워크스페이스의 VM 목록을 조회합니다.
- **헤더**: `Authorization: Bearer <token>`
- **쿼리 파라미터**: `workspace_id`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": [
      {
        "id": "vm-uuid",
        "workspace_id": "workspace-uuid",
        "name": "vm-name",
        "provider": "aws",
        "instance_id": "i-1234567890abcdef0",
        "status": "running",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
  ```

### VM 조회
- **GET** `/api/v1/vms/:id`
- **설명**: 특정 VM 정보를 조회합니다.
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": {
      "id": "vm-uuid",
      "workspace_id": "workspace-uuid",
      "name": "vm-name",
      "provider": "aws",
      "instance_id": "i-1234567890abcdef0",
      "status": "running",
      "public_ip": "1.2.3.4",
      "private_ip": "10.0.1.100",
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
  ```

### VM 시작
- **POST** `/api/v1/vms/:id/start`
- **설명**: VM을 시작합니다.
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "message": "VM이 시작되었습니다."
  }
  ```

### VM 중지
- **POST** `/api/v1/vms/:id/stop`
- **설명**: VM을 중지합니다.
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "message": "VM이 중지되었습니다."
  }
  ```

### VM 삭제
- **DELETE** `/api/v1/vms/:id`
- **설명**: VM을 삭제합니다.
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "message": "VM이 삭제되었습니다."
  }
  ```

---

## 프로바이더 API

### 프로바이더 목록 조회
- **GET** `/api/v1/providers`
- **설명**: 사용 가능한 클라우드 프로바이더 목록을 조회합니다.
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": {
      "providers": [
        {
          "name": "AWS",
          "version": "1.0.0"
        },
        {
          "name": "GCP",
          "version": "1.0.0"
        }
      ]
    }
  }
  ```

### 프로바이더 조회
- **GET** `/api/v1/providers/:name`
- **설명**: 특정 프로바이더 정보를 조회합니다.
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": {
      "name": "AWS",
      "version": "1.0.0"
    }
  }
  ```

### 인스턴스 목록 조회
- **GET** `/api/v1/providers/:name/instances`
- **설명**: 프로바이더의 인스턴스 목록을 조회합니다.
- **헤더**: `Authorization: Bearer <token>`
- **쿼리 파라미터**: `region` (선택사항)
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": [
      {
        "id": "i-1234567890abcdef0",
        "name": "instance-name",
        "status": "running",
        "type": "t3.micro",
        "region": "us-east-1",
        "public_ip": "1.2.3.4",
        "private_ip": "10.0.1.100",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
  ```

### 인스턴스 조회
- **GET** `/api/v1/providers/:name/instances/:instanceID`
- **설명**: 특정 인스턴스 정보를 조회합니다.
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": {
      "id": "i-1234567890abcdef0",
      "name": "instance-name",
      "status": "running",
      "type": "t3.micro",
      "region": "us-east-1",
      "public_ip": "1.2.3.4",
      "private_ip": "10.0.1.100",
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
  ```

### 리전 목록 조회
- **GET** `/api/v1/providers/:name/regions`
- **설명**: 프로바이더의 리전 목록을 조회합니다.
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": [
      {
        "name": "us-east-1",
        "display_name": "US East (N. Virginia)"
      },
      {
        "name": "us-west-2",
        "display_name": "US West (Oregon)"
      }
    ]
  }
  ```

### 비용 추정 조회
- **GET** `/api/v1/providers/:name/cost-estimates`
- **설명**: 프로바이더의 비용 추정 목록을 조회합니다.
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": []
  }
  ```

### 비용 추정 생성
- **POST** `/api/v1/providers/:name/cost-estimates`
- **설명**: 새 비용 추정을 생성합니다.
- **헤더**: `Authorization: Bearer <token>`
- **요청 본문**:
  ```json
  {
    "instance_type": "t3.micro",
    "region": "us-east-1",
    "duration_hours": 720
  }
  ```
- **응답**: `201 Created`
  ```json
  {
    "success": true,
    "data": {
      "id": "estimate-uuid",
      "instance_type": "t3.micro",
      "region": "us-east-1",
      "estimated_cost": 8.64,
      "currency": "USD",
      "created_at": "2024-01-01T00:00:00Z"
    }
  }
  ```

---

## IaC API

### OpenTofu 계획
- **POST** `/api/v1/iac/plan`
- **설명**: OpenTofu 계획을 실행합니다.
- **헤더**: `Authorization: Bearer <token>`
- **요청 본문**:
  ```json
  {
    "workspace_id": "workspace-uuid",
    "config": "terraform configuration"
  }
  ```
- **응답**: `201 Created`
  ```json
  {
    "success": true,
    "data": {
      "id": "execution-uuid",
      "workspace_id": "workspace-uuid",
      "command": "plan",
      "status": "running",
      "started_at": "2024-01-01T00:00:00Z"
    }
  }
  ```

### OpenTofu 적용
- **POST** `/api/v1/iac/apply`
- **설명**: OpenTofu 적용을 실행합니다.
- **헤더**: `Authorization: Bearer <token>`
- **요청 본문**:
  ```json
  {
    "workspace_id": "workspace-uuid",
    "config": "terraform configuration"
  }
  ```
- **응답**: `201 Created`
  ```json
  {
    "success": true,
    "data": {
      "id": "execution-uuid",
      "workspace_id": "workspace-uuid",
      "command": "apply",
      "status": "running",
      "started_at": "2024-01-01T00:00:00Z"
    }
  }
  ```

### OpenTofu 파괴
- **POST** `/api/v1/iac/destroy`
- **설명**: OpenTofu 파괴를 실행합니다.
- **헤더**: `Authorization: Bearer <token>`
- **요청 본문**:
  ```json
  {
    "workspace_id": "workspace-uuid",
    "config": "terraform configuration"
  }
  ```
- **응답**: `201 Created`
  ```json
  {
    "success": true,
    "data": {
      "id": "execution-uuid",
      "workspace_id": "workspace-uuid",
      "command": "destroy",
      "status": "running",
      "started_at": "2024-01-01T00:00:00Z"
    }
  }
  ```

### 실행 목록 조회
- **GET** `/api/v1/iac/executions`
- **설명**: OpenTofu 실행 목록을 조회합니다.
- **헤더**: `Authorization: Bearer <token>`
- **쿼리 파라미터**: `workspace_id`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": [
      {
        "id": "execution-uuid",
        "workspace_id": "workspace-uuid",
        "command": "plan",
        "status": "completed",
        "started_at": "2024-01-01T00:00:00Z",
        "completed_at": "2024-01-01T00:05:00Z"
      }
    ]
  }
  ```

### 실행 조회
- **GET** `/api/v1/iac/executions/:id`
- **설명**: 특정 OpenTofu 실행 정보를 조회합니다.
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": {
      "id": "execution-uuid",
      "workspace_id": "workspace-uuid",
      "command": "plan",
      "status": "completed",
      "output": "Plan: 2 to add, 0 to change, 0 to destroy.",
      "error": "",
      "started_at": "2024-01-01T00:00:00Z",
      "completed_at": "2024-01-01T00:05:00Z"
    }
  }
  ```

---

## OIDC API

### OIDC 인증 URL
- **GET** `/api/v1/oidc/:provider/auth`
- **설명**: OIDC 프로바이더의 인증 URL을 생성합니다.
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": {
      "auth_url": "https://accounts.google.com/oauth/authorize?...",
      "state": "random-state-string"
    }
  }
  ```

### OIDC 콜백 처리
- **POST** `/api/v1/oidc/:provider/callback`
- **설명**: OIDC 인증 콜백을 처리합니다.
- **헤더**: `Authorization: Bearer <token>`
- **요청 본문**:
  ```json
  {
    "code": "authorization-code",
    "state": "random-state-string"
  }
  ```
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": {
      "token": "jwt-token",
      "user": {
        "id": "user-uuid",
        "email": "user@example.com",
        "name": "사용자 이름"
      }
    }
  }
  ```

---

## 시스템 API

### 헬스 체크
- **GET** `/health`
- **설명**: 시스템 상태를 확인합니다.
- **응답**: `200 OK`
  ```json
  {
    "status": "healthy",
    "timestamp": "2024-01-01T00:00:00Z",
    "version": "1.0.0"
  }
  ```

### 설정 조회
- **GET** `/debug/config`
- **설명**: 현재 설정을 조회합니다. (디버그용)
- **헤더**: `Authorization: Bearer <token>`
- **응답**: `200 OK`
  ```json
  {
    "success": true,
    "data": {
      "database": {
        "host": "localhost",
        "port": 5432,
        "name": "skyclust"
      },
      "redis": {
        "host": "localhost",
        "port": 6379
      }
    }
  }
  ```

---

## 🔧 공통 응답 형식

### 성공 응답
```json
{
  "success": true,
  "data": { ... },
  "message": "성공 메시지" // 선택사항
}
```

### 에러 응답
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "에러 메시지",
    "details": { ... } // 선택사항
  }
}
```

### 에러 코드
- `VALIDATION_ERROR`: 입력 데이터 검증 실패
- `UNAUTHORIZED`: 인증 실패
- `FORBIDDEN`: 권한 없음
- `NOT_FOUND`: 리소스 없음
- `CONFLICT`: 리소스 충돌
- `INTERNAL_ERROR`: 내부 서버 오류

---

## 🔐 인증

대부분의 API는 JWT 토큰 인증이 필요합니다.

### 헤더 형식
```
Authorization: Bearer <jwt-token>
```

### 토큰 획득
1. `/api/v1/auth/login` 엔드포인트로 로그인
2. 응답에서 `token` 필드 사용
3. 모든 API 호출 시 `Authorization` 헤더에 포함

---

## 📝 참고사항

- 모든 API는 JSON 형식을 사용합니다.
- 날짜/시간은 ISO 8601 형식(UTC)을 사용합니다.
- 페이지네이션은 `limit`과 `offset` 쿼리 파라미터를 사용합니다.
- 필터링은 쿼리 파라미터를 통해 수행됩니다.
- 모든 API 호출은 감사 로그에 기록됩니다.

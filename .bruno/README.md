# SkyClust API Bruno Collection

이 폴더는 [Bruno](https://www.usebruno.com/) API 클라이언트를 위한 SkyClust API 테스트 컬렉션입니다.

## 📁 구조

```
.bruno/
├── bruno.json                 # 메인 컬렉션 파일
├── environments/              # 환경 변수
│   ├── Development.bru       # 개발 환경
│   └── Production.bru        # 프로덕션 환경
└── [폴더별 API 요청들]
```

## 🚀 사용 방법

1. **Bruno 설치**: [Bruno 공식 사이트](https://www.usebruno.com/)에서 다운로드
2. **컬렉션 열기**: Bruno에서 이 `.bruno` 폴더를 열기
3. **환경 설정**: Development 또는 Production 환경 선택
4. **API 테스트**: 각 폴더의 요청들을 실행

## 📋 API 엔드포인트

### 🔐 인증 (Authentication)
- 사용자 등록/로그인/로그아웃
- OIDC (Google, GitHub, Azure AD) 인증

### ☁️ 클라우드 프로바이더 (Cloud Providers)
- AWS, GCP, Azure 등 클라우드 프로바이더 관리
- 인스턴스 조회 및 관리

### 🏢 워크스페이스 (Workspaces)
- 워크스페이스 CRUD 작업
- 사용자별 워크스페이스 관리

### 🔑 자격증명 (Credentials)
- 클라우드 자격증명 관리
- 암호화된 저장 및 관리

### 💰 비용 분석 (Cost Analysis)
- 비용 분석 및 예측
- 비용 추정 및 최적화

### 🔔 알림 (Notifications)
- 실시간 알림 시스템
- 알림 설정 및 관리

### 📊 내보내기 (Exports)
- 데이터 내보내기 (CSV, JSON)
- 내보내기 상태 추적

### 📡 실시간 통신 (SSE)
- Server-Sent Events를 통한 실시간 모니터링
- 실시간 알림 스트림

## 🔧 환경 변수

### Development
- `baseUrl`: http://localhost:8081
- `apiVersion`: v1

### Production
- `baseUrl`: http://localhost:8080
- `apiVersion`: v1

## 📝 사용 예시

1. **인증 플로우**:
   - User Registration → User Login → API 요청들

2. **워크스페이스 관리**:
   - Create Workspace → Get Workspaces → Update/Delete

3. **클라우드 관리**:
   - Create Credential → Get Providers → Get Instances

## 🧪 테스트

각 요청에는 기본적인 테스트가 포함되어 있습니다:
- HTTP 상태 코드 검증
- 응답 구조 검증
- 필수 필드 존재 확인

## 📚 참고 자료

- [Bruno 공식 문서](https://docs.usebruno.com/)
- [SkyClust API 문서](./docs/API.md)
- [환경 설정 가이드](./docs/CONFIGURATION.md)

# GCP IAM 정책 설정 가이드

이 문서는 SkyClust 플랫폼에서 GCP 자격증명을 설정하기 위한 프로젝트, API, 서비스 계정 및 IAM 역할 설정 방법을 단계별로 안내합니다.

---

## 목차

1. [프로젝트 생성](#1-프로젝트-생성)
2. [필수 API 활성화](#2-필수-api-활성화)
3. [서비스 계정 생성](#3-서비스-계정-생성)
4. [IAM 역할 부여](#4-iam-역할-부여)
5. [서비스 계정 키 생성](#5-서비스-계정-키-생성)
6. [설정 확인](#6-설정-확인)

---

## 1. 프로젝트 생성

### 1.1 GCP Console 접속

1. [GCP Console](https://console.cloud.google.com/)에 로그인
2. 상단의 프로젝트 선택 드롭다운 클릭

### 1.2 새 프로젝트 생성

1. **새 프로젝트(NEW PROJECT)** 버튼 클릭

### 1.3 프로젝트 정보 입력

1. **프로젝트 이름(Project name)** 입력: `skyclust-project`
   - 프로젝트 ID는 자동으로 생성됩니다
2. **조직(Organization)** 선택 (선택사항)
   - 조직이 있는 경우 선택
3. **위치(Location)** 선택 (선택사항)
   - 조직의 폴더 또는 조직 자체 선택

### 1.4 프로젝트 생성

1. **만들기(CREATE)** 버튼 클릭
2. 프로젝트 생성 완료까지 몇 초 소요될 수 있습니다

### 1.5 프로젝트 ID 확인

1. 프로젝트 선택 드롭다운에서 생성된 프로젝트 선택
2. 프로젝트 ID 확인 (예: `skyclust-project-123456`)
   - 이 프로젝트 ID는 나중에 사용됩니다

---

## 2. 필수 API 활성화

SkyClust가 GCP 서비스를 사용하기 위해 필요한 API를 활성화합니다.

### 방법 1: GCP Console 사용

#### 2.1 API 및 서비스 메뉴 이동

1. 왼쪽 햄버거 메뉴 클릭
2. **API 및 서비스(APIs & Services)** > **라이브러리(Library)** 선택

#### 2.2 Kubernetes Engine API 활성화

1. 검색창에 **Kubernetes Engine API** 입력
2. **Kubernetes Engine API** 선택
3. **사용 설정(ENABLE)** 버튼 클릭
4. 활성화 완료까지 대기

#### 2.3 Compute Engine API 활성화

1. **라이브러리(Library)**로 돌아가기
2. 검색창에 **Compute Engine API** 입력
3. **Compute Engine API** 선택
4. **사용 설정(ENABLE)** 버튼 클릭
5. 활성화 완료까지 대기

#### 2.4 Identity and Access Management API 활성화

1. **라이브러리(Library)**로 돌아가기
2. 검색창에 **Identity and Access Management API** 입력
3. **Identity and Access Management (IAM) API** 선택
4. **사용 설정(ENABLE)** 버튼 클릭
5. 활성화 완료까지 대기

#### 2.5 Service Usage API 활성화

1. **라이브러리(Library)**로 돌아가기
2. 검색창에 **Service Usage API** 입력
3. **Service Usage API** 선택
4. **사용 설정(ENABLE)** 버튼 클릭
5. 활성화 완료까지 대기

#### 2.6 Cloud Billing API 활성화 (비용 분석 기능용)

1. **라이브러리(Library)**로 돌아가기
2. 검색창에 **Cloud Billing API** 입력
3. **Cloud Billing API** 선택
4. **사용 설정(ENABLE)** 버튼 클릭
5. 활성화 완료까지 대기

### 방법 2: gcloud CLI 사용

터미널에서 다음 명령어를 실행합니다:

```bash
# 프로젝트 설정
gcloud config set project <PROJECT_ID>

# 필수 API 활성화
gcloud services enable container.googleapis.com
gcloud services enable compute.googleapis.com
gcloud services enable iam.googleapis.com
gcloud services enable serviceusage.googleapis.com
gcloud services enable cloudbilling.googleapis.com

# 활성화 확인
gcloud services list --enabled
```

---

## 3. 서비스 계정 생성

### 3.1 서비스 계정 메뉴 이동

1. 왼쪽 햄버거 메뉴 클릭
2. **IAM 및 관리자(IAM & Admin)** > **서비스 계정(Service Accounts)** 선택

### 3.2 서비스 계정 생성 시작

1. **서비스 계정 만들기(CREATE SERVICE ACCOUNT)** 버튼 클릭

### 3.3 서비스 계정 정보 입력

1. **서비스 계정 이름(Service account name)** 입력: `skyclust-service-account`
2. **서비스 계정 ID(Service account ID)** 확인
   - 자동으로 `skyclust-service-account`로 생성됩니다
3. **설명(Description)** 입력 (선택사항): `SkyClust application service account for GKE and network management`
4. **만들고 계속하기(CREATE AND CONTINUE)** 버튼 클릭

### 3.4 역할 부여 (선택사항 - 여기서는 건너뛰고 다음 단계에서 부여)

1. 이 단계에서는 역할을 부여하지 않고 **건너뛰기(SKIP)** 버튼 클릭
   - 역할은 다음 단계에서 상세하게 부여합니다

### 3.5 서비스 계정 생성 완료

1. **완료(DONE)** 버튼 클릭
2. 생성된 서비스 계정이 목록에 표시됩니다

---

## 4. IAM 역할 부여

서비스 계정에 필요한 IAM 역할을 부여합니다.

### 방법 1: GCP Console 사용

#### 4.1 IAM 메뉴 이동

1. 왼쪽 햄버거 메뉴 클릭
2. **IAM 및 관리자(IAM & Admin)** > **IAM** 선택

#### 4.2 권한 부여 시작

1. **권한 부여(GRANT ACCESS)** 버튼 클릭

#### 4.3 구성원 추가

1. **새 구성원(New principals)** 필드에 서비스 계정 이메일 입력:
   - 형식: `skyclust-service-account@<PROJECT_ID>.iam.gserviceaccount.com`
   - 예: `skyclust-service-account@skyclust-project-123456.iam.gserviceaccount.com`

#### 4.4 역할 선택

**역할 선택(Select a role)** 드롭다운에서 다음 역할들을 하나씩 추가:

1. **Kubernetes Engine 관리자** (`roles/container.admin`)
   - 역할 선택 드롭다운 클릭
   - 검색창에 `Kubernetes Engine 관리자` 또는 `container.admin` 입력
   - 선택 후 **역할 추가(ADD ANOTHER ROLE)** 클릭

2. **Compute 네트워크 관리자** (`roles/compute.networkAdmin`)
   - 검색창에 `Compute 네트워크 관리자` 또는 `compute.networkAdmin` 입력
   - 선택 후 **역할 추가(ADD ANOTHER ROLE)** 클릭

3. **Compute 인스턴스 관리자** (`roles/compute.instanceAdmin`)
   - 검색창에 `Compute 인스턴스 관리자` 또는 `compute.instanceAdmin` 입력
   - 선택 후 **역할 추가(ADD ANOTHER ROLE)** 클릭

4. **Compute Engine 서비스 에이전트** (`roles/compute.serviceAgent`)
   - 검색창에 `Compute Engine 서비스 에이전트` 또는 `compute.serviceAgent` 입력
   - 선택 후 **역할 추가(ADD ANOTHER ROLE)** 클릭

5. **Compute 보안 관리자** (`roles/compute.securityAdmin`)
   - 검색창에 `Compute 보안 관리자` 또는 `compute.securityAdmin` 입력
   - 선택 후 **역할 추가(ADD ANOTHER ROLE)** 클릭

6. **서비스 계정 관리자** (`roles/iam.serviceAccountAdmin`)
   - 검색창에 `서비스 계정 관리자` 또는 `iam.serviceAccountAdmin` 입력
   - 선택 후 **역할 추가(ADD ANOTHER ROLE)** 클릭

7. **뷰어** (`roles/viewer`)
   - 검색창에 `뷰어` 또는 `viewer` 입력
   - 선택 후 **역할 추가(ADD ANOTHER ROLE)** 클릭

8. **서비스 사용량 관리자** (`roles/serviceusage.serviceUsageAdmin`)
   - 검색창에 `서비스 사용량 관리자` 또는 `serviceusage.serviceUsageAdmin` 입력
   - 선택

#### 4.5 권한 부여 완료

1. 모든 역할이 추가되었는지 확인
2. **저장(SAVE)** 버튼 클릭
3. 권한 부여 완료 메시지 확인

### 방법 2: gcloud CLI 사용

터미널에서 다음 명령어를 실행합니다:

```bash
# 변수 설정 (본인의 프로젝트 ID로 변경)
PROJECT_ID="your-project-id"
SERVICE_ACCOUNT="skyclust-service-account@${PROJECT_ID}.iam.gserviceaccount.com"

# Kubernetes Engine 관리자
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/container.admin"

# Compute 네트워크 관리자
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/compute.networkAdmin"

# Compute 인스턴스 관리자
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/compute.instanceAdmin"

# Compute Engine 서비스 에이전트
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/compute.serviceAgent"

# Compute 보안 관리자
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/compute.securityAdmin"

# 서비스 계정 관리자
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/iam.serviceAccountAdmin"

# 서비스 사용량 관리자
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/serviceusage.serviceUsageAdmin"

# 뷰어
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/viewer"
```

#### 역할 부여 확인

```bash
# 서비스 계정의 역할 확인
gcloud projects get-iam-policy ${PROJECT_ID} \
    --flatten="bindings[].members" \
    --filter="bindings.members:serviceAccount:${SERVICE_ACCOUNT}"
```

---

## 5. 서비스 계정 키 생성

SkyClust에서 사용할 서비스 계정 키(JSON 파일)를 생성합니다.

### 5.1 서비스 계정 선택

1. **IAM 및 관리자(IAM & Admin)** > **서비스 계정(Service Accounts)** 이동
2. `skyclust-service-account` 클릭

### 5.2 키 생성 시작

1. **키(KEYS)** 탭 클릭
2. **키 추가(ADD KEY)** 버튼 클릭
3. **새 키 만들기(Create new key)** 선택

### 5.3 키 유형 선택

1. **JSON** 선택
2. **만들기(CREATE)** 버튼 클릭

### 5.4 키 파일 다운로드

1. JSON 키 파일이 자동으로 다운로드됩니다
2. 파일 이름 형식: `skyclust-project-123456-xxxxx.json`
3. **이 파일을 안전하게 보관**하세요
   - 이 파일은 나중에 다시 다운로드할 수 없습니다
   - 보안이 중요한 정보를 포함하고 있으므로 안전한 곳에 보관하세요

### 5.5 키 파일 내용 확인

다운로드한 JSON 파일을 열어서 다음 정보가 포함되어 있는지 확인:

```json
{
  "type": "service_account",
  "project_id": "skyclust-project-123456",
  "private_key_id": "...",
  "private_key": "-----BEGIN PRIVATE KEY-----\n...",
  "client_email": "skyclust-service-account@skyclust-project-123456.iam.gserviceaccount.com",
  "client_id": "...",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "...",
  "universe_domain": "googleapis.com"
}
```

---

## 6. 설정 확인

### 6.1 프로젝트 확인

1. 프로젝트 선택 드롭다운에서 프로젝트가 올바르게 선택되어 있는지 확인
2. 프로젝트 ID 기록

### 6.2 API 활성화 확인

1. **API 및 서비스(APIs & Services)** > **사용 설정된 API(Enabled APIs)** 이동
2. 다음 API가 활성화되어 있는지 확인:
   - Kubernetes Engine API
   - Compute Engine API
   - Identity and Access Management (IAM) API
   - Service Usage API
   - Cloud Billing API

### 6.3 서비스 계정 확인

1. **IAM 및 관리자(IAM & Admin)** > **서비스 계정(Service Accounts)** 이동
2. `skyclust-service-account`가 존재하는지 확인
3. 서비스 계정 이메일 주소 기록

### 6.4 IAM 역할 확인

1. **IAM 및 관리자(IAM & Admin)** > **IAM** 이동
2. `skyclust-service-account@...` 검색
3. 다음 역할들이 부여되어 있는지 확인:
   - Kubernetes Engine 관리자
   - Compute 네트워크 관리자
   - Compute 인스턴스 관리자
   - Compute Engine 서비스 에이전트
   - Compute 보안 관리자
   - 서비스 계정 관리자
   - 뷰어
   - 서비스 사용량 관리자

### 6.5 서비스 계정 키 확인

1. 서비스 계정 > **키(KEYS)** 탭
2. 생성된 키가 있는지 확인
3. JSON 키 파일이 안전하게 보관되어 있는지 확인

---

## 다음 단계

이제 SkyClust 플랫폼에서 자격증명을 등록할 수 있습니다:

1. **프로젝트 ID**: 위에서 생성한 프로젝트 ID
2. **서비스 계정 JSON 키 파일**: 위에서 다운로드한 JSON 파일의 내용
   - 또는 JSON 파일의 내용을 직접 복사하여 사용

---

## 추가 설정 (선택사항)

### Billing 계정 연결 (비용 분석 기능용)

비용 분석 기능을 사용하려면 프로젝트에 Billing 계정을 연결해야 합니다.

1. 왼쪽 햄버거 메뉴 클릭
2. **결제(Billing)** 선택
3. **계정 연결(Link a billing account)** 클릭
4. 기존 Billing 계정 선택 또는 새로 생성
5. 프로젝트에 Billing 계정 연결 확인

---

## 문제 해결

### API 활성화 실패

- **문제**: API 활성화가 실패하거나 시간이 오래 걸림
- **해결**: 
  - 프로젝트에 Billing 계정이 연결되어 있는지 확인
  - 몇 분 후 다시 시도

### 서비스 계정 생성 실패

- **문제**: 서비스 계정 ID가 이미 사용 중
- **해결**: 다른 이름으로 서비스 계정 생성

### 권한 부족 오류

- **문제**: `PERMISSION_DENIED` 오류 발생
- **해결**: 
  - 서비스 계정에 모든 필수 역할이 부여되었는지 확인
  - IAM 역할이 올바르게 부여되었는지 확인

### 키 파일 손실

- **문제**: JSON 키 파일을 분실함
- **해결**: 
  - 기존 키를 삭제하고 새 키 생성
  - 서비스 계정 > **키(KEYS)** 탭에서 **키 추가(ADD KEY)** > **새 키 만들기(Create new key)**

---

## 참고 자료

- [GCP Kubernetes Engine 문서](https://cloud.google.com/kubernetes-engine/docs)
- [GCP IAM 역할 참조](https://cloud.google.com/iam/docs/understanding-roles)
- [GCP 서비스 계정 가이드](https://cloud.google.com/iam/docs/service-accounts)
- [GCP Cloud Billing API 문서](https://cloud.google.com/billing/docs/reference/rest)


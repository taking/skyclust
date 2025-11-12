# Azure IAM 정책 설정 가이드

이 문서는 SkyClust 플랫폼에서 Azure 자격증명을 설정하기 위한 Subscription, Service Principal 및 RBAC 역할 설정 방법을 단계별로 안내합니다.

---

## 목차

1. [Subscription 확인](#1-subscription-확인)
2. [Resource Group 생성](#2-resource-group-생성)
3. [필수 리소스 공급자 등록](#3-필수-리소스-공급자-등록)
4. [Service Principal 생성](#4-service-principal-생성)
5. [RBAC 역할 부여](#5-rbac-역할-부여)
6. [Service Principal 인증 정보 확인](#6-service-principal-인증-정보-확인)
7. [설정 확인](#7-설정-확인)

---

## 1. Subscription 확인

### 1.1 Azure Portal 접속

1. [Azure Portal](https://portal.azure.com/)에 로그인
2. 상단 검색창에 **Subscriptions (구독)** 입력 후 선택

### 1.2 Subscription 확인

1. 사용 가능한 Subscription 목록 확인
2. SkyClust에서 사용할 Subscription 선택
3. **Subscription ID (구독 ID)** 복사
   - 형식: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`
   - Subscription 페이지에서 직접 확인 가능

### 1.3 Tenant ID 확인

1. Azure Portal 상단 검색창에 **Azure Active Directory** 또는 **Microsoft Entra ID** 입력
2. **개요(Overview)** 탭 클릭
3. **테넌트 ID(Tenant ID)** 복사
   - 형식: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`

---

## 2. Resource Group 생성

Resource Group은 Azure 리소스를 논리적으로 그룹화하는 컨테이너입니다. AKS 클러스터와 네트워크 리소스를 관리하기 위해 생성하는 것을 권장합니다.

### 2.1 Resource Group 생성 시작

1. Azure Portal 상단 검색창에 **Resource groups** 입력 후 선택
2. **만들기(Create)** 버튼 클릭

### 2.2 Resource Group 정보 입력

1. **구독(Subscription)** 선택
   - 위에서 확인한 Subscription 선택
2. **리소스 그룹 이름(Resource group name)** 입력: `skyclust-rg`
3. **지역(Region)** 선택
   - 예: `Korea Central`, `East US` 등
   - AKS 클러스터를 배포할 지역 선택

### 2.3 Resource Group 생성

1. **검토 + 만들기(Review + Create)** 버튼 클릭
2. 설정 확인 후 **만들기(Create)** 버튼 클릭
3. 생성 완료까지 대기

### 2.4 Resource Group 이름 확인

1. 생성된 Resource Group 클릭
2. Resource Group 이름 기록
   - 나중에 SkyClust에서 사용됩니다

---

## 3. 필수 리소스 공급자 등록

SkyClust가 Azure 서비스를 사용하기 위해 필요한 리소스 공급자를 등록합니다.

### 방법 1: Azure Portal 사용

#### 3.1 Subscriptions 메뉴 이동

1. Azure Portal 상단 검색창에 **Subscriptions** 입력 후 선택
2. 사용할 Subscription 클릭

#### 3.2 리소스 공급자 메뉴 이동

1. 왼쪽 메뉴에서 **리소스 공급자(Resource providers)** 클릭

#### 3.3 필수 리소스 공급자 등록

다음 리소스 공급자들을 검색하여 **등록(Register)** 클릭:

1. **Microsoft.ContainerService**
   - 검색창에 `Microsoft.ContainerService` 입력
   - 선택 후 **등록(Register)** 클릭
   - 등록 완료까지 대기

2. **Microsoft.Network**
   - 검색창에 `Microsoft.Network` 입력
   - 선택 후 **등록(Register)** 클릭
   - 등록 완료까지 대기

3. **Microsoft.Compute**
   - 검색창에 `Microsoft.Compute` 입력
   - 선택 후 **등록(Register)** 클릭
   - 등록 완료까지 대기

4. **Microsoft.Resources**
   - 검색창에 `Microsoft.Resources` 입력
   - 선택 후 **등록(Register)** 클릭
   - 등록 완료까지 대기

### 방법 2: Azure CLI 사용

터미널에서 다음 명령어를 실행합니다:

```bash
# Azure에 로그인
az login

# Subscription 설정
az account set --subscription <SUBSCRIPTION_ID>

# 필수 리소스 공급자 등록
az provider register --namespace Microsoft.ContainerService
az provider register --namespace Microsoft.Network
az provider register --namespace Microsoft.Compute
az provider register --namespace Microsoft.Resources

# 등록 상태 확인
az provider list --query "[?namespace=='Microsoft.ContainerService' || namespace=='Microsoft.Network' || namespace=='Microsoft.Compute' || namespace=='Microsoft.Resources'].{Namespace:namespace, State:registrationState}" --output table
```

---

## 4. Service Principal 생성

Service Principal은 SkyClust 애플리케이션이 Azure 리소스에 접근하기 위해 사용하는 보안 주체입니다.

### 방법 1: Azure Portal 사용

#### 4.1 Azure Active Directory 메뉴 이동

1. Azure Portal 상단 검색창에 **Azure Active Directory** 또는 **Microsoft Entra ID** 입력 후 선택
2. 왼쪽 메뉴에서 **앱 등록(App registrations)** 클릭

#### 4.2 새 등록 시작

1. **새 등록(New registration)** 버튼 클릭

#### 4.3 앱 등록 정보 입력

1. **이름(Name)** 입력: `skyclust-service-principal`
2. **지원되는 계정 유형(Supported account types)** 선택:
   - **이 조직 디렉터리만의 계정(단일 테넌트)** 선택 (권장)
3. **리디렉션 URI(Redirect URI)**는 비워둡니다
4. **등록(Register)** 버튼 클릭

#### 4.4 Application (Client) ID 확인

1. 등록 완료 후 **개요(Overview)** 페이지로 이동
2. **애플리케이션(클라이언트) ID(Application (client) ID)** 복사
   - 형식: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`
   - 이것이 `client_id`입니다

#### 4.5 클라이언트 암호 생성

1. 왼쪽 메뉴에서 **인증서 및 암호(Certificates & secrets)** 클릭
2. **새 클라이언트 암호(New client secret)** 버튼 클릭
3. **설명(Description)** 입력: `SkyClust Service Principal Secret`
4. **만료(Expires)** 선택:
   - **24개월** 또는 원하는 기간 선택
5. **추가(Add)** 버튼 클릭
6. **값(Value)** 복사
   - **이 페이지를 떠나면 다시 확인할 수 없으므로 안전하게 보관**하세요
   - 이것이 `client_secret`입니다

### 방법 2: Azure CLI 사용

터미널에서 다음 명령어를 실행합니다:

```bash
# Azure에 로그인
az login

# Subscription ID 확인
SUBSCRIPTION_ID=$(az account show --query id -o tsv)
echo "Subscription ID: $SUBSCRIPTION_ID"

# Tenant ID 확인
TENANT_ID=$(az account show --query tenantId -o tsv)
echo "Tenant ID: $TENANT_ID"

# Service Principal 생성
az ad sp create-for-rbac \
  --name "skyclust-service-principal" \
  --role contributor \
  --scopes /subscriptions/$SUBSCRIPTION_ID

# 출력 예시:
# {
#   "appId": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
#   "displayName": "skyclust-service-principal",
#   "password": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
#   "tenant": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
# }
```

> **참고**: Azure CLI를 사용하면 `appId`가 `client_id`, `password`가 `client_secret`, `tenant`가 `tenant_id`입니다.

---

## 5. RBAC 역할 부여

Service Principal에 Azure 리소스를 관리하기 위한 RBAC 역할을 부여합니다.

### 방법 1: Azure Portal 사용

#### 5.1 Subscription 레벨 역할 부여

1. Azure Portal 상단 검색창에 **Subscriptions** 입력 후 선택
2. 사용할 Subscription 클릭
3. 왼쪽 메뉴에서 **액세스 제어(IAM)** 클릭
4. **역할 할당 추가(Add role assignment)** 버튼 클릭

#### 5.2 Contributor 역할 부여

1. **역할(Role)** 탭에서 **Contributor** 검색 후 선택
2. **다음(Next)** 버튼 클릭
3. **구성원(Members)** 탭에서 **+ 구성원 선택(Select members)** 클릭
4. 검색창에 `skyclust-service-principal` 입력
5. Service Principal 선택 후 **선택(Select)** 클릭
6. **검토 + 할당(Review + assign)** 버튼 클릭

#### 5.3 Resource Group 레벨 역할 부여 (선택사항)

더 세분화된 권한 관리를 원하는 경우:

1. Azure Portal에서 **Resource groups** 이동
2. `skyclust-rg` 클릭
3. 왼쪽 메뉴에서 **액세스 제어(IAM)** 클릭
4. **역할 할당 추가(Add role assignment)** 버튼 클릭
5. 다음 역할들을 하나씩 추가:

   - **Contributor** (또는 다음 역할들)
   - **Azure Kubernetes Service Cluster Admin Role** (`Azure Kubernetes Service Cluster Admin Role`)
   - **Network Contributor** (`Network Contributor`)
   - **Virtual Machine Contributor** (`Virtual Machine Contributor`)

### 방법 2: Azure CLI 사용

터미널에서 다음 명령어를 실행합니다:

```bash
# 변수 설정
SUBSCRIPTION_ID="your-subscription-id"
RESOURCE_GROUP="skyclust-rg"
SERVICE_PRINCIPAL_NAME="skyclust-service-principal"

# Service Principal의 Object ID 가져오기
SP_OBJECT_ID=$(az ad sp list --display-name $SERVICE_PRINCIPAL_NAME --query [0].id -o tsv)

# Subscription 레벨에 Contributor 역할 부여
az role assignment create \
  --assignee $SP_OBJECT_ID \
  --role "Contributor" \
  --scope /subscriptions/$SUBSCRIPTION_ID

# Resource Group 레벨에 역할 부여 (선택사항)
az role assignment create \
  --assignee $SP_OBJECT_ID \
  --role "Contributor" \
  --scope /subscriptions/$SUBSCRIPTION_ID/resourceGroups/$RESOURCE_GROUP

# Azure Kubernetes Service Cluster Admin Role 부여
az role assignment create \
  --assignee $SP_OBJECT_ID \
  --role "Azure Kubernetes Service Cluster Admin Role" \
  --scope /subscriptions/$SUBSCRIPTION_ID

# Network Contributor 역할 부여
az role assignment create \
  --assignee $SP_OBJECT_ID \
  --role "Network Contributor" \
  --scope /subscriptions/$SUBSCRIPTION_ID

# Virtual Machine Contributor 역할 부여
az role assignment create \
  --assignee $SP_OBJECT_ID \
  --role "Virtual Machine Contributor" \
  --scope /subscriptions/$SUBSCRIPTION_ID
```

#### 역할 부여 확인

```bash
# Service Principal의 역할 확인
az role assignment list \
  --assignee $SP_OBJECT_ID \
  --all \
  --output table
```

---

## 6. Service Principal 인증 정보 확인

SkyClust에서 사용할 모든 인증 정보를 확인합니다.

### 6.1 Application (Client) ID 확인

1. Azure Portal > **Azure Active Directory** > **앱 등록(App registrations)**
2. `skyclust-service-principal` 클릭
3. **개요(Overview)** 탭에서 **애플리케이션(클라이언트) ID** 복사
   - 이것이 `client_id`입니다

### 6.2 Tenant ID 확인

1. Azure Portal > **Azure Active Directory** > **개요(Overview)**
2. **테넌트 ID(Tenant ID)** 복사
   - 이것이 `tenant_id`입니다

### 6.3 Client Secret 확인

1. Azure Portal > **Azure Active Directory** > **앱 등록(App registrations)**
2. `skyclust-service-principal` 클릭
3. **인증서 및 암호(Certificates & secrets)** 탭
4. 생성한 클라이언트 암호의 **값(Value)** 확인
   - 만료되지 않은 암호의 값이 `client_secret`입니다
   - 만료된 경우 새로 생성해야 합니다

### 6.4 Subscription ID 확인

1. Azure Portal > **Subscriptions**
2. 사용할 Subscription 클릭
3. **개요(Overview)** 탭에서 **Subscription ID** 복사

### 6.5 인증 정보 요약

다음 정보를 안전하게 보관하세요:

- **Subscription ID**: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`
- **Client ID (Application ID)**: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`
- **Client Secret**: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`
- **Tenant ID**: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`
- **Resource Group** (선택사항): `skyclust-rg`

---

## 7. 설정 확인

### 7.1 Subscription 확인

1. Azure Portal > **Subscriptions**
2. Subscription이 활성 상태인지 확인
3. Subscription ID 기록

### 7.2 Resource Group 확인

1. Azure Portal > **Resource groups**
2. `skyclust-rg`가 존재하는지 확인
3. Resource Group 이름 기록

### 7.3 리소스 공급자 등록 확인

1. Azure Portal > **Subscriptions** > 사용할 Subscription
2. **리소스 공급자(Resource providers)** 메뉴
3. 다음 리소스 공급자가 **등록됨(Registered)** 상태인지 확인:
   - Microsoft.ContainerService
   - Microsoft.Network
   - Microsoft.Compute
   - Microsoft.Resources

### 7.4 Service Principal 확인

1. Azure Portal > **Azure Active Directory** > **앱 등록(App registrations)**
2. `skyclust-service-principal`이 존재하는지 확인
3. Application (Client) ID 기록

### 7.5 RBAC 역할 확인

1. Azure Portal > **Subscriptions** > 사용할 Subscription
2. **액세스 제어(IAM)** 메뉴
3. **역할 할당(Role assignments)** 탭
4. `skyclust-service-principal` 검색
5. 다음 역할들이 부여되어 있는지 확인:
   - Contributor (또는 필요한 역할들)

### 7.6 인증 테스트 (선택사항)

Azure CLI를 사용하여 Service Principal 인증이 올바르게 작동하는지 테스트:

```bash
# Service Principal로 로그인
az login --service-principal \
  --username <CLIENT_ID> \
  --password <CLIENT_SECRET> \
  --tenant <TENANT_ID>

# Subscription 확인
az account show

# 리소스 그룹 목록 확인
az group list
```

---

## 다음 단계

이제 SkyClust 플랫폼에서 자격증명을 등록할 수 있습니다:

1. **Subscription ID**: 위에서 확인한 Subscription ID
2. **Client ID**: Service Principal의 Application (Client) ID
3. **Client Secret**: Service Principal의 클라이언트 암호 값
4. **Tenant ID**: Azure Active Directory의 Tenant ID
5. **Resource Group** (선택사항): 기본으로 사용할 Resource Group 이름

---

## 추가 설정 (선택사항)

### Managed Identity 사용 (권장)

프로덕션 환경에서는 Service Principal 대신 Managed Identity를 사용하는 것을 권장합니다. Managed Identity는 암호를 관리할 필요가 없으며 더 안전합니다.

#### System-assigned Managed Identity

1. Azure Portal에서 리소스 생성 시 **관리 ID(Managed identity)** 활성화
2. 자동으로 생성된 Identity에 역할 부여

#### User-assigned Managed Identity

1. Azure Portal > **Managed identities**
2. **만들기(Create)** 버튼 클릭
3. Identity 정보 입력 후 생성
4. 필요한 역할 부여

### 비용 분석 기능 사용

비용 분석 기능을 사용하려면:

1. Azure Portal > **Cost Management + Billing**
2. **Cost Management** 메뉴에서 비용 데이터 확인
3. Service Principal에 **Cost Management Reader** 역할 부여 (필요한 경우)

---

## 문제 해결

### 리소스 공급자 등록 실패

- **문제**: 리소스 공급자 등록이 실패하거나 시간이 오래 걸림
- **해결**: 
  - Subscription이 활성 상태인지 확인
  - 몇 분 후 다시 시도
  - Azure CLI를 사용하여 등록 상태 확인

### Service Principal 생성 실패

- **문제**: Service Principal 이름이 이미 사용 중
- **해결**: 다른 이름으로 Service Principal 생성

### 권한 부족 오류

- **문제**: `AuthorizationFailed` 또는 `Forbidden` 오류 발생
- **해결**: 
  - Service Principal에 필요한 역할이 모두 부여되었는지 확인
  - Subscription 레벨과 Resource Group 레벨 모두 확인
  - 역할 부여 후 몇 분 정도 대기 (전파 시간)

### Client Secret 만료

- **문제**: `InvalidAuthenticationTokenTenant` 오류 발생
- **해결**: 
  - Azure Portal > **Azure Active Directory** > **앱 등록** > Service Principal
  - **인증서 및 암호** 탭에서 새 클라이언트 암호 생성
  - SkyClust에서 자격증명 정보 업데이트

### AKS 클러스터 생성 실패

- **문제**: `ResourceGroupNotFound` 또는 `InvalidResourceGroup` 오류
- **해결**: 
  - Resource Group이 존재하는지 확인
  - Resource Group 이름이 올바른지 확인
  - Service Principal에 해당 Resource Group에 대한 권한이 있는지 확인

### 네트워크 리소스 생성 실패

- **문제**: `AuthorizationFailed` 오류 발생
- **해결**: 
  - Service Principal에 **Network Contributor** 역할이 부여되었는지 확인
  - Subscription 레벨 또는 Resource Group 레벨에 역할 부여

---

## 참고 자료

- [Azure Kubernetes Service (AKS) 문서](https://docs.microsoft.com/azure/aks/)
- [Azure RBAC 역할 참조](https://docs.microsoft.com/azure/role-based-access-control/role-definitions-list)
- [Azure Service Principal 가이드](https://docs.microsoft.com/azure/active-directory/develop/howto-create-service-principal-portal)
- [Azure CLI 참조](https://docs.microsoft.com/cli/azure/)
- [Azure 리소스 공급자 및 리소스 유형](https://docs.microsoft.com/azure/azure-resource-manager/management/resource-providers-and-types)


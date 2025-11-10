# SkyClust 클라우드 자격증명 설정 가이드

이 문서는 SkyClust 플랫폼에서 AWS와 GCP 자격증명을 설정하는 방법을 안내합니다.

---

## AWS 자격증명 설정

### 1. IAM Policy 생성

SkyClust가 EKS 클러스터와 관련 리소스를 관리하기 위해 필요한 권한을 포함한 IAM Policy를 생성합니다.

#### 1.1 Policy 생성 절차

1. AWS IAM Console 접속
2. **Policies** > **Create Policy** 선택
3. **JSON** 탭 선택 후 아래 Policy 내용 입력
4. Policy 이름: `SkyClustEKSFullAccess`

#### 1.2 Policy 내용

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "EKSClusterManagement",
            "Effect": "Allow",
            "Action": [
                "eks:CreateCluster",
                "eks:DescribeCluster",
                "eks:ListClusters",
                "eks:DeleteCluster",
                "eks:UpdateClusterConfig",
                "eks:TagResource",
                "eks:UntagResource",
                "eks:ListTagsForResource"
            ],
            "Resource": "*"
        },
        {
            "Sid": "EKSNodeGroupManagement",
            "Effect": "Allow",
            "Action": [
                "eks:CreateNodegroup",
                "eks:DescribeNodegroup",
                "eks:ListNodegroups",
                "eks:DeleteNodegroup",
                "eks:UpdateNodegroupConfig",
                "eks:UpdateNodegroupVersion"
            ],
            "Resource": "*"
        },
        {
            "Sid": "EKSConsoleAccess",
            "Effect": "Allow",
            "Action": [
                "eks:AccessKubernetesApi",
                "eks:ListUpdates",
                "eks:DescribeUpdate"
            ],
            "Resource": "*"
        },
        {
            "Sid": "IAMRoleManagement",
            "Effect": "Allow",
            "Action": [
                "iam:PassRole",
                "iam:ListAttachedRolePolicies",
                "iam:GetRole",
                "iam:ListRoles",
                "iam:CreateRole",
                "iam:AttachRolePolicy",
                "iam:DetachRolePolicy",
                "iam:PutRolePolicy",
                "iam:DeleteRolePolicy",
                "iam:GetRolePolicy",
                "iam:CreateServiceLinkedRole",
                "iam:DeleteServiceLinkedRole",
                "iam:GetServiceLinkedRoleDeletionStatus"
            ],
            "Resource": [
                "arn:aws:iam::*:role/EKSClusterRole",
                "arn:aws:iam::*:role/EKSNodeRole",
                "arn:aws:iam::*:role/*",
                "arn:aws:iam::*:role/aws-service-role/eks-nodegroup.amazonaws.com/*"
            ]
        },
        {
            "Sid": "EC2NetworkManagement",
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeVpcs",
                "ec2:DescribeSubnets",
                "ec2:DescribeSecurityGroups",
                "ec2:DescribeAvailabilityZones",
                "ec2:CreateTags",
                "ec2:DeleteTags",
                "ec2:DescribeTags",
                "ec2:DescribeInstances",
                "ec2:DescribeInstanceStatus",
                "ec2:DescribeInstanceAttribute"
            ],
            "Resource": "*"
        },
        {
            "Sid": "EC2NetworkCreation",
            "Effect": "Allow",
            "Action": [
                "ec2:CreateVpc",
                "ec2:DeleteVpc",
                "ec2:ModifyVpcAttribute",
                "ec2:CreateSubnet",
                "ec2:DeleteSubnet",
                "ec2:ModifySubnetAttribute",
                "ec2:CreateSecurityGroup",
                "ec2:DeleteSecurityGroup",
                "ec2:AuthorizeSecurityGroupIngress",
                "ec2:AuthorizeSecurityGroupEgress",
                "ec2:RevokeSecurityGroupIngress",
                "ec2:RevokeSecurityGroupEgress",
                "ec2:UpdateSecurityGroupRuleDescriptionsIngress",
                "ec2:UpdateSecurityGroupRuleDescriptionsEgress"
            ],
            "Resource": "*"
        },
        {
            "Sid": "CloudFormationSupport",
            "Effect": "Allow",
            "Action": [
                "cloudformation:CreateStack",
                "cloudformation:DeleteStack",
                "cloudformation:DescribeStacks",
                "cloudformation:DescribeStackEvents",
                "cloudformation:DescribeStackResources",
                "cloudformation:GetTemplate",
                "cloudformation:ListStacks"
            ],
            "Resource": "*"
        },
        {
            "Sid": "LogsManagement",
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:DescribeLogGroups",
                "logs:DeleteLogGroup",
                "logs:PutRetentionPolicy",
                "logs:DescribeLogStreams",
                "logs:GetLogEvents"
            ],
            "Resource": "*"
        },
        {
            "Sid": "SecurityGroupManagement",
            "Effect": "Allow",
            "Action": [
                "ec2:AuthorizeSecurityGroupIngress",
                "ec2:AuthorizeSecurityGroupEgress",
                "ec2:RevokeSecurityGroupIngress",
                "ec2:RevokeSecurityGroupEgress",
                "ec2:DescribeSecurityGroups",
                "ec2:DescribeVpcs",
                "ec2:DescribeSubnets"
            ],
            "Resource": "*"
        },
        {
            "Sid": "EKSAccessEntryManagement",
            "Effect": "Allow",
            "Action": [
                "eks:CreateAccessEntry",
                "eks:UpdateAccessEntry",
                "eks:DeleteAccessEntry",
                "eks:DescribeAccessEntry",
                "eks:ListAccessEntries",
                "eks:AssociateAccessPolicy",
                "eks:DisassociateAccessPolicy",
                "eks:ListAssociatedAccessPolicies",
                "eks:ListAccessPolicies"
            ],
            "Resource": "*"
        },
        {
            "Sid": "CostExplorerAccess",
            "Effect": "Allow",
            "Action": [
                "ce:GetCostAndUsage",
                "ce:DescribeCostCategoryDefinition"
            ],
            "Resource": "*"
        }
    ]
}
```

#### 1.3 Policy 설명

- **EKSClusterManagement**: EKS 클러스터 생성, 조회, 삭제, 업데이트 권한
- **EKSNodeGroupManagement**: 노드 그룹 관리 권한
- **EKSConsoleAccess**: Kubernetes API 접근 및 업데이트 조회 권한
- **IAMRoleManagement**: EKS 클러스터 및 노드 그룹에 필요한 IAM 역할 생성 및 관리 권한
- **EC2NetworkManagement**: VPC, 서브넷, 보안 그룹 조회 및 태그 관리 권한
- **EC2NetworkCreation**: VPC, 서브넷, 보안 그룹 생성, 삭제, 수정 권한 (네트워크 리소스 생성/관리용)
- **CloudFormationSupport**: CloudFormation 스택 관리 권한 (EKS 생성 시 내부적으로 사용)
- **LogsManagement**: CloudWatch Logs 관리 권한
- **SecurityGroupManagement**: 보안 그룹 규칙 관리 권한
- **EKSAccessEntryManagement**: EKS Access Entry 관리 권한
- **CostExplorerAccess**: AWS Cost Explorer API 접근 권한 (비용 분석 기능용)

---

### 2. EKS Cluster Role 생성

EKS 클러스터가 사용할 IAM 역할을 생성합니다.

#### 2.1 Role 생성 절차

1. AWS IAM Console > **역할** > **역할 만들기** 선택
2. **AWS 서비스** 선택 후 **EKS** 선택
3. **다음** 클릭
4. 역할 이름: `EKSClusterRole`

#### 2.2 신뢰 정책 (Trust Policy)

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "eks.amazonaws.com"
      },
      "Action": [
        "sts:AssumeRole",
        "sts:TagSession"
      ]
    }
  ]
}
```

#### 2.3 권한 정책 (Permissions Policy)

다음 AWS 관리형 정책들을 연결:

- `AmazonEKSClusterPolicy`
- `AmazonEKSServicePolicy`
- `AmazonEKSBlockStoragePolicy`
- `AmazonEKSComputePolicy`
- `AmazonEKSLoadBalancingPolicy`
- `AmazonEKSNetworkingPolicy`

#### 2.4 Role ARN 확인

역할 생성 후 ARN을 확인합니다. 형식:
```
arn:aws:iam::<ACCOUNT_ID>:role/EKSClusterRole
```

---

### 3. EKS Node Role 생성

EKS 노드 그룹이 사용할 IAM 역할을 생성합니다.

#### 3.1 Role 생성 절차

1. AWS IAM Console > **역할** > **역할 만들기** 선택
2. **AWS 서비스** 선택 후 **EC2** 선택
3. **다음** 클릭
4. 역할 이름: `EKSNodeRole`

#### 3.2 신뢰 정책 (Trust Policy)

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
```

#### 3.3 권한 정책 (Permissions Policy)

다음 AWS 관리형 정책들을 연결:

- `AmazonEKSWorkerNodePolicy`
- `AmazonEKS_CNI_Policy`
- `AmazonEC2ContainerRegistryReadOnly`

#### 3.4 Role ARN 확인

역할 생성 후 ARN을 확인합니다. 형식:
```
arn:aws:iam::<ACCOUNT_ID>:role/EKSNodeRole
```

---

### 4. IAM 사용자에 Policy 부여

#### 4.1 절차

1. AWS IAM Console > **사용자** 이동
2. SkyClust에서 사용할 사용자 선택
3. **권한** 탭 > **권한 추가** 클릭
4. **직접 정책 연결** 선택
5. 생성한 `SkyClustEKSFullAccess` Policy 선택
6. **다음** > **권한 추가** 클릭

#### 4.2 Access Key 생성

1. 사용자 > **보안 자격 증명** 탭 이동
2. **액세스 키 만들기** 클릭
3. **애플리케이션 실행 외부에서 코드 실행** 선택 (또는 필요에 따라 선택)
4. Access Key ID와 Secret Access Key 확인 및 안전하게 보관

---

## GCP 자격증명 설정

### 1. 프로젝트 생성

#### 1.1 절차

1. [GCP Console](https://console.cloud.google.com/) 접속
2. 프로젝트 선택 드롭다운 > **새 프로젝트** 클릭
3. 프로젝트 정보 입력:
   - **프로젝트 이름**: `skyclust-project` (또는 원하는 이름)
   - **조직**: 선택사항 (조직이 있는 경우)
4. **만들기** 클릭
5. 프로젝트 ID 확인 (프로젝트 선택 드롭다운에서 확인 가능)

---

### 2. 필수 API 활성화

SkyClust가 GCP 서비스를 사용하기 위해 필요한 API를 활성화합니다.

#### 2.1 GCP Console을 통한 활성화

1. **API 및 서비스** > **라이브러리** 이동
2. 다음 API를 검색하여 활성화:
   - **Kubernetes Engine API**
   - **Compute Engine API**
   - **Identity and Access Management (IAM) API**
   - **Service Usage API**
   - **Cloud Billing API** (비용 분석 기능용)

#### 2.2 gcloud CLI를 통한 활성화

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

### 3. 서비스 계정 생성 및 설정

#### 3.1 서비스 계정 생성

1. **IAM 및 관리자** > **서비스 계정** 이동
2. **서비스 계정 만들기** 클릭
3. 서비스 계정 정보 입력:
   - **서비스 계정 이름**: `skyclust-service-account`
   - **서비스 계정 ID**: `skyclust-service-account` (자동 생성됨)
   - **설명**: `SkyClust application service account for GKE and network management`
4. **만들고 계속하기** 클릭

#### 3.2 역할 부여

다음 역할을 서비스 계정에 부여:

- **Kubernetes Engine 관리자** (`roles/container.admin`)
- **Compute 네트워크 관리자** (`roles/compute.networkAdmin`)
- **Compute 인스턴스 관리자** (`roles/compute.instanceAdmin`)
- **Compute Engine 서비스 에이전트** (`roles/compute.serviceAgent`)
- **Compute 보안 관리자** (`roles/compute.securityAdmin`)
- **서비스 계정 사용자** (`roles/iam.serviceAccountUser`)
- **뷰어** (`roles/viewer`)
- **서비스 사용량 관리자** (`roles/serviceusage.serviceUsageAdmin`)
- **Billing 계정 사용자** (`roles/billing.user`) - 비용 분석 기능용
- **Billing 계정 뷰어** (`roles/billing.viewer`) - 비용 분석 기능용

**역할 부여 방법:**
1. 역할을 선택할 때마다 **역할 추가** 클릭
2. 모든 역할 선택 후 **계속** 클릭
3. **완료** 클릭

#### 3.3 서비스 계정 키 생성

1. 생성된 서비스 계정 클릭
2. **키** 탭 이동
3. **키 추가** > **새 키 만들기** 클릭
4. **JSON** 선택 후 **만들기** 클릭
5. JSON 키 파일이 자동으로 다운로드됨 (안전하게 보관)
6. 다운로드된 파일을 SkyClust 자격증명 등록 시 사용

---

### 4. gcloud CLI를 통한 역할 부여 (선택사항)

#### 4.1 절차

```bash
# 변수 설정
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

# 서비스 계정 사용자
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/iam.serviceAccountUser"

# 서비스 사용량 관리자
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/serviceusage.serviceUsageAdmin"

# 뷰어
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/viewer"

# Billing 계정 사용자 (비용 분석 기능용)
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/billing.user"

# Billing 계정 뷰어 (비용 분석 기능용)
gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member="serviceAccount:${SERVICE_ACCOUNT}" \
    --role="roles/billing.viewer"
```

#### 4.2 역할 부여 확인

```bash
# 서비스 계정의 역할 확인
gcloud projects get-iam-policy ${PROJECT_ID} \
    --flatten="bindings[].members" \
    --filter="bindings.members:serviceAccount:${SERVICE_ACCOUNT}"
```

---

### 5. Billing 계정 연결 (비용 분석 기능용)

비용 분석 기능을 사용하려면 프로젝트에 Billing 계정을 연결해야 합니다.

#### 5.1 절차

1. **결제** > **계정 연결** 이동
2. 기존 Billing 계정 선택 또는 새로 생성
3. 프로젝트에 Billing 계정 연결 확인

---

## SkyClust에서 자격증명 등록

### AWS 자격증명 등록

SkyClust API를 통해 자격증명을 등록합니다:

```bash
POST /api/v1/credentials
Content-Type: application/json
Authorization: Bearer <token>

{
  "workspace_id": "workspace-uuid",
  "name": "AWS Production",
  "provider": "aws",
  "data": {
    "access_key": "AKIA...",
    "secret_key": "your-secret-access-key",
    "region": "us-west-2"
  }
}
```

**필수 정보:**
- `workspace_id`: 워크스페이스 UUID
- `name`: 자격증명 이름
- `provider`: `aws`
- `data.access_key`: AWS Access Key ID
- `data.secret_key`: AWS Secret Access Key
- `data.region`: 기본 리전 (선택사항)

---

### GCP 자격증명 등록

#### 방법 1: JSON 키 파일 내용 직접 입력

```bash
POST /api/v1/credentials
Content-Type: application/json
Authorization: Bearer <token>

{
  "workspace_id": "workspace-uuid",
  "name": "GCP Production",
  "provider": "gcp",
  "data": {
    "type": "service_account",
    "project_id": "your-project-id",
    "private_key_id": "key-id-from-json",
    "private_key": "-----BEGIN PRIVATE KEY-----\n...",
    "client_email": "skyclust-service-account@your-project-id.iam.gserviceaccount.com",
    "client_id": "123456789",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "token_uri": "https://oauth2.googleapis.com/token",
    "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
    "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/...",
    "universe_domain": "googleapis.com"
  }
}
```

#### 방법 2: 파일 업로드

```bash
POST /api/v1/credentials/upload
Content-Type: multipart/form-data
Authorization: Bearer <token>

workspace_id: workspace-uuid
name: GCP Production
provider: gcp
file: <json-key-file>
```

**필수 정보:**
- `workspace_id`: 워크스페이스 UUID
- `name`: 자격증명 이름
- `provider`: `gcp`
- `data` 또는 `file`: 서비스 계정 JSON 키 파일 내용

---

## 보안 권장 사항

### AWS

1. **최소 권한 원칙**: 필요한 최소한의 권한만 부여
2. **정기적인 Access Key 로테이션**: 주기적으로 Access Key를 교체
3. **MFA 활성화**: IAM 사용자에 MFA 활성화 권장
4. **CloudTrail 모니터링**: 모든 API 호출을 CloudTrail로 모니터링

### GCP

1. **서비스 계정 키 관리**: 생성한 JSON 키 파일을 안전하게 보관
2. **정기적인 키 로테이션**: 서비스 계정 키를 주기적으로 교체
3. **최소 권한 원칙**: 필요한 역할만 부여
4. **Audit Logs 모니터링**: Cloud Audit Logs로 모든 활동 모니터링

---

## 문제 해결

### AWS

**문제**: `AccessDeniedException` 또는 `UnauthorizedOperation` 발생
- **원인**: IAM Policy 권한 부족
- **해결**: 
  - `SkyClustEKSFullAccess` Policy가 올바르게 연결되었는지 확인
  - Policy에 `EC2NetworkCreation` 섹션이 포함되어 있는지 확인
  - 필요한 권한: `ec2:CreateVpc`, `ec2:CreateSubnet`, `ec2:DeleteVpc`, `ec2:DeleteSubnet` 등

**문제**: VPC/Subnet 생성 실패 (`ec2:CreateVpc` 또는 `ec2:CreateSubnet` 권한 없음)
- **원인**: `EC2NetworkCreation` 섹션이 Policy에 없거나 권한이 부족함
- **해결**: 
  - `SkyClustEKSFullAccess` Policy에 `EC2NetworkCreation` 섹션 추가
  - 또는 `AmazonEC2FullAccess` 정책을 임시로 연결하여 테스트 (프로덕션에서는 최소 권한 원칙 준수)

**문제**: EKS 클러스터 생성 실패
- **원인**: `EKSClusterRole` 또는 `EKSNodeRole`이 올바르게 설정되지 않음
- **해결**: 역할의 신뢰 정책과 권한 정책 확인

### GCP

**문제**: `PERMISSION_DENIED` 오류
- **원인**: 필요한 IAM 역할이 부여되지 않음
- **해결**: 서비스 계정에 모든 필수 역할이 부여되었는지 확인

**문제**: API 비활성화 오류
- **원인**: 필수 API가 활성화되지 않음
- **해결**: 모든 필수 API가 활성화되었는지 확인

**문제**: 비용 분석 API 오류
- **원인**: Billing API 미활성화 또는 Billing 계정 미연결
- **해결**: Cloud Billing API 활성화 및 프로젝트에 Billing 계정 연결

---

## 추가 참고 자료

- [AWS EKS 사용자 가이드](https://docs.aws.amazon.com/eks/latest/userguide/)
- [AWS IAM 정책 참조](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies.html)
- [GCP Kubernetes Engine 문서](https://cloud.google.com/kubernetes-engine/docs)
- [GCP IAM 역할 참조](https://cloud.google.com/iam/docs/understanding-roles)
- [GCP Cloud Billing API 문서](https://cloud.google.com/billing/docs/reference/rest)


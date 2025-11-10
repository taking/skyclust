# AWS IAM 정책 설정 가이드

이 문서는 SkyClust 플랫폼에서 AWS 자격증명을 설정하기 위한 IAM 정책 및 역할 설정 방법을 단계별로 안내합니다.

---

## 목차

1. [IAM Policy 생성](#1-iam-policy-생성)
2. [EKS Cluster Role 생성](#2-eks-cluster-role-생성)
3. [EKS Node Role 생성](#3-eks-node-role-생성)
4. [IAM 사용자에 Policy 부여](#4-iam-사용자에-policy-부여)
5. [설정 확인](#5-설정-확인)

---

## 1. IAM Policy 생성

### 1.1 AWS IAM Console 접속

1. [AWS Console](https://console.aws.amazon.com/)에 로그인
2. 검색창에 **IAM** 입력 후 선택
3. 왼쪽 메뉴에서 **정책(Policies)** 클릭

### 1.2 Policy 생성 시작

1. **정책 만들기(Create Policy)** 버튼 클릭
2. **JSON** 탭 선택

### 1.3 Policy 내용 입력

아래 JSON 정책을 복사하여 붙여넣기:

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
                "arn:aws:iam::481665112450:role/EKSClusterRole",
                "arn:aws:iam::481665112450:role/EKSNodeRole",
                "arn:aws:iam::481665112450:role/*",
                "arn:aws:iam::481665112450:role/aws-service-role/eks-nodegroup.amazonaws.com/*"
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
        },
        {
            "Sid": "EC2RegionAndZoneAccess",
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeRegions",
                "ec2:DescribeAvailabilityZones"
            ],
            "Resource": "*"
        }
    ]
}
```

> **참고**: 위 정책의 `IAMRoleManagement` 섹션에 있는 Account ID (`481665112450`)를 본인의 AWS Account ID로 변경해야 합니다.

### 1.4 Policy 이름 설정

1. **다음(Next)** 버튼 클릭
2. **정책 이름(Policy name)** 입력: `SkyClustEKSFullAccess`
3. **설명(Description)** 입력 (선택사항): `SkyClust EKS 및 네트워크 리소스 관리 권한`
4. **정책 만들기(Create Policy)** 버튼 클릭

### 1.5 Policy 생성 확인

- 정책 목록에서 `SkyClustEKSFullAccess` 정책이 생성되었는지 확인

---

## 2. EKS Cluster Role 생성

EKS 클러스터가 사용할 IAM 역할을 생성합니다.

### 2.1 역할 생성 시작

1. IAM Console 왼쪽 메뉴에서 **역할(Roles)** 클릭
2. **역할 만들기(Create Role)** 버튼 클릭

### 2.2 신뢰 정책 설정

1. **AWS 서비스(AWS service)** 선택
2. **EKS** 선택
3. **사용 사례(Use case)** 선택:
   - **EKS - Cluster** 선택
4. **다음(Next)** 버튼 클릭

### 2.3 사용자 지정 신뢰 정책 설정

1. **사용자 지정 신뢰 정책(Custom trust policy)** 선택
2. 아래 JSON을 복사하여 붙여넣기:

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

3. **다음(Next)** 버튼 클릭

### 2.4 권한 정책 연결

다음 AWS 관리형 정책들을 검색하여 선택:

1. **권한 추가(Add permissions)** 클릭
2. 검색창에 각 정책 이름 입력 후 선택:
   - `AmazonEKSClusterPolicy`
   - `AmazonEKSServicePolicy`
   - `AmazonEKSBlockStoragePolicy`
   - `AmazonEKSComputePolicy`
   - `AmazonEKSLoadBalancingPolicy`
   - `AmazonEKSNetworkingPolicy`

3. **다음(Next)** 버튼 클릭

### 2.5 역할 이름 설정

1. **역할 이름(Role name)** 입력: `EKSClusterRole`
2. **설명(Description)** 입력 (선택사항): `EKS 클러스터용 IAM 역할`
3. **역할 만들기(Create Role)** 버튼 클릭

### 2.6 Role ARN 확인

1. 생성된 `EKSClusterRole` 클릭
2. **요약(Summary)** 섹션에서 **ARN** 복사
   - 형식: `arn:aws:iam::<ACCOUNT_ID>:role/EKSClusterRole`
3. 이 ARN을 나중에 EKS 클러스터 생성 시 사용합니다

---

## 3. EKS Node Role 생성

EKS 노드 그룹이 사용할 IAM 역할을 생성합니다.

### 3.1 역할 생성 시작

1. IAM Console 왼쪽 메뉴에서 **역할(Roles)** 클릭
2. **역할 만들기(Create Role)** 버튼 클릭

### 3.2 신뢰 정책 설정

1. **AWS 서비스(AWS service)** 선택
2. **EC2** 선택
3. **사용 사례(Use case)** 선택:
   - **EC2** 선택
4. **다음(Next)** 버튼 클릭

### 3.3 사용자 지정 신뢰 정책 설정

1. **사용자 지정 신뢰 정책(Custom trust policy)** 선택
2. 아래 JSON을 복사하여 붙여넣기:

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

3. **다음(Next)** 버튼 클릭

### 3.4 권한 정책 연결

다음 AWS 관리형 정책들을 검색하여 선택:

1. **권한 추가(Add permissions)** 클릭
2. 검색창에 각 정책 이름 입력 후 선택:
   - `AmazonEKSWorkerNodePolicy`
   - `AmazonEKS_CNI_Policy`
   - `AmazonEC2ContainerRegistryReadOnly`

3. **다음(Next)** 버튼 클릭

### 3.5 역할 이름 설정

1. **역할 이름(Role name)** 입력: `EKSNodeRole`
2. **설명(Description)** 입력 (선택사항): `EKS 노드 그룹용 IAM 역할`
3. **역할 만들기(Create Role)** 버튼 클릭

### 3.6 Role ARN 확인

1. 생성된 `EKSNodeRole` 클릭
2. **요약(Summary)** 섹션에서 **ARN** 복사
   - 형식: `arn:aws:iam::<ACCOUNT_ID>:role/EKSNodeRole`
3. 이 ARN을 나중에 EKS 노드 그룹 생성 시 사용합니다

---

## 4. IAM 사용자에 Policy 부여

SkyClust 애플리케이션이 사용할 IAM 사용자에 정책을 부여합니다.

### 4.1 사용자 선택

1. IAM Console 왼쪽 메뉴에서 **사용자(Users)** 클릭
2. SkyClust에서 사용할 사용자 선택
   - 사용자가 없다면 먼저 사용자를 생성해야 합니다

### 4.2 권한 추가

1. **권한(Permissions)** 탭 클릭
2. **권한 추가(Add permissions)** 버튼 클릭

### 4.3 정책 연결

1. **직접 정책 연결(Attach policies directly)** 선택
2. 검색창에 `SkyClustEKSFullAccess` 입력
3. `SkyClustEKSFullAccess` 정책 체크박스 선택
4. **다음(Next)** 버튼 클릭

### 4.4 권한 추가 확인

1. 선택한 정책이 올바른지 확인
2. **권한 추가(Add permissions)** 버튼 클릭

### 4.5 Access Key 생성

1. 사용자 페이지에서 **보안 자격 증명(Security credentials)** 탭 클릭
2. **액세스 키 만들기(Create access key)** 버튼 클릭
3. **사용 사례(Use case)** 선택:
   - **애플리케이션 실행 외부에서 코드 실행(Application running outside AWS)** 선택
4. **다음(Next)** 버튼 클릭
5. **설명 태그 추가(Add description tag)** (선택사항)
6. **액세스 키 만들기(Create access key)** 버튼 클릭

### 4.6 Access Key 저장

1. **Access Key ID**와 **Secret Access Key** 확인
2. **Secret Access Key**는 이 페이지를 떠나면 다시 확인할 수 없으므로 안전하게 보관
3. **CSV 파일 다운로드** 또는 **복사**하여 보관
4. **완료(Done)** 버튼 클릭

---

## 5. 설정 확인

### 5.1 Policy 확인

1. IAM Console > **정책(Policies)**
2. `SkyClustEKSFullAccess` 정책이 존재하는지 확인
3. 정책을 클릭하여 내용이 올바른지 확인

### 5.2 Role 확인

1. IAM Console > **역할(Roles)**
2. `EKSClusterRole`과 `EKSNodeRole`이 존재하는지 확인
3. 각 역할의 ARN을 확인하고 기록

### 5.3 사용자 권한 확인

1. IAM Console > **사용자(Users)**
2. 사용자를 선택하여 **권한(Permissions)** 탭 확인
3. `SkyClustEKSFullAccess` 정책이 연결되어 있는지 확인

### 5.4 Access Key 확인

1. 사용자 > **보안 자격 증명(Security credentials)** 탭
2. **액세스 키(Access keys)** 섹션에서 생성된 Access Key 확인

---

## 다음 단계

이제 SkyClust 플랫폼에서 자격증명을 등록할 수 있습니다:

1. **Access Key ID**: 위에서 생성한 Access Key ID
2. **Secret Access Key**: 위에서 생성한 Secret Access Key
3. **EKS Cluster Role ARN**: `EKSClusterRole`의 ARN
4. **EKS Node Role ARN**: `EKSNodeRole`의 ARN

---

## 문제 해결

### Policy 생성 실패

- **문제**: JSON 정책이 유효하지 않음
- **해결**: JSON 형식이 올바른지 확인 (쉼표, 중괄호 등)

### Role 생성 실패

- **문제**: 신뢰 정책 오류
- **해결**: Principal Service가 올바른지 확인 (`eks.amazonaws.com`, `ec2.amazonaws.com`)

### 권한 부족 오류

- **문제**: `AccessDeniedException` 또는 `UnauthorizedOperation` 발생
- **해결**: 
  - `SkyClustEKSFullAccess` 정책이 사용자에 연결되어 있는지 확인
  - 정책에 필요한 권한이 모두 포함되어 있는지 확인
  - Account ID가 정책의 `IAMRoleManagement` 섹션에 올바르게 설정되었는지 확인

### VPC/Subnet 생성 실패

- **문제**: `ec2:CreateVpc` 또는 `ec2:CreateSubnet` 권한 없음
- **해결**: `SkyClustEKSFullAccess` 정책에 `EC2NetworkCreation` 섹션이 포함되어 있는지 확인

---

## 참고 자료

- [AWS EKS 사용자 가이드](https://docs.aws.amazon.com/eks/latest/userguide/)
- [AWS IAM 정책 참조](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies.html)
- [AWS IAM 역할 생성 가이드](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create.html)


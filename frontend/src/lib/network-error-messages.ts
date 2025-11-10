/**
 * Network Error Messages
 * VPC 및 Subnet 생성/삭제 시 에러를 구분하여 명확한 메시지를 제공
 */

import { ServerError } from './error-handler';

/**
 * 에러 메시지에서 IAM 권한 관련 키워드 확인
 */
function isIAMPermissionError(message: string): boolean {
  const iamKeywords = [
    'UnauthorizedOperation',
    'AccessDenied',
    'not authorized',
    'no identity-based policy',
    'ec2:CreateVpc',
    'ec2:CreateSubnet',
    'ec2:DeleteVpc',
    'ec2:DeleteSubnet',
    'PROVIDER_ERROR',
  ];
  
  return iamKeywords.some(keyword => 
    message.toLowerCase().includes(keyword.toLowerCase())
  );
}

/**
 * 에러 메시지에서 네트워크 관련 키워드 확인
 */
function isNetworkError(message: string): boolean {
  const networkKeywords = [
    'cidr',
    'overlap',
    'conflict',
    'already exists',
    'duplicate',
    'invalid cidr',
    'address space',
    'subnet',
    'vpc',
  ];
  
  return networkKeywords.some(keyword => 
    message.toLowerCase().includes(keyword.toLowerCase())
  );
}

/**
 * 에러 메시지에서 검증 관련 키워드 확인
 */
function isValidationError(message: string): boolean {
  const validationKeywords = [
    'required',
    'invalid',
    'validation',
    'missing',
    'bad request',
  ];
  
  return validationKeywords.some(keyword => 
    message.toLowerCase().includes(keyword.toLowerCase())
  );
}

/**
 * 에러에서 메시지 추출
 */
function extractErrorMessage(error: unknown): string {
  // ServiceError인 경우 (service-base.ts에서 생성)
  // ServiceError는 message 필드에 이미 백엔드 에러 메시지가 포함되어 있음
  if (error && typeof error === 'object' && 'message' in error) {
    const serviceError = error as { 
      message?: string;
      code?: string;
      data?: {
        error?: { message?: string; code?: string };
        message?: string;
      } | unknown;
    };
    
    // ServiceError의 message 필드가 이미 백엔드 에러 메시지를 포함하고 있음
    // service-base.ts에서 responseData?.error?.message를 message로 설정함
    if (serviceError.message) {
      return serviceError.message;
    }
    
    // data 필드에서도 확인 (fallback)
    if (serviceError.data && typeof serviceError.data === 'object') {
      const data = serviceError.data as {
        error?: { message?: string; code?: string };
        message?: string;
      };
      
      if (data.error?.message) {
        return data.error.message;
      }
      if (data.message) {
        return data.message;
      }
    }
  }
  
  // 일반 Error 객체
  if (error instanceof Error) {
    return error.message;
  }
  
  if (error instanceof ServerError) {
    return error.message;
  }
  
  // Axios 에러 (직접 axios를 사용한 경우)
  if (error && typeof error === 'object' && 'response' in error) {
    const axiosError = error as { 
      response?: { 
        data?: { 
          error?: { message?: string; code?: string }; 
          message?: string 
        } 
      } 
    };
    
    return (
      axiosError.response?.data?.error?.message ||
      axiosError.response?.data?.message ||
      'An error occurred'
    );
  }
  
  if (typeof error === 'string') {
    return error;
  }
  
  return 'An unexpected error occurred';
}

/**
 * Provider별 IAM 권한 에러 메시지 생성
 */
function getIAMPermissionMessage(provider: string | undefined, resource: 'VPC' | 'Subnet'): string {
  const providerName = provider?.toUpperCase() || 'Cloud Provider';
  const action = resource === 'VPC' ? 'CreateVpc' : 'CreateSubnet';
  
  return `[${providerName} IAM 권한 오류] ${resource} 생성을 위한 IAM 권한이 없습니다.\n\n` +
    `필요한 권한: ec2:${action}\n\n` +
    `해결 방법:\n` +
    `1. AWS IAM Console에서 사용자 또는 역할의 정책을 확인하세요\n` +
    `2. SkyClustEKSFullAccess 정책에 EC2NetworkCreation 섹션이 포함되어 있는지 확인하세요\n` +
    `3. 필요한 권한: ec2:${action}, ec2:Delete${resource}, ec2:Modify${resource}Attribute`;
}

/**
 * VPC 생성 에러 메시지 생성
 */
export function getVPCCreationErrorMessage(error: unknown, provider?: string): string {
  const message = extractErrorMessage(error);
  
  // IAM 권한 에러
  if (isIAMPermissionError(message)) {
    return getIAMPermissionMessage(provider, 'VPC');
  }
  
  // 네트워크 에러 (CIDR 충돌 등)
  if (isNetworkError(message)) {
    if (message.toLowerCase().includes('cidr') || message.toLowerCase().includes('overlap')) {
      return `[네트워크 오류] VPC CIDR 블록이 기존 VPC와 겹치거나 유효하지 않습니다.\n\n` +
        `오류: ${message}\n\n` +
        `해결 방법:\n` +
        `1. 다른 CIDR 블록을 사용하세요 (예: 10.1.0.0/16, 172.16.0.0/16)\n` +
        `2. 기존 VPC의 CIDR 블록을 확인하세요`;
    }
    
    if (message.toLowerCase().includes('already exists') || message.toLowerCase().includes('duplicate')) {
      return `[중복 오류] 동일한 이름의 VPC가 이미 존재합니다.\n\n` +
        `오류: ${message}\n\n` +
        `해결 방법: 다른 이름을 사용하세요`;
    }
    
    return `[네트워크 오류] VPC 생성 중 네트워크 오류가 발생했습니다.\n\n` +
      `오류: ${message}`;
  }
  
  // 검증 에러
  if (isValidationError(message)) {
    return `[입력 오류] VPC 생성에 필요한 정보가 올바르지 않습니다.\n\n` +
      `오류: ${message}\n\n` +
      `해결 방법:\n` +
      `1. 필수 필드가 모두 입력되었는지 확인하세요\n` +
      `2. CIDR 블록 형식이 올바른지 확인하세요 (예: 10.0.0.0/16)`;
  }
  
  // Provider별 특수 에러
  if (provider === 'azure') {
    if (message.toLowerCase().includes('resource group')) {
      return `[Azure 오류] Resource Group이 존재하지 않거나 접근할 수 없습니다.\n\n` +
        `오류: ${message}\n\n` +
        `해결 방법:\n` +
        `1. Resource Group 이름이 올바른지 확인하세요\n` +
        `2. Resource Group이 해당 Location에 존재하는지 확인하세요`;
    }
    
    if (message.toLowerCase().includes('location')) {
      return `[Azure 오류] Location이 유효하지 않습니다.\n\n` +
        `오류: ${message}\n\n` +
        `해결 방법: 유효한 Azure Location을 입력하세요 (예: eastus, westus2)`;
    }
  }
  
  if (provider === 'gcp') {
    if (message.toLowerCase().includes('project')) {
      return `[GCP 오류] Project ID가 유효하지 않거나 접근할 수 없습니다.\n\n` +
        `오류: ${message}\n\n` +
        `해결 방법:\n` +
        `1. Project ID가 올바른지 확인하세요\n` +
        `2. 서비스 계정에 필요한 권한이 부여되어 있는지 확인하세요`;
    }
  }
  
  // 일반 에러
  return `[VPC 생성 실패] VPC 생성 중 오류가 발생했습니다.\n\n` +
    `오류: ${message}`;
}

/**
 * Subnet 생성 에러 메시지 생성
 */
export function getSubnetCreationErrorMessage(error: unknown, provider?: string): string {
  const message = extractErrorMessage(error);
  
  // IAM 권한 에러
  if (isIAMPermissionError(message)) {
    return getIAMPermissionMessage(provider, 'Subnet');
  }
  
  // 네트워크 에러 (CIDR 충돌 등)
  if (isNetworkError(message)) {
    if (message.toLowerCase().includes('cidr') || message.toLowerCase().includes('overlap')) {
      return `[네트워크 오류] Subnet CIDR 블록이 VPC CIDR 범위를 벗어나거나 다른 Subnet과 겹칩니다.\n\n` +
        `오류: ${message}\n\n` +
        `해결 방법:\n` +
        `1. VPC의 CIDR 블록 범위 내에서 Subnet CIDR을 설정하세요\n` +
        `2. 기존 Subnet의 CIDR 블록과 겹치지 않는지 확인하세요\n` +
        `3. 예: VPC가 10.0.0.0/16이면 Subnet은 10.0.1.0/24, 10.0.2.0/24 등 사용 가능`;
    }
    
    if (message.toLowerCase().includes('availability zone') || message.toLowerCase().includes('zone')) {
      return `[가용 영역 오류] 선택한 Availability Zone이 유효하지 않거나 VPC와 호환되지 않습니다.\n\n` +
        `오류: ${message}\n\n` +
        `해결 방법:\n` +
        `1. VPC가 생성된 Region의 유효한 Availability Zone을 선택하세요\n` +
        `2. ${provider === 'gcp' ? 'Zone' : 'Availability Zone'} 형식이 올바른지 확인하세요`;
    }
    
    if (message.toLowerCase().includes('already exists') || message.toLowerCase().includes('duplicate')) {
      return `[중복 오류] 동일한 이름의 Subnet이 이미 존재합니다.\n\n` +
        `오류: ${message}\n\n` +
        `해결 방법: 다른 이름을 사용하세요`;
    }
    
    return `[네트워크 오류] Subnet 생성 중 네트워크 오류가 발생했습니다.\n\n` +
      `오류: ${message}`;
  }
  
  // 검증 에러
  if (isValidationError(message)) {
    return `[입력 오류] Subnet 생성에 필요한 정보가 올바르지 않습니다.\n\n` +
      `오류: ${message}\n\n` +
      `해결 방법:\n` +
      `1. 필수 필드가 모두 입력되었는지 확인하세요 (Name, VPC ID, CIDR Block, Region, Availability Zone)\n` +
      `2. CIDR 블록 형식이 올바른지 확인하세요 (예: 10.0.1.0/24)\n` +
      `3. VPC ID가 올바른지 확인하세요`;
  }
  
  // Provider별 특수 에러
  if (provider === 'azure') {
    if (message.toLowerCase().includes('virtual network') || message.toLowerCase().includes('vnet')) {
      return `[Azure 오류] Virtual Network (VPC)를 찾을 수 없거나 접근할 수 없습니다.\n\n` +
        `오류: ${message}\n\n` +
        `해결 방법:\n` +
        `1. Virtual Network ID가 올바른지 확인하세요\n` +
        `2. Virtual Network가 해당 Resource Group과 Location에 존재하는지 확인하세요`;
    }
  }
  
  if (provider === 'gcp') {
    if (message.toLowerCase().includes('vpc') && message.toLowerCase().includes('not found')) {
      return `[GCP 오류] VPC를 찾을 수 없습니다.\n\n` +
        `오류: ${message}\n\n` +
        `해결 방법:\n` +
        `1. VPC ID가 올바른지 확인하세요\n` +
        `2. VPC가 해당 Project와 Region에 존재하는지 확인하세요`;
    }
  }
  
  // 일반 에러
  return `[Subnet 생성 실패] Subnet 생성 중 오류가 발생했습니다.\n\n` +
    `오류: ${message}`;
}

/**
 * VPC 삭제 에러 메시지 생성
 */
export function getVPCDeletionErrorMessage(error: unknown, provider?: string): string {
  const message = extractErrorMessage(error);
  
  // ServiceError의 code 필드 확인 (PROVIDER_ERROR 등)
  let errorCode: string | undefined;
  if (error && typeof error === 'object' && 'code' in error) {
    errorCode = (error as { code?: string }).code;
  }
  
  // 의존성 위반 에러 (DependencyViolation)
  // 백엔드에서 PROVIDER_ERROR로 반환하지만 메시지에 DependencyViolation이 포함됨
  if (message.toLowerCase().includes('dependencyviolation') || 
      message.toLowerCase().includes('has dependencies') ||
      message.toLowerCase().includes('cannot be deleted') ||
      message.toLowerCase().includes('연결되어 있어')) {
    return `[VPC 삭제 불가] VPC에 연결된 리소스가 있어 삭제할 수 없습니다.\n\n` +
      `오류: ${message}\n\n` +
      `해결 방법:\n` +
      `1. VPC에 연결된 모든 서브넷을 먼저 삭제하세요\n` +
      `2. VPC에 연결된 보안 그룹을 삭제하거나 다른 리소스에서 분리하세요\n` +
      `3. VPC에 연결된 인터넷 게이트웨이를 분리하고 삭제하세요\n` +
      `4. VPC에 연결된 네트워크 인터페이스를 삭제하세요\n` +
      `5. VPC 피어링 연결이 있다면 삭제하세요\n` +
      `6. 모든 의존 리소스를 제거한 후 다시 시도하세요`;
  }
  
  // 리소스를 찾을 수 없음
  if (message.toLowerCase().includes('not found') || 
      message.toLowerCase().includes('찾을 수 없습니다')) {
    return `[VPC 삭제 불가] VPC를 찾을 수 없습니다.\n\n` +
      `오류: ${message}\n\n` +
      `VPC가 이미 삭제되었거나 존재하지 않습니다.`;
  }
  
  // 권한 에러
  if (message.toLowerCase().includes('accessdenied') || 
      message.toLowerCase().includes('unauthorized') ||
      message.toLowerCase().includes('권한이 없습니다')) {
    return `[권한 오류] VPC 삭제 권한이 없습니다.\n\n` +
      `오류: ${message}\n\n` +
      `해결 방법:\n` +
      `1. IAM 정책에서 'ec2:DeleteVpc' 권한을 확인하세요\n` +
      `2. 필요한 권한이 부여되어 있는지 확인하세요`;
  }
  
  // IAM 권한 에러
  if (isIAMPermissionError(message)) {
    return `[IAM 권한 오류] VPC 삭제를 위한 IAM 권한이 없습니다.\n\n` +
      `오류: ${message}\n\n` +
      `해결 방법:\n` +
      `1. AWS IAM Console에서 사용자 또는 역할의 정책을 확인하세요\n` +
      `2. SkyClustEKSFullAccess 정책에 EC2NetworkDeletion 섹션이 포함되어 있는지 확인하세요\n` +
      `3. 필요한 권한: ec2:DeleteVpc, ec2:DescribeVpcs`;
  }
  
  // 일반 에러
  return `[VPC 삭제 실패] VPC 삭제 중 오류가 발생했습니다.\n\n` +
    `오류: ${message}`;
}


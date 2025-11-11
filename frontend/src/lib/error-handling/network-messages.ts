/**
 * Network Error Messages
 * VPC 및 Subnet 생성/삭제 시 에러를 구분하여 명확한 메시지를 제공
 */

import { ServerError } from './types';

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
 * 백엔드에서 반환하는 상세한 에러 메시지를 우선적으로 추출
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
    // 백엔드에서 상세한 메시지를 반환하는 경우 (예: "VPC 'xxx' cannot be deleted...")
    if (serviceError.message && serviceError.message.trim().length > 0) {
      return serviceError.message;
    }
    
    // data 필드에서도 확인 (fallback)
    if (serviceError.data && typeof serviceError.data === 'object') {
      const data = serviceError.data as {
        error?: { message?: string; code?: string };
        message?: string;
      };
      
      // data.error.message가 가장 상세한 백엔드 메시지일 가능성이 높음
      if (data.error?.message && data.error.message.trim().length > 0) {
        return data.error.message;
      }
      if (data.message && data.message.trim().length > 0) {
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
    
    // 백엔드 응답에서 에러 메시지 추출 (우선순위: error.message > message)
    const backendMessage = 
      axiosError.response?.data?.error?.message ||
      axiosError.response?.data?.message;
    
    if (backendMessage && backendMessage.trim().length > 0) {
      return backendMessage;
    }
    
    return 'An error occurred';
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
 * Backend에서 이미 상세한 영어 메시지를 반환하므로, 백엔드 메시지를 우선 사용
 */
export function getVPCDeletionErrorMessage(error: unknown, provider?: string): string {
  const message = extractErrorMessage(error);
  
  // 백엔드에서 이미 상세한 메시지를 반환하는 경우 (DependencyViolation, Resolution steps 포함)
  // 백엔드 메시지에 "Resolution steps" 또는 "cannot be deleted"가 포함되어 있으면 그대로 사용
  if (message.includes('Resolution steps') || 
      message.includes('cannot be deleted') ||
      message.includes('Please delete or detach')) {
    // 백엔드 메시지가 이미 충분히 상세하므로 그대로 반환
    return message;
  }
  
  // 의존성 위반 에러 (DependencyViolation) - 백엔드 메시지가 없는 경우 fallback
  if (message.toLowerCase().includes('dependencyviolation') || 
      message.toLowerCase().includes('has dependencies')) {
    return `[VPC Deletion Failed] VPC cannot be deleted because it has attached resources.\n\n` +
      `Error: ${message}\n\n` +
      `Resolution steps:\n` +
      `1. Delete all subnets attached to the VPC\n` +
      `2. Delete or detach security groups from other resources\n` +
      `3. Detach and delete internet gateways attached to the VPC\n` +
      `4. Delete network interfaces attached to the VPC\n` +
      `5. Delete any VPC peering connections\n` +
      `6. Remove all dependent resources and try again`;
  }
  
  // 리소스를 찾을 수 없음
  if (message.toLowerCase().includes('not found') || 
      message.toLowerCase().includes('does not exist')) {
    return `[VPC Deletion Failed] VPC not found.\n\n` +
      `Error: ${message}\n\n` +
      `The VPC may have already been deleted or does not exist.`;
  }
  
  // 권한 에러
  if (message.toLowerCase().includes('accessdenied') || 
      message.toLowerCase().includes('unauthorized') ||
      message.toLowerCase().includes('permission denied')) {
    return `[Permission Error] Permission denied for VPC deletion.\n\n` +
      `Error: ${message}\n\n` +
      `Resolution steps:\n` +
      `1. Check the 'ec2:DeleteVpc' permission in your IAM policy\n` +
      `2. Verify that the required permissions are granted`;
  }
  
  // IAM 권한 에러
  if (isIAMPermissionError(message)) {
    return `[IAM Permission Error] IAM permissions are missing for VPC deletion.\n\n` +
      `Error: ${message}\n\n` +
      `Resolution steps:\n` +
      `1. Check the user or role policy in AWS IAM Console\n` +
      `2. Verify that the SkyClustEKSFullAccess policy includes the EC2NetworkDeletion section\n` +
      `3. Required permissions: ec2:DeleteVpc, ec2:DescribeVpcs`;
  }
  
  // 일반 에러 - 백엔드 메시지가 있으면 그대로 사용, 없으면 기본 메시지
  if (message && message.trim().length > 0) {
    return `[VPC Deletion Failed] An error occurred while deleting the VPC.\n\n` +
      `Error: ${message}`;
  }
  
  return `[VPC Deletion Failed] An unexpected error occurred while deleting the VPC.`;
}


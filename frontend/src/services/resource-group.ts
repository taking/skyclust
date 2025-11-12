/**
 * Resource Group Service
 * Azure Resource Group 관련 API 호출
 */

import { BaseService } from '@/lib/api';
import { API_ENDPOINTS } from '@/lib/api';

export interface ResourceGroupInfo {
  id: string;
  name: string;
  location: string;
  provisioning_state: string;
  tags?: Record<string, string>;
}

class ResourceGroupService extends BaseService {
  /**
   * Azure Resource Group 목록 조회
   * 
   * @param credentialId - 자격 증명 ID
   * @param limit - 조회할 최대 개수 (선택사항, 기본값: 10, Sidebar용으로는 100 권장)
   * @returns Resource Group 배열
   * 
   * @example
   * ```tsx
   * // Sidebar용: 모든 Resource Group 조회
   * const resourceGroups = await resourceGroupService.listResourceGroups('credential-id', 100);
   * 
   * // 목록 페이지용: Pagination 사용 (limit은 페이지 크기)
   * const resourceGroups = await resourceGroupService.listResourceGroups('credential-id', 20);
   * ```
   */
  async listResourceGroups(credentialId: string, limit?: number): Promise<ResourceGroupInfo[]> {
    // Backend의 OKWithPagination은 data에 배열을 직접 반환합니다
    const data = await this.get<ResourceGroupInfo[]>(
      API_ENDPOINTS.azure.iam.resourceGroups.list(credentialId, limit)
    );
    
    return Array.isArray(data) ? data : [];
  }

  /**
   * 특정 Resource Group 조회
   * 
   * @param name - Resource Group 이름
   * @param credentialId - 자격 증명 ID
   * @returns Resource Group 정보
   * 
   * @example
   * ```tsx
   * const rg = await resourceGroupService.getResourceGroup('my-rg', 'credential-id');
   * ```
   */
  async getResourceGroup(name: string, credentialId: string): Promise<ResourceGroupInfo> {
    return this.get<ResourceGroupInfo>(
      API_ENDPOINTS.azure.iam.resourceGroups.detail(name, credentialId)
    );
  }

  /**
   * Resource Group 생성
   * 
   * @param data - Resource Group 생성 데이터
   * @returns 생성된 Resource Group 정보
   */
  async createResourceGroup(data: {
    credential_id: string;
    name: string;
    location: string;
    tags?: Record<string, string>;
  }): Promise<ResourceGroupInfo> {
    return this.post<ResourceGroupInfo>(
      API_ENDPOINTS.azure.iam.resourceGroups.create(),
      data
    );
  }

  /**
   * Resource Group 삭제
   * 
   * @param name - Resource Group 이름
   * @param credentialId - 자격 증명 ID
   */
  async deleteResourceGroup(name: string, credentialId: string): Promise<void> {
    return this.delete(
      API_ENDPOINTS.azure.iam.resourceGroups.delete(name, credentialId)
    );
  }
}

export const resourceGroupService = new ResourceGroupService();


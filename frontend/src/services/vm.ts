/**
 * VM Service
 * Virtual Machine 관련 API 호출
 */

import { BaseService } from '@/lib/api';
import { API_ENDPOINTS } from '@/lib/api';
import type { VM, CreateVMForm } from '@/lib/types';

class VMService extends BaseService {
  /**
   * 워크스페이스의 VM 목록 조회
   * 
   * @param workspaceId - 워크스페이스 ID
   * @returns VM 배열
   * 
   * @example
   * ```tsx
   * const vms = await vmService.getVMs('workspace-id');
   * ```
   */
  async getVMs(workspaceId: string): Promise<VM[]> {
    return this.get<VM[]>(API_ENDPOINTS.vms.list(workspaceId));
  }

  /**
   * ID로 VM 조회
   * 
   * @param id - VM ID
   * @returns VM 정보
   * 
   * @example
   * ```tsx
   * const vm = await vmService.getVM('vm-id');
   * ```
   */
  async getVM(id: string): Promise<VM> {
    return this.get<VM>(API_ENDPOINTS.vms.detail(id));
  }

  /**
   * VM 생성
   * 
   * @param data - VM 생성 데이터 (name, instance_type, image_id 등)
   * @returns 생성된 VM 정보
   * 
   * @example
   * ```tsx
   * const vm = await vmService.createVM({
   *   workspace_id: 'workspace-id',
   *   name: 'my-vm',
   *   instance_type: 't2.micro',
   *   image_id: 'ami-123',
   * });
   * ```
   */
  async createVM(data: CreateVMForm): Promise<VM> {
    return this.post<VM>(API_ENDPOINTS.vms.create(), data);
  }

  /**
   * VM 정보 업데이트
   * 
   * @param id - VM ID
   * @param data - 업데이트할 VM 데이터 (부분 업데이트 지원)
   * @returns 업데이트된 VM 정보
   * 
   * @example
   * ```tsx
   * const updated = await vmService.updateVM('vm-id', {
   *   name: 'Updated Name',
   * });
   * ```
   */
  async updateVM(id: string, data: Partial<CreateVMForm>): Promise<VM> {
    return this.put<VM>(API_ENDPOINTS.vms.update(id), data);
  }

  /**
   * VM 삭제
   * 
   * @param id - VM ID
   * 
   * @example
   * ```tsx
   * await vmService.deleteVM('vm-id');
   * ```
   */
  async deleteVM(id: string): Promise<void> {
    return this.delete<void>(API_ENDPOINTS.vms.delete(id));
  }

  /**
   * VM 시작
   * 
   * @param id - VM ID
   * 
   * @example
   * ```tsx
   * await vmService.startVM('vm-id');
   * ```
   */
  async startVM(id: string): Promise<void> {
    return this.post<void>(API_ENDPOINTS.vms.start(id));
  }

  /**
   * VM 중지
   * 
   * @param id - VM ID
   * 
   * @example
   * ```tsx
   * await vmService.stopVM('vm-id');
   * ```
   */
  async stopVM(id: string): Promise<void> {
    return this.post<void>(API_ENDPOINTS.vms.stop(id));
  }
}

export const vmService = new VMService();

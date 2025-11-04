/**
 * VM Service
 * Virtual Machine 관련 API 호출
 */

import { BaseService } from '@/lib/service-base';
import type { VM, CreateVMForm } from '@/lib/types';

class VMService extends BaseService {
  // Get VMs by workspace
  async getVMs(workspaceId: string): Promise<VM[]> {
    return this.get<VM[]>(`/api/v1/vms?workspace_id=${workspaceId}`);
  }

  // Get VM by ID
  async getVM(id: string): Promise<VM> {
    return this.get<VM>(`/api/v1/vms/${id}`);
  }

  // Create VM
  async createVM(data: CreateVMForm): Promise<VM> {
    return this.post<VM>('/api/v1/vms', data);
  }

  // Update VM
  async updateVM(id: string, data: Partial<CreateVMForm>): Promise<VM> {
    return this.put<VM>(`/api/v1/vms/${id}`, data);
  }

  // Delete VM
  async deleteVM(id: string): Promise<void> {
    return this.delete<void>(`/api/v1/vms/${id}`);
  }

  // Start VM
  async startVM(id: string): Promise<void> {
    return this.post<void>(`/api/v1/vms/${id}/start`);
  }

  // Stop VM
  async stopVM(id: string): Promise<void> {
    return this.post<void>(`/api/v1/vms/${id}/stop`);
  }
}

export const vmService = new VMService();

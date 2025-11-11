/**
 * VM Service
 * Virtual Machine 관련 API 호출
 */

import { BaseService } from '@/lib/api';
import { API_ENDPOINTS } from '@/lib/api';
import type { VM, CreateVMForm } from '@/lib/types';

class VMService extends BaseService {
  // Get VMs by workspace
  async getVMs(workspaceId: string): Promise<VM[]> {
    return this.get<VM[]>(API_ENDPOINTS.vms.list(workspaceId));
  }

  // Get VM by ID
  async getVM(id: string): Promise<VM> {
    return this.get<VM>(API_ENDPOINTS.vms.detail(id));
  }

  // Create VM
  async createVM(data: CreateVMForm): Promise<VM> {
    return this.post<VM>(API_ENDPOINTS.vms.create(), data);
  }

  // Update VM
  async updateVM(id: string, data: Partial<CreateVMForm>): Promise<VM> {
    return this.put<VM>(API_ENDPOINTS.vms.update(id), data);
  }

  // Delete VM
  async deleteVM(id: string): Promise<void> {
    return this.delete<void>(API_ENDPOINTS.vms.delete(id));
  }

  // Start VM
  async startVM(id: string): Promise<void> {
    return this.post<void>(API_ENDPOINTS.vms.start(id));
  }

  // Stop VM
  async stopVM(id: string): Promise<void> {
    return this.post<void>(API_ENDPOINTS.vms.stop(id));
  }
}

export const vmService = new VMService();

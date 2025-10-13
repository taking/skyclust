import api from '@/lib/api';
import { ApiResponse, VM, CreateVMForm } from '@/lib/types';

export const vmService = {
  // Get VMs by workspace
  getVMs: async (workspaceId: string): Promise<VM[]> => {
    const response = await api.get<ApiResponse<VM[]>>(`/api/v1/vms?workspace_id=${workspaceId}`);
    return response.data.data!;
  },

  // Get VM by ID
  getVM: async (id: string): Promise<VM> => {
    const response = await api.get<ApiResponse<VM>>(`/api/v1/vms/${id}`);
    return response.data.data!;
  },

  // Create VM
  createVM: async (data: CreateVMForm): Promise<VM> => {
    const response = await api.post<ApiResponse<VM>>('/api/v1/vms', data);
    return response.data.data!;
  },

  // Update VM
  updateVM: async (id: string, data: Partial<CreateVMForm>): Promise<VM> => {
    const response = await api.put<ApiResponse<VM>>(`/api/v1/vms/${id}`, data);
    return response.data.data!;
  },

  // Delete VM
  deleteVM: async (id: string): Promise<void> => {
    await api.delete(`/api/v1/vms/${id}`);
  },

  // Start VM
  startVM: async (id: string): Promise<void> => {
    await api.post(`/api/v1/vms/${id}/start`);
  },

  // Stop VM
  stopVM: async (id: string): Promise<void> => {
    await api.post(`/api/v1/vms/${id}/stop`);
  },
};

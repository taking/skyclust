import api from '@/lib/api';
import { ApiResponse, Credential, CreateCredentialForm } from '@/lib/types';

export const credentialService = {
  // Get credentials by workspace
  getCredentials: async (workspaceId: string): Promise<Credential[]> => {
    const response = await api.get<ApiResponse<Credential[]>>(`/api/v1/credentials?workspace_id=${workspaceId}`);
    return response.data.data!;
  },

  // Get credential by ID
  getCredential: async (id: string): Promise<Credential> => {
    const response = await api.get<ApiResponse<Credential>>(`/api/v1/credentials/${id}`);
    return response.data.data!;
  },

  // Create credential
  createCredential: async (data: CreateCredentialForm): Promise<Credential> => {
    const response = await api.post<ApiResponse<Credential>>('/api/v1/credentials', data);
    return response.data.data!;
  },

  // Update credential
  updateCredential: async (id: string, data: Partial<CreateCredentialForm>): Promise<Credential> => {
    const response = await api.put<ApiResponse<Credential>>(`/api/v1/credentials/${id}`, data);
    return response.data.data!;
  },

  // Delete credential
  deleteCredential: async (id: string): Promise<void> => {
    await api.delete(`/api/v1/credentials/${id}`);
  },
};

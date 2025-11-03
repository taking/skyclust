import api from '@/lib/api';
import { ApiResponse, Credential, CreateCredentialForm } from '@/lib/types';

export const credentialService = {
  // Get credentials by workspace
  getCredentials: async (workspaceId: string): Promise<Credential[]> => {
    const response = await api.get<ApiResponse<{ credentials: Credential[] }>>(`/api/v1/credentials?workspace_id=${workspaceId}`);
    // Backend returns { credentials: [...] } inside data field
    const data = response.data.data;
    if (!data || typeof data !== 'object') {
      return [];
    }
    // Check if data has credentials property
    if ('credentials' in data && Array.isArray((data as { credentials: Credential[] }).credentials)) {
      return (data as { credentials: Credential[] }).credentials;
    }
    // Fallback: if data is directly an array
    if (Array.isArray(data)) {
      return data;
    }
    return [];
  },

  // Get credential by ID
  getCredential: async (id: string): Promise<Credential> => {
    const response = await api.get<ApiResponse<Credential>>(`/api/v1/credentials/${id}`);
    return response.data.data!;
  },

  // Create credential
  createCredential: async (data: CreateCredentialForm & { workspace_id: string; name?: string }): Promise<Credential> => {
    const response = await api.post<ApiResponse<Credential>>('/api/v1/credentials', {
      workspace_id: data.workspace_id,
      name: data.name || `${data.provider.toUpperCase()} Credential`,
      provider: data.provider,
      data: data.credentials || {},
    });
    return response.data.data!;
  },

  // Create credential from file (multipart/form-data)
  createCredentialFromFile: async (workspaceId: string, name: string, provider: string, file: File): Promise<Credential> => {
    const formData = new FormData();
    formData.append('workspace_id', workspaceId);
    formData.append('name', name);
    formData.append('provider', provider);
    formData.append('file', file);
    
    // Axios automatically sets Content-Type to multipart/form-data with boundary
    // when FormData is used, so we don't need to set it manually
    const response = await api.post<ApiResponse<Credential>>('/api/v1/credentials/upload', formData);
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

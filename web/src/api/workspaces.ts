import apiClient from './client'

export interface Workspace {
  id: string
  name: string
  owner_id: string
  settings: Record<string, any>
  created_at: string
  updated_at: string
}

export interface CreateWorkspaceRequest {
  name: string
}

export const workspacesApi = {
  list: async (): Promise<{ workspaces: Workspace[] }> => {
    const response = await apiClient.get('/workspaces')
    return response.data
  },

  create: async (data: CreateWorkspaceRequest): Promise<{ workspace: Workspace }> => {
    const response = await apiClient.post('/workspaces', data)
    return response.data
  },

  get: async (id: string): Promise<{ workspace: Workspace }> => {
    const response = await apiClient.get(`/workspaces/${id}`)
    return response.data
  },

  update: async (id: string, data: Partial<CreateWorkspaceRequest>): Promise<{ workspace: Workspace }> => {
    const response = await apiClient.put(`/workspaces/${id}`, data)
    return response.data
  },

  delete: async (id: string): Promise<void> => {
    await apiClient.delete(`/workspaces/${id}`)
  },
}


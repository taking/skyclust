import apiClient from './client'

export interface User {
  id: string
  email: string
  name: string
  created_at: string
  updated_at: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
  name: string
}

export interface AuthResponse {
  token: string
  user: User
}

export const authApi = {
  login: async (data: LoginRequest): Promise<AuthResponse> => {
    const response = await apiClient.post('/auth/login', data)
    return response.data
  },

  register: async (data: RegisterRequest): Promise<{ user: User }> => {
    const response = await apiClient.post('/auth/register', data)
    return response.data
  },

  getCurrentUser: async (): Promise<{ user: User }> => {
    const response = await apiClient.get('/auth/me')
    return response.data
  },

  logout: async (): Promise<void> => {
    await apiClient.post('/auth/logout')
  },
}


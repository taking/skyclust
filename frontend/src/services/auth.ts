import api from '@/lib/api';
import { ApiResponse, LoginForm, RegisterForm, AuthResponse, User } from '@/lib/types';

export const authService = {
  // Login
  login: async (data: LoginForm): Promise<AuthResponse> => {
    const response = await api.post<ApiResponse<AuthResponse>>('/api/v1/auth/login', data);
    return response.data.data!;
  },

  // Register
  register: async (data: RegisterForm): Promise<User> => {
    const response = await api.post<ApiResponse<User>>('/api/v1/auth/register', data);
    return response.data.data!;
  },

  // Logout
  logout: async (): Promise<void> => {
    await api.post('/api/v1/auth/logout');
  },

  // Get current user
  getCurrentUser: async (): Promise<User> => {
    const response = await api.get<ApiResponse<User>>('/api/v1/users/me');
    return response.data.data!;
  },

  // Update user
  updateUser: async (id: string, data: Partial<User>): Promise<User> => {
    const response = await api.put<ApiResponse<User>>(`/api/v1/users/${id}`, data);
    return response.data.data!;
  },
};

/**
 * Auth Service
 * 인증 관련 API 호출
 */

import { BaseService } from '@/lib/service-base';
import type { LoginForm, RegisterForm, AuthResponse, User } from '@/lib/types';

class AuthService extends BaseService {
  // Login
  async login(data: LoginForm): Promise<AuthResponse> {
    const responseData = await this.post<{ token: string; user: User }>('auth/login', data);
    return {
      token: responseData.token,
      expires_at: '', // Backend doesn't return expires_at, but we can calculate it if needed
      user: responseData.user,
    };
  }

  // Register
  async register(data: RegisterForm): Promise<User> {
    return this.post<User>('auth/register', data);
  }

  // Logout
  async logout(): Promise<void> {
    return this.post<void>('auth/logout');
  }

  // Get current user
  async getCurrentUser(): Promise<User> {
    return this.get<User>('auth/me');
  }

  // Update user
  async updateUser(id: string, data: Partial<User>): Promise<User> {
    return this.put<User>(`users/${id}`, data);
  }
}

export const authService = new AuthService();

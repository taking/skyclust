/**
 * Auth Service
 * 인증 관련 API 호출
 */

import { BaseService } from '@/lib/api';
import { API_ENDPOINTS } from '@/lib/api';
import type { LoginForm, RegisterForm, AuthResponse, User } from '@/lib/types';

class AuthService extends BaseService {
  /**
   * 사용자 로그인
   * 
   * @param data - 로그인 폼 데이터 (email, password)
   * @returns 인증 토큰, 만료 시간, 사용자 정보
   * 
   * @example
   * ```tsx
   * const response = await authService.login({
   *   email: 'user@example.com',
   *   password: 'password123',
   * });
   * ```
   */
  async login(data: LoginForm): Promise<AuthResponse> {
    // 1. 로그인 API 호출
    const responseData = await this.post<{ 
      access_token: string; 
      refresh_token: string; 
      expires_in: number;
      token_type: string;
      user: User;
    }>(API_ENDPOINTS.auth.login(), data);
    
    // 2. 응답 데이터를 AuthResponse 형식으로 변환
    const expiresAt = new Date(Date.now() + (responseData.expires_in * 1000)).toISOString();
    return {
      token: responseData.access_token,
      refreshToken: responseData.refresh_token,
      expires_at: expiresAt,
      user: responseData.user,
    };
  }

  /**
   * 사용자 회원가입
   * 
   * @param data - 회원가입 폼 데이터 (name, email, password 등)
   * @returns 인증 토큰 및 사용자 정보
   * 
   * @example
   * ```tsx
   * const response = await authService.register({
   *   name: 'John Doe',
   *   email: 'john@example.com',
   *   password: 'password123',
   * });
   * ```
   */
  async register(data: RegisterForm): Promise<AuthResponse> {
    // 백엔드는 username 필드를 요구하므로 name을 username으로 매핑
    const requestData = {
      username: data.name,
      email: data.email,
      password: data.password,
    };
    const responseData = await this.post<{ 
      access_token: string; 
      refresh_token: string; 
      expires_in: number;
      token_type: string;
      user: User;
    }>(API_ENDPOINTS.auth.register(), requestData);
    const expiresAt = new Date(Date.now() + (responseData.expires_in * 1000)).toISOString();
    return {
      token: responseData.access_token,
      refreshToken: responseData.refresh_token,
      expires_at: expiresAt,
      user: responseData.user,
    };
  }

  /**
   * Refresh Token을 사용하여 새로운 Access Token과 Refresh Token을 발급합니다
   * 
   * @param refreshToken - Refresh Token
   * @returns 새로운 Access Token과 Refresh Token
   */
  async refreshToken(refreshToken: string): Promise<{ access_token: string; refresh_token: string; expires_in: number }> {
    return this.post<{ access_token: string; refresh_token: string; expires_in: number }>(
      API_ENDPOINTS.auth.refresh(),
      { refresh_token: refreshToken }
    );
  }

  /**
   * 사용자 로그아웃
   * 
   * 서버 세션을 종료하고 토큰을 무효화합니다.
   * 
   * @example
   * ```tsx
   * await authService.logout();
   * ```
   */
  async logout(): Promise<void> {
    return this.post<void>(API_ENDPOINTS.auth.logout());
  }

  /**
   * 현재 로그인한 사용자 정보 조회
   * 
   * @returns 현재 사용자 정보
   * 
   * @example
   * ```tsx
   * const user = await authService.getCurrentUser();
   * // logger.debug('Current user', { user });
   * ```
   */
  async getCurrentUser(): Promise<User> {
    return this.get<User>(API_ENDPOINTS.auth.me());
  }

  /**
   * 사용자 정보 업데이트
   * 
   * @param id - 사용자 ID
   * @param data - 업데이트할 사용자 정보 (부분 업데이트 지원)
   * @returns 업데이트된 사용자 정보
   * 
   * @example
   * ```tsx
   * const updatedUser = await authService.updateUser('user-id', {
   *   name: 'New Name',
   * });
   * ```
   */
  async updateUser(id: string, data: Partial<User>): Promise<User> {
    return this.put<User>(API_ENDPOINTS.users.update(id), data);
  }
}

export const authService = new AuthService();

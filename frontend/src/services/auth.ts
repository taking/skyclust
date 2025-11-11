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
    const responseData = await this.post<{ token: string; user: User }>(API_ENDPOINTS.auth.login(), data);
    
    // 2. 응답 데이터를 AuthResponse 형식으로 변환
    // 백엔드에서 expires_at을 반환하지 않으므로 빈 문자열로 설정
    // 필요시 토큰 만료 시간을 계산할 수 있음
    return {
      token: responseData.token,
      expires_at: '', // 백엔드에서 expires_at을 반환하지 않지만, 필요시 계산할 수 있음
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
    const responseData = await this.post<{ token: string; user: User }>(API_ENDPOINTS.auth.register(), requestData);
    return {
      token: responseData.token,
      expires_at: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(), // 7일
      user: responseData.user,
    };
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

/**
 * 인증 관련 타입 정의
 */

import { User } from './user';

export interface AuthResponse {
  token: string; // Access Token
  refreshToken?: string; // Refresh Token (optional for backward compatibility)
  expires_at: string;
  user: User;
}

export interface LoginForm {
  email: string;
  password: string;
}

export interface RegisterForm {
  email: string;
  password: string;
  name: string;
}


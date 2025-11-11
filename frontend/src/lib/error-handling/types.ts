/**
 * Error Handling Types
 * 모든 에러 관련 타입 정의
 */

/**
 * API 에러 인터페이스
 */
export interface ApiError {
  message: string;
  code?: string;
  status?: number;
  details?: Record<string, unknown>;
}

/**
 * 에러 로그 인터페이스
 */
export interface ErrorLog {
  /**
   * 에러 ID
   */
  id: string;
  
  /**
   * 에러 메시지
   */
  message: string;
  
  /**
   * 에러 스택
   */
  stack?: string;
  
  /**
   * 에러 타입
   */
  type: string;
  
  /**
   * 에러 코드
   */
  code?: string;
  
  /**
   * HTTP 상태 코드
   */
  statusCode?: number;
  
  /**
   * 타임스탬프
   */
  timestamp: string;
  
  /**
   * 사용자 에이전트
   */
  userAgent?: string;
  
  /**
   * 현재 URL
   */
  url?: string;
  
  /**
   * 추가 컨텍스트
   */
  context?: Record<string, unknown>;
  
  /**
   * 컴포넌트 스택 (Error Boundary에서)
   */
  componentStack?: string;
}

/**
 * Network Error
 * 네트워크 연결 실패 시 발생하는 에러
 */
export class NetworkError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'NetworkError';
  }
}

/**
 * Server Error
 * 서버에서 반환하는 에러
 */
export class ServerError extends Error {
  status: number;
  code?: string;
  details?: Record<string, unknown>;

  constructor(message: string, status: number, code?: string, details?: Record<string, unknown>) {
    super(message);
    this.name = 'ServerError';
    this.status = status;
    this.code = code;
    this.details = details;
  }
}


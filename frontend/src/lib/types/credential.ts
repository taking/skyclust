/**
 * Credential 관련 타입 정의
 */

export interface Credential {
  id: string;
  workspace_id: string;
  provider: string;
  name?: string; // Credential name for display
  created_at: string;
  updated_at: string;
}

export interface CreateCredentialForm {
  name?: string;
  provider: string;
  credentials: Record<string, unknown>;
}


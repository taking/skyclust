/**
 * 공통 타입 정의
 */

export interface Provider {
  name: string;
  version: string;
}

export interface Instance {
  id: string;
  name: string;
  status: string;
  type: string;
  region: string;
  public_ip?: string;
  private_ip?: string;
  created_at: string;
  tags?: Record<string, string>;
}

export interface Region {
  name: string;
  display_name: string;
}


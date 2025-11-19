/**
 * Workspace 관련 타입 정의
 */

import { User } from './user';

export interface Workspace {
  id: string;
  name: string;
  description: string;
  owner_id: string;
  credential_count?: number;
  member_count?: number;
  settings?: Record<string, unknown> | null;
  created_at: string;
  updated_at: string;
  is_active: boolean;
}

export interface WorkspaceMember {
  user_id: string;
  workspace_id: string;
  role: 'owner' | 'admin' | 'member';
  joined_at: string;
  user: User;
}

export interface CreateWorkspaceForm {
  name: string;
  description: string;
}

export interface UpdateWorkspaceForm {
  name?: string;
  description?: string;
}


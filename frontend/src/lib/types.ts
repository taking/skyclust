// API Response types
export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  message?: string;
  error?: string;
  code?: string;
}

// User types
export interface User {
  id: string;
  username: string;
  email: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  token: string;
  expires_at: string;
  user: User;
}

// Workspace types
export interface Workspace {
  id: string;
  name: string;
  description: string;
  owner_id: string;
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

// VM types
export interface VM {
  id: string;
  workspace_id: string;
  name: string;
  provider: string;
  instance_id: string;
  status: string;
  instance_type: string;
  region: string;
  public_ip?: string;
  private_ip?: string;
  created_at: string;
  updated_at: string;
}

// Credential types
export interface Credential {
  id: string;
  workspace_id: string;
  provider: string;
  created_at: string;
  updated_at: string;
}

// Provider types
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

// Form types
export interface LoginForm {
  username: string;
  password: string;
}

export interface RegisterForm {
  email: string;
  password: string;
  name: string;
}

export interface CreateWorkspaceForm {
  name: string;
  description: string;
}

export interface CreateVMForm {
  name: string;
  provider: string;
  instance_type: string;
  region: string;
  image_id: string;
}

export interface CreateCredentialForm {
  provider: string;
  credentials: Record<string, string>;
}

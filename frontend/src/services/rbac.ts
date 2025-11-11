/**
 * RBAC Service
 * RBAC (역할 기반 접근 제어) 관련 API 호출
 */

import { BaseService } from '@/lib/api';
import { API_ENDPOINTS } from '@/lib/api';

export type Role = 'admin' | 'user' | 'viewer';

export type Permission = 
  | 'user:create' | 'user:read' | 'user:update' | 'user:delete' | 'user:manage'
  | 'system:read' | 'system:update' | 'system:manage'
  | 'audit:read' | 'audit:export' | 'audit:manage'
  | 'workspace:create' | 'workspace:read' | 'workspace:update' | 'workspace:delete' | 'workspace:manage'
  | 'provider:read' | 'provider:manage';

export interface AssignRoleRequest {
  role: Role;
}

export interface GrantPermissionRequest {
  permission: Permission;
}

export interface UserRolesResponse {
  user_id: string;
  roles: string[];
}

export interface RolePermissionsResponse {
  role: string;
  permissions: string[];
}

export interface CheckPermissionResponse {
  user_id: string;
  permission: string;
  has_permission: boolean;
}

export interface UserEffectivePermissionsResponse {
  user_id: string;
  permissions: string[];
}

class RBACService extends BaseService {
  /**
   * 사용자에게 역할 할당
   * 
   * @param userId - 사용자 ID
   * @param role - 할당할 역할
   * 
   * @example
   * ```tsx
   * await rbacService.assignRole('user-id', 'admin');
   * ```
   */
  async assignRole(userId: string, role: Role): Promise<void> {
    await this.post<void>(API_ENDPOINTS.rbac.assignRole(userId), { role });
  }

  /**
   * 사용자로부터 역할 제거
   * 
   * @param userId - 사용자 ID
   * @param role - 제거할 역할
   * 
   * @example
   * ```tsx
   * await rbacService.removeRole('user-id', 'admin');
   * ```
   */
  async removeRole(userId: string, role: Role): Promise<void> {
    await this.delete<void>(API_ENDPOINTS.rbac.removeRole(userId), { role });
  }

  /**
   * 사용자의 모든 역할 조회
   * 
   * @param userId - 사용자 ID
   * @returns 사용자 역할 목록
   * 
   * @example
   * ```tsx
   * const roles = await rbacService.getUserRoles('user-id');
   * ```
   */
  async getUserRoles(userId: string): Promise<UserRolesResponse> {
    return this.get<UserRolesResponse>(API_ENDPOINTS.rbac.getUserRoles(userId));
  }

  /**
   * 역할에 권한 부여
   * 
   * @param role - 역할
   * @param permission - 부여할 권한
   * 
   * @example
   * ```tsx
   * await rbacService.grantPermission('admin', 'user:create');
   * ```
   */
  async grantPermission(role: Role, permission: Permission): Promise<void> {
    await this.post<void>(API_ENDPOINTS.rbac.grantPermission(role), { permission });
  }

  /**
   * 역할로부터 권한 제거
   * 
   * @param role - 역할
   * @param permission - 제거할 권한
   * 
   * @example
   * ```tsx
   * await rbacService.revokePermission('admin', 'user:create');
   * ```
   */
  async revokePermission(role: Role, permission: Permission): Promise<void> {
    await this.delete<void>(API_ENDPOINTS.rbac.revokePermission(role, permission));
  }

  /**
   * 역할의 모든 권한 조회
   * 
   * @param role - 역할
   * @returns 역할 권한 목록
   * 
   * @example
   * ```tsx
   * const permissions = await rbacService.getRolePermissions('admin');
   * ```
   */
  async getRolePermissions(role: Role): Promise<RolePermissionsResponse> {
    return this.get<RolePermissionsResponse>(API_ENDPOINTS.rbac.getRolePermissions(role));
  }

  /**
   * 사용자가 특정 권한을 가지고 있는지 확인
   * 
   * @param userId - 사용자 ID
   * @param permission - 확인할 권한
   * @returns 권한 보유 여부
   * 
   * @example
   * ```tsx
   * const hasPermission = await rbacService.checkUserPermission('user-id', 'user:create');
   * ```
   */
  async checkUserPermission(userId: string, permission: Permission): Promise<CheckPermissionResponse> {
    return this.get<CheckPermissionResponse>(API_ENDPOINTS.rbac.checkUserPermission(userId, permission));
  }

  /**
   * 사용자의 모든 유효 권한 조회 (역할 상속 포함)
   * 
   * @param userId - 사용자 ID
   * @returns 사용자의 모든 유효 권한 목록
   * 
   * @example
   * ```tsx
   * const permissions = await rbacService.getUserEffectivePermissions('user-id');
   * ```
   */
  async getUserEffectivePermissions(userId: string): Promise<UserEffectivePermissionsResponse> {
    return this.get<UserEffectivePermissionsResponse>(API_ENDPOINTS.rbac.getUserEffectivePermissions(userId));
  }
}

export const rbacService = new RBACService();


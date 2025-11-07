/**
 * Repository Interfaces
 * 데이터 접근 계층 추상화
 */

import type { WorkspaceMember } from './workspace';

/**
 * Base Repository Interface
 * 모든 Repository가 구현해야 하는 기본 인터페이스
 */
export interface IRepository<T, TCreate = T, TUpdate = Partial<TCreate>> {
  /**
   * 목록 조회
   */
  findAll(params?: Record<string, unknown>): Promise<T[]>;

  /**
   * ID로 단일 조회
   */
  findById(id: string, params?: Record<string, unknown>): Promise<T | null>;

  /**
   * 생성
   */
  create(data: TCreate): Promise<T>;

  /**
   * 업데이트
   */
  update(id: string, data: TUpdate): Promise<T>;

  /**
   * 삭제
   */
  delete(id: string, params?: Record<string, unknown>): Promise<void>;
}

/**
 * VPC Repository Interface
 */
export interface IVPCRepository extends Omit<IRepository<import('@/lib/types').VPC, import('@/lib/types').CreateVPCForm>, 'create' | 'update' | 'delete' | 'findAll' | 'findById'> {
  list(provider: string, credentialId: string, region?: string): Promise<import('@/lib/types').VPC[]>;
  getById(provider: string, vpcId: string, credentialId: string, region: string): Promise<import('@/lib/types').VPC>;
  create(provider: string, data: import('@/lib/types').CreateVPCForm): Promise<import('@/lib/types').VPC>;
  update(provider: string, vpcId: string, data: Partial<import('@/lib/types').CreateVPCForm>, credentialId: string, region: string): Promise<import('@/lib/types').VPC>;
  delete(provider: string, vpcId: string, credentialId: string, region: string): Promise<void>;
  findAll(): Promise<import('@/lib/types').VPC[]>;
  findById(id: string): Promise<import('@/lib/types').VPC | null>;
}

/**
 * Credential Repository Interface
 */
export interface ICredentialRepository extends IRepository<import('@/lib/types').Credential, import('@/lib/types').CreateCredentialForm> {
  list(workspaceId: string): Promise<import('@/lib/types').Credential[]>;
  getById(id: string): Promise<import('@/lib/types').Credential>;
  create(data: import('@/lib/types').CreateCredentialForm & { workspace_id: string; name?: string }): Promise<import('@/lib/types').Credential>;
  createFromFile(workspaceId: string, name: string, provider: string, file: File): Promise<import('@/lib/types').Credential>;
  update(id: string, data: Partial<import('@/lib/types').CreateCredentialForm>): Promise<import('@/lib/types').Credential>;
  delete(id: string): Promise<void>;
}

/**
 * Workspace Repository Interface
 */
export interface IWorkspaceRepository extends IRepository<import('@/lib/types').Workspace, import('@/lib/types').CreateWorkspaceForm> {
  list(): Promise<import('@/lib/types').Workspace[]>;
  getById(id: string): Promise<import('@/lib/types').Workspace>;
  create(data: import('@/lib/types').CreateWorkspaceForm): Promise<import('@/lib/types').Workspace>;
  update(id: string, data: Partial<import('@/lib/types').CreateWorkspaceForm>): Promise<import('@/lib/types').Workspace>;
  delete(id: string): Promise<void>;
  addMember(workspaceId: string, email: string, role?: string): Promise<void>;
  getMembers(workspaceId: string): Promise<WorkspaceMember[]>;
  removeMember(workspaceId: string, userId: string): Promise<void>;
}


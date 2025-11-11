/**
 * Credential Service
 * Credential 관련 API 호출
 */

import { BaseService } from '@/lib/api';
import { API_ENDPOINTS } from '@/lib/api';
import type { Credential, CreateCredentialForm } from '@/lib/types';

class CredentialService extends BaseService {
  /**
   * 워크스페이스의 자격 증명 목록 조회
   * 
   * @param workspaceId - 워크스페이스 ID
   * @returns 자격 증명 배열
   * 
   * @example
   * ```tsx
   * const credentials = await credentialService.getCredentials('workspace-id');
   * ```
   */
  async getCredentials(workspaceId: string): Promise<Credential[]> {
    const data = await this.get<Credential[]>(
      API_ENDPOINTS.credentials.list(workspaceId)
    );
    
    return Array.isArray(data) ? data : [];
  }

  /**
   * ID로 자격 증명 조회
   * 
   * @param id - 자격 증명 ID
   * @returns 자격 증명 정보
   * 
   * @example
   * ```tsx
   * const credential = await credentialService.getCredential('credential-id');
   * ```
   */
  async getCredential(id: string): Promise<Credential> {
    return this.get<Credential>(API_ENDPOINTS.credentials.detail(id));
  }

  /**
   * 자격 증명 생성
   * 
   * @param data - 자격 증명 생성 데이터 (workspace_id, provider, credentials 등)
   * @returns 생성된 자격 증명 정보
   * 
   * @example
   * ```tsx
   * const credential = await credentialService.createCredential({
   *   workspace_id: 'workspace-id',
   *   provider: 'aws',
   *   credentials: { accessKey: '...', secretKey: '...' },
   * });
   * ```
   */
  async createCredential(data: CreateCredentialForm & { workspace_id: string; name?: string }): Promise<Credential> {
    // 1. API 요청 데이터 구성
    // name이 없으면 기본 이름 생성 (예: "AWS Credential")
    return this.post<Credential>(API_ENDPOINTS.credentials.create(), {
      workspace_id: data.workspace_id,
      name: data.name || `${data.provider.toUpperCase()} Credential`,
      provider: data.provider,
      data: data.credentials || {},
    });
  }

  /**
   * 파일로부터 자격 증명 생성 (multipart/form-data)
   * 
   * 파일 업로드를 통해 자격 증명을 생성합니다.
   * FormData를 사용하므로 BaseService의 request 메서드를 직접 사용합니다.
   * 
   * @param workspaceId - 워크스페이스 ID
   * @param name - 자격 증명 이름
   * @param provider - 클라우드 프로바이더 (aws, gcp, azure 등)
   * @param file - 업로드할 자격 증명 파일
   * @returns 생성된 자격 증명 정보
   * 
   * @example
   * ```tsx
   * const file = event.target.files[0];
   * const credential = await credentialService.createCredentialFromFile(
   *   'workspace-id',
   *   'My AWS Credential',
   *   'aws',
   *   file
   * );
   * ```
   */
  async createCredentialFromFile(workspaceId: string, name: string, provider: string, file: File): Promise<Credential> {
    // 1. FormData 생성 및 필드 추가
    const formData = new FormData();
    formData.append('workspace_id', workspaceId);
    formData.append('name', name);
    formData.append('provider', provider);
    formData.append('file', file);
    
    // 2. FormData는 BaseService의 request 메서드를 직접 사용
    // (일반 post 메서드는 JSON만 지원하므로)
    const url = this.buildApiUrl(API_ENDPOINTS.credentials.upload());
    return this.request<Credential>('post', url, formData);
  }

  /**
   * 자격 증명 업데이트
   * 
   * @param id - 자격 증명 ID
   * @param data - 업데이트할 자격 증명 데이터 (부분 업데이트 지원)
   * @returns 업데이트된 자격 증명 정보
   * 
   * @example
   * ```tsx
   * const updated = await credentialService.updateCredential('credential-id', {
   *   name: 'Updated Name',
   * });
   * ```
   */
  async updateCredential(id: string, data: Partial<CreateCredentialForm>): Promise<Credential> {
    return this.put<Credential>(API_ENDPOINTS.credentials.update(id), data);
  }

  /**
   * 자격 증명 삭제
   * 
   * @param id - 자격 증명 ID
   * 
   * @example
   * ```tsx
   * await credentialService.deleteCredential('credential-id');
   * ```
   */
  async deleteCredential(id: string): Promise<void> {
    return this.delete<void>(API_ENDPOINTS.credentials.delete(id));
  }
}

export const credentialService = new CredentialService();

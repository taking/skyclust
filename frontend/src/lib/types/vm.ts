/**
 * VM (Virtual Machine) 관련 타입 정의
 */

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

export interface CreateVMForm {
  name: string;
  provider: string;
  instance_type: string;
  region: string;
  image_id: string;
  workspace_id?: string;
}


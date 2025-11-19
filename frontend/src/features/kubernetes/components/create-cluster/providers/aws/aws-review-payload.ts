/**
 * AWS EKS Cluster Creation Payload Builder
 * AWS EKS 클러스터 생성 API 요청 payload 생성
 */

import type { CreateClusterForm } from '@/lib/types';

export function buildAWSPayload(data: CreateClusterForm): Record<string, unknown> {
  const payload: Record<string, unknown> = {};

  if (data.credential_id) payload.credential_id = data.credential_id;
  if (data.name) payload.name = data.name;
  if (data.version) payload.version = data.version;
  if (data.region) payload.region = data.region;
  if (data.subnet_ids && data.subnet_ids.length > 0) payload.subnet_ids = data.subnet_ids;
  if (data.vpc_id) payload.vpc_id = data.vpc_id;
  if (data.role_arn) payload.role_arn = data.role_arn;
  if (data.tags && Object.keys(data.tags).length > 0) payload.tags = data.tags;
  if (data.access_config) payload.access_config = data.access_config;

  return payload;
}

